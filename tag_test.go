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

type fakeTagClient struct {
	items []client.Tag
	err   error
}

func (c *fakeTagClient) ListAllTags(ctx context.Context, opts client.ListTagsOptions, handler func(context.Context, client.Tag) error) error {
	for _, i := range c.items {
		if err := handler(ctx, i); err != nil {
			return err
		}
	}

	return c.err
}

func TestTag(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeTagClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeTagClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "tags",
			cl: fakeTagClient{
				items: []client.Tag{
					{ID: 8463},
					{ID: 8463},
					{ID: 338},
					{ID: 11768},
					{ID: 30619},
					{ID: 27086},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newTagCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTagCollect(t *testing.T) {
	cl := fakeTagClient{}

	c := newMultiCollector(newTagCollector(&cl))

	testutil.CollectAndCompare(t, c, ``)

	cl.items = append(cl.items, []client.Tag{
		{ID: 8463, Name: "aaa", Slug: "aslug"},
		{ID: 338, Name: "three-three-eight", DocumentCount: 13},
		{ID: 26429, Name: "last"},
	}...)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_tag_document_count Number of documents associated with a tag.
# TYPE paperless_tag_document_count gauge
paperless_tag_document_count{id="26429"} 0
paperless_tag_document_count{id="338"} 13
paperless_tag_document_count{id="8463"} 0
# HELP paperless_tag_info Static information about a tag.
# TYPE paperless_tag_info gauge
paperless_tag_info{id="26429",name="last",slug=""} 1
paperless_tag_info{id="338",name="three-three-eight",slug=""} 1
paperless_tag_info{id="8463",name="aaa",slug="aslug"} 1
`)
}
