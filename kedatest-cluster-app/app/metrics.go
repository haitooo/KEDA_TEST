package main

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MetricExporter は GCP Monitoring にカスタムメトリクスを書き込む薄いラッパ
type MetricExporter struct {
	client     *monitoring.MetricClient
	projectID  string
	metricType string
}

// コンストラクタ
func newMetricExporter(ctx context.Context, projectID, metricType string) (*MetricExporter, error) {
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric client: %w", err)
	}
	return &MetricExporter{
		client:     c,
		projectID:  projectID,
		metricType: metricType,
	}, nil
}

func (e *MetricExporter) Close() error {
	return e.client.Close()
}

// ExportRequestDelta は「この間に増えたリクエスト数（delta）」を
// custom.googleapis.com/... のメトリクスとして送信する。
func (e *MetricExporter) ExportRequestDelta(ctx context.Context, delta int64) error {
	if delta <= 0 {
		return nil
	}

	now := timestamppb.Now()

	ts := &monitoringpb.TimeSeries{
		Metric: &metricpb.Metric{
			Type: e.metricType,
			Labels: map[string]string{
				"endpoint": "work",
			},
		},
		// シンプルに global リソースとして送信
		Resource: &monitoredres.MonitoredResource{
			Type: "global",
			Labels: map[string]string{
				"project_id": e.projectID,
			},
		},
		Points: []*monitoringpb.Point{{
			Interval: &monitoringpb.TimeInterval{
				EndTime: now,
			},
			Value: &monitoringpb.TypedValue{
				Value: &monitoringpb.TypedValue_Int64Value{
					Int64Value: delta,
				},
			},
		}},
	}

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name:       "projects/" + e.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{ts},
	}

	if err := e.client.CreateTimeSeries(ctx, req); err != nil {
		return fmt.Errorf("CreateTimeSeries failed: %w", err)
	}
	return nil
}
