package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"

	pb "github.com/alexei38/monitoring/internal/grpc"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run() error {
	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %w", err)
	}
	defer ui.Close()
	uiEvents := ui.PollEvents()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for {
			select {
			case e := <-uiEvents:
				switch e.ID {
				case "q", "<C-c>":
					cancel()
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	host := viper.GetString("clientHost")
	port := viper.GetString("clientPort")
	interval := viper.GetInt32("interval")
	counter := viper.GetInt32("counter")
	hostPort := net.JoinHostPort(host, port)
	stream, err := MonitoringClient(ctx, hostPort, interval, counter)
	if err != nil {
		return err
	}
	go func() {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				log.Info("connection closed")
				cancel()
				return
			}
			if err != nil {
				log.Errorf("cannot receive %v", err)
				cancel()
				return
			}
			drawLoadTable(resp.Load)
			drawCPUTable(resp.CPU)
			drawIOTable(resp.IOStat)
		}
	}()
	<-ctx.Done()
	return nil
}

func drawLoadTable(metrics *pb.LoadMetric) {
	if metrics == nil {
		return
	}
	table := widgets.NewTable()
	table.Rows = [][]string{
		{"LA 1min", "LA 5min", "LA 15min"},
		{
			fmt.Sprintf("%.2f %%", metrics.Load1),
			fmt.Sprintf("%.2f %%", metrics.Load5),
			fmt.Sprintf("%.2f %%", metrics.Load15),
		},
	}
	table.Title = "Load Average"
	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, 45, 5)
	ui.Render(table)
}

func drawCPUTable(metrics []*pb.CPUMetric) {
	if metrics == nil {
		return
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].CPU < metrics[j].CPU
	})
	cpuData := make(map[string]map[string]float32)
	header := []string{""}
	for _, metric := range metrics {
		var cpuName string
		if metric.CPU == "all" {
			cpuName = "AVG"
		} else {
			cpuName = fmt.Sprintf("CPU %s", metric.CPU)
		}
		header = append(header, cpuName)
		cpuData[cpuName] = map[string]float32{
			"user":   metric.User,
			"system": metric.System,
			"idle":   metric.Idle,
		}
	}
	table := widgets.NewTable()
	table.Title = "CPU Load"
	table.Rows = [][]string{header}

	for _, loadType := range []string{"user", "system", "idle"} {
		var row []string
		for _, cpuName := range header {
			if cpuName == "" {
				row = append(row, loadType)
			} else {
				row = append(row, fmt.Sprintf("%.2f %%", cpuData[cpuName][loadType]))
			}
		}
		table.Rows = append(table.Rows, row)
	}

	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, len(table.Rows)*4, len(header)*10, 5)
	ui.Render(table)
}

func drawIOTable(metrics []*pb.IOMetric) {
	if metrics == nil {
		return
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Device < metrics[j].Device
	})
	header := []string{"device", "rkb/s", "wkb/s", "util%"}

	table := widgets.NewTable()
	table.Title = "IO Load"
	table.Rows = [][]string{header}

	for _, metric := range metrics {
		row := []string{
			metric.Device,
			fmt.Sprintf("%.2f", metric.Rkbs),
			fmt.Sprintf("%.2f", metric.Wkbs),
			fmt.Sprintf("%.2f %%", metric.Util),
		}
		table.Rows = append(table.Rows, row)
	}

	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, len(table.Rows)*4, len(header)*15, 16)
	ui.Render(table)
}

func MonitoringClient(ctx context.Context, hostPort string, interval, counter int32) (pb.StreamService_FetchResponseClient, error) { // nolint:lll
	credentials := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, hostPort, credentials)
	if err != nil {
		return nil, fmt.Errorf("failed connect to server %s: %w", hostPort, err)
	}
	client := pb.NewStreamServiceClient(conn)
	go func() {
		defer conn.Close()
		<-ctx.Done()
	}()

	in := &pb.ClientRequest{
		Interval: interval,
		Counter:  counter,
	}
	return client.FetchResponse(ctx, in)
}
