package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hansmi/paperhooks/pkg/client"
	"github.com/hansmi/prometheus-paperless-exporter/internal/testutil"
)

type fakeStatusClient struct {
	err error
}

func (c *fakeStatusClient) GetStatus(ctx context.Context) (*client.SystemStatus, *client.Response, error) {
	return &client.SystemStatus{
		PNGXVersion: "2.14.7",
		ServerOS:    "Linux-6.8.12-8-pve-x86_64-with-glibc2.36",
		InstallType: "bare-metal",
		Storage: client.SystemStatusStorage{
			Total:     21474836480,
			Available: 13406437376,
		},
		Database: client.SystemStatusDatabase{
			Type:   "postgresql",
			URL:    "paperlessdb",
			Status: "OK",
			Error:  "",
			MigrationStatus: client.SystemStatusDatabaseMigration{
				LatestMigration:     "mfa.0003_authenticator_type_uniq",
				UnappliedMigrations: []string{},
			},
		},
		Tasks: client.SystemStatusTasks{
			RedisURL:              "redis://localhost:6379",
			RedisStatus:           "OK",
			RedisError:            "",
			CeleryStatus:          "OK",
			IndexStatus:           "OK",
			IndexLastModified:     time.Date(2025, time.February, 21, 0, 1, 54, 773392000, time.UTC),
			IndexError:            "",
			ClassifierStatus:      "OK",
			ClassifierLastTrained: time.Date(2025, time.February, 21, 20, 5, 1, 589548000, time.UTC),
			ClassifierError:       "",
			SanityCheckStatus:     "OK",
			SanityCheckLastRun:    time.Date(2025, time.February, 22, 15, 30, 0, 0, time.UTC),
			SanityCheckError:      "",
		},
	}, &client.Response{}, c.err
}

func TestStatus(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeStatusClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "get status fails",
			cl: fakeStatusClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "get status succeeds",
			cl:   fakeStatusClient{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newStatusCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStatusCollect(t *testing.T) {
	cl := fakeStatusClient{}

	c := newMultiCollectorForTest(t, newStatusCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_status_celery_status Status of celery. 1 is OK, 0 is not OK.
# TYPE paperless_status_celery_status gauge
paperless_status_celery_status 1
# HELP paperless_status_classifier_last_trained_timestamp_seconds Number of seconds since 1970-01-01 since the last time the classifier has been trained.
# TYPE paperless_status_classifier_last_trained_timestamp_seconds gauge
paperless_status_classifier_last_trained_timestamp_seconds 1.740168301e+09
# HELP paperless_status_classifier_status Status of the classifier. 1 is OK, 0 is not OK.
# TYPE paperless_status_classifier_status gauge
paperless_status_classifier_status 1
# HELP paperless_status_database_status Status of the database. 1 is OK, 0 is not OK.
# TYPE paperless_status_database_status gauge
paperless_status_database_status 1
# HELP paperless_status_database_unapplied_migrations Number of unapplied database migrations.
# TYPE paperless_status_database_unapplied_migrations gauge
paperless_status_database_unapplied_migrations 0
# HELP paperless_status_index_last_modified_timestamp_seconds Number of seconds since 1970-01-01 since the last time the index has been modified.
# TYPE paperless_status_index_last_modified_timestamp_seconds gauge
paperless_status_index_last_modified_timestamp_seconds 1.740096114e+09
# HELP paperless_status_index_status Status of the index. 1 is OK, 0 is not OK.
# TYPE paperless_status_index_status gauge
paperless_status_index_status 1
# HELP paperless_status_redis_status Status of redis. 1 is OK, 0 is not OK.
# TYPE paperless_status_redis_status gauge
paperless_status_redis_status 1
# HELP paperless_status_sanity_check_last_run_timestamp_seconds Number of seconds since 1970-01-01 since the last time the sanity check has been run.
# TYPE paperless_status_sanity_check_last_run_timestamp_seconds gauge
paperless_status_sanity_check_last_run_timestamp_seconds 1.7402382e+09
# HELP paperless_status_sanity_check_status Status of the sanity check. 1 is OK, 0 is not OK.
# TYPE paperless_status_sanity_check_status gauge
paperless_status_sanity_check_status 1
# HELP paperless_status_storage_available_bytes Available storage of Paperless in bytes.
# TYPE paperless_status_storage_available_bytes gauge
paperless_status_storage_available_bytes 1.3406437376e+10
# HELP paperless_status_storage_total_bytes Total storage of Paperless in bytes.
# TYPE paperless_status_storage_total_bytes gauge
paperless_status_storage_total_bytes 2.147483648e+10
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total{category="unspecified"} 0
`)
}
