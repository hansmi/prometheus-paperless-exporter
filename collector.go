package main

import (
	"fmt"
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

var knownCollectors = map[string]func(*client.Client) multiCollectorMember{
	"tag":            func(c *client.Client) multiCollectorMember { return newTagCollector(c) },
	"correspondent":  func(c *client.Client) multiCollectorMember { return newCorrespondentCollector(c) },
	"document_type":  func(c *client.Client) multiCollectorMember { return newDocumentTypeCollector(c) },
	"storage_path":   func(c *client.Client) multiCollectorMember { return newStoragePathCollector(c) },
	"task":           func(c *client.Client) multiCollectorMember { return newTaskCollector(c) },
	"log":            func(c *client.Client) multiCollectorMember { return newLogCollector(c) },
	"group":          func(c *client.Client) multiCollectorMember { return newGroupCollector(c) },
	"user":           func(c *client.Client) multiCollectorMember { return newUserCollector(c) },
	"document":       func(c *client.Client) multiCollectorMember { return newDocumentCollector(c) },
	"status":         func(c *client.Client) multiCollectorMember { return newStatusCollector(c) },
	"statistics":     func(c *client.Client) multiCollectorMember { return newStatisticsCollector(c) },
	"remote_version": func(c *client.Client) multiCollectorMember { return newRemoteVersionCollector(c) },
}

func newCollector(cl *client.Client, timeout time.Duration, enableRemoteNetwork bool, enabledIDs []string) (prometheus.Collector, error) {
	var members []multiCollectorMember

	add := func(id string, fn func(*client.Client) multiCollectorMember) {
		// Remote collector is treated specially since it depends on external
		// network and should only be enabled when requested.
		if id == remoteVersionCollectorID && !enableRemoteNetwork {
			return
		}

		members = append(members, fn(cl))
	}

	if len(enabledIDs) == 0 {
		for id, fn := range knownCollectors {
			add(id, fn)
		}
	} else {
		for _, id := range enabledIDs {
			fn := knownCollectors[id]
			if fn == nil {
				return nil, fmt.Errorf("unknown collector: %s", id)
			}

			add(id, fn)
		}
	}

	c := newMultiCollector(members...)
	c.timeout = timeout

	return c, nil
}
