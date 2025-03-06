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

type fakeStatisticsClient struct {
	err error
}

func (c *fakeStatisticsClient) GetStatistics(ctx context.Context) (*client.Statistics, *client.Response, error) {
	statistics := &client.Statistics{
		DocumentsTotal: 1447,
		DocumentsInbox: 273,
		InboxTag:       1,
		InboxTags:      []int64{1},
		DocumentFileTypeCounts: []client.StatisticsDocumentFileType{
			{
				MimeType:      "application/pdf",
				MimeTypeCount: 1397,
			},
			{
				MimeType:      "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
				MimeTypeCount: 37,
			},
		},
		CharacterCount:     11048372,
		TagCount:           55,
		CorrespondentCount: 201,
		DocumentTypeCount:  42,
		StoragePathCount:   0,
		CurrentAsn:         0,
	}
	return statistics, &client.Response{}, c.err
}

func TestStatistics(t *testing.T) {
	errTest := errors.New("test error")

	for _, tc := range []struct {
		name    string
		cl      fakeStatisticsClient
		wantErr error
	}{
		{
			name: "empty",
		},
		{
			name: "get statistics fails",
			cl: fakeStatisticsClient{
				err: errTest,
			},
			wantErr: errTest,
		},
		{
			name: "get statistics works",
			cl: fakeStatisticsClient{
				err: nil,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newStatisticsCollector(&tc.cl)

			err := c.collect(context.Background(), testutil.DiscardMetrics(t))

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStatisticsCollect(t *testing.T) {
	cl := fakeStatisticsClient{}

	c := newMultiCollector(newStatisticsCollector(&cl))

	testutil.CollectAndCompare(t, c, `
# HELP paperless_statistics_character_count Number of characters stored across the total number of documents.
# TYPE paperless_statistics_character_count gauge
paperless_statistics_character_count 1.1048372e+07
# HELP paperless_statistics_correspondent_count Total number of correspondents.
# TYPE paperless_statistics_correspondent_count gauge
paperless_statistics_correspondent_count 201
# HELP paperless_statistics_document_type_count Total number of document types.
# TYPE paperless_statistics_document_type_count gauge
paperless_statistics_document_type_count 42
# HELP paperless_statistics_documents_file_type_counts Total number of documents per MIME type.
# TYPE paperless_statistics_documents_file_type_counts gauge
paperless_statistics_documents_file_type_counts{mime_type="application/pdf"} 1397
paperless_statistics_documents_file_type_counts{mime_type="application/vnd.openxmlformats-officedocument.wordprocessingml.document"} 37
# HELP paperless_statistics_documents_inbox_count Total number of documents that have the defined 'Inbox' tag.
# TYPE paperless_statistics_documents_inbox_count gauge
paperless_statistics_documents_inbox_count 273
# HELP paperless_statistics_documents_total Total number of documents.
# TYPE paperless_statistics_documents_total gauge
paperless_statistics_documents_total 1447
# HELP paperless_statistics_storage_path_count Total number of storage pathes.
# TYPE paperless_statistics_storage_path_count gauge
paperless_statistics_storage_path_count 0
# HELP paperless_statistics_tag_count Total number of tags.
# TYPE paperless_statistics_tag_count gauge
paperless_statistics_tag_count 55
# HELP paperless_warnings_total Number of warnings generated while scraping metrics.
# TYPE paperless_warnings_total gauge
paperless_warnings_total 0
`)
}
