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
	"github.com/alexei38/monitoring/pkg/cli/client"
	"github.com/alexei38/monitoring/pkg/cli/server"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Metrics struct {
	cpuMetrics       int32
	loadMetrics      int32
	ioMetrics        int32
	diskUsageMetrics int32
	diskInodeMetrics int32
}

func (m *Metrics) Append(item *pb.Metrics) {
	switch {
	case item.CPU != nil:
		atomic.AddInt32(&m.cpuMetrics, 1)
	case item.Load != nil:
		atomic.AddInt32(&m.loadMetrics, 1)
	case item.IOStat != nil:
		atomic.AddInt32(&m.ioMetrics, 1)
	case item.DiskUsage != nil:
		atomic.AddInt32(&m.diskUsageMetrics, 1)
	case item.DiskInode != nil:
		atomic.AddInt32(&m.diskInodeMetrics, 1)
	}
}

func (m *Metrics) Len() int32 {
	return m.cpuMetrics + m.loadMetrics + m.ioMetrics + m.diskUsageMetrics + m.diskInodeMetrics
}

// getEvent слушает поток данных из grpc, и сохраняет результат в Metrics.
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

// waitMetrics ждет, когда наберется нужное количество метрик за определенное количество времени.
func waitMetrics(t *testing.T, cfg *config.Config, interval int32, count int32, metrics *Metrics) {
	t.Helper()
	// Время с запасом несколько секунд и ждем пока наберем нужное количество ответов (count)
	repeatCheck := time.Second * time.Duration(interval+3)
	if count > 0 {
		// если у нас 0 метрик, то подождем хотя бы пару секунд,
		// чтобы Eventually завершил работу корректно
		repeatCheck *= time.Duration(count)
	}
	require.Eventually(t, func() bool {
		// Проверка, что каждой метрики набрали по количеству count
		var results []bool
		if cfg.Metrics.CPU {
			results = append(results, atomic.LoadInt32(&metrics.cpuMetrics) == count)
		}
		if cfg.Metrics.Load {
			results = append(results, atomic.LoadInt32(&metrics.loadMetrics) == count)
		}
		if cfg.Metrics.IO {
			results = append(results, atomic.LoadInt32(&metrics.ioMetrics) == count)
		}
		if cfg.Metrics.DiskUsage {
			results = append(results, atomic.LoadInt32(&metrics.diskUsageMetrics) == count)
		}
		if cfg.Metrics.DiskInode {
			results = append(results, atomic.LoadInt32(&metrics.diskInodeMetrics) == count)
		}
		for _, result := range results {
			if !result {
				return false
			}
		}
		return true
	}, repeatCheck, time.Millisecond*10)
}

// пропускаем линтер, т.к нет необходимости в t.Cleanup
// nolint: tparallel
func TestGRPCMetrics(t *testing.T) {
	defer goleak.VerifyNone(t)
	// Раз в interval секунд, получать агрегированную метрику за counter метрик
	var interval int32 = 1
	var counter int32 = 2

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
					CPU:       true,
					Load:      true,
					IO:        true,
					DiskUsage: true,
					DiskInode: true,
				},
			},
			count:  2,  // по count метрик каждого типа
			expect: 10, // сумма всех метрик
		},
		{
			name: "load metric only",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:       false,
					Load:      true,
					IO:        false,
					DiskUsage: false,
					DiskInode: false,
				},
			},
			count:  2,
			expect: 2, // только load метрики - 2
		},
		{
			name: "cpu and disk metric only",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:       true,
					Load:      false,
					IO:        false,
					DiskUsage: true,
					DiskInode: true,
				},
			},
			count:  2,
			expect: 6, // только load метрики - 2
		},
		{
			name: "no metrics",
			cfg: &config.Config{
				Listen: config.ListenConfig{
					Host: "127.0.0.1",
					Port: "0",
				},
				Metrics: config.Metrics{
					CPU:       false,
					Load:      false,
					IO:        false,
					DiskUsage: false,
					DiskInode: false,
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

			waitMetrics(t, tc.cfg, interval, tc.count, metrics)

			cancel()
			wg.Wait()
			require.Equal(t, metrics.Len(), tc.expect)
		})
	}
}

func TestGRPCMultipleClients(t *testing.T) {
	defer goleak.VerifyNone(t)

	cfg := &config.Config{
		Listen: config.ListenConfig{
			Host: "127.0.0.1",
			Port: "0",
		},
		Metrics: config.Metrics{
			CPU:       true,
			Load:      true,
			IO:        true,
			DiskUsage: true,
			DiskInode: true,
		},
	}
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// старт сервера
	lis, err := server.MonitoringServer(ctx, cancel, wg, cfg)
	require.NoError(t, err)

	// Первый клиент
	metrics1 := &Metrics{}
	var client1Interval int32 = 2     // Интервал получения метрики
	var client1Counter int32 = 1      // Сколько метрик агрегировать
	var client1WaitMetrics int32 = 10 // Сколько групп метрик ждем в тесте, перед остановкой клиента
	var client1Expected int32 = 50    // Сколько в сумме всех метрик должны получить в тесте

	client1Elapsed := make(chan time.Duration, 1)
	defer close(client1Elapsed)

	client1, err := client.MonitoringClient(ctx, lis.Addr().String(), client1Interval, client1Counter)
	require.NoError(t, err)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := getEvent(ctx, metrics1, client1)
		require.NoError(t, err)
	}()

	start := time.Now()
	wgClient := sync.WaitGroup{}
	wgClient.Add(1)
	go func() {
		defer wgClient.Done()
		// набираем минимум по 10 метрики каждого типа
		waitMetrics(t, cfg, client1Interval, client1WaitMetrics, metrics1)
		client1Elapsed <- time.Since(start)
	}()

	// Второй клиент
	metrics2 := &Metrics{}
	var client2Interval int32 = 4    // Интервал получения метрики
	var client2Counter int32 = 1     // Сколько метрик агрегировать
	var client2WaitMetrics int32 = 5 // Сколько групп метрик ждем в тесте, перед остановкой клиента
	var client2Expected int32 = 25   // Сколько в сумме всех метрик должны получить в тесте
	client2Elapsed := make(chan time.Duration, 1)
	defer close(client2Elapsed)

	client2, err := client.MonitoringClient(ctx, lis.Addr().String(), client2Interval, client2Counter)
	require.NoError(t, err)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := getEvent(ctx, metrics2, client2)
		require.NoError(t, err)
	}()

	wgClient.Add(1)
	go func() {
		defer wgClient.Done()
		// набираем минимум по 5 метрик каждого типа
		waitMetrics(t, cfg, client2Interval, client2WaitMetrics, metrics2)
		client2Elapsed <- time.Since(start)
	}()
	wgClient.Wait()
	cancel()
	wg.Wait()

	// тест должен закончится в одну секунду
	elapsed1 := <-client1Elapsed
	elapsed2 := <-client2Elapsed
	require.Equal(t, int64(elapsed1.Seconds()), int64(elapsed2.Seconds()))

	// проверяем ожидаемое количество метрик
	require.Equal(t, client1Expected, metrics1.Len())
	require.Equal(t, client2Expected, metrics2.Len())
}
