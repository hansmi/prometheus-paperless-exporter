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

type fakeUserClient struct {
	count int64
	err   error
}

func (c *fakeUserClient) ListUsers(ctx context.Context, opts client.ListUsersOptions) ([]client.User, *client.Response, error) {
	return nil, &client.Response{
		ItemCount: c.count,
	}, c.err
}

func TestUser(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeUserClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeUserClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "users",
			cl: fakeUserClient{
				count: 987,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newUserCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUserCollect(t *testing.T) {
	cl := fakeUserClient{
		count: client.ItemCountUnknown,
	}

	c := newMultiCollector(newUserCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)

	cl.count = 6799

	testutil.CollectAndCompare(t, c, `
# HELP paperless_users Number of users.
# TYPE paperless_users gauge
paperless_users 6799
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
