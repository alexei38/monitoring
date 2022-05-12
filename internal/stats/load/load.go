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

func (s *Stats) Get() error {
	f, err := os.Open(loadAveragePath)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		count, err := fmt.Sscanf(
			line,
			"%f %f %f",
			&s.Load1, &s.Load5, &s.Load15,
		)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed parse %s. line: %q err: %w", loadAveragePath, line, err)
		}
		if count == 0 {
			return fmt.Errorf("failed parse %s. line: %q", loadAveragePath, line)
		}
	}
	return nil
}

func NewStat() *Stats {
	stat := &Stats{}
	return stat
}
