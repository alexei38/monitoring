package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/iostat"
	log "github.com/sirupsen/logrus"
)

func sendIOStat(ctx context.Context, ch <-chan *iostat.Stats, srv pb.StreamService_FetchResponseServer) {
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
			// TODO CLIENT
			log.Infof("send iostat metric (%v) to client %s", metrics.IOStat, "client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				return
			}
		}
	}
}
