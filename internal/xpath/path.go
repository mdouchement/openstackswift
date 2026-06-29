package xpath

import (
	"net/url"
	"path"
	"strings"
)

// Entities takes the path p and extracts the container and the object.
func Entities(p string) (container, object string) {
	cp, err := url.PathUnescape(p)
	if err == nil {
		p = cp
	}

	// Swift's X-Copy-From and Destination headers are conventionally written
	// with a leading slash ("/container/object"); drop it so the container is
	// not parsed as an empty segment.
	p = strings.TrimPrefix(p, "/")

	artifacts := strings.Split(p, "/")
	if len(artifacts) < 2 {
		return artifacts[0], ""
	}
	return artifacts[0], path.Join(artifacts[1:]...)
}
