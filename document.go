package main

import (
	"context"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type documentClient interface {
	ListDocuments(context.Context, client.ListDocumentsOptions) ([]client.Document, *client.Response, error)
}

type documentCollector struct {
	cl documentClient

	countDesc *prometheus.Desc
}

func newDocumentCollector(cl documentClient) *documentCollector {
	return &documentCollector{
		cl: cl,

		countDesc: prometheus.NewDesc("paperless_documents",
			"Number of documents.",
			nil, nil),
	}
}

func (c *documentCollector) id() string {
	return "document"
}

func (c *documentCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

func (c *documentCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	_, response, err := c.cl.ListDocuments(ctx, client.ListDocumentsOptions{})

	if err != nil {
		return err
	}

	if response.ItemCount != client.ItemCountUnknown {
		ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue,
			float64(response.ItemCount))
	}

	return nil
}
