package cmd

import (
	"github.com/alexei38/monitoring/pkg/cli/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command.
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "monitoring server",
	Version: GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		err := server.Run()
		if err != nil {
			log.Error(err)
		}
	},
}

func init() {
	serverCmd.Flags().String(
		"host",
		"127.0.0.1",
		"connection host",
	)
	viper.BindPFlag("serverHost", serverCmd.Flags().Lookup("host"))
	serverCmd.Flags().String(
		"port",
		"9080",
		"connection port",
	)
	viper.BindPFlag("serverPort", serverCmd.Flags().Lookup("port"))

	serverCmd.Flags().String(
		"config",
		"",
		"config file (default is $HOME/.monitoring.yaml, /etc/monitoring/config.yaml)",
	)
	viper.BindPFlag("config", serverCmd.Flags().Lookup("config"))
	rootCmd.AddCommand(serverCmd)
}
