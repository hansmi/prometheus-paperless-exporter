package main

import (
	"context"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type statisticsClient interface {
	GetStatistics(context.Context) (*client.Statistics, *client.Response, error)
}

type statisticsCollector struct {
	cl statisticsClient

	documentsTotalDesc         *prometheus.Desc
	documentsInboxDesc         *prometheus.Desc
	documentFileTypeCountsDesc *prometheus.Desc
	characterCountDesc         *prometheus.Desc
	tagCountDesc               *prometheus.Desc
	correspondentCountDesc     *prometheus.Desc
	documentTypeCountDesc      *prometheus.Desc
	storagePathCountDesc       *prometheus.Desc

	// Missing fields from statistics: InboxTag, InboxTags, CurrentAsn
}

func newStatisticsCollector(cl statisticsClient) *statisticsCollector {
	return &statisticsCollector{
		cl: cl,

		documentsTotalDesc:         prometheus.NewDesc("paperless_statistics_documents_total", "Total number of documents.", nil, nil),
		documentsInboxDesc:         prometheus.NewDesc("paperless_statistics_documents_inbox_count", "Total number of documents that have the defined 'Inbox' tag.", nil, nil),
		documentFileTypeCountsDesc: prometheus.NewDesc("paperless_statistics_documents_file_type_counts", "Total number of documents per MIME type.", []string{"mime_type"}, nil),
		characterCountDesc:         prometheus.NewDesc("paperless_statistics_character_count", "Number of characters stored across the total number of documents.", nil, nil),
		tagCountDesc:               prometheus.NewDesc("paperless_statistics_tag_count", "Total number of tags.", nil, nil),
		correspondentCountDesc:     prometheus.NewDesc("paperless_statistics_correspondent_count", "Total number of correspondents.", nil, nil),
		documentTypeCountDesc:      prometheus.NewDesc("paperless_statistics_document_type_count", "Total number of document types.", nil, nil),
		storagePathCountDesc:       prometheus.NewDesc("paperless_statistics_storage_path_count", "Total number of storage pathes.", nil, nil),
	}
}

func (c *statisticsCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.documentsTotalDesc
	ch <- c.documentsInboxDesc
	ch <- c.documentFileTypeCountsDesc
	ch <- c.characterCountDesc
	ch <- c.tagCountDesc
	ch <- c.correspondentCountDesc
	ch <- c.documentTypeCountDesc
	ch <- c.storagePathCountDesc
}

func (c *statisticsCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	statistics, _, err := c.cl.GetStatistics(ctx)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.documentsTotalDesc, prometheus.GaugeValue, float64(statistics.DocumentsTotal))
	ch <- prometheus.MustNewConstMetric(c.documentsInboxDesc, prometheus.GaugeValue, float64(statistics.DocumentsInbox))

	for _, documentFileTypeCount := range statistics.DocumentFileTypeCounts {
		ch <- prometheus.MustNewConstMetric(c.documentFileTypeCountsDesc, prometheus.GaugeValue, float64(documentFileTypeCount.MimeTypeCount), documentFileTypeCount.MimeType)
	}

	ch <- prometheus.MustNewConstMetric(c.characterCountDesc, prometheus.GaugeValue, float64(statistics.CharacterCount))
	ch <- prometheus.MustNewConstMetric(c.tagCountDesc, prometheus.GaugeValue, float64(statistics.TagCount))
	ch <- prometheus.MustNewConstMetric(c.correspondentCountDesc, prometheus.GaugeValue, float64(statistics.CorrespondentCount))
	ch <- prometheus.MustNewConstMetric(c.documentTypeCountDesc, prometheus.GaugeValue, float64(statistics.DocumentTypeCount))
	ch <- prometheus.MustNewConstMetric(c.storagePathCountDesc, prometheus.GaugeValue, float64(statistics.StoragePathCount))

	return nil
}
