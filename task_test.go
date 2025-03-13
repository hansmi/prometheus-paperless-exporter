package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/ref"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

type fakeTaskClient struct {
	tasks   []client.Task
	listErr error
}

func (c *fakeTaskClient) ListTasks(context.Context) ([]client.Task, *client.Response, error) {
	return c.tasks, nil, c.listErr
}

func TestTask(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeTaskClient
		wantErr error
	}{
		{name: "empty"},
		{
			name: "listing fails",
			cl: fakeTaskClient{
				listErr: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "tasks",
			cl: fakeTaskClient{
				tasks: []client.Task{
					{ID: 2942},
					{ID: 27064},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTaskCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTaskCollect(t *testing.T) {
	cl := fakeTaskClient{}

	c := newMultiCollectorForTest(t, newTaskCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_task_status_info Task status names.
# TYPE paperless_task_status_info gauge
paperless_task_status_info{status="success"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)

	cl.tasks = append(cl.tasks, client.Task{
		ID:      31563,
		Created: ref.Ref(time.Date(1980, time.January, 1, 0, 0, 0, 0, time.UTC)),
	})

	testutil.CollectAndCompare(t, c, `
# HELP paperless_task_created_timestamp_seconds Number of seconds since 1970 of the task creation.
# TYPE paperless_task_created_timestamp_seconds gauge
paperless_task_created_timestamp_seconds{id="31563"} 3.155328e+08
# HELP paperless_task_done_timestamp_seconds Number of seconds since 1970 of when the task finished.
# TYPE paperless_task_done_timestamp_seconds gauge
paperless_task_done_timestamp_seconds{id="31563"} 0
# HELP paperless_task_filename Filename associated with the task (if any).
# TYPE paperless_task_filename gauge
paperless_task_filename{filename="",id="31563"} 1
# HELP paperless_task_info Static information about a task.
# TYPE paperless_task_info gauge
paperless_task_info{id="31563",task_id="",type=""} 1
# HELP paperless_task_status Task status.
# TYPE paperless_task_status gauge
paperless_task_status{id="31563",status="statusunspecified"} 1
# HELP paperless_task_status_info Task status names.
# TYPE paperless_task_status_info gauge
paperless_task_status_info{status="statusunspecified"} 1
paperless_task_status_info{status="success"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)
}
