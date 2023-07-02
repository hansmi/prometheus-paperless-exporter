package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

func TestCollector(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "{}")
	}))
	t.Cleanup(ts.Close)

	cl := client.New(client.Options{
		BaseURL: ts.URL,
	})

	c := newCollector(cl, time.Minute)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_task_status_info Task status names.
# TYPE paperless_task_status_info gauge
paperless_task_status_info{status="success"} 1
`)
}
