package model

// A Manifest represents aggregates an blob accross several Objects used by chunked upload.
type Manifest struct {
	Base `json:",inline" storm:"inline"`

	ContainerID string `json:"container_id" storm:"index"`

	// URI         string `json:"uri"`
	Key         string `json:"key"          storm:"index"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	// FilePath    string `json:"file_path"`
	Checksum string `json:"checksum"`
}
