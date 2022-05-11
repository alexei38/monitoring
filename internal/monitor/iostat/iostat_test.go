package cpu_test

import (
	"context"
	"sync"
	"testing"

	miostat "github.com/alexei38/monitoring/internal/monitor/iostat"
	siostat "github.com/alexei38/monitoring/internal/stats/iostat"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestIOMetric(t *testing.T) {
	defer goleak.VerifyNone(t)
	counter := 1
	interval := 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	statCh := make(chan *siostat.Stats)
	defer close(statCh)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		miostat.AvgStat(ctx, statCh, interval, counter)
	}()

	stat := <-statCh
	cancel()
	wg.Wait()

	require.NotNil(t, stat)
	for _, disk := range stat.Disk {
		require.GreaterOrEqual(t, disk.Rkbs, float32(0.0))
		require.GreaterOrEqual(t, disk.Wkbs, float32(0.0))
		require.GreaterOrEqual(t, disk.Util, float32(0.0))
	}
}
