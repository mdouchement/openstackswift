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
	err := c.db.AllByIndex("Name", &containers)
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

func (c *strm) FindObjectsByContainerID(id string, limit int, prefix string) ([]*model.Object, error) {
	objects := make([]*model.Object, 0)
	if (limit == 0) {
		limit = -1
	}
	err := c.db.Select(q.Eq("ContainerID", id), q.Re("Key", "^" + prefix)).Limit(limit).OrderBy("Key").Find(&objects)
	return objects, errors.Wrap(err, "could not get objects by container_id")
}

func (c *strm) FindObjectsByManifestID(id string) ([]*model.Object, error) {
	objects := make([]*model.Object, 0)
	err := c.db.Select(q.Eq("ManifestID", id)).OrderBy("CreatedAt").OrderBy("Key").Find(&objects)
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

//
// Meta
//
func (c *strm) AddMeta(cid, okey string, key string, value string) (*model.Meta, error) {
	var meta_model = new(model.Meta)
	meta_model.ContainerID = cid
	meta_model.ObjectKey = okey
	meta_model.Key = key
	meta_model.Value = value
	if err := c.Save(meta_model); err != nil {
		return nil, errors.Wrap(err, "could not save meta")
	}
	return meta_model, nil
}

func (c *strm) FindMeta(cid, okey string) ([]*model.Meta, error) {
	var metas = make([]*model.Meta, 0)
	err := c.db.Select(q.Eq("ContainerID", cid), q.Eq("ObjectKey", okey)).Find(&metas)
	return metas, errors.Wrap(err, "could not find metas")
}

func (c *strm) DeleteMeta(cid, okey string, key string) (error) {
	err := c.db.Select(q.Eq("ContainerID", cid), q.Eq("ObjectKey", okey), q.Eq("Key", key)).Delete(&model.Meta{})
	return errors.Wrap(err, "could not delete meta")
}

func (c *strm) DeleteAllMetas(cid, okey string) (error) {
	err := c.db.Select(q.Eq("ContainerID", cid), q.Eq("ObjectKey", okey)).Delete(&model.Meta{})
	return errors.Wrap(err, "could not delete all metas")
}
