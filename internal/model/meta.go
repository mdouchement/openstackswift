package model

// An Object represents the meta data for object or container stored in the filesystem.
type Meta struct {
	Base `json:",inline" storm:"inline"`

	// container or object
	ContainerID string    `json:"container_id" storm:"index"`
	ObjectKey   string    `json:"object_key"   storm:"index"`
	// meta data key and value for the object
	Key         string    `json:"key"          storm:"index"`
	Value       string    `json:"value"`
}
