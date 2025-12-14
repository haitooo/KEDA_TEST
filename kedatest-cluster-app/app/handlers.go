package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// リクエスト数カウンタ（/work 用）
var (
	reqMu         sync.Mutex
	totalRequests int64
	lastExported  int64
)

// Stats は /stats で返す情報（最小構成）
type Stats struct {
	TotalRequests int64 `json:"total_requests"`
}

// リクエスト数を +1
func incrementRequests() {
	reqMu.Lock()
	defer reqMu.Unlock()
	totalRequests++
}

// メトリクス送信用に「前回送信からの増分」を取得
func snapshotRequestDelta() int64 {
	reqMu.Lock()
	defer reqMu.Unlock()
	delta := totalRequests - lastExported
	lastExported = totalRequests
	return delta
}

// /healthz: シンプルなヘルスチェック
func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// /work: CPU / メモリ負荷をかけるエンドポイント
// 例: POST /work?cpu_ms=20&mem_mb=50
func handleWork(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	cpuMs, _ := strconv.Atoi(q.Get("cpu_ms"))
	if cpuMs < 0 {
		cpuMs = 0
	}
	memMB, _ := strconv.Atoi(q.Get("mem_mb"))
	if memMB < 0 {
		memMB = 0
	}

	// CPU負荷
	if cpuMs > 0 {
		doCPULoadMs(cpuMs)
	}
	// メモリ負荷（確保して触るだけ。参照を捨てるので GC 対象）
	if memMB > 0 {
		allocateMemoryOnce(memMB)
	}

	// リクエスト数カウント
	incrementRequests()

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]interface{}{
		"message": "work done",
		"cpu_ms":  cpuMs,
		"mem_mb":  memMB,
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// /stats: 内部状態の確認（トータルリクエストだけ）
func handleStats(w http.ResponseWriter, r *http.Request) {
	reqMu.Lock()
	tr := totalRequests
	reqMu.Unlock()

	s := Stats{
		TotalRequests: tr,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s)
}

// 指定ミリ秒間 busy-loop で CPU を使う
func doCPULoadMs(ms int) {
	if ms <= 0 {
		return
	}
	d := time.Duration(ms) * time.Millisecond
	end := time.Now().Add(d)
	var x uint64
	for time.Now().Before(end) {
		x++
		if x == 0 {
			x = 1
		}
	}
	_ = x
}

// memMB MB のバイトスライスを一度だけ確保してちょっと触る
// キャッシュはせず、そのまま関数を抜けて GC 対象になる。
func allocateMemoryOnce(memMB int) {
	if memMB <= 0 {
		return
	}
	size := memMB * 1024 * 1024
	buf := make([]byte, size)

	// 物理メモリを使わせるためにページごとに軽く触る
	for i := 0; i < len(buf); i += 4096 {
		buf[i] = 1
	}
}
