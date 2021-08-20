package service

import (
	"crypto/md5"
	"encoding/hex"
	"io"

	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/storage"
)

// An ObjectUploader performs upload and metrics.
type ObjectUploader struct {
	storage   storage.Backend
	container *model.Container
	object    *model.Object
}

// NewObjectUploader returns a new ObjectUploader.
func NewObjectUploader(storage storage.Backend, container *model.Container, object *model.Object) *ObjectUploader {
	return &ObjectUploader{
		storage:   storage,
		container: container,
		object:    object,
	}
}

// Upload performs the upload and update the inner Object.
func (s *ObjectUploader) Upload(r io.Reader) error {
	wc, err := s.storage.Writer(s.container.Name, s.object.Key)
	if err != nil {
		return err
	}
	defer wc.Close()

	h := md5.New()
	w := io.MultiWriter(h, wc)

	n, err := io.Copy(w, r)
	if err != nil {
		return err
	}

	s.object.Size = n
	s.object.Checksum = hex.EncodeToString(h.Sum(nil))
	return nil
}
