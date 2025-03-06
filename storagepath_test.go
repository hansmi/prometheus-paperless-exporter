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

type fakeStoragePathClient struct {
	items []client.StoragePath
	err   error
}

func (c *fakeStoragePathClient) ListAllStoragePaths(ctx context.Context, opts client.ListStoragePathsOptions, handler func(context.Context, client.StoragePath) error) error {
	for _, i := range c.items {
		if err := handler(ctx, i); err != nil {
			return err
		}
	}

	return c.err
}

func TestStoragePath(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeStoragePathClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeStoragePathClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "storagePaths",
			cl: fakeStoragePathClient{
				items: []client.StoragePath{
					{ID: 70},
					{ID: 20667},
					{ID: 12805},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newStoragePathCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStoragePathCollect(t *testing.T) {
	cl := fakeStoragePathClient{}

	c := newMultiCollector(newStoragePathCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.items = append(cl.items, []client.StoragePath{
		{ID: 23547, Name: "personal", Slug: "personal"},
		{ID: 704, Name: "work", DocumentCount: 13},
	}...)

	testutil.CollectAndCompare(t, c, `
# HELP paperless_storage_path_document_count Number of documents associated with a storage path.
# TYPE paperless_storage_path_document_count gauge
paperless_storage_path_document_count{id="23547"} 0
paperless_storage_path_document_count{id="704"} 13
# HELP paperless_storage_path_info Static information about a storage path.
# TYPE paperless_storage_path_info gauge
paperless_storage_path_info{id="23547",name="personal",slug="personal"} 1
paperless_storage_path_info{id="704",name="work",slug=""} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
