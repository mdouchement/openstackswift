package database

import (
	"github.com/mdouchement/openstackswift/internal/model"
)

type (
	// A Client can interacts with the database.
	Client interface {
		// Save inserts or updates the entry in database with the given model.
		Save(m model.Model) error
		// Delete deletes the entry in database with the given model.
		Delete(m model.Model) error
		// Close the database.
		Close() error
		// IsNotFound returns true if err is nil or a not found error.
		IsNotFound(err error) bool

		ContainerInteraction
		ManifestInteraction
		ObjectInteraction
	}

	// A ContainerInteraction defines all the methods used to interact with a container record.
	ContainerInteraction interface {
		ListContainers() ([]*model.Container, error)
		FindContainer(id string) (*model.Container, error)
		FindContainerByName(name string) (*model.Container, error)
		DeleteContainer(id string) error
	}

	// A ManifestInteraction defines all the methods used to interact with a manifest record.
	ManifestInteraction interface {
		FindManifestByKey(cid, key string) (*model.Manifest, error)
		DeleteManifest(id string) error
	}

	// A ObjectInteraction defines all the methods used to interact with a object record.
	ObjectInteraction interface {
		AllObjects() ([]*model.Object, error)
		FindObjectsByContainerID(id string) ([]*model.Object, error)
		FindObjectsByManifestID(id string) ([]*model.Object, error)
		FindObjectByKey(cid, key string) (*model.Object, error)
		DeleteObject(id string) error
	}
)
