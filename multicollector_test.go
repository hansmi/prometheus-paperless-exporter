package main

import (
	"io"
	"log"
	"testing"
)

func newMultiCollectorForTest(t *testing.T, m multiCollectorMember) *multiCollector {
	c := newMultiCollector(m)
	c.logger = log.New(io.Discard, "", 0)

	return c
}
