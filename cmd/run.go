package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/v2rayA/dae-wing/cmd/internal"
	"github.com/v2rayA/dae-wing/db"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "/etc/dae/wing/", "config dir")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "disable timestamp")
}

var (
	cfgDir           string
	disableTimestamp bool

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run dae in the foreground",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgDir == "" {
				logrus.Fatalln("Argument \"--config\" or \"-c\" is required but not provided.")
			}

			// Require "sudo" if necessary.
			internal.AutoSu()

			// Read config from --config cfgDir.
			if err := db.InitDatabase(cfgDir); err != nil {
				logrus.Fatalln("Failed to init db:", err)
			}
		},
	}
)
