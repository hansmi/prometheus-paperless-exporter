package main

import (
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func newCollector(cl *client.Client, timeout time.Duration) prometheus.Collector {
	return &multiCollector{
		timeout: timeout,
		members: []multiCollectorMember{
			newTagCollector(cl),
			newCorrespondentCollector(cl),
			newDocumentTypeCollector(cl),
			newStoragePathCollector(cl),
			newTaskCollector(cl),
			newLogCollector(cl),
			newGroupCollector(cl),
			newUserCollector(cl),
			newDocumentCollector(cl),
		},
	}
}
