package integration_test

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alexei38/monitoring/internal/config"
	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/logger"
	"github.com/alexei38/monitoring/pkg/cli/client"
	"github.com/alexei38/monitoring/pkg/cli/server"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Metrics struct {
	cpuMetrics  int32
	loadMetrics int32
	ioMetrics   int32
}

func (m *Metrics) Append(item *pb.Metrics) {
	switch {
	case item.CPU != nil:
		atomic.AddInt32(&m.cpuMetrics, 1)
	case item.Load != nil:
		atomic.AddInt32(&m.loadMetrics, 1)
	case item.IOStat != nil:
		atomic.AddInt32(&m.ioMetrics, 1)
	}
}

func (m *Metrics) Len() int32 {
	return m.cpuMetrics + m.loadMetrics + m.ioMetrics
}

func getEvent(ctx context.Context, metrics *Metrics, stream pb.StreamService_FetchResponseClient) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				return nil
			}
			if e, ok := status.FromError(err); ok {
				// пропускаем линтер, нам не нужно обрабатывать другие ошибки в этом месте
				switch e.Code() { // nolint: exhaustive
				case codes.Canceled, codes.Unavailable:
					// ctx.Done
					return nil
				}
			}
			if err != nil {
				return err
			}
			metrics.Append(resp)
		}
	}
}

// пропускаем линтер, т.к нет необходимости в t.Cleanup
// nolint: tparallel
func TestGRPCMetrics(t *testing.T) {
	defer goleak.VerifyNone(t)
	tests := []struct {
		name   string
		cfg    *config.Config
		count  int32
		expect int32
	}{
		{
			name: "all metrics",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:  true,
					Load: true,
					IO:   true,
				},
			},
			count:  2, // по count метрик каждого типа
			expect: 6, // сума всех метрик
		},
		{
			name: "load metric only",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:  false,
					Load: true,
					IO:   false,
				},
			},
			count:  2,
			expect: 2, // только load метрики - 2
		},
		{
			name: "no metrics",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:  false,
					Load: false,
					IO:   false,
				},
			},
			count:  0,
			expect: 0, // не должны набрать ни одной метрики
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			err := logger.New(config.LoggerConf{
				Level: "DEBUG",
			})
			require.NoError(t, err)

			// Раз в interval секунд, получать агрегированную метрику за counter метрик
			var interval int32 = 1
			var counter int32 = 2
			wg := &sync.WaitGroup{}
			lis, err := server.MonitoringServer(ctx, cancel, wg, tc.cfg)
			require.NoError(t, err)

			stream, err := client.MonitoringClient(ctx, lis.Addr().String(), interval, counter)
			require.NoError(t, err)
			metrics := &Metrics{}

			wg.Add(1)
			go func() {
				defer wg.Done()
				err := getEvent(ctx, metrics, stream)
				require.NoError(t, err)
			}()
			// Время с запасом несколько секунд и ждем пока наберем нужное количество ответов (count)
			repeatCheck := time.Second * time.Duration(interval+3)
			if tc.count > 0 {
				// если у нас 0 метрик, то подождем хотя бы пару секунд,
				// чтобы Eventually завершил работу корректно
				repeatCheck *= time.Duration(tc.count)
			}
			require.Eventually(t, func() bool {
				// Проверка, что каждой метрики набрали по количеству count
				var results []bool
				if tc.cfg.Metrics.CPU {
					results = append(results, atomic.LoadInt32(&metrics.cpuMetrics) == tc.count)
				}
				if tc.cfg.Metrics.Load {
					results = append(results, atomic.LoadInt32(&metrics.loadMetrics) == tc.count)
				}
				if tc.cfg.Metrics.IO {
					results = append(results, atomic.LoadInt32(&metrics.ioMetrics) == tc.count)
				}
				for _, result := range results {
					if !result {
						return false
					}
				}
				return true
			}, repeatCheck, time.Millisecond*10)
			cancel()
			wg.Wait()

			require.Equal(t, metrics.Len(), tc.expect)
		})
	}
}

// todo multiple clients test
