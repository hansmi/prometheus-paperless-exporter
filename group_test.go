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

type fakeGroupClient struct {
	count int64
	err   error
}

func (c *fakeGroupClient) ListGroups(ctx context.Context, opts client.ListGroupsOptions) ([]client.Group, *client.Response, error) {
	return nil, &client.Response{
		ItemCount: c.count,
	}, c.err
}

func TestGroup(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeGroupClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "listing fails",
			cl: fakeGroupClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "groups",
			cl: fakeGroupClient{
				count: 987,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newGroupCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGroupCollect(t *testing.T) {
	cl := fakeGroupClient{
		count: client.ItemCountUnknown,
	}

	c := newMultiCollector(newGroupCollector(&cl))

	testutil.CollectAndCompare(t, c, ``)

	cl.count = 321

	testutil.CollectAndCompare(t, c, `
# HELP paperless_groups Number of user groups.
# TYPE paperless_groups gauge
paperless_groups 321
`)
}
