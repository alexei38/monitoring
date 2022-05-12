package logger

import (
	"errors"
	"os"

	"github.com/alexei38/monitoring/internal/config"
	log "github.com/sirupsen/logrus"
)

var ErrNotFountLogLevel = errors.New("log level not found")

func New(config config.LoggerConf) error {
	err := setLogLevel(config.Level)
	if err != nil {
		return err
	}
	err = setLogOutput(config.File)
	if err != nil {
		return err
	}
	formatter := &log.TextFormatter{
		// enable time logging
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	log.SetFormatter(formatter)
	return nil
}

func setLogOutput(logFile string) error {
	if logFile == "" {
		log.SetOutput(os.Stdout)
		return nil
	}
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}
	log.SetOutput(f)
	return nil
}

func setLogLevel(logLevel string) error {
	if logLevel == "" {
		return ErrNotFountLogLevel
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		return err
	}
	log.SetLevel(level)
	return nil
}
