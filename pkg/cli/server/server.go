package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/alexei38/monitoring/internal/config"
	pb "github.com/alexei38/monitoring/internal/grpc"
	mcpu "github.com/alexei38/monitoring/internal/monitor/cpu"
	mload "github.com/alexei38/monitoring/internal/monitor/load"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	"github.com/alexei38/monitoring/internal/stats/load"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type server struct {
	config *config.Config
	pb.UnimplementedStreamServiceServer
}

// Пропускаем линтер, т.к. длинные имена аргументов у grpc
// nolint:lll
func sendLoadStat(ctx context.Context, cancel context.CancelFunc, loadCh <-chan *load.Stats, srv pb.StreamService_FetchResponseServer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case loadStat := <-loadCh:
			metrics := pb.Metrics{
				Load: &pb.LoadMetric{
					Load1:  loadStat.Load1,
					Load5:  loadStat.Load5,
					Load15: loadStat.Load15,
				},
			}
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				cancel()
				return
			}
		}
	}
}

// Пропускаем линтер, т.к. длинные имена аргументов у grpc
// nolint:lll
func sendCPUStat(ctx context.Context, cancel context.CancelFunc, cpuCh <-chan *cpu.Stats, srv pb.StreamService_FetchResponseServer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case cpuStats := <-cpuCh:
			metrics := pb.Metrics{}
			for _, cpuStat := range cpuStats.CPU {
				metrics.CPU = append(metrics.CPU, &pb.CPUMetric{
					CPU:    cpuStat.CPU,
					User:   cpuStat.Usr,
					System: cpuStat.Sys,
					Idle:   cpuStat.Idle,
				})
			}
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				cancel()
				return
			}
		}
	}
}

func (s server) FetchResponse(client *pb.ClientRequest, srv pb.StreamService_FetchResponseServer) error {
	log.WithField("interval", client.Interval).WithField("counter", client.Counter)
	log.Info("Client connected")
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if s.config.Metrics.Load {
		loadCh := make(chan *load.Stats)
		defer close(loadCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			mload.AvgStat(ctx, loadCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendLoadStat(ctx, cancel, loadCh, srv)
		}()
	}

	if s.config.Metrics.CPU {
		cpuCh := make(chan *cpu.Stats)
		defer close(cpuCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			mcpu.AvgStat(ctx, cpuCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendCPUStat(ctx, cancel, cpuCh, srv)
		}()
	}
	wg.Wait()
	return nil
}

func Run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed read config file: %w", err)
	}
	log.Info("Monitoring server started")

	hostPort := net.JoinHostPort(cfg.Listen.Host, cfg.Listen.Port)
	lis, err := net.Listen("tcp", hostPort)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	pb.RegisterStreamServiceServer(s, server{config: cfg})

	log.Println("Start grpc server")
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
