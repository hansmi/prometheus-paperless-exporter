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

type fakeDocumentTypeClient struct {
	items []client.DocumentType
	err   error
}

func (c *fakeDocumentTypeClient) ListAllDocumentTypes(ctx context.Context, opts client.ListDocumentTypesOptions, handler func(context.Context, client.DocumentType) error) error {
	for _, i := range c.items {
		if err := handler(ctx, i); err != nil {
			return err
		}
	}

	return c.err
}

func TestDocumentType(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeDocumentTypeClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeDocumentTypeClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "documentTypes",
			cl: fakeDocumentTypeClient{
				items: []client.DocumentType{
					{ID: 23758},
					{ID: 22848},
					{ID: 10504},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newDocumentTypeCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDocumentTypeCollect(t *testing.T) {
	cl := fakeDocumentTypeClient{}

	c := newMultiCollectorForTest(t, newDocumentTypeCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.items = append(cl.items, []client.DocumentType{
		{ID: 3760, Name: "Contract", Slug: "contract"},
		{ID: 5558, Name: "Purchase order", Slug: "po", DocumentCount: 20},
	}...)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_document_type_document_count Number of documents associated with a document type.
# TYPE paperless_document_type_document_count gauge
paperless_document_type_document_count{id="3760"} 0
paperless_document_type_document_count{id="5558"} 20
# HELP paperless_document_type_info Static information about a document type.
# TYPE paperless_document_type_info gauge
paperless_document_type_info{id="3760",name="Contract",slug="contract"} 1
paperless_document_type_info{id="5558",name="Purchase order",slug="po"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
