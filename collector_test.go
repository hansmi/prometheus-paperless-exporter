package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

func TestCollector(t *testing.T) {
	contentTypeJson := mime.FormatMediaType("application/json", nil)

	for _, enableRemoteNetwork := range []bool{false, true} {
		t.Run(fmt.Sprint(enableRemoteNetwork), func(t *testing.T) {
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
			mux.HandleFunc("/api/remote_version/", func(w http.ResponseWriter, r *http.Request) {
				if !enableRemoteNetwork {
					w.WriteHeader(http.StatusTeapot)
					return
				}

				w.Header().Set("Content-Type", contentTypeJson)
				io.WriteString(w, `{"version": "v2.14.7", "update_available": true}`)
			})

			ts := httptest.NewServer(mux)
			t.Cleanup(ts.Close)

			cl := client.New(client.Options{
				BaseURL: ts.URL,
			})

			var want strings.Builder

			want.WriteString(`
# HELP paperless_task_status_info Task status names.
# TYPE paperless_task_status_info gauge
paperless_task_status_info{status="success"} 1
# HELP paperless_groups Number of user groups.
# TYPE paperless_groups gauge
paperless_groups 10
# HELP paperless_statistics_character_count Number of characters stored across the total number of documents.
# TYPE paperless_statistics_character_count gauge
paperless_statistics_character_count 0
# HELP paperless_statistics_correspondent_count Total number of correspondents.
# TYPE paperless_statistics_correspondent_count gauge
paperless_statistics_correspondent_count 0
# HELP paperless_statistics_document_type_count Total number of document types.
# TYPE paperless_statistics_document_type_count gauge
paperless_statistics_document_type_count 0
# HELP paperless_statistics_documents_inbox_count Total number of documents that have the defined 'Inbox' tag.
# TYPE paperless_statistics_documents_inbox_count gauge
paperless_statistics_documents_inbox_count 0
# HELP paperless_statistics_documents_total Total number of documents.
# TYPE paperless_statistics_documents_total gauge
paperless_statistics_documents_total 0
# HELP paperless_statistics_storage_path_count Total number of storage pathes.
# TYPE paperless_statistics_storage_path_count gauge
paperless_statistics_storage_path_count 0
# HELP paperless_statistics_tag_count Total number of tags.
# TYPE paperless_statistics_tag_count gauge
paperless_statistics_tag_count 0
# HELP paperless_status_celery_status Status of celery. 1 is OK, 0 is not OK.
# TYPE paperless_status_celery_status gauge
paperless_status_celery_status 0
# HELP paperless_status_classifier_last_trained_timestamp_seconds Number of seconds since 1970-01-01 since the last time the classifier has been trained.
# TYPE paperless_status_classifier_last_trained_timestamp_seconds gauge
paperless_status_classifier_last_trained_timestamp_seconds -6.21355968e+10
# HELP paperless_status_classifier_status Status of the classifier. 1 is OK, 0 is not OK.
# TYPE paperless_status_classifier_status gauge
paperless_status_classifier_status 0
# HELP paperless_status_database_status Status of the database. 1 is OK, 0 is not OK.
# TYPE paperless_status_database_status gauge
paperless_status_database_status 0
# HELP paperless_status_database_unapplied_migrations Number of unapplied database migrations.
# TYPE paperless_status_database_unapplied_migrations gauge
paperless_status_database_unapplied_migrations 0
# HELP paperless_status_index_last_modified_timestamp_seconds Number of seconds since 1970-01-01 since the last time the index has been modified.
# TYPE paperless_status_index_last_modified_timestamp_seconds gauge
paperless_status_index_last_modified_timestamp_seconds -6.21355968e+10
# HELP paperless_status_index_status Status of the index. 1 is OK, 0 is not OK.
# TYPE paperless_status_index_status gauge
paperless_status_index_status 0
# HELP paperless_status_redis_status Status of redis. 1 is OK, 0 is not OK.
# TYPE paperless_status_redis_status gauge
paperless_status_redis_status 0
# HELP paperless_status_sanity_check_last_run_timestamp_seconds Number of seconds since 1970-01-01 since the last time the sanity check has been run.
# TYPE paperless_status_sanity_check_last_run_timestamp_seconds gauge
paperless_status_sanity_check_last_run_timestamp_seconds -6.21355968e+10
# HELP paperless_status_sanity_check_status Status of the sanity check. 1 is OK, 0 is not OK.
# TYPE paperless_status_sanity_check_status gauge
paperless_status_sanity_check_status 0
# HELP paperless_status_storage_available_bytes Available storage of Paperless in bytes.
# TYPE paperless_status_storage_available_bytes gauge
paperless_status_storage_available_bytes 0
# HELP paperless_status_storage_total_bytes Total storage of Paperless in bytes.
# TYPE paperless_status_storage_total_bytes gauge
paperless_status_storage_total_bytes 0
# HELP paperless_users Number of users.
# TYPE paperless_users gauge
paperless_users 20
# HELP paperless_documents Number of documents.
# TYPE paperless_documents gauge
paperless_documents 30
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)

			if enableRemoteNetwork {
				want.WriteString(`
# HELP paperless_remote_version_update_available Whether an update is available.
# TYPE paperless_remote_version_update_available gauge
paperless_remote_version_update_available{version="v2.14.7"} 1
`)
			}

			c, err := newCollector(cl, time.Minute, enableRemoteNetwork, nil)
			if err != nil {
				t.Errorf("newCollector() failed: %v", err)
			}

			testutil.CollectAndCompare(t, c, want.String())
		})
	}
}

func TestCollectorError(t *testing.T) {
	for _, tc := range []struct {
		name                string
		enableRemoteNetwork bool
		enabled             []string
		wantErr             error
	}{
		{name: "default"},
		{
			name:                "default with remote",
			enableRemoteNetwork: true,
		},
		{
			name:    "unknown",
			enabled: []string{"bad", "", "foo"},
			wantErr: cmpopts.AnyError,
		},
		{
			name:    "remote version only",
			enabled: []string{"remote_version"},
		},
		{
			name:                "enabled remote version only",
			enableRemoteNetwork: true,
			enabled:             []string{"remote_version"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newCollector(nil, 0, tc.enableRemoteNetwork, tc.enabled)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}
