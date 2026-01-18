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

type CPUStats struct {
	User   uint64
	Nice   uint64
	System uint64
	Idle   uint64
	Total  uint64
}

func readCPUStats() (*CPUStats, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("failed to read /proc/stat")
	}

	line := scanner.Text()
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return nil, fmt.Errorf("invalid /proc/stat format")
	}

	stats := &CPUStats{}
	stats.User, _ = strconv.ParseUint(fields[1], 10, 64)
	stats.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
	stats.System, _ = strconv.ParseUint(fields[3], 10, 64)
	stats.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
	stats.Total = stats.User + stats.Nice + stats.System + stats.Idle

	return stats, nil
}

func calculateUsage(prev, curr *CPUStats) float64 {
	totalDiff := curr.Total - prev.Total
	idleDiff := curr.Idle - prev.Idle
	if totalDiff == 0 {
		return 0
	}
	return float64(totalDiff-idleDiff) / float64(totalDiff) * 100
}

func main() {
	interval := 60

	prevStats, err := readCPUStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read CPU stats: %v\n", err)
		return
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		currStats, err := readCPUStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read CPU stats: %v\n", err)
			continue
		}

		usage := calculateUsage(prevStats, currStats)
		prevStats = currStats

		msg := map[string]interface{}{
			"type": "metric",
			"data": map[string]interface{}{
				"name":      "cpu_usage",
				"value":     usage,
				"timestamp": time.Now().Unix(),
			},
		}

		data, _ := json.Marshal(msg)
		fmt.Println(string(data))

		prevStats = currStats
	}
}
