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
	return ObjectsWithDelimiter(objects, prefix, "")
}

// ObjectsWithDelimiter serializes the listing, applying Swift's delimiter
// roll-up: object keys that contain the delimiter after the prefix are
// collapsed into deduplicated {"subdir": "<prefix + segment up to and
// including the delimiter>"} entries, and only keys without the delimiter
// are returned as objects.  An empty delimiter lists every object.
func ObjectsWithDelimiter(objects []*model.Object, prefix, delimiter string) []map[string]interface{} {
	sl := make([]map[string]interface{}, 0, len(objects))
	seen := make(map[string]bool)

	for _, object := range objects {
		if !strings.HasPrefix(object.Key, prefix) {
			continue
		}

		if delimiter != "" {
			rest := object.Key[len(prefix):]
			if idx := strings.Index(rest, delimiter); idx >= 0 {
				subdir := prefix + rest[:idx+len(delimiter)]
				if !seen[subdir] {
					seen[subdir] = true
					sl = append(sl, map[string]interface{}{"subdir": subdir})
				}
				continue
			}
		}

		sl = append(sl, Object(object))
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
