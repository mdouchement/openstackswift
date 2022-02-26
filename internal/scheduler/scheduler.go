package scheduler

import (
	"path"
	"time"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/robfig/cron/v3"
)

// A Controller is an Iversion Of Control pattern used to init the server package.
type Controller struct {
	Logger        logger.Logger
	Database      database.Client
	Storage       storage.Backend
	Specification string
}

// Start lauches the scheduler asynchronously.
func Start(c Controller) {
	cron := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DiscardLogger),
	))

	log := c.Logger.WithPrefix("[scheduler]")

	_, err := cron.AddFunc(c.Specification, func() {
		log = c.Logger.WithPrefix("[TTL]")

		objects, err := c.Database.AllObjects()
		if err != nil {
			log.Error(err)
			return
		}

		for _, object := range objects {
			if object.TTL.IsZero() {
				continue
			}

			if object.TTL.After(time.Now()) {
				continue
			}

			container, err := c.Database.FindContainer(object.ContainerID)
			if err != nil {
				log.Error(err)
				return
			}

			err = c.Storage.Remove(container.Name, object.Key)
			if err != nil {
				log.Error(err)
				return
			}

			err = c.Database.Delete(object)
			if err != nil {
				log.Error(err)
				return
			}

			log.Infof("Removed %s", path.Join(container.Name, object.Key))
		}

		log.Info("Storage cleanup")
		err = c.Storage.Cleanup()
		if err != nil {
			log.Error(err)
			return
		}
	})
	if err != nil {
		panic(err)
	}
	log.Info("TTL object task registred")

	cron.Start()
	log.Info("Scheduler is running")
}
