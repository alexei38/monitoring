package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Logger  LoggerConf
	Listen  ListenConfig
	Metrics Metrics
}

type ListenConfig struct {
	Host string
	Port string
}

type LoggerConf struct {
	Level string
	File  string
}

type Metrics struct {
	CPU  bool
	Load bool
	IO   bool
}

// NewConfig read configs and return Config
// read from flag --config if exists
// else find config.yaml file in:
// - /etc/calendar
// - $HOME/.calendar
// - $PWD/configs
// - current dir.
func NewConfig() (*Config, error) {
	config := &Config{
		Logger: LoggerConf{
			Level: "DEBUG",
		},
		Metrics: Metrics{
			CPU:  true,
			Load: true,
			IO:   true,
		},
		Listen: ListenConfig{
			Host: "127.0.0.1",
			Port: "9080",
		},
	}
	cfgFile := viper.GetString("config")
	viper.SetConfigType("yaml")

	host := viper.GetString("serverHost")
	if host != "" {
		config.Listen.Host = host
	}

	port := viper.GetString("serverPort")
	if port != "" {
		config.Listen.Port = port
	}

	if cfgFile != "" {
		// Если указан конфиг, то читаем только его
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("Config %s not found", cfgFile)
		}
		viper.AddConfigPath(filepath.Dir(cfgFile))
		viper.SetConfigName(filepath.Base(cfgFile))
	} else {
		// Если не указан, то ищем конфиги в "default" каталогах
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("/etc/monitoring")
		viper.AddConfigPath("$HOME/.monitoring")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	err = viper.Unmarshal(config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
