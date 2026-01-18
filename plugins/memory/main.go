package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type MemoryStats struct {
	Total     uint64
	Available uint64
	Used      uint64
	UsageRate float64
}

func readMemoryStats() (*MemoryStats, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stats := &MemoryStats{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseUint(fields[1], 10, 64)

		switch key {
		case "MemTotal":
			stats.Total = value
		case "MemAvailable":
			stats.Available = value
		}
	}

	if stats.Total > 0 {
		stats.Used = stats.Total - stats.Available
		stats.UsageRate = float64(stats.Used) / float64(stats.Total) * 100
	}

	return stats, nil
}

func main() {
	interval := 60

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats, err := readMemoryStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read memory stats: %v\n", err)
			continue
		}

		msg := map[string]interface{}{
			"type": "metric",
			"data": map[string]interface{}{
				"name":       "memory_usage",
				"value":      stats.UsageRate,
				"used_kb":    stats.Used,
				"available_kb": stats.Available,
				"timestamp":  time.Now().Unix(),
			},
		}

		data, _ := json.Marshal(msg)
		fmt.Println(string(data))
	}
}
