package metrics

import (
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type Stats struct {
	Timestamp int64       `json:"timestamp"`
	CPU       CPUStats    `json:"cpu"`
	Memory    MemoryStats `json:"memory"`
}

type CPUStats struct {
	Total   float64   `json:"total"`
	PerCore []float64 `json:"perCore"`
}

type MemoryStats struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
}

func Snapshot() (Stats, error) {
	perCore, err := cpu.Percent(0, true)
	if err != nil {
		return Stats{}, err
	}
	totalValues, err := cpu.Percent(0, false)
	if err != nil {
		return Stats{}, err
	}
	total := 0.0
	if len(totalValues) > 0 {
		total = totalValues[0]
	} else if len(perCore) > 0 {
		sum := 0.0
		for _, value := range perCore {
			sum += value
		}
		total = sum / float64(len(perCore))
	}
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return Stats{}, err
	}
	return Stats{
		Timestamp: time.Now().UnixMilli(),
		CPU: CPUStats{
			Total:   total,
			PerCore: perCore,
		},
		Memory: MemoryStats{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		},
	}, nil
}
