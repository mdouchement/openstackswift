package serializer

import (
	"strings"

	"github.com/mdouchement/openstackswift/internal/model"
)

// TextContainers returns the text serialized form of the given models.
func TextContainers(containers []*model.Container) string {
	sl := make([]string, 0, len(containers))

	for _, container := range containers {
		sl = append(sl, container.Name)
	}

	return strings.Join(sl, "\n")
}

// Containers returns the serialized form of the given models.
func Containers(containers []*model.Container) []map[string]interface{} {
	sl := make([]map[string]interface{}, 0, len(containers))

	for _, container := range containers {
		sl = append(sl, Container(container))
	}

	return sl
}

// Container returns the serialized form of the given model.
func Container(container *model.Container) map[string]interface{} {
	return map[string]interface{}{
		"name":         container.Name,
		"count":        container.Count,
		"bytes":        container.Bytes,
		"last_updated": container.UpdatedAt,
	}
}
