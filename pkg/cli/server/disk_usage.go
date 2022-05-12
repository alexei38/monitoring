package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/disk/usage"
	log "github.com/sirupsen/logrus"
)

func sendDiskUsageStat(ctx context.Context, log *log.Entry, ch <-chan *usage.Stats, srv pb.StreamService_FetchResponseServer) { // nolint: lll
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
			for _, stat := range stats.Stat {
				metrics.DiskUsage = append(metrics.DiskUsage, &pb.DiskUsageMetric{
					Device:    stat.Device,
					Mount:     stat.Mount,
					Typefs:    stat.TypeFS,
					Used:      stat.Used,
					Available: stat.Available,
				})
			}
			log.Info("send metric to client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("connection closed")
				return
			}
		}
	}
}
