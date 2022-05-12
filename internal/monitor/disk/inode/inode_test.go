package inode_test

import (
	"context"
	"sync"
	"testing"

	mDisk "github.com/alexei38/monitoring/internal/monitor/disk/inode"
	sDisk "github.com/alexei38/monitoring/internal/stats/disk/inode"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestDiskUsageMetric(t *testing.T) {
	defer goleak.VerifyNone(t)
	counter := 1
	interval := 1
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan *sDisk.Stats)
	defer close(ch)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		mDisk.AvgStat(ctx, log.WithContext(ctx), ch, interval, counter)
	}()
	stat := <-ch
	cancel()
	wg.Wait()

	require.NotNil(t, stat)
	for _, s := range stat.Stat {
		require.NotEmpty(t, s.Mount)
		require.NotEmpty(t, s.Device)
		require.NotEmpty(t, s.TypeFS)
		if s.TypeFS == "devtmpfs" || s.TypeFS == "tmpfs" || s.TypeFS == "squashfs" {
			require.GreaterOrEqual(t, s.Available, int64(0))
			require.GreaterOrEqual(t, s.Used, int64(0))
		} else {
			require.Greater(t, s.Available, int64(0))
			require.Greater(t, s.Used, int64(0))
		}
	}
}
