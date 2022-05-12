package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sort"
	"strconv"
	"strings"

	pb "github.com/alexei38/monitoring/internal/grpc"
	"github.com/dustin/go-humanize"
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
	recieveData(cancel, stream)
	<-ctx.Done()
	return nil
}

func recieveData(cancel context.CancelFunc, stream pb.StreamService_FetchResponseClient) {
	drawLoadTable(nil)
	drawCPUTable(nil)
	drawIOTable(nil)
	drawDiskUsageTable(nil)
	drawDiskInodeTable(nil)
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
			if resp.Load != nil {
				drawLoadTable(resp.Load)
			}
			if resp.CPU != nil {
				drawCPUTable(resp.CPU)
			}
			if resp.IOStat != nil {
				drawIOTable(resp.IOStat)
			}
			if resp.DiskUsage != nil {
				drawDiskUsageTable(resp.DiskUsage)
			}
			if resp.DiskInode != nil {
				drawDiskInodeTable(resp.DiskInode)
			}
		}
	}()
}

func drawLoadTable(metrics *pb.LoadMetric) {
	table := widgets.NewTable()
	table.Rows = [][]string{
		{"LA 1min", "LA 5min", "LA 15min"},
	}
	if metrics != nil {
		table.Rows = append(table.Rows, []string{
			fmt.Sprintf("%.2f %%", metrics.Load1),
			fmt.Sprintf("%.2f %%", metrics.Load5),
			fmt.Sprintf("%.2f %%", metrics.Load15),
		})
	}
	table.Title = "Load Average"
	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 0, 45, 5)
	ui.Render(table)
}

func drawCPUTable(metrics []*pb.CPUMetric) {
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].CPU < metrics[j].CPU
	})
	table := widgets.NewTable()
	table.Title = "CPU Load"
	header := []string{"CPU", "User", "System", "Idle"}
	table.Rows = [][]string{header}
	for _, metric := range metrics {
		table.Rows = append(table.Rows, []string{
			metric.CPU,
			fmt.Sprintf("%.2f %%", metric.User),
			fmt.Sprintf("%.2f %%", metric.System),
			fmt.Sprintf("%.2f %%", metric.Idle),
		})
	}
	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(0, 26, 45, 5)
	ui.Render(table)
}

func drawIOTable(metrics []*pb.IOMetric) {
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Device < metrics[j].Device
	})
	header := []string{"device", "rkb/s", "wkb/s", "util%"}

	table := widgets.NewTable()
	table.Title = "IO Load"
	table.Rows = [][]string{header}

	for _, metric := range metrics {
		if strings.Contains(metric.Device, "loop") {
			continue
		}
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
	table.SetRect(0, 26, 45, 40)
	ui.Render(table)
}

func drawDiskUsageTable(metrics []*pb.DiskUsageMetric) {
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Device < metrics[j].Device
	})
	header := []string{"device", "all", "used", "available", "mountpoint"}

	table := widgets.NewTable()
	table.Title = "Disk Usage"
	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(46, 0, 150, 15)
	table.Rows = [][]string{header}

	for _, metric := range metrics {
		if metric.Typefs == "devtmpfs" || metric.Typefs == "tmpfs" || metric.Typefs == "squashfs" {
			continue
		}
		table.Rows = append(table.Rows, []string{
			metric.Device,
			humanize.Bytes(uint64(metric.Available + metric.Used)),
			humanize.Bytes(uint64(metric.Available)),
			humanize.Bytes(uint64(metric.Used)),
			metric.Mount,
		})
	}
	ui.Render(table)
}

func drawDiskInodeTable(metrics []*pb.DiskInodeMetric) {
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].Device < metrics[j].Device
	})
	header := []string{"device", "count", "used", "available", "mountpoint"}

	table := widgets.NewTable()
	table.Title = "Disk Inode"
	table.Rows = [][]string{header}

	for _, metric := range metrics {
		if metric.Typefs == "devtmpfs" || metric.Typefs == "tmpfs" || metric.Typefs == "squashfs" {
			continue
		}
		table.Rows = append(table.Rows, []string{
			metric.Device,
			strconv.Itoa(int(metric.Available + metric.Used)),
			strconv.Itoa(int(metric.Available)),
			strconv.Itoa(int(metric.Used)),
			metric.Mount,
		})
	}

	table.TextAlignment = ui.AlignCenter
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.SetRect(46, 15, 150, 35)
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
