package load

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
)

const loadAveragePath = "/proc/loadavg"

type Stats struct {
	Load1  float32
	Load5  float32
	Load15 float32
}

func GetStat() (*Stats, error) {
	stat := &Stats{}
	f, err := os.Open(loadAveragePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		count, err := fmt.Sscanf(
			line,
			"%f %f %f",
			&stat.Load1, &stat.Load5, &stat.Load15,
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("failed parse %s. line: %q err: %w", loadAveragePath, line, err)
		}
		if count == 0 {
			return nil, fmt.Errorf("failed parse %s. line: %q", loadAveragePath, line)
		}
	}
	return stat, nil
}
