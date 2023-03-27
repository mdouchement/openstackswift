package model

// A Container holds several Objects and Manifests.
type Container struct {
	Base `json:",inline" storm:"inline"`

	Name   string `json:"name" storm:"unique"`
	Count  int    `json:"count"`
	Bytes  int64  `json:"bytes"`
}
