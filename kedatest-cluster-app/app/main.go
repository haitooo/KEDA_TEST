package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	// GCP プロジェクトID
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	// カスタムメトリクス名 (HPA/KEDA から参照)
	metricType := os.Getenv("METRIC_TYPE")
	if metricType == "" {
		metricType = "custom.googleapis.com/testapi/request_count"
	}

	// メトリクス送信間隔（秒）
	intervalSec := 10
	if v := os.Getenv("METRIC_PUSH_INTERVAL_SEC"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			intervalSec = n
		}
	}

	ctx := context.Background()
	var exporter *MetricExporter

	if projectID == "" {
		log.Println("[WARN] GCP_PROJECT / GOOGLE_CLOUD_PROJECT が未設定のため、Cloud Monitoring への送信は行いません。")
	} else {
		var err error
		exporter, err = newMetricExporter(ctx, projectID, metricType)
		if err != nil {
			log.Printf("[ERROR] MetricExporter 初期化失敗: %v", err)
		} else {
			defer exporter.Close()
			log.Printf("MetricExporter initialized. projectID=%s metricType=%s interval=%ds",
				projectID, metricType, intervalSec)

			// リクエスト数の増分を定期的に Cloud Monitoring に送信
			go func() {
				ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
				defer ticker.Stop()
				for range ticker.C {
					delta := snapshotRequestDelta()
					if delta <= 0 {
						continue
					}
					if err := exporter.ExportRequestDelta(ctx, delta); err != nil {
						log.Printf("[ERROR] export metric failed: %v", err)
					} else {
						log.Printf("[INFO] exported request_count delta=%d", delta)
					}
				}
			}()
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/work", handleWork)
	mux.HandleFunc("/stats", handleStats)

	addr := ":8080"
	log.Printf("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
