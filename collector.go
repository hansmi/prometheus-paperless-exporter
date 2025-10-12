package main

import (
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func newCollector(cl *client.Client, timeout time.Duration, enableRemoteNetwork bool, enabledIDs []string) prometheus.Collector {
	// Map of all available collectors. Keys are collector IDs.
	constructors := map[string]func(*client.Client) multiCollectorMember{
		"tag":           func(c *client.Client) multiCollectorMember { return newTagCollector(c) },
		"correspondent": func(c *client.Client) multiCollectorMember { return newCorrespondentCollector(c) },
		"document_type": func(c *client.Client) multiCollectorMember { return newDocumentTypeCollector(c) },
		"storage_path":  func(c *client.Client) multiCollectorMember { return newStoragePathCollector(c) },
		"task":          func(c *client.Client) multiCollectorMember { return newTaskCollector(c) },
		"log":           func(c *client.Client) multiCollectorMember { return newLogCollector(c) },
		"group":         func(c *client.Client) multiCollectorMember { return newGroupCollector(c) },
		"user":          func(c *client.Client) multiCollectorMember { return newUserCollector(c) },
		"document":      func(c *client.Client) multiCollectorMember { return newDocumentCollector(c) },
		"status":        func(c *client.Client) multiCollectorMember { return newStatusCollector(c) },
		"statistics":    func(c *client.Client) multiCollectorMember { return newStatisticsCollector(c) },
		// remote_version is only included when enableRemoteNetwork is true (see below)
	}

	// If enabledIDs is empty, enable all standard collectors.
	enableAll := len(enabledIDs) == 0
	enabled := map[string]bool{}
	if enableAll {
		for id := range constructors {
			enabled[id] = true
		}
	} else {
		for _, id := range enabledIDs {
			enabled[id] = true
		}
	}

	members := []multiCollectorMember{}
	for id, collectorFunc := range constructors {
		if enabled[id] {
			members = append(members, collectorFunc(cl))
		}
	}

	// Handle remote collector separately since it depends on external network and should
	// only be included when requested.
	if enableRemoteNetwork && (enableAll || enabled["remote_version"]) {
		members = append(members, newRemoteVersionCollector(cl))
	}

	c := newMultiCollector(members...)
	c.timeout = timeout

	return c
}
