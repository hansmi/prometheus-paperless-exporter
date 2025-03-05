package main

import (
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func newCollector(cl *client.Client, timeout time.Duration, enableRemoteNetwork bool) prometheus.Collector {
	members := []multiCollectorMember{
		newTagCollector(cl),
		newCorrespondentCollector(cl),
		newDocumentTypeCollector(cl),
		newStoragePathCollector(cl),
		newTaskCollector(cl),
		newLogCollector(cl),
		newGroupCollector(cl),
		newUserCollector(cl),
		newDocumentCollector(cl),
		newStatusCollector(cl),
		newStatisticsCollector(cl),
	}

	if enableRemoteNetwork {
		members = append(members, newRemoteVersionCollector(cl))
	}

	return &multiCollector{
		timeout: timeout,
		members: members,
	}
}
