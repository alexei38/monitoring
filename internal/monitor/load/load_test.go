package load_test

import (
	"context"
	"sync"
	"testing"
	"time"

	mload "github.com/alexei38/monitoring/internal/monitor/load"
	sload "github.com/alexei38/monitoring/internal/stats/load"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestCPUMetric(t *testing.T) {
	defer goleak.VerifyNone(t)
	counter := 1
	interval := 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	statCh := make(chan *sload.Stats)
	defer close(statCh)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		mload.AvgStat(ctx, statCh, interval, counter)
	}()

	var stat *sload.Stats
	require.Eventually(t, func() bool {
		select {
		case stat = <-statCh:
			return true
		default:
			return false
		}
	}, time.Second*20, time.Second)

	cancel()
	wg.Wait()

	require.NotNil(t, stat)
	// load average должен быть больше 0
	require.GreaterOrEqual(t, stat.Load1, float32(0.0))
	require.GreaterOrEqual(t, stat.Load5, float32(0.0))
	require.GreaterOrEqual(t, stat.Load15, float32(0.0))
}
