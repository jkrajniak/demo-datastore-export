package cloud_functions

import (
	"context"
	"fmt"
	"strings"

	"github.com/nordcloud/ncerrors/errors"
	"github.com/sirupsen/logrus"
	bigqueryApi "google.golang.org/api/bigquery/v2"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

type BucketEvent struct {
	EventKind string `json:"kind"`
	Name      string `json:"name"`
	Bucket    string `json:"bucket"`
}

func WatchBucket(ctx context.Context, event BucketEvent) error {
	projectID := LoadEnvVarOrPanic(gcpProjectEnvVarName)
	outputDatasetName := LoadEnvVarOrPanic(outputBigQueryDatasetNameEnvVarName)

	if event.EventKind == "storage#object" {
		// 2021_03_09_05_00_01/Kind2/all_namespaces/kind_Kind2/all_namespaces_kind_Kind2.export_metadata
		if strings.HasSuffix(event.Name, "export_metadata") {
			sourceUri := fmt.Sprintf("gs://%s/%s", event.Bucket, event.Name)
			kindName := strings.Split(event.Name, "/")[1]
			if err := ScheduleBigQueryImport(ctx, projectID, outputDatasetName, sourceUri, kindName); err != nil {
				return errors.WithContext(err, "failed to schedule big query", errors.Fields{
					"sourceUri": sourceUri,
				})
			}
		}
	}

	return nil
}

func ScheduleBigQueryImport(ctx context.Context, projectId, outputDatasetName, sourceUri, kindName string) error {
	service, err := bigqueryApi.NewService(ctx)
	if err != nil {
		return err
	}

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
