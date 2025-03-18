package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

type fakeDocumentClient struct {
	count int64
	err   error
}

func (c *fakeDocumentClient) ListDocuments(ctx context.Context, opts client.ListDocumentsOptions) ([]client.Document, *client.Response, error) {
	return nil, &client.Response{
		ItemCount: c.count,
	}, c.err
}

func TestDocument(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeDocumentClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeDocumentClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "documents",
			cl: fakeDocumentClient{
				count: 987,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newDocumentCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDocumentCollect(t *testing.T) {
	cl := fakeDocumentClient{
		count: client.ItemCountUnknown,
	}

	c := newMultiCollectorForTest(t, newDocumentCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)

	cl.count = 11921

	testutil.CollectAndCompare(t, c, `
# HELP paperless_documents Number of documents.
# TYPE paperless_documents gauge
paperless_documents 11921
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)
}
