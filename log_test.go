package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

type fakeLogClient struct {
	names   []string
	entries map[string][]client.LogEntry

	listErr error
	getErr  error
}

func (c *fakeLogClient) addEntries(name string, e []client.LogEntry) {
	if c.entries == nil {
		c.entries = map[string][]client.LogEntry{}
	}

	c.entries[name] = append(c.entries[name], e...)
}

func (c *fakeLogClient) ListLogs(context.Context) ([]string, *client.Response, error) {
	return c.names, nil, c.listErr
}

func (c *fakeLogClient) GetLog(_ context.Context, name string) ([]client.LogEntry, *client.Response, error) {
	if c.getErr != nil {
		return nil, nil, c.getErr
	}

	entries, ok := c.entries[name]
	if !ok {
		return nil, nil, &client.RequestError{StatusCode: http.StatusNotFound}
	}

	return entries, nil, nil
}

func TestLog(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeLogClient
		wantErr error
	}{
		{name: "empty"},
		{
			name: "listing fails",
			cl: fakeLogClient{
				listErr: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "get fails",
			cl: fakeLogClient{
				names:  []string{"foo", "bar"},
				getErr: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "empty logs",
			cl: fakeLogClient{
				names: []string{"first", "second", "error404"},
				entries: map[string][]client.LogEntry{
					"first":  nil,
					"second": nil,
				},
			},
		},
		{
			name: "entries",
			cl: fakeLogClient{
				names: []string{"first", "second"},
				entries: map[string][]client.LogEntry{
					"first": []client.LogEntry{
						{
							Time:    time.Date(2020, time.March, 1, 0, 0, 0, 0, time.UTC),
							Message: "aaa",
						},
						{
							Time:    time.Date(2020, time.March, 2, 0, 0, 0, 0, time.UTC),
							Message: "bbb",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newLogCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLogCollect(t *testing.T) {
	cl := fakeLogClient{
		entries: map[string][]client.LogEntry{},
	}

	c := newMultiCollectorForTest(t, newLogCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.names = append(cl.names, "server", "db", "not found")

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.addEntries("server", []client.LogEntry{
		{
			Time:   time.Date(2020, time.March, 2, 0, 0, 0, 0, time.UTC),
			Module: "storage",
		},
		{
			Time:   time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
			Module: "storage",
			Level:  "another",
		},
	})

	testutil.CollectAndCompare(t, c, `
# HELP paperless_log_entries_total Best-effort count of log entries.
# TYPE paperless_log_entries_total counter
paperless_log_entries_total{level="",module="storage",name="server"} 1
paperless_log_entries_total{level="another",module="storage",name="server"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.addEntries("server", []client.LogEntry{
		{
			Time:   time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			Module: "storage",
		},
	})

	for range [3]int{} {
		cl.addEntries("db", []client.LogEntry{
			{
				Time: time.Date(2021, time.January, 1, 2, 3, 0, 0, time.UTC),
			},
		})

		testutil.CollectAndCompare(t, c, `
# HELP paperless_log_entries_total Best-effort count of log entries.
# TYPE paperless_log_entries_total counter
paperless_log_entries_total{level="",module="",name="db"} 1
paperless_log_entries_total{level="",module="storage",name="server"} 2
paperless_log_entries_total{level="another",module="storage",name="server"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

		// Reset logs
		cl.entries = nil
	}
}
