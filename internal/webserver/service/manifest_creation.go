package service

import (
	"crypto/md5"
	"encoding/hex"
	pathpkg "path"

	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/mdouchement/openstackswift/internal/xpath"
	"github.com/pkg/errors"
)

// A ManifestCreation is used to create manifest from uploaded objects.
type ManifestCreation struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	manifest  *model.Manifest
}

// NewObjectUploader returns a new ObjectUploader.
func NewManifestCreation(database database.Client, storage storage.Backend, container *model.Container, manifest *model.Manifest) *ManifestCreation {
	return &ManifestCreation{
		database:  database,
		storage:   storage,
		container: container,
		manifest:  manifest,
	}
}

func (s *ManifestCreation) Create(path string) (err error) {
	containername, basekey := xpath.Entities(path)

	//

	container := s.container
	if container.Name != containername {
		container, err = s.database.FindContainerByName(containername)
		if err != nil {
			return errors.Wrap(err, "X-Object-Manifest")
		}
	}

	//

	filenames, err := s.storage.FilenamesFrom(path)
	if err != nil {
		return errors.Wrap(err, "could not get filenames")
	}

	var size int64
	h := md5.New()
	for _, filename := range filenames {
		object, err := s.database.FindObjectByKey(container.ID, pathpkg.Join(basekey, filename))
		if err != nil {
			return errors.Wrap(err, "X-Object-Manifest")
		}

		object.ManifestID = s.manifest.ID
		if err = s.database.Save(object); err != nil {
			return errors.Wrap(err, "X-Object-Manifest")
		}

		//

		if s.manifest.ContentType == "" {
			s.manifest.ContentType = object.ContentType
		}

		size += object.Size
		h.Write([]byte(object.Checksum))
	}

	s.manifest.Size = size
	s.manifest.Checksum = hex.EncodeToString(h.Sum(nil))
	return nil
}
