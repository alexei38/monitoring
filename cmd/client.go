package cmd

import (
	"github.com/alexei38/monitoring/pkg/cli/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// clientCmd represents the client command.
var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "monitoring client",
	Version: GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		err := client.Run()
		if err != nil {
			log.Fatalln(err)
		}
	},
}

func init() {
	clientCmd.Flags().String(
		"host",
		"127.0.0.1",
		"connection host",
	)
	viper.BindPFlag("clientHost", clientCmd.Flags().Lookup("host"))
	clientCmd.Flags().Int(
		"port",
		9080,
		"connection port",
	)
	viper.BindPFlag("clientPort", clientCmd.Flags().Lookup("port"))

	clientCmd.Flags().Int32(
		"interval",
		5,
		"interval scrape metrics",
	)
	viper.BindPFlag("interval", clientCmd.Flags().Lookup("interval"))
	clientCmd.Flags().Int32(
		"counter",
		15,
		"time avg counter",
	)
	viper.BindPFlag("counter", clientCmd.Flags().Lookup("counter"))
	rootCmd.AddCommand(clientCmd)
}
