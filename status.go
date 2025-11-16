package main

import (
	"context"
	"strings"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type statusClient interface {
	GetStatus(ctx context.Context) (*client.SystemStatus, *client.Response, error)
}

type statusCollector struct {
	cl statusClient

	storageTotalDesc                *prometheus.Desc
	storageAvailableDesc            *prometheus.Desc
	databaseStatusDesc              *prometheus.Desc
	databaseUnappliedMigrationsDesc *prometheus.Desc
	redisStatusDesc                 *prometheus.Desc
	celeryStatusDesc                *prometheus.Desc
	indexStatusDesc                 *prometheus.Desc
	indexLastModifiedDesc           *prometheus.Desc
	classifierStatusDesc            *prometheus.Desc
	classifierLastTrainedDesc       *prometheus.Desc
	sanityCheckStatusDesc           *prometheus.Desc
	sanityCheckLastRunDesc          *prometheus.Desc
}

func newStatusCollector(cl statusClient) *statusCollector {
	return &statusCollector{
		cl: cl,

		storageTotalDesc:                prometheus.NewDesc("paperless_status_storage_total_bytes", "Total storage of Paperless in bytes.", nil, nil),
		storageAvailableDesc:            prometheus.NewDesc("paperless_status_storage_available_bytes", "Available storage of Paperless in bytes.", nil, nil),
		databaseStatusDesc:              prometheus.NewDesc("paperless_status_database_status", "Status of the database. 1 is OK, 0 is not OK.", nil, nil),
		databaseUnappliedMigrationsDesc: prometheus.NewDesc("paperless_status_database_unapplied_migrations", "Number of unapplied database migrations.", nil, nil),
		redisStatusDesc:                 prometheus.NewDesc("paperless_status_redis_status", "Status of redis. 1 is OK, 0 is not OK.", nil, nil),
		celeryStatusDesc:                prometheus.NewDesc("paperless_status_celery_status", "Status of celery. 1 is OK, 0 is not OK.", nil, nil),
		indexStatusDesc:                 prometheus.NewDesc("paperless_status_index_status", "Status of the index. 1 is OK, 0 is not OK.", nil, nil),
		indexLastModifiedDesc:           prometheus.NewDesc("paperless_status_index_last_modified_timestamp_seconds", "Number of seconds since 1970-01-01 since the last time the index has been modified.", nil, nil),
		classifierStatusDesc:            prometheus.NewDesc("paperless_status_classifier_status", "Status of the classifier. 1 is OK, 0 is not OK.", nil, nil),
		classifierLastTrainedDesc:       prometheus.NewDesc("paperless_status_classifier_last_trained_timestamp_seconds", "Number of seconds since 1970-01-01 since the last time the classifier has been trained.", nil, nil),
		sanityCheckStatusDesc:           prometheus.NewDesc("paperless_status_sanity_check_status", "Status of the sanity check. 1 is OK, 0 is not OK.", nil, nil),
		sanityCheckLastRunDesc:          prometheus.NewDesc("paperless_status_sanity_check_last_run_timestamp_seconds", "Number of seconds since 1970-01-01 since the last time the sanity check has been run.", nil, nil),
	}
}

func (c *statusCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.storageTotalDesc
	ch <- c.storageAvailableDesc
	ch <- c.databaseStatusDesc
	ch <- c.databaseUnappliedMigrationsDesc
	ch <- c.redisStatusDesc
	ch <- c.celeryStatusDesc
	ch <- c.indexStatusDesc
	ch <- c.indexLastModifiedDesc
	ch <- c.classifierStatusDesc
	ch <- c.classifierLastTrainedDesc
	ch <- c.sanityCheckStatusDesc
	ch <- c.sanityCheckLastRunDesc
}

func (c *statusCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	status, _, err := c.cl.GetStatus(ctx)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.storageTotalDesc, prometheus.GaugeValue, float64(status.Storage.Total))
	ch <- prometheus.MustNewConstMetric(c.storageAvailableDesc, prometheus.GaugeValue, float64(status.Storage.Available))
	ch <- prometheus.MustNewConstMetric(c.databaseStatusDesc, prometheus.GaugeValue, c.isOK(status.Database.Status))
	ch <- prometheus.MustNewConstMetric(c.databaseUnappliedMigrationsDesc, prometheus.GaugeValue, float64(len(status.Database.MigrationStatus.UnappliedMigrations)))
	ch <- prometheus.MustNewConstMetric(c.redisStatusDesc, prometheus.GaugeValue, c.isOK(status.Tasks.RedisStatus))
	ch <- prometheus.MustNewConstMetric(c.celeryStatusDesc, prometheus.GaugeValue, c.isOK(status.Tasks.CeleryStatus))
	ch <- prometheus.MustNewConstMetric(c.indexStatusDesc, prometheus.GaugeValue, c.isOK(status.Tasks.IndexStatus))
	ch <- prometheus.MustNewConstMetric(c.indexLastModifiedDesc, prometheus.GaugeValue, float64(status.Tasks.IndexLastModified.Unix()))
	ch <- prometheus.MustNewConstMetric(c.classifierStatusDesc, prometheus.GaugeValue, c.isOK(status.Tasks.ClassifierStatus))
	ch <- prometheus.MustNewConstMetric(c.classifierLastTrainedDesc, prometheus.GaugeValue, float64(status.Tasks.ClassifierLastTrained.Unix()))
	ch <- prometheus.MustNewConstMetric(c.sanityCheckStatusDesc, prometheus.GaugeValue, c.isOK(status.Tasks.SanityCheckStatus))
	ch <- prometheus.MustNewConstMetric(c.sanityCheckLastRunDesc, prometheus.GaugeValue, float64(status.Tasks.SanityCheckLastRun.Unix()))

	return nil
}

func (c *statusCollector) isOK(status string) float64 {
	if strings.EqualFold(status, "OK") {
		return 1
	}

	return 0
}
