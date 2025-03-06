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

type fakeRemoteVersionClient struct {
	err error
}

func (c *fakeRemoteVersionClient) GetRemoteVersion(ctx context.Context) (*client.RemoteVersion, *client.Response, error) {
	return &client.RemoteVersion{
		UpdateAvailable: true,
		Version:         "1.2.3",
	}, &client.Response{}, c.err
}

func TestRemoteVersion(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeRemoteVersionClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "remote version fails",
			cl: fakeRemoteVersionClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "remote version suceeds",
			cl:   fakeRemoteVersionClient{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newRemoteVersionCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRemoteVersionCollect(t *testing.T) {
	cl := fakeRemoteVersionClient{}

	c := newMultiCollector(newRemoteVersionCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_remote_version_update_available Whether an update is available.
# TYPE paperless_remote_version_update_available gauge
paperless_remote_version_update_available{version="1.2.3"} 1
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
