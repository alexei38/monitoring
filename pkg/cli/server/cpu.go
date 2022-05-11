package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/cpu"
	log "github.com/sirupsen/logrus"
)

// Пропускаем линтер, т.к. длинные имена аргументов у grpc
// nolint:lll
func sendCPUStat(ctx context.Context, ch <-chan *cpu.Stats, srv pb.StreamService_FetchResponseServer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		select {
		case <-ctx.Done():
			return
		case stats := <-ch:
			metrics := pb.Metrics{}
			for _, stat := range stats.CPU {
				metrics.CPU = append(metrics.CPU, &pb.CPUMetric{
					CPU:    stat.CPU,
					User:   stat.Usr,
					System: stat.Sys,
					Idle:   stat.Idle,
				})
			}
			// TODO CLIENT
			log.Infof("send cpuload metric (%v) to client %s", metrics.CPU, "client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				return
			}
		}
	}
}
