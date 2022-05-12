package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/iostat"
	log "github.com/sirupsen/logrus"
)

func sendIOStat(ctx context.Context, log *log.Entry, ch <-chan *iostat.Stats, srv pb.StreamService_FetchResponseServer) { // nolint: lll
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
			for _, stat := range stats.Disk {
				metrics.IOStat = append(metrics.IOStat, &pb.IOMetric{
					Device: stat.Device,
					Rkbs:   stat.Rkbs,
					Wkbs:   stat.Wkbs,
					Util:   stat.Util,
				})
			}
			log.Info("send metric to client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				return
			}
		}
	}
}
