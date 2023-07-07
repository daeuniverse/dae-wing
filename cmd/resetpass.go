package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/daeuniverse/dae-wing/cmd/internal"
	"github.com/daeuniverse/dae-wing/db"
	"github.com/daeuniverse/dae-wing/graphql"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	resetpassCmd = &cobra.Command{
		Use:   "resetpass",
		Short: "Set random passwords for every accounts",
		Run: func(cmd *cobra.Command, args []string) {
			if cfgDir == "" {
				logrus.Fatalln("Argument \"--config\" or \"-c\" is required but not provided.")
			}
			if _, err := os.Stat(cfgDir); err != nil {
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

			// Remove all users.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			var users []db.User
			if err := db.DB(ctx).Find(&users).Error; err != nil {
				logrus.Fatalln(err)
			}
			for _, u := range users {
				password := gonanoid.Must(8)
				if _, err := graphql.UpdatePassword(ctx, &struct {
					CurrentPassword string
					NewPassword     string
				}{
					NewPassword: password,
				}, &u, true); err != nil {
					logrus.Fatalf("Username: %v: %v", u.Username, err)
				}
				fmt.Printf("Username: %v, Password: %v\n", u.Username, password)
			}

		},
	}
)

func init() {
	resetpassCmd.PersistentFlags().StringVarP(&cfgDir, "config", "c", filepath.Join("/etc", db.AppName), "config directory")
}
