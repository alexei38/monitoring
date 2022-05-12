package load

import (
	"context"
	"time"

	"github.com/alexei38/monitoring/internal/monitor"
	"github.com/alexei38/monitoring/internal/stats/load"
	"github.com/alexei38/monitoring/internal/storage/memory"
	log "github.com/sirupsen/logrus"
)

func avgLoad(store memory.Storage) *load.Stats {
	load1 := make([]float32, store.Len())
	load5 := make([]float32, store.Len())
	load15 := make([]float32, store.Len())
	for i, item := range store.List() {
		stat := item.Value.(*load.Stats)
		load1[i] = stat.Load1
		load5[i] = stat.Load5
		load15[i] = stat.Load15
	}
	result := &load.Stats{}
	result.Load1 = monitor.AvgFloat(load1...)
	result.Load5 = monitor.AvgFloat(load5...)
	result.Load15 = monitor.AvgFloat(load15...)
	return result
}

func AvgStat(ctx context.Context, log *log.Entry, ch chan<- *load.Stats, interval int, counter int) {
	var iter int
	store := memory.NewStorage()
	tickerSec := time.NewTicker(time.Second)
	stat := load.NewStat()
	for {
		select {
		case <-ctx.Done():
			log.Info("stop collect")
			return
		default:
		}
		select {
		case <-ctx.Done():
			log.Info("stop collect")
			return
		case <-tickerSec.C:
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
					case ch <- avgLoad(store):
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
