package main

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

//go:generate go run golang.org/x/tools/cmd/stringer -linecomment -type=warningCategory -output=warning_string.go
type warningCategory int

const (
	warningCategoryUnspecified      warningCategory = iota // unspecified
	warningCategoryGetRemoteVersion                        // get_remote_version
)

// warning is a special form of a metric and suitable for reporting non-fatal
// errors during a scrape. Warnings are only logged and not forwarded to the
// registry.
type warning struct {
	err error
}

var _ prometheus.Metric = (*warning)(nil)

func newWarning(err error) *warning {
	return &warning{err}
}

func (*warning) Desc() *prometheus.Desc {
	return nil
}

func (*warning) Write(*dto.Metric) error {
	return os.ErrInvalid
}
