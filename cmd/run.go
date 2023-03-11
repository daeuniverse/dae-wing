package cmd

import (
	"context"
	"fmt"
	"github.com/daeuniverse/dae-wing/cmd/internal"
	"github.com/daeuniverse/dae-wing/dae"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql"
	"github.com/daeuniverse/dae-wing/graphql/service/config"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
					dae.EmptyConfig,
					disableTimestamp,
					apiOnly,
				))
			}()
			// Reload with running state.
			if err := restoreRunningState(); err != nil {
				logrus.Warnln("Failed to restore last running state:", err)
			}

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

func restoreRunningState() (err error) {
	reload, err := shouldReload()
	if err != nil {
		return err
	}
	if !reload {
		return nil
	}
	tx := db.BeginTx(context.TODO())
	// Reload.
	if _, err = config.Run(tx, false); err != nil {
		tx.Rollback()

		// Another tx.
		// Set running = false.
		tx2 := db.BeginTx(context.TODO())
		var sys db.System
		if err2 := tx2.Model(&sys).Select("id").First(&sys).Error; err2 != nil {
			tx2.Rollback()
			return fmt.Errorf("%w; %v", err, err2)
		}
		if err2 := tx2.Model(&sys).Updates(map[string]interface{}{
			"running":  false,
			"modified": false,
		}).Error; err2 != nil {
			tx2.Rollback()
			return fmt.Errorf("%w; %v", err, err2)
		}
		tx2.Commit()
		return err
	}
	tx.Commit()
	return nil
}

func shouldReload() (ok bool, err error) {
	var sys db.System
	if err := db.DB(context.TODO()).Model(&db.System{}).FirstOrCreate(&sys).Error; err != nil {
		return false, err
	}
	if !sys.Running {
		return false, nil
	}
	var m db.Config
	q := db.DB(context.TODO()).Model(&db.Config{}).
		Where("selected = ?", true).
		First(&m)
	if q.Error != nil {
		return false, q.Error
	}
	if q.RowsAffected == 0 {
		// Data inconsistency.
		logrus.Warnln("Data inconsistency detected: no selected config but last state is running")
		_ = db.DB(context.TODO()).Model(&sys).Update("running", false).Error
		return false, nil
	}
	return true, nil
}
