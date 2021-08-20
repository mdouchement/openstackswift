package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/scheduler"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/mdouchement/openstackswift/internal/webserver"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const dbname = "swift.db"

var (
	version  = "dev"
	revision = "none"
	date     = "unknown"

	binding string
	port    string
)

func main() {
	c := &cobra.Command{
		Use:     "swift",
		Short:   "Lightweight OpenStack Swift server",
		Version: fmt.Sprintf("%s - build %.7s @ %s - %s", version, revision, date, runtime.Version()),
		Args:    cobra.ExactArgs(0),
	}
	c.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Version for swift",
		Args:  cobra.NoArgs,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(c.Version)
		},
	})
	c.AddCommand(initCmd)
	c.AddCommand(reindexCmd)

	serverCmd.Flags().StringVarP(&binding, "binding", "b", "0.0.0.0", "Server's binding")
	serverCmd.Flags().StringVarP(&port, "port", "p", "5000", "Server's port")
	c.AddCommand(serverCmd)

	if err := c.Execute(); err != nil {
		log.Fatalf("%+v", err)
	}
}

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Init the database",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			return database.StormInit(nameWithEnv("DATABASE_PATH", dbname))
		},
	}

	//

	reindexCmd = &cobra.Command{
		Use:   "reindex",
		Short: "Reindex the database",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			return database.StormReIndex(nameWithEnv("DATABASE_PATH", dbname))
		},
	}

	//

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start server",
		Args:  cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, _ []string) error {
			ctrl := webserver.Controller{
				Version: c.Parent().Version,
				//
				Tenant:   envORdefault("SWIFT_STORAGE_TENANT", "test"),
				Domain:   envORdefault("SWIFT_STORAGE_DOMAIN", "Default"),
				Username: envORdefault("SWIFT_STORAGE_USERNAME", "tester"),
				Password: envORdefault("SWIFT_STORAGE_PASSWORD", "testing"),
			}

			//

			log := logrus.New()
			log.SetFormatter(&logger.LogrusTextFormatter{
				DisableColors:   false,
				ForceColors:     true,
				ForceFormatting: true,
				PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
			ctrl.Logger = logger.WrapLogrus(log)

			//

			db, err := database.StormOpen(nameWithEnv("DATABASE_PATH", dbname))
			if err != nil {
				return errors.Wrap(err, "could not open database")
			}
			defer db.Close()
			ctrl.Database = db

			//

			ctrl.Storage = storage.NewFileSystem(nameWithEnv("STORAGE_PATH", "storage"))

			//

			scheduler.Start(scheduler.Controller{
				Logger:        ctrl.Logger,
				Database:      ctrl.Database,
				Storage:       ctrl.Storage,
				Specification: "@every 30s",
			})

			//

			engine := webserver.EchoEngine(ctrl)
			webserver.PrintRoutes(engine)

			listen := fmt.Sprintf("%s:%s", binding, port)
			log.Printf("Server listening on %s", listen)
			return errors.Wrap(
				engine.Start(listen),
				"could not run server",
			)
		},
	}
)

func nameWithEnv(env, name string) string {
	p := os.Getenv(env)
	if len(p) == 0 {
		return name
	}
	return filepath.Join(p, name)
}

func envORdefault(name, fallback string) string {
	p := os.Getenv(name)
	if len(p) == 0 {
		return fallback
	}
	return p
}
