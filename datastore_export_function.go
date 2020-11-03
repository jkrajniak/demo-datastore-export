package p

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"
	datastoreApi "google.golang.org/api/datastore/v1beta1"
	"net/http"
	"os"
)

func DatastoreExport(w http.ResponseWriter, r *http.Request) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Error("DatastoreExport")

	projectId := os.Getenv("GCP_PROJECT")

	ctx := context.Background()

	datastoreClient, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := datastore.NewQuery("__kind__").KeysOnly()
	keys, err := datastoreClient.GetAll(ctx, query, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logrus.Info(keys)

	w.WriteHeader(http.StatusOK)
}

//exportKind - export specific kind to the bucket.
func exportKind(kind, outputBucket, projectId string) error {
	ctxWithDeadline := context.Background()

	client, err := google.DefaultClient(ctxWithDeadline, datastoreApi.DatastoreScope)
	if err != nil {
		return err
	}

	service, err := datastoreApi.New(client)
	if err != nil {
		return err
	}

	op, err := service.Projects.Export(projectId, &datastoreApi.GoogleDatastoreAdminV1beta1ExportEntitiesRequest{
		EntityFilter: &datastoreApi.GoogleDatastoreAdminV1beta1EntityFilter{
			NamespaceIds: []string{},
			Kinds:        []string{kind},
		},
		OutputUrlPrefix: fmt.Sprintf("gs://%s", outputBucket),
	}).Do()

	logrus.Debug(op)
	if err != nil {
		return err
	}
	return nil
}