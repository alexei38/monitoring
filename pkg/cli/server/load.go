package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/load"
	log "github.com/sirupsen/logrus"
)

// Пропускаем линтер, т.к. длинные имена аргументов у grpc
// nolint:lll
func sendLoadStat(ctx context.Context, ch <-chan *load.Stats, srv pb.StreamService_FetchResponseServer) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-srv.Context().Done():
			return
		case loadStat := <-ch:
			metrics := pb.Metrics{
				Load: &pb.LoadMetric{
					Load1:  loadStat.Load1,
					Load5:  loadStat.Load5,
					Load15: loadStat.Load15,
				},
			}
			// TODO CLIENT
			log.Infof("send load average metric (%v) to client %s", metrics.Load, "client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				return
			}
		}
	}
}
