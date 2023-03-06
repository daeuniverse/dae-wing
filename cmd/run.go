package cmd

import (
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/daeuniverse/dae-wing/cmd/internal"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	runCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", "/etc/dae/", "config directory")
	runCmd.PersistentFlags().StringVarP(&listen, "listen", "l", "0.0.0.0:2023", "listening address")
	runCmd.PersistentFlags().BoolVar(&apiOnly, "api-only", false, "run graphql backend without dae")
	runCmd.PersistentFlags().BoolVarP(&disableTimestamp, "disable-timestamp", "", false, "disable timestamp")
}

var (
	cfgDir           string
	disableTimestamp bool
	listen           string
	apiOnly          bool

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
			if !apiOnly {
				internal.AutoSu()
			}

			// Read config from --config cfgDir.
			if err := db.InitDatabase(cfgDir); err != nil {
				logrus.Fatalln("Failed to init db:", err)
			}

			// Run dae.
			go func() {
				logrus.Fatalln(dae.Run(
					logrus.StandardLogger(),
					dae.EmptyConfig, // TODO: boot with running.
					disableTimestamp,
					apiOnly,
				))
			}()

			// ListenAndServe GraphQL.
			schema, err := graphql.Schema()
			if err != nil {
				logrus.Errorf("Exiting: %v", err)
				dae.ChReloadConfigs <- nil
				os.Exit(1)
			}
			http.Handle("/graphql", cors.AllowAll().Handler(&relay.Handler{Schema: schema}))

			go func() {
				if err = http.ListenAndServe(listen, nil); err != nil {
					// Notify to Close().
					logrus.Errorf("Exiting: %v", err)
					dae.ChReloadConfigs <- nil
					os.Exit(1)
				}
			}()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGILL)
			for sig := range sigs {
				// Notify to Close().
				logrus.Errorf("Exiting: %v", sig.String())
				dae.ChReloadConfigs <- nil
				return
			}
		},
	}
)
