package cmd

import (
	"strings"
)

var (
	release   = "UNKNOWN"
	buildDate = "UNKNOWN"
	gitHash   = "UNKNOWN"
)

func GetVersion() string {
	v := []string{release, buildDate, gitHash}
	return strings.Join(v, "_")
}
