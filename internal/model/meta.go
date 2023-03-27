package model

// An Object represents the blob stored on the filesystem.
type Meta struct {
	Base `json:",inline" storm:"inline"`

	// container and object
	ContainerID string    `json:"container_id" storm:"index"`
	ObjectKey   string    `json:"object_key"   storm:"index"`
	Key         string    `json:"key"          storm:"index"`
	Value       string    `json:"value"`
}
