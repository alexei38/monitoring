package cpu

import (
	"context"
	"time"

	"github.com/alexei38/monitoring/internal/monitor"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	"github.com/alexei38/monitoring/internal/storage/memory"
	log "github.com/sirupsen/logrus"
)

func avgCPU(store memory.Storage) *cpu.Stats {
	// stats[cpuID][sys|usr|idle][10.2, 13.3, 15.6]
	stats := map[string]map[string][]float32{}
	for _, item := range store.List() {
		stat := item.Value.(*cpu.Stats)
		for _, cpuStat := range stat.CPU {
			if _, ok := stats[cpuStat.CPU]; !ok {
				stats[cpuStat.CPU] = map[string][]float32{}
			}
			stats[cpuStat.CPU]["sys"] = append(stats[cpuStat.CPU]["sys"], cpuStat.Sys)
			stats[cpuStat.CPU]["usr"] = append(stats[cpuStat.CPU]["usr"], cpuStat.Usr)
			stats[cpuStat.CPU]["idle"] = append(stats[cpuStat.CPU]["idle"], cpuStat.Idle)
		}
	}
	result := &cpu.Stats{}
	for cpuID, params := range stats {
		stat := cpu.Stat{CPU: cpuID}
		for metric, values := range params {
			switch metric {
			case "sys":
				stat.Sys = monitor.AvgFloat(values...)
			case "usr":
				stat.Usr = monitor.AvgFloat(values...)
			case "idle":
				stat.Idle = monitor.AvgFloat(values...)
			}
		}
		result.CPU = append(result.CPU, stat)
	}
	return result
}

func AvgStat(ctx context.Context, statCh chan<- *cpu.Stats, interval int, counter int) {
	var iter int
	store := memory.NewStorage()
	countErrors := 0
	tickerSec := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case <-tickerSec.C:
			laStat, err := cpu.GetStat()
			if err != nil {
				log.Errorf("failed get cpu statistic: %v", err)
				countErrors++
				if countErrors >= monitor.MaxErrors {
					statCh <- avgCPU(store)
				}
				continue
			}
			if store.Len() >= counter && store.Len() > 0 {
				store.Remove(store.Back())
			}
			store.PushFront(laStat)
			if store.Len() >= counter-interval {
				if iter == interval {
					statCh <- avgCPU(store)
					iter = 0
				}
				iter++
			}
		}
	}
}
