package iostat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type ioStat struct {
	Sysstat struct {
		Hosts []struct {
			Statistics []struct {
				Disk []Stat `json:"disk"`
			} `json:"statistics"`
		} `json:"hosts"`
	} `json:"sysstat"`
}

type Stat struct {
	// пропускаем линтер, т.к нет возможности влиять на имена полей в json
	// nolint:tagliatelle
	Device string  `json:"disk_device"`
	Rkbs   float32 `json:"rkB/s"`
	Wkbs   float32 `json:"wkB/s"`
	Util   float32 `json:"%util"`
}

type Stats struct {
	Disk []Stat
}

func (s *Stats) Get() error {
	s.clear()
	iostat := &ioStat{}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("iostat", "-dx", "1", "2", "-o", "JSON")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed run mpstat: %w", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), iostat); err != nil {
		return fmt.Errorf("failed parse mpstat output: %w", err)
	}
	if len(iostat.Sysstat.Hosts) < 1 {
		return fmt.Errorf("iostat. data not found. check iostat -dx")
	}
	if len(iostat.Sysstat.Hosts) > 1 {
		return fmt.Errorf("unsuported multiple hosts")
	}
	statistics := iostat.Sysstat.Hosts[0].Statistics
	if len(statistics) <= 1 {
		return fmt.Errorf("iostat. not found data statistics. check iostat -dx")
	}
	res := statistics[len(statistics)-1]
	s.Disk = res.Disk
	return nil
}

func (s *Stats) clear() {
	s.Disk = []Stat{}
}

func NewStat() *Stats {
	stat := &Stats{}
	return stat
}
