package inode

import (
	"context"
	"time"

	"github.com/alexei38/monitoring/internal/monitor"
	"github.com/alexei38/monitoring/internal/stats/disk/inode"
	"github.com/alexei38/monitoring/internal/storage/memory"
	log "github.com/sirupsen/logrus"
)

// avgDiskInode получает все элементы из storage,
// по каждому примонтированному устройству, каждой метрики вычисляет среднее число.
func avgDiskInode(store memory.Storage) *inode.Stats {
	stats := map[string]map[string][]interface{}{}
	for _, item := range store.List() {
		storeStat := item.Value.(*inode.Stats)
		for _, stat := range storeStat.Stat {
			// Mount более уникальное для агрегации,
			// т.к Device может быть примонтирован в несколько каталогов
			if _, ok := stats[stat.Mount]; !ok {
				stats[stat.Mount] = map[string][]interface{}{}
			}
			stats[stat.Mount]["available"] = append(stats[stat.Mount]["available"], stat.Available)
			stats[stat.Mount]["used"] = append(stats[stat.Mount]["used"], stat.Used)
			stats[stat.Mount]["mount"] = append(stats[stat.Mount]["mount"], stat.Mount)
			stats[stat.Mount]["type"] = append(stats[stat.Mount]["type"], stat.TypeFS)
			stats[stat.Mount]["device"] = append(stats[stat.Mount]["device"], stat.Device)
		}
	}
	result := &inode.Stats{}
	for mount, params := range stats {
		stat := inode.Stat{Mount: mount}
		for metric, values := range params {
			switch metric {
			case "used":
				var usedData []int64
				for _, v := range values {
					usedData = append(usedData, v.(int64))
				}
				stat.Used = monitor.AvgInt64(usedData...)
			case "available":
				var availableData []int64
				for _, v := range values {
					availableData = append(availableData, v.(int64))
				}
				stat.Available = monitor.AvgInt64(availableData...)
			case "mount":
				stat.Mount = values[0].(string)
			case "type":
				stat.TypeFS = values[0].(string)
			case "device":
				stat.Device = values[0].(string)
			}
		}
		result.Stat = append(result.Stat, stat)
	}
	return result
}

func AvgStat(ctx context.Context, log *log.Entry, ch chan<- *inode.Stats, interval int, counter int) {
	var iter int
	store := memory.NewStorage()
	stat := inode.NewStat()
	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info("stop collect")
			return
		case <-tick.C:
			err := stat.Get()
			if err != nil {
				log.Errorf("failed get metric: %v", err)
				continue
			}

			if store.Len() >= counter && store.Len() > 0 {
				log.Debug("remove last metric from storage")
				store.Remove(store.Back())
			}
			log.Debugf("save metric to store: {%v}", stat)
			store.PushFront(stat)
			if store.Len() >= counter-interval {
				if iter+1 == interval {
					select {
					case ch <- avgDiskInode(store):
						log.Debugf("send to channel")
						break
					case <-ctx.Done():
						log.Info("stop collect")
						return
					}
					iter = 0
				} else {
					log.Debugf("continue waiting interval")
					iter++
				}
			}
		}
	}
}
