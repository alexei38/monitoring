package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/alexei38/monitoring/internal/config"
	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/logger"
	mcpu "github.com/alexei38/monitoring/internal/monitor/cpu"
	miostat "github.com/alexei38/monitoring/internal/monitor/iostat"
	mload "github.com/alexei38/monitoring/internal/monitor/load"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	"github.com/alexei38/monitoring/internal/stats/iostat"
	"github.com/alexei38/monitoring/internal/stats/load"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type server struct {
	ctx    context.Context
	cancel context.CancelFunc
	config *config.Config
	pb.UnimplementedStreamServiceServer
}

func (s *server) FetchResponse(client *pb.ClientRequest, srv pb.StreamService_FetchResponseServer) error {
	log.WithField("interval", client.Interval).WithField("counter", client.Counter)
	log.Info("Client connected")
	wg := &sync.WaitGroup{}

	if s.config.Metrics.IO {
		ioCh := make(chan *iostat.Stats)
		defer close(ioCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			miostat.AvgStat(s.ctx, ioCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendIOStat(s.ctx, ioCh, srv)
		}()
	}

	if s.config.Metrics.CPU {
		cpuCh := make(chan *cpu.Stats)
		defer close(cpuCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			mcpu.AvgStat(s.ctx, cpuCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendCPUStat(s.ctx, cpuCh, srv)
		}()
	}

	if s.config.Metrics.Load {
		loadCh := make(chan *load.Stats)
		defer close(loadCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			mload.AvgStat(s.ctx, loadCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendLoadStat(s.ctx, loadCh, srv)
		}()
	}
	wg.Wait()
	return nil
}

func MonitoringServer(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, cfg *config.Config) (net.Listener, error) { // nolint:lll
	// start server
	hostPort := net.JoinHostPort(cfg.Listen.Host, cfg.Listen.Port)
	lis, err := net.Listen("tcp", hostPort)
	if err != nil {
		return nil, err
	}
	gSrv := grpc.NewServer()
	srv := &server{config: cfg, ctx: ctx, cancel: cancel}
	pb.RegisterStreamServiceServer(gSrv, srv)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()
		err := gSrv.Serve(lis)
		if err != nil {
			log.Errorf("Server stoped with error %v", err)
		}
	}()

	// graceful shutdown
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case s := <-exit:
			log.Debugf("got signal %v, attempting graceful shutdown", s)
			break
		case <-ctx.Done():
			log.Debug("parent context done")
			break
		}
		cancel()
		log.Infof("Gracefull stopping server")
		gSrv.GracefulStop()
	}()
	return lis, nil
}

func Run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed read config file: %w", err)
	}
	err = logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed initialize logger: %v", err)
	}

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	lis, err := MonitoringServer(ctx, cancel, wg, cfg)
	if err != nil {
		return fmt.Errorf("failed start server: %w", err)
	}

	log.Infof("server started on %s", lis.Addr().String())
	wg.Wait()
	return nil
}
