package tests

import (
	"fmt"
	"net/http/httptest"
	"os"
	"regexp"
	"time"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/mdouchement/openstackswift/internal/webserver"
	"github.com/ncw/swift/v2"
	"github.com/sirupsen/logrus"
)

func setup() (*swift.Connection, func()) {
	log := logrus.New()
	log.SetFormatter(&logger.LogrusTextFormatter{
		DisableColors:   false,
		ForceColors:     true,
		ForceFormatting: true,
		PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	//

	dbname, err := os.CreateTemp(os.TempDir(), "swift.db.")
	if err != nil {
		panic(err)
	}

	db, err := database.StormOpen(dbname.Name())
	if err != nil {
		panic(err)
	}

	//

	workspace, err := os.MkdirTemp(os.TempDir(), "swift.")
	if err != nil {
		panic(err)
	}

	//

	ctrl := webserver.Controller{
		Logger:   logger.WrapLogrus(log),
		Database: db,
		Storage:  storage.NewFileSystem(workspace),

		Tenant:   "test",
		Domain:   "Default",
		Username: "tester",
		Password: "testing",
	}
	engine := webserver.EchoEngine(ctrl)

	server := httptest.NewUnstartedServer(engine)
	server.Config.ReadTimeout = 20 * time.Second
	server.Config.WriteTimeout = 20 * time.Second
	server.Start()

	//

	fmt.Println("Listen:", server.URL+"/v3")
	c := &swift.Connection{
		AuthUrl:  server.URL + "/v3",
		Tenant:   ctrl.Tenant,
		Domain:   ctrl.Domain,
		UserName: ctrl.Username,
		ApiKey:   ctrl.Password,
		Region:   "RegionOne",
	}

	return c, func() {
		server.Close()
		db.Close()
		dbname.Close()

		os.RemoveAll(dbname.Name())
		os.RemoveAll(workspace)
	}
}
