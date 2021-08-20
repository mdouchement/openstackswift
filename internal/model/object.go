package model

import "time"

// An Object represents the blob stored on the filesystem.
type Object struct {
	Base `json:",inline" storm:"inline"`

	ContainerID string `json:"container_id" storm:"index"`
	ManifestID  string `json:"manifest_id"  storm:"index"`

	Key         string    `json:"key"          storm:"index"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	Checksum    string    `json:"checksum"`
	TTL         time.Time `json:"ttl"          storm:"index"`
}
