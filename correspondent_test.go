package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/ref"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

type fakeCorrespondentClient struct {
	items []client.Correspondent
	err   error
}

func (c *fakeCorrespondentClient) ListAllCorrespondents(ctx context.Context, opts client.ListCorrespondentsOptions, handler func(context.Context, client.Correspondent) error) error {
	for _, i := range c.items {
		if err := handler(ctx, i); err != nil {
			return err
		}
	}

	return c.err
}

func TestCorrespondent(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeCorrespondentClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeCorrespondentClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "correspondents",
			cl: fakeCorrespondentClient{
				items: []client.Correspondent{
					{ID: 21383},
					{ID: 3096},
					{ID: 22044},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCorrespondentCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCorrespondentCollect(t *testing.T) {
	cl := fakeCorrespondentClient{}

	c := newMultiCollectorForTest(t, newCorrespondentCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.items = append(cl.items, []client.Correspondent{
		{ID: 15818, Name: "bank", Slug: "aslug"},
		{
			ID:                 167,
			Name:               "insurance",
			DocumentCount:      2,
			LastCorrespondence: ref.Ref(time.Date(2019, time.July, 1, 0, 0, 0, 0, time.UTC)),
		},
		{ID: 24467, Name: "employer", DocumentCount: 121},
	}...)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_correspondent_document_count Number of documents associated with a correspondent.
# TYPE paperless_correspondent_document_count gauge
paperless_correspondent_document_count{id="15818"} 0
paperless_correspondent_document_count{id="167"} 2
paperless_correspondent_document_count{id="24467"} 121
# HELP paperless_correspondent_info Static information about a correspondent.
# TYPE paperless_correspondent_info gauge
paperless_correspondent_info{id="15818",name="bank",slug="aslug"} 1
paperless_correspondent_info{id="167",name="insurance",slug=""} 1
paperless_correspondent_info{id="24467",name="employer",slug=""} 1
# HELP paperless_correspondent_last_correspondence_timestamp_seconds Number of seconds since 1970 of the most recent correspondence.
# TYPE paperless_correspondent_last_correspondence_timestamp_seconds gauge
paperless_correspondent_last_correspondence_timestamp_seconds{id="15818"} 0
paperless_correspondent_last_correspondence_timestamp_seconds{id="167"} 1.5619392e+09
paperless_correspondent_last_correspondence_timestamp_seconds{id="24467"} 0
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
