package main

import (
	"sync"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func newCollector(cl *client.Client, timeout time.Duration, enableRemoteNetwork bool, remoteVersionInterval time.Duration) (prometheus.Collector, func()) {
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
		members = append(members, newRemoteVersionCollector(cl, remoteVersionInterval))
	}

	c := newMultiCollector(members...)
	c.timeout = timeout

	// build stop function which calls Stop() on members that implement it
	var once sync.Once
	stop := func() {
		once.Do(func() {
			for _, m := range members {
				if s, ok := m.(interface{ Stop() }); ok {
					s.Stop()
				}
			}
		})
	}

	return c, stop
}
