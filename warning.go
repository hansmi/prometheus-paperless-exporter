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
// errors during a scrape. Warning messages are logged and counts per category
// reported as a metric.
type warning struct {
	category warningCategory
	err      error
}

var _ prometheus.Metric = (*warning)(nil)

func newWarning(category warningCategory, err error) *warning {
	return &warning{category, err}
}

func (*warning) Desc() *prometheus.Desc {
	return nil
}

func (*warning) Write(*dto.Metric) error {
	return os.ErrInvalid
}
