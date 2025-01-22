package main

import (
	"context"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

type userClient interface {
	ListUsers(context.Context, client.ListUsersOptions) ([]client.User, *client.Response, error)
}

type userCollector struct {
	cl userClient

	countDesc *prometheus.Desc
}

func newUserCollector(cl userClient) *userCollector {
	return &userCollector{
		cl: cl,

		countDesc: prometheus.NewDesc("paperless_users",
			"Number of users.",
			nil, nil),
	}
}

func (c *userCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.countDesc
}

func (c *userCollector) collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	_, response, err := c.cl.ListUsers(ctx, client.ListUsersOptions{})

	if err != nil {
		return err
	}

	if response.ItemCount != client.ItemCountUnknown {
		ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue,
			float64(response.ItemCount))
	}

	return nil
}
