package service

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/pkg/errors"
)

// A Destroyer removes file(s) from storage.
type Destroyer interface {
	Destroy() error
}

// SetupObjectTTL configures the time to live to live according the requests headers.
func SetupObjectTTL(m *model.Object, r *http.Request) error {
	m.TTL = time.Time{} // Reset

	delete := r.Header.Get("X-Delete-After")
	if delete != "" {
		seconds, err := strconv.ParseInt(delete, 10, 64)
		if err != nil {
			return errors.Wrap(err, "X-Delete-After")
		}

		m.TTL = time.Now().Add(time.Duration(seconds) * time.Second)
		return nil
	}

	delete = r.Header.Get("X-Delete-At")
	if delete != "" {
		unix, err := strconv.ParseInt(delete, 10, 64)
		if err != nil {
			return errors.Wrap(err, "X-Delete-At")
		}

		m.TTL = time.Unix(unix, 0)
	}

	return nil
}

//
//-----
//

// An ObjectDestroyer removes the file from storage.
type ObjectDestroyer struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	object    *model.Object
}

// NewObjectDestroyer returns a new ObjectDestroyer.
func NewObjectDestroyer(database database.Client, storage storage.Backend, container *model.Container, object *model.Object) Destroyer {
	return &ObjectDestroyer{
		database:  database,
		storage:   storage,
		container: container,
		object:    object,
	}
}

func (s *ObjectDestroyer) Destroy() error {
	err := s.storage.Remove(s.container.Name, s.object.Key)
	if err != nil {
		return errors.Wrap(err, "ObjectDestroyer")
	}

	err = s.database.DeleteObject(s.object.ID)
	return errors.Wrap(err, "ObjectDestroyer")
}

//
//-----
//

// An ManifestDestroyer removes the segment objects from storage.
type ManifestDestroyer struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	manifest  *model.Manifest
}

// NewManifestDestroyer returns a new ManifestDestroyer.
func NewManifestDestroyer(database database.Client, storage storage.Backend, container *model.Container, manifest *model.Manifest) Destroyer {
	return &ManifestDestroyer{
		database:  database,
		storage:   storage,
		container: container,
		manifest:  manifest,
	}
}

func (s *ManifestDestroyer) Destroy() error {
	objects, err := s.database.FindObjectsByManifestID(s.manifest.ID)
	if err != nil {
		return errors.Wrap(err, "ManifestDestroyer")
	}

	for _, object := range objects {
		container, err := s.database.FindContainer(object.ContainerID)
		if err != nil {
			return errors.Wrap(err, "ManifestDestroyer")
		}

		err = s.storage.Remove(container.Name, object.Key)
		if err != nil {
			return errors.Wrap(err, "ManifestDestroyer")
		}

		err = s.database.DeleteObject(object.ID)
		if err != nil {
			return errors.Wrap(err, "ManifestDestroyer")
		}
	}

	//

	err = s.database.DeleteManifest(s.manifest.ID)
	return errors.Wrap(err, "ManifestDestroyer")
}
