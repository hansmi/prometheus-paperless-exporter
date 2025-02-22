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
# HELP paperless_users Number of users.
# TYPE paperless_users gauge
paperless_users 20
# HELP paperless_documents Number of documents.
# TYPE paperless_documents gauge
paperless_documents 30
`)
}
