package cpu_test

import (
	"context"
	"runtime"
	"sync"
	"testing"

	mcpu "github.com/alexei38/monitoring/internal/monitor/cpu"
	scpu "github.com/alexei38/monitoring/internal/stats/cpu"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestCPUMetric(t *testing.T) {
	defer goleak.VerifyNone(t)
	counter := 1
	interval := 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	statCh := make(chan *scpu.Stats)
	defer close(statCh)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		mcpu.AvgStat(ctx, log.WithContext(ctx), statCh, interval, counter)
	}()

	stat := <-statCh
	cancel()
	wg.Wait()

	require.NotNil(t, stat)
	// количество CPU + общая статистика по всем
	require.Len(t, stat.CPU, runtime.NumCPU()+1)
	for _, cpu := range stat.CPU {
		require.GreaterOrEqual(t, cpu.Usr, float32(0.0))
		require.GreaterOrEqual(t, cpu.Sys, float32(0.0))
		require.GreaterOrEqual(t, cpu.Usr, float32(0.0))
	}
}
