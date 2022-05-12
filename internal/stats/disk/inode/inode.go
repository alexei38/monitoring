package inode

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Stat struct {
	Device    string
	TypeFS    string
	Mount     string
	Used      int64
	Available int64
}

type Stats struct {
	Stat []Stat
}

func (s *Stats) Get() error {
	s.clear()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("df", "-PTi")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed run df commands: %w", err)
	}
	scanner := bufio.NewScanner(&stdout)
	scanner.Scan() // Skip first header line
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 7 {
			return fmt.Errorf("unknown format. need 7 fields for df -PT")
		}
		available, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return fmt.Errorf("failed parse available size: %w", err)
		}
		used, err := strconv.ParseInt(fields[3], 10, 64)
		if err != nil {
			return fmt.Errorf("failed parse used size: %w", err)
		}
		// df выдает информацию в килобайтах, переводим в байты, а клиент решит как отображать
		s.Stat = append(s.Stat, Stat{
			Device:    fields[0],
			TypeFS:    fields[1],
			Mount:     fields[6],
			Used:      used * 1024,
			Available: available * 1024,
		})
	}
	return nil
}

func (s *Stats) clear() {
	s.Stat = []Stat{}
}

func NewStat() *Stats {
	stat := &Stats{}
	return stat
}
