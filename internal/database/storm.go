package database

import (
	"time"

	"github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/codec/json"
	"github.com/asdine/storm/v3/q"
	"github.com/gofrs/uuid"
	"github.com/mdouchement/openstackswift/internal/model"
	"github.com/pkg/errors"
)

type strm struct {
	db *storm.DB
}

// StormCodec is the format used to store data in the database.
var StormCodec = storm.Codec(json.Codec)

// StormInit initializes Storm database.
func StormInit(database string) error {
	db, err := storm.Open(database, StormCodec)
	if err != nil {
		return errors.Wrap(err, "could not get database connection")
	}

	if err := db.Init(&model.Container{}); err != nil {
		return errors.Wrap(err, "could not init container index")
	}

	if err := db.Init(&model.Manifest{}); err != nil {
		return errors.Wrap(err, "could not init manifest index")
	}

	err = db.Init(&model.Object{})
	return errors.Wrap(err, "could not init object index")
}

func StormReIndex(database string) error {
	db, err := storm.Open(database, StormCodec)
	if err != nil {
		return errors.Wrap(err, "could not get database connection")
	}

	if err := db.ReIndex(&model.Container{}); err != nil {
		return errors.Wrap(err, "could not ReIndex containers")
	}

	if err := db.ReIndex(&model.Manifest{}); err != nil {
		return errors.Wrap(err, "could not ReIndex manifests")
	}

	err = db.ReIndex(&model.Object{})
	return errors.Wrap(err, "could not ReIndex objects")
}

func StormOpen(database string) (Client, error) {
	db, err := storm.Open(database, StormCodec)
	if err != nil {
		return nil, errors.Wrap(err, "could not get database connection")
	}

	return &strm{
		db: db,
	}, nil
}

func (c *strm) Save(m model.Model) error {
	t := time.Now().UTC()
	m.SetUpdatedAt(t)

	if m.GetID() == "" {
		m.SetID(uuid.Must(uuid.NewV4()).String())
		m.SetCreatedAt(t)
	}

	return errors.Wrap(c.db.Save(m), "could not save the model")
}

func (c *strm) Delete(m model.Model) error {
	return errors.Wrap(c.db.DeleteStruct(m), "could not delete the model")
}

func (c *strm) Close() error {
	return c.db.Close()
}

func (c *strm) IsNotFound(err error) bool {
	return errors.Cause(err) == storm.ErrNotFound
}

//
// Container
//

func (c *strm) ListContainers() ([]*model.Container, error) {
	containers := make([]*model.Container, 0)
	err := c.db.All(&containers)
	return containers, errors.Wrap(err, "could not get all containers")
}

func (c *strm) FindContainer(id string) (*model.Container, error) {
	var container model.Container
	err := c.db.One("ID", id, &container)
	return &container, errors.Wrap(err, "could not find container")
}

func (c *strm) FindContainerByName(name string) (*model.Container, error) {
	var container model.Container
	err := c.db.One("Name", name, &container)
	return &container, errors.Wrap(err, "could not find container")
}

func (c *strm) DeleteContainer(id string) error {
	err := c.db.Select(q.Eq("ID", id)).Delete(&model.Container{})
	return errors.Wrap(err, "could not delete container")
}

//
// Object
//

func (c *strm) AllObjects() ([]*model.Object, error) {
	objects := make([]*model.Object, 0)
	err := c.db.All(&objects)
	return objects, errors.Wrap(err, "could not get all objects")
}

func (c *strm) FindObjectsByContainerID(id string) ([]*model.Object, error) {
	objects := make([]*model.Object, 0)
	err := c.db.Select(q.Eq("ContainerID", id)).Find(&objects)
	return objects, errors.Wrap(err, "could not get objects by container_id")
}

func (c *strm) FindObjectsByManifestID(id string) ([]*model.Object, error) {
	objects := make([]*model.Object, 0)
	err := c.db.Select(q.Eq("ManifestID", id)).OrderBy("CreatedAt").Find(&objects)
	return objects, errors.Wrap(err, "could not get objects by manifest_id")
}

func (c *strm) FindObjectByKey(cid, key string) (*model.Object, error) {
	var object model.Object
	err := c.db.Select(q.Eq("ContainerID", cid), q.Eq("Key", key)).First(&object)
	return &object, errors.Wrap(err, "could not find object")
}

func (c *strm) DeleteObject(id string) error {
	err := c.db.Select(q.Eq("ID", id)).Delete(&model.Object{})
	return errors.Wrap(err, "could not delete object")
}

//
// Manifest
//

func (c *strm) FindManifestByKey(cid, key string) (*model.Manifest, error) {
	var manifest model.Manifest
	err := c.db.Select(q.Eq("ContainerID", cid), q.Eq("Key", key)).First(&manifest)
	return &manifest, errors.Wrap(err, "could not find manifest")
}

func (c *strm) DeleteManifest(id string) error {
	err := c.db.Select(q.Eq("ID", id)).Delete(&model.Manifest{})
	return errors.Wrap(err, "could not delete manifest")
}
