package serializer

import (
	"strings"

	"github.com/mdouchement/openstackswift/internal/model"
)

// TextObjects returns the text serialized form of the given models.
func TextObjects(objects []*model.Object, prefix string) string {
	sl := make([]string, 0, len(objects))

	for _, object := range objects {
		if strings.HasPrefix(object.Key, prefix) {
			sl = append(sl, object.Key)
		}
	}

	return strings.Join(sl, "\n")
}

// Objects returns the serialized form of the given models.
func Objects(objects []*model.Object, prefix string) []map[string]interface{} {
	sl := make([]map[string]interface{}, 0, len(objects))

	for _, object := range objects {
		if strings.HasPrefix(object.Key, prefix) {
			sl = append(sl, Object(object))
		}
	}

	return sl
}

// Object returns the serialized form of the given model.
func Object(object *model.Object) map[string]interface{} {
	return map[string]interface{}{
		"name":          object.Key,
		"content_type":  object.ContentType,
		"bytes":         object.Size,
		"last_modified": object.UpdatedAt,
		"hash":          object.Checksum,
	}
}
