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

func GetStat() (*Stats, error) {
	mstat := &mpStat{}
	result := &Stats{}
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("mpstat", "-P", "ALL", "-o", "JSON")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed run mpstat: %w", err)
	}
	if err := json.Unmarshal(stdout.Bytes(), mstat); err != nil {
		return nil, fmt.Errorf("failed parse mpstat output: %w", err)
	}
	if len(mstat.Sysstat.Hosts) < 1 {
		return nil, fmt.Errorf("mpstat data not found. check mpstat -P ALL")
	}
	if len(mstat.Sysstat.Hosts) > 1 {
		return nil, fmt.Errorf("unsuported multiple hosts")
	}
	for _, s := range mstat.Sysstat.Hosts[0].Statistics {
		result.CPU = append(result.CPU, s.CPULoad...)
	}
	return result, nil
}
