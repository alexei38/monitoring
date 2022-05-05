package cpu

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type mpStat struct {
	Sysstat struct {
		Hosts []struct {
			Statistics []struct {
				// пропускаем линтер, т.к нет возможности влиять на имена полей в json
				// nolint:tagliatelle
				CPULoad []Stat `json:"cpu-load"`
			} `json:"statistics"`
		} `json:"hosts"`
	} `json:"sysstat"`
}

type Stat struct {
	CPU  string  `json:"cpu"`
	Usr  float32 `json:"usr"`
	Sys  float32 `json:"sys"`
	Idle float32 `json:"idle"`
}

type Stats struct {
	CPU []Stat
}

func (s *Stats) Get() error {
	s.clear()
	mstat := &mpStat{}
	var stdout bytes.Buffer
	cmd := exec.Command("mpstat", "-P", "ALL", "1", "1", "-o", "JSON")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed run mpstat: %w", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), mstat); err != nil {
		return fmt.Errorf("failed parse mpstat output: %w", err)
	}
	if len(mstat.Sysstat.Hosts) < 1 {
		return fmt.Errorf("mpstat data not found. check mpstat -P ALL")
	}
	if len(mstat.Sysstat.Hosts) > 1 {
		return fmt.Errorf("unsuported multiple hosts")
	}
	for _, res := range mstat.Sysstat.Hosts[0].Statistics {
		s.CPU = append(s.CPU, res.CPULoad...)
	}
	return nil
}

func (s *Stats) clear() {
	s.CPU = []Stat{}
}

func NewStat() *Stats {
	stat := &Stats{}
	return stat
}
