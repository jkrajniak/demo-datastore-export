package bigquery_loader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	bigqueryApi "google.golang.org/api/bigquery/v2"
	datastoreApi "google.golang.org/api/datastore/v1"
)

const (
	operationCheckDelaySeconds          = 10
	gcpProjectEnvVarName                = "GCP_PROJECT"
	datastoreExportTopicIdEnvVarName    = "TOPIC_ID"
	outputBigQueryDatasetNameEnvVarName = "OUTPUT_DATASET"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type DatastoreExportOperation struct {
	ID string `json:"id"`
}

type DatastoreExportRequest struct {
	OutputURLPrefix string `json:"outputUrlPrefix"`
	EntityFilter    struct {
		Kinds []string `json:"kinds"`
	} `json:"entityFilter"`
	ProjectID string `json:"projectId"`
}

type DatastoreExportProtoPayload struct {
	Request DatastoreExportRequest `json:"request"`
}

type DatastoreExportLogEntry struct {
	Operation    DatastoreExportOperation    `json:"operation"`
	ProtoPayload DatastoreExportProtoPayload `json:"protoPayload"`
}

type PubSubEvent struct {
	Data []byte `json:"data"`
}

func DatastoreExportPubsubHandler(ctx context.Context, m PubSubEvent) error {
	logrus.WithField("m", string(m.Data)).Debug("Datastore export pubsub handler")

	projectID := LoadEnvVarOrPanic(gcpProjectEnvVarName)
	pubsubTopic := LoadEnvVarOrPanic(datastoreExportTopicIdEnvVarName)
	outputDatasetName := LoadEnvVarOrPanic(outputBigQueryDatasetNameEnvVarName)

	var logEntry DatastoreExportLogEntry
	if err := json.Unmarshal(m.Data, &logEntry); err != nil {
		logrus.WithError(err).Error("error during unmarshal")
		return err
	}

	if len(logEntry.ProtoPayload.Request.EntityFilter.Kinds) != 1 {
		logrus.WithField("logEntry", logEntry).Error("invalid log entry")
		return errors.New("invalid log entry")
	}

	logrus.WithField("logEntry", logEntry).Debug("log entry")

	result, err := GetOperationResult(ctx, logEntry.Operation.ID)
	if err != nil {
		logrus.WithError(err).Error("failed to get the operation results")
		return err
	}
	logrus.WithField("result", fmt.Sprintf("%+v", result)).Debug("long run operation status")

	// If the operation is still in progress then republish the message and check again after a while
	if !result.Done {
		logrus.Debug("sleep for %d s", operationCheckDelaySeconds)
		time.Sleep(operationCheckDelaySeconds * time.Second)

		logrus.Debug("resend message")
		pubsubClient, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			return err
		}
		m := &pubsub.Message{
			Data: m.Data,
		}
		publishResult := pubsubClient.Topic(pubsubTopic).Publish(ctx, m)
		id, err := publishResult.Get(ctx)
		if err != nil {
			return err
		}
		logrus.WithField("id", id).Debug("Message resend")
		return nil
	}

	kindName := logEntry.ProtoPayload.Request.EntityFilter.Kinds[0]
	logrus.Debug("operation done - schedule BigQuery import of kind %s", kindName)
	outputUrl := logEntry.ProtoPayload.Request.OutputURLPrefix
	if err := ScheduleBigQueryImport(ctx, projectID, outputDatasetName, outputUrl, kindName); err != nil {
		logrus.WithError(err).Error("problem with scheduling bigquery import")
		return err
	}

	return nil
}

//GetOperationResult - get datastore export operation result.
func GetOperationResult(ctx context.Context, operationID string) (*datastoreApi.GoogleLongrunningOperation, error) {
	client, err := google.DefaultClient(ctx, datastoreApi.DatastoreScope)
	if err != nil {
		return nil, err
	}

	service, err := datastoreApi.New(client)
	if err != nil {
		return nil, err
	}

	call := service.Projects.Operations.Get(operationID)
	result, err := call.Do()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ScheduleBigQueryImport(ctx context.Context, projectId, outputDatasetName, outputUrl, kindName string) error {
	service, err := bigqueryApi.NewService(ctx)
	if err != nil {
		return err
	}

	sourceUri := fmt.Sprintf("%s/all_namespaces/kind_%s/all_namespaces_kind_%s.export_metadata", outputUrl, kindName, kindName)

	jobs := bigqueryApi.NewJobsService(service)
	jobConfig := &bigqueryApi.Job{
		Configuration: &bigqueryApi.JobConfiguration{
			Load: &bigqueryApi.JobConfigurationLoad{
				SourceFormat: "DATASTORE_BACKUP",
				SourceUris: []string{
					sourceUri,
				},
				DestinationTable: &bigqueryApi.TableReference{
					DatasetId: outputDatasetName,
					ProjectId: projectId,
					TableId:   kindName,
				},
			},
		},
	}
	logrus.WithFields(logrus.Fields{
		"sourceUri": sourceUri,
		"jobConfig": jobConfig,
	}).Debug("run bigquery load job")

	call := jobs.Insert(projectId, jobConfig)
	jobStatus, err := call.Do()
	if err != nil {
		return err
	}
	logrus.WithField("job", jobStatus).Debug("Run import from %s to %s.%s", sourceUri, outputDatasetName, kindName)

	return nil
}
