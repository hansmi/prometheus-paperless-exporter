package main

import (
	"time"

	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/prometheus/client_golang/prometheus"
)

func newCollector(cl *client.Client, timeout time.Duration, enableRemoteNetwork bool, enabledIDs []string) prometheus.Collector {
	constructors := knownCollectors()

	// If enabledIDs is empty, enable all standard collectors.
	enableAll := len(enabledIDs) == 0
	enabled := map[string]struct{}{}
	if enableAll {
		for id := range constructors {
			enabled[id] = struct{}{}
		}
	} else {
		for _, id := range enabledIDs {
			enabled[id] = struct{}{}
		}
	}

	members := []multiCollectorMember{}
	for id, collectorFunc := range constructors {
		if _, ok := enabled[id]; ok {
			members = append(members, collectorFunc(cl))
		}
	}

	// Handle remote collector separately since it depends on external network and should
	// only be included when requested.
	_, remoteVersionEnabled := enabled[remoteVersionCollectorID]
	if enableRemoteNetwork && (enableAll || remoteVersionEnabled) {
		members = append(members, newRemoteVersionCollector(cl))
	}

	c := newMultiCollector(members...)
	c.timeout = timeout

	return c
}

// knownCollectors returns the map of standard collector constructors keyed by their ID.
// This can be used for validation or listing available collectors.
func knownCollectors() map[string]func(*client.Client) multiCollectorMember {
	return map[string]func(*client.Client) multiCollectorMember{
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
	}
}

// validateCollectorIDs checks the provided collector ids and returns a slice of
// unknown IDs (those not present in KnownCollectors). The remote collector is
// intentionally excluded from KnownCollectors since it requires network and is
// handled separately.
func validateCollectorIDs(ids []string) []string {
	known := knownCollectors()
	var unknown []string
	for _, id := range ids {
		if len(id) == 0 {
			continue
		}
		if _, ok := known[id]; !ok && id != remoteVersionCollectorID {
			unknown = append(unknown, id)
		}
	}
	return unknown
}
