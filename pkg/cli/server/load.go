package server

import (
	"context"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/alexei38/monitoring/internal/stats/load"
	log "github.com/sirupsen/logrus"
)

func sendLoadStat(ctx context.Context, log *log.Entry, ch <-chan *load.Stats, srv pb.StreamService_FetchResponseServer) { // nolint: lll
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
			log.Info("send metric to client")
			if err := srv.Send(&metrics); err != nil {
				log.Info("Connection closed")
				return
			}
		}
	}
}
