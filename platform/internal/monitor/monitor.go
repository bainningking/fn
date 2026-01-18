package monitor

import (
	"runtime"
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// 系统指标
	Goroutines   int
	MemoryAlloc  uint64
	MemorySys    uint64
	GCPauseTotal uint64

	// 业务指标
	ActiveAgents  int
	TotalRequests int64
	FailedRequests int64
	AvgResponseTime float64

	LastUpdate time.Time
}

var globalMetrics = &Metrics{}

func GetMetrics() *Metrics {
	globalMetrics.mu.RLock()
	defer globalMetrics.mu.RUnlock()

	m := &Metrics{}
	*m = *globalMetrics
	return m
}

func UpdateSystemMetrics() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()

	globalMetrics.Goroutines = runtime.NumGoroutine()
	globalMetrics.MemoryAlloc = mem.Alloc
	globalMetrics.MemorySys = mem.Sys
	globalMetrics.GCPauseTotal = mem.PauseTotalNs
	globalMetrics.LastUpdate = time.Now()
}

func IncrementRequests() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.TotalRequests++
}

func IncrementFailedRequests() {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.FailedRequests++
}

func UpdateActiveAgents(count int) {
	globalMetrics.mu.Lock()
	defer globalMetrics.mu.Unlock()
	globalMetrics.ActiveAgents = count
}

func StartMonitoring() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			UpdateSystemMetrics()
		}
	}()
}
