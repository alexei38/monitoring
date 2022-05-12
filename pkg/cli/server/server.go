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
	monitorCPU "github.com/alexei38/monitoring/internal/monitor/cpu"
	monitorInode "github.com/alexei38/monitoring/internal/monitor/disk/inode"
	monitorUsage "github.com/alexei38/monitoring/internal/monitor/disk/usage"
	monitorIostat "github.com/alexei38/monitoring/internal/monitor/iostat"
	monitorLoad "github.com/alexei38/monitoring/internal/monitor/load"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	statInode "github.com/alexei38/monitoring/internal/stats/disk/inode"
	statUsage "github.com/alexei38/monitoring/internal/stats/disk/usage"
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
	logF := log.WithField("interval", client.Interval).WithField("counter", client.Counter)
	logF.Info("Client connected")
	wg := &sync.WaitGroup{}
	if s.config.Metrics.IO {
		logF := logF.WithField("metric", "iostat")
		logF.Infof("start collection metric")

		ioCh := make(chan *iostat.Stats)
		defer close(ioCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			monitorIostat.AvgStat(s.ctx, logF, ioCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendIOStat(s.ctx, logF, ioCh, srv)
		}()
	}

	if s.config.Metrics.CPU {
		logF := logF.WithField("metric", "cpu usage")
		logF.Infof("start collection metric")

		cpuCh := make(chan *cpu.Stats)
		defer close(cpuCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			monitorCPU.AvgStat(s.ctx, logF, cpuCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendCPUStat(s.ctx, logF, cpuCh, srv)
		}()
	}

	if s.config.Metrics.Load {
		logF := logF.WithField("metric", "load average")
		logF.Infof("start collection metric")

		loadCh := make(chan *load.Stats)
		defer close(loadCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			monitorLoad.AvgStat(s.ctx, logF, loadCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendLoadStat(s.ctx, logF, loadCh, srv)
		}()
	}

	if s.config.Metrics.DiskUsage {
		logF := logF.WithField("metric", "disk usage")
		logF.Infof("start collection metric")

		diskCh := make(chan *statUsage.Stats)
		defer close(diskCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			monitorUsage.AvgStat(s.ctx, logF, diskCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendDiskUsageStat(s.ctx, logF, diskCh, srv)
		}()
	}

	if s.config.Metrics.DiskInode {
		logF := logF.WithField("metric", "disk inode")
		logF.Infof("start collection metric")

		diskCh := make(chan *statInode.Stats)
		defer close(diskCh)

		wg.Add(2)
		go func() {
			defer wg.Done()
			monitorInode.AvgStat(s.ctx, logF, diskCh, int(client.Interval), int(client.Counter))
		}()
		go func() {
			defer wg.Done()
			sendDiskInodeStat(s.ctx, logF, diskCh, srv)
		}()
	}
	wg.Wait()
	return nil
}

func MonitoringServer(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup, cfg *config.Config) (net.Listener, error) { // nolint:lll
	// start server
	log.Infof("starting server")
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
		log.Infof("shutdown server")
		gSrv.GracefulStop()
	}()
	log.Infof("server started on %s", lis.Addr().String())
	return lis, nil
}

func Run() error {
	log.Info("reading config")
	cfg, err := config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed read config file: %w", err)
	}
	err = logger.New(cfg.Logger)
	if err != nil {
		return fmt.Errorf("failed initialize logger: %w", err)
	}

	wg := &sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())
	_, err = MonitoringServer(ctx, cancel, wg, cfg)
	if err != nil {
		return fmt.Errorf("failed start server: %w", err)
	}

	wg.Wait()
	return nil
}
