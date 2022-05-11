package cpu

import (
	"context"

	"github.com/alexei38/monitoring/internal/monitor"
	"github.com/alexei38/monitoring/internal/stats/iostat"
	"github.com/alexei38/monitoring/internal/storage/memory"
	log "github.com/sirupsen/logrus"
)

// avgIO получает все элементы из storage,
// по каждому типу (rkbs, wkbs, util) всех элементов вычисляет среднее число
// в результате отдает один элемент iostat.Stats со средними числами всех элементов.
func avgIO(store memory.Storage) *iostat.Stats {
	stats := map[string]map[string][]float32{}
	for _, item := range store.List() {
		stat := item.Value.(*iostat.Stats)
		for _, ioStat := range stat.Disk {
			if _, ok := stats[ioStat.Device]; !ok {
				stats[ioStat.Device] = map[string][]float32{}
			}
			stats[ioStat.Device]["rkbs"] = append(stats[ioStat.Device]["rkbs"], ioStat.Rkbs)
			stats[ioStat.Device]["wkbs"] = append(stats[ioStat.Device]["wkbs"], ioStat.Wkbs)
			stats[ioStat.Device]["util"] = append(stats[ioStat.Device]["util"], ioStat.Util)
		}
	}
	result := &iostat.Stats{}
	for device, params := range stats {
		stat := iostat.Stat{Device: device}
		for metric, values := range params {
			switch metric {
			case "rkbs":
				stat.Rkbs = monitor.AvgFloat(values...)
			case "wkbs":
				stat.Wkbs = monitor.AvgFloat(values...)
			case "util":
				stat.Util = monitor.AvgFloat(values...)
			}
		}
		result.Disk = append(result.Disk, stat)
	}
	return result
}

// AvgStat каждую секунду собирает статистику по утилизации IO%,
// сохраняет в storage метрику, если метрик >= counter, то удаляет самую старую
// Как только накопилось количество метрик == counter,
// то пишем в канал statCh среднее значение всех сохраненных метрик.
func AvgStat(ctx context.Context, ch chan<- *iostat.Stats, interval int, counter int) {
	var iter int
	store := memory.NewStorage()
	stat := iostat.NewStat()
	// tick := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			err := stat.Get()
			if err != nil {
				log.Errorf("failed get cpu statistic: %v", err)
				continue
			}

			if store.Len() >= counter && store.Len() > 0 {
				store.Remove(store.Back())
			}
			store.PushFront(stat)
			if store.Len() >= counter-interval {
				if iter == interval {
					select {
					case ch <- avgIO(store):
						break
					case <-ctx.Done():
						log.Info("Cancel send iostat metric")
						return
					}
					iter = 0
				}
				iter++
			}
		}
	}
}
