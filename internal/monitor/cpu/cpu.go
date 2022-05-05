package cpu

import (
	"context"
	"time"

	"github.com/alexei38/monitoring/internal/monitor"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	"github.com/alexei38/monitoring/internal/storage/memory"
	log "github.com/sirupsen/logrus"
)

// avgCPU получает все элементы из storage,
// по каждому типу (sys, usr, idle) всех элементов вычисляет среднее число
// в результате отдает один элемент cpu.Stats со средними числами всех элементов.
func avgCPU(store memory.Storage) *cpu.Stats {
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

// AvgStat каждую секунду собирает статистику по утилизации CPU%,
// сохраняет в storage метрику, если метрик >= counter, то удаляет самую старую
// Как только накопилось количество метрик == counter,
// то пишем в канал statCh среднее значение всех сохраненных метрик.
func AvgStat(ctx context.Context, statCh chan<- *cpu.Stats, interval int, counter int) {
	var iter int
	store := memory.NewStorage()
	countErrors := 0
	tickerSec := time.NewTicker(time.Second)
	stat := cpu.NewStat()
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
			err := stat.Get()
			if err != nil {
				countErrors++
				if countErrors >= monitor.MaxErrors && store.Len() > 0 {
					// todo wrap msg
					log.Errorf("send cpu metrics with failed failed get cpu statistic: %v", err)
					statCh <- avgCPU(store)
				} else {
					log.Errorf("failed get cpu statistic: %v", err)
				}
				continue
			} else {
				// как получили успешно метрику, сбрасываем счетчик
				countErrors = 0
			}
			if store.Len() >= counter && store.Len() > 0 {
				store.Remove(store.Back())
			}
			store.PushFront(stat)
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
