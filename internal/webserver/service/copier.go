package service

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"time"

	"github.com/mdouchement/openstackswift/internal/database"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/mdouchement/openstackswift/internal/storage"
	"github.com/ncw/swift/v2"
	"github.com/pkg/errors"
)

type Copier interface {
	Copy(container, object string) error
	CreatedAt() time.Time
	Checksum() string
}

//
//-----
//

type ObjectCopier struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	object    *model.Object

	createdAt time.Time
}

func NewObjectCopier(database database.Client, storage storage.Backend, container *model.Container, object *model.Object) Copier {
	return &ObjectCopier{
		database:  database,
		storage:   storage,
		container: container,
		object:    object,
	}
}

func (s *ObjectCopier) Copy(containername, objectname string) error {
	container, err := s.database.FindContainerByName(containername)
	if err != nil {
		return errors.Wrap(err, "ObjectCopier")
	}

	err = s.storage.Copy(s.container.Name, s.object.Key, containername, objectname)
	if err != nil {
		return errors.Wrap(err, "ObjectCopier")
	}

	object := new(model.Object)
	object.ContainerID = container.ID
	object.Key = objectname
	object.ContentType = s.object.ContentType
	object.Checksum = s.object.Checksum
	object.Size = s.object.Size

	err = s.database.Save(object)
	s.createdAt = *object.CreatedAt
	return errors.Wrap(err, "ObjectCopier")
}

func (s *ObjectCopier) CreatedAt() time.Time {
	return s.createdAt
}

func (s *ObjectCopier) Checksum() string {
	return s.object.Checksum
}

//
//-----
//

// A ManifestCopier handles copie of a manifest.
// The semantics are the same for both static and dynamic large objects.
// When copying large objects, the COPY operation does not create
// a manifest object but a normal object with content same as
//what you would get on a GET request to the original manifest object.
// https://docs.openstack.org/swift/latest/api/large_objects.html
type ManifestCopier struct {
	database  database.Client
	storage   storage.Backend
	container *model.Container
	manifest  *model.Manifest
	object    *model.Object
}

// NewManifestCopier returns a new ManifestCopier.
func NewManifestCopier(database database.Client, storage storage.Backend, container *model.Container, manifest *model.Manifest) Copier {
	return &ManifestCopier{
		database:  database,
		storage:   storage,
		container: container,
		manifest:  manifest,
		object:    new(model.Object),
	}
}

// If you make a COPY request by using a manifest object as the source,
// the new object is a normal, and not a segment, object.
// If the total size of the source segment objects exceeds 5 GB, the COPY request fails.
// However, you can make a duplicate of the manifest object and this new object can be larger than 5 GB.
func (s *ManifestCopier) Copy(containername, objectname string) error {
	if s.manifest.Size > 5<<30 {
		return swift.TooLargeObject
	}

	container, err := s.database.FindContainerByName(containername)
	if err != nil {
		return errors.Wrap(err, "ManifestCopier")
	}

	//

	s.object.ContainerID = container.ID
	s.object.Key = objectname
	s.object.ContentType = s.manifest.ContentType

	//

	objects, err := s.database.FindObjectsByManifestID(s.manifest.ID)
	if err != nil {
		return errors.Wrap(err, "ManifestCopier")
	}

	//

	wc, err := s.storage.Writer(container.Name, s.object.Key)
	if err != nil {
		return errors.Wrap(err, "ManifestCopier")
	}
	defer wc.Close()

	h := md5.New()
	w := io.MultiWriter(h, wc)

	//

	for _, o := range objects {
		scontainer, err := s.database.FindContainer(o.ContainerID)
		if err != nil {
			return errors.Wrap(err, "ManifestCopier")
		}

		r, err := s.storage.Reader(scontainer.Name, o.Key)
		if err != nil {
			return errors.Wrap(err, "ManifestCopier")
		}

		n, err := io.Copy(w, r)
		if err != nil {
			return errors.Wrap(err, "ManifestCopier")
		}
		s.object.Size += n
	}
	s.object.Checksum = hex.EncodeToString(h.Sum(nil))

	if s.object.Size != s.manifest.Size {
		return swift.ObjectCorrupted
	}

	err = s.database.Save(s.object)
	return errors.Wrap(err, "ManifestCopier")
}

func (s *ManifestCopier) CreatedAt() time.Time {
	return *s.object.CreatedAt
}

func (s *ManifestCopier) Checksum() string {
	return s.object.Checksum
}
