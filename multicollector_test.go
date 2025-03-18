package main

import (
	"errors"
	"io"
	"log"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func newMultiCollectorForTest(t *testing.T, m multiCollectorMember) *multiCollector {
	t.Helper()

	c := newMultiCollector(m)
	c.logger = log.New(io.Discard, "", 0)

	return c
}

func TestFormatWarnings(t *testing.T) {
	errTest := errors.New("test error")
	errTest2 := errors.New("second error")

	for _, tc := range []struct {
		name     string
		warnings map[warningCategory][]error
		want     string
	}{
		{name: "empty"},
		{
			name: "empty slices",
			warnings: map[warningCategory][]error{
				warningCategoryUnspecified:      nil,
				warningCategoryGetRemoteVersion: nil,
			},
		},
		{
			name: "single",
			warnings: map[warningCategory][]error{
				warningCategoryUnspecified:      nil,
				warningCategoryGetRemoteVersion: []error{errTest},
			},
			want: `get_remote_version: ["test error"]`,
		},
		{
			name: "multiple",
			warnings: map[warningCategory][]error{
				warningCategoryUnspecified:      nil,
				warningCategoryGetRemoteVersion: []error{errTest, errTest2},
			},
			want: `get_remote_version: ["second error" "test error"]`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, formatWarnings(tc.warnings)); diff != "" {
				t.Errorf("formatWarnings(%q) diff (-want +got):\n%s", tc.warnings, diff)
			}
		})
	}
}
