package cloud_functions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	datastoreApi "google.golang.org/api/datastore/v1beta1"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

//DatastoreExport - export all kinds from all namespaces
func DatastoreExport(w http.ResponseWriter, r *http.Request) {
	logrus.Debug("DatastoreExport")

	projectId := LoadEnvVarOrPanic(gcpProjectEnvVarName)
	outputBucket := LoadEnvVarOrPanic(exportOutputBucketEnvVarName)

	ctx := context.Background()

	logFields := logrus.Fields{
		"project_id":    projectId,
		"output_bucket": outputBucket}

	datastoreClient, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		logrus.WithFields(logFields).WithError(err).Error("failed to create datastore client")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := datastore.NewQuery("__kind__").KeysOnly()
	keys, err := datastoreClient.GetAll(ctx, query, nil)
	if err != nil {
		logrus.WithFields(logFields).WithError(err).Error("failed to get datastore kinds")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// export to specific datetime object in the destination bucket
	today := time.Now().UTC()

	logrus.WithField("keys", keys).Debug("All Kinds")
	for _, k := range keys {
		// Ignore kinds starts with __Stat
		if !strings.HasPrefix(k.Name, "__Stat") {
			logrus.WithField("kind", k.Name).Debug("Export kind")
			if err := exportKind(k.Name, outputBucket, projectId, today.Format(dateTimeFormat)); err != nil {
				logrus.WithFields(logFields).WithField("key", k).WithError(err).Error("failed to export kind")
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}

//exportKind - export specific kind to the bucket.
func exportKind(kind, outputBucket, projectId, outputObject string) error {
	ctxWithDeadline := context.Background()

	client, err := google.DefaultClient(ctxWithDeadline, datastoreApi.DatastoreScope)
	if err != nil {
		return err
	}

	service, err := datastoreApi.New(client)
	if err != nil {
		return err
	}

	exportRequest := datastoreApi.GoogleDatastoreAdminV1beta1ExportEntitiesRequest{
		EntityFilter: &datastoreApi.GoogleDatastoreAdminV1beta1EntityFilter{
			NamespaceIds: []string{},
			Kinds:        []string{kind},
		},
		OutputUrlPrefix: fmt.Sprintf("gs://%s/%s/%s", outputBucket, outputObject, kind),
	}
	logrus.WithField("kind", kind).Info("outputBucket:", outputBucket)

	_, err = service.Projects.Export(projectId, &exportRequest).Do()
	if err != nil {
		return err
	}
	return nil
}
