package main

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func optionalTimestamp(t *time.Time) float64 {
	if t == nil || t.IsZero() {
		return 0
	}

	return float64(t.UnixMilli()) / 1000
}

type taskClient interface {
	ListTasks(context.Context) ([]client.Task, *client.Response, error)
}

type taskCollector struct {
	cl taskClient

	infoDesc     *prometheus.Desc
	createdDesc  *prometheus.Desc
	doneDesc     *prometheus.Desc
	statusDesc   *prometheus.Desc
	filenameDesc *prometheus.Desc

	statusInfoVec *prometheus.GaugeVec
}

func newTaskCollector(cl taskClient) *taskCollector {
	c := &taskCollector{
		cl: cl,

		infoDesc: prometheus.NewDesc("paperless_task_info",
			"Static information about a task.",
			[]string{"id", "task_id", "type"}, nil),
		createdDesc: prometheus.NewDesc("paperless_task_created_timestamp_seconds",
			"Number of seconds since 1970 of the task creation.",
			[]string{"id"}, nil),
		doneDesc: prometheus.NewDesc("paperless_task_done_timestamp_seconds",
			"Number of seconds since 1970 of when the task finished.",
			[]string{"id"}, nil),
		statusDesc: prometheus.NewDesc("paperless_task_status",
			"Task status.",
			[]string{"id", "status"}, nil),
		filenameDesc: prometheus.NewDesc("paperless_task_filename",
			"Filename associated with the task (if any).",
			[]string{"id", "filename"}, nil),

		statusInfoVec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "paperless_task_status_info",
			Help: "Task status names.",
		}, []string{"status"}),
	}

	c.ensureStatusInfo(client.TaskSuccess)

	return c
}

func (c *taskCollector) id() string {
	return "task"
}

// Returns a canonicalized status string for labels.
func (c *taskCollector) ensureStatusInfo(s client.TaskStatus) string {
	status := strings.ToLower(s.String())

	c.statusInfoVec.With(prometheus.Labels{
		"status": status,
	}).Set(1)

	return status
}

func (c *taskCollector) describe(ch chan<- *prometheus.Desc) {
	c.statusInfoVec.Describe(ch)

	ch <- c.infoDesc
	ch <- c.createdDesc
	ch <- c.doneDesc
	ch <- c.statusDesc
	ch <- c.filenameDesc
}

func (c *taskCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	tasks, _, err := c.cl.ListTasks(ctx)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		var filename string

		if task.TaskFileName != nil {
			filename = *task.TaskFileName
		}

		id := strconv.FormatInt(task.ID, 10)

		ch <- prometheus.MustNewConstMetric(c.infoDesc, prometheus.GaugeValue, 1,
			id,
			task.TaskID,
			task.Type,
		)

		ch <- prometheus.MustNewConstMetric(c.createdDesc, prometheus.GaugeValue,
			optionalTimestamp(task.Created), id)

		ch <- prometheus.MustNewConstMetric(c.doneDesc, prometheus.GaugeValue,
			optionalTimestamp(task.Done), id)

		ch <- prometheus.MustNewConstMetric(c.statusDesc, prometheus.GaugeValue,
			1, id, c.ensureStatusInfo(task.Status))

		ch <- prometheus.MustNewConstMetric(c.filenameDesc, prometheus.GaugeValue,
			1, id, filename)
	}

	c.statusInfoVec.Collect(ch)

	return nil
}
