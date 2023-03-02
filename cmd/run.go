package cmd

import (
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/v2rayA/dae-wing/db"
	"github.com/v2rayA/dae-wing/graphql"
	"net/http"
	"os"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "/etc/dae/wing.db", "database file")
	runCmd.PersistentFlags().StringVarP(&listen, "listen", "l", "0.0.0.0:2023", "listening address")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "disable timestamp")
}

var (
	cfgDir           string
	disableTimestamp bool
	listen           string

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run dae in the foreground",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgDir == "" {
				logrus.Fatalln("Argument \"--config\" or \"-c\" is required but not provided.")
			}
			if err := os.MkdirAll(cfgDir, 0750); err != nil && !os.IsExist(err) {
				logrus.Fatalln(err)
			}

			// Require "sudo" if necessary.
			//internal.AutoSu()

			// Read config from --config cfgDir.
			if err := db.InitDatabase(cfgDir); err != nil {
				logrus.Fatalln("Failed to init db:", err)
			}

			// ListenAndServe.
			schema, err := graphql.Schema()
			if err != nil {
				return
			}
			http.Handle("/graphql", cors.AllowAll().Handler(&relay.Handler{Schema: schema}))
			logrus.Fatal(http.ListenAndServe(listen, nil))
		},
	}
)
