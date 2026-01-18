package monitor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMetrics(t *testing.T) {
	metrics := GetMetrics()
	assert.NotNil(t, metrics)
}

func TestUpdateSystemMetrics(t *testing.T) {
	UpdateSystemMetrics()

	metrics := GetMetrics()
	assert.Greater(t, metrics.Goroutines, 0)
	assert.Greater(t, metrics.MemoryAlloc, uint64(0))
	assert.Greater(t, metrics.MemorySys, uint64(0))
	assert.False(t, metrics.LastUpdate.IsZero())
}

func TestIncrementRequests(t *testing.T) {
	globalMetrics = &Metrics{}

	IncrementRequests()
	IncrementRequests()

	metrics := GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalRequests)
}

func TestIncrementFailedRequests(t *testing.T) {
	globalMetrics = &Metrics{}

	IncrementFailedRequests()

	metrics := GetMetrics()
	assert.Equal(t, int64(1), metrics.FailedRequests)
}

func TestUpdateActiveAgents(t *testing.T) {
	globalMetrics = &Metrics{}

	UpdateActiveAgents(5)

	metrics := GetMetrics()
	assert.Equal(t, 5, metrics.ActiveAgents)
}

func TestStartMonitoring(t *testing.T) {
	globalMetrics = &Metrics{}

	StartMonitoring()

	time.Sleep(100 * time.Millisecond)

	metrics := GetMetrics()
	assert.NotNil(t, metrics)
}
