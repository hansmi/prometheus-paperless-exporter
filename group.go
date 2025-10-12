package main

import (
	"context"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type groupClient interface {
	ListGroups(context.Context, client.ListGroupsOptions) ([]client.Group, *client.Response, error)
}

type groupCollector struct {
	cl groupClient

	countDesc *prometheus.Desc
}

func newGroupCollector(cl groupClient) *groupCollector {
	return &groupCollector{
		cl: cl,

		countDesc: prometheus.NewDesc("paperless_groups",
			"Number of user groups.",
			nil, nil),
	}
}

func (c *groupCollector) id() string {
	return "group"
}

func (c *groupCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

func (c *groupCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	_, response, err := c.cl.ListGroups(ctx, client.ListGroupsOptions{})

	if err != nil {
		return err
	}

	if response.ItemCount != client.ItemCountUnknown {
		ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue,
			float64(response.ItemCount))
	}

	return nil
}
