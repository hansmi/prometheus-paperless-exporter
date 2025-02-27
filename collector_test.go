package main

import (
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

func TestCollector(t *testing.T) {
	// Mocking time.Since
	timeSince = func(t time.Time) time.Duration {
		return time.Duration(5000000000) // 5 seconds
	}

	contentTypeJson := mime.FormatMediaType("application/json", nil)

	mux := http.NewServeMux()
	mux.Handle("/", http.NotFoundHandler())
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "{}")
	})
	mux.HandleFunc("/api/groups/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentTypeJson)
		io.WriteString(w, `{"count": 10}`)
	})
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentTypeJson)
		io.WriteString(w, `{"count": 20}`)
	})
	mux.HandleFunc("/api/documents/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentTypeJson)
		io.WriteString(w, `{"count": 30}`)
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	cl := client.New(client.Options{
		BaseURL: ts.URL,
	})

	c := newCollector(cl, time.Minute)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_task_status_info Task status names.
# TYPE paperless_task_status_info gauge
paperless_task_status_info{status="success"} 1
# HELP paperless_groups Number of user groups.
# TYPE paperless_groups gauge
paperless_groups 10
# HELP paperless_status_celery_status Status of celery. 1 is OK, 0 is not OK.
# TYPE paperless_status_celery_status gauge
paperless_status_celery_status 0
# HELP paperless_status_classifier_last_trained Seconds since the last time the classifier has been trained.
# TYPE paperless_status_classifier_last_trained gauge
paperless_status_classifier_last_trained 5
# HELP paperless_status_classifier_status Status of the classifier. 1 is OK, 0 is not OK.
# TYPE paperless_status_classifier_status gauge
paperless_status_classifier_status 0
# HELP paperless_status_database_status Status of the database. 1 is OK, 0 is not OK.
# TYPE paperless_status_database_status gauge
paperless_status_database_status 0
# HELP paperless_status_database_unapplied_migrations Number of unapplied database migrations.
# TYPE paperless_status_database_unapplied_migrations gauge
paperless_status_database_unapplied_migrations 0
# HELP paperless_status_index_last_modified Seconds since the last time the index has been modified.
# TYPE paperless_status_index_last_modified gauge
paperless_status_index_last_modified 5
# HELP paperless_status_index_status Status of the index. 1 is OK, 0 is not OK.
# TYPE paperless_status_index_status gauge
paperless_status_index_status 0
# HELP paperless_status_redis_status Status of redis. 1 is OK, 0 is not OK.
# TYPE paperless_status_redis_status gauge
paperless_status_redis_status 0
# HELP paperless_status_storage_available Available storage of Paperless in bytes.
# TYPE paperless_status_storage_available gauge
paperless_status_storage_available 0
# HELP paperless_status_storage_total Total storage of Paperless in bytes.
# TYPE paperless_status_storage_total gauge
paperless_status_storage_total 0
# HELP paperless_users Number of users.
# TYPE paperless_users gauge
paperless_users 20
# HELP paperless_documents Number of documents.
# TYPE paperless_documents gauge
paperless_documents 30
`)
}
