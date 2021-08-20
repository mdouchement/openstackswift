package service

import (
	"io"

	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/pkg/errors"
)

type Downloader interface {
	Stream() (io.ReadCloser, error)
	ContentType() string
	Size() int64
	Checksum() string
}

//
//-----
//

type ObjectDownloader struct {
	storage   storage.Backend
	container *model.Container
	object    *model.Object
}

func NewObjectDownloader(storage storage.Backend, container *model.Container, object *model.Object) Downloader {
	return &ObjectDownloader{
		storage:   storage,
		container: container,
		object:    object,
	}
}

func (s *ObjectDownloader) Stream() (io.ReadCloser, error) {
	return s.storage.Reader(s.container.Name, s.object.Key)
}

func (s *ObjectDownloader) ContentType() string {
	return s.object.ContentType
}

func (s *ObjectDownloader) Size() int64 {
	return s.object.Size
}

func (s *ObjectDownloader) Checksum() string {
	return s.object.Checksum
}

//
//-----
//

type ManifestDownloader struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	manifest  *model.Manifest
}

func NewManifestDownloader(database database.Client, storage storage.Backend, container *model.Container, manifest *model.Manifest) Downloader {
	return &ManifestDownloader{
		database:  database,
		storage:   storage,
		container: container,
		manifest:  manifest,
	}
}

func (s *ManifestDownloader) Stream() (io.ReadCloser, error) {
	objects, err := s.database.FindObjectsByManifestID(s.manifest.ID)
	if err != nil {
		return nil, errors.Wrap(err, "ManifestDownloader")
	}

	reader := &mreader{}
	var readers []io.Reader
	for _, object := range objects {
		container, err := s.database.FindContainer(object.ContainerID)
		if err != nil {
			return nil, errors.Wrap(err, "ManifestDownloader")
		}

		r, err := s.storage.Reader(container.Name, object.Key)
		if err != nil {
			return nil, errors.Wrap(err, "ManifestDownloader")
		}
		readers = append(readers, r)
		reader.closers = append(reader.closers, r)
	}

	reader.Reader = io.MultiReader(readers...)
	return reader, nil
}

func (s *ManifestDownloader) ContentType() string {
	return s.manifest.ContentType
}

func (s *ManifestDownloader) Size() int64 {
	return s.manifest.Size
}

func (s *ManifestDownloader) Checksum() string {
	return s.manifest.Checksum
}

//
//-----
//

type mreader struct {
	io.Reader
	closers []io.Closer
}

func (r *mreader) Close() error {
	for _, closer := range r.closers {
		closer.Close()
	}
	return nil
}
