package storage

import (
	"io"
	fspkg "io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type fs struct {
	workspace string
}

// NewFileSystem returns a new File System backend.
func NewFileSystem(workspace string) Backend {
	return &fs{
		workspace: workspace,
	}
}

func (b *fs) Name() string {
	return "file_system"
}

func (b *fs) Reader(container, object string) (io.ReadCloser, error) {
	rc, err := os.Open(filepath.Join(b.workspace, container, object))
	if err != nil {
		return rc, errors.Wrap(err, "could not open file")
	}
	return rc, err
}

func (b *fs) Writer(container, object string) (io.WriteCloser, error) {
	b.mkdirAllWithFilename(container, object)

	wc, err := os.Create(filepath.Join(b.workspace, container, object))
	if err != nil {
		return wc, errors.Wrap(err, "could not create file")
	}
	return wc, err
}

func (b *fs) Copy(sc, so, dc, do string) error {
	src, err := os.Open(filepath.Join(b.workspace, sc, so))
	if err != nil {
		return errors.Wrap(err, "copy: source")
	}
	defer src.Close()

	//

	b.mkdirAllWithFilename(dc, do)

	dst, err := os.Create(filepath.Join(b.workspace, dc, do))
	if err != nil {
		return errors.Wrap(err, "copy: destination")
	}
	defer dst.Close()

	//

	_, err = io.Copy(dst, src)
	if err != nil {
		return errors.Wrap(err, "copy")
	}

	err = dst.Sync()
	return errors.Wrap(err, "copy: destination")
}

func (b *fs) FilenamesFrom(prefix string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(b.workspace, prefix))
	if err != nil {
		return nil, err
	}

	var filenames []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filenames = append(filenames, entry.Name())
	}

	return filenames, nil
}

func (b *fs) Exist(container, object string) bool {
	_, err := os.Stat(filepath.Join(b.workspace, container, object))
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true // ignoring error
}

func (b *fs) RemoveAll(path string) error {
	return b.Remove(path, "")
}

func (b *fs) Remove(container, object string) error {
	err := os.RemoveAll(filepath.Join(b.workspace, container, object))
	if err != nil {
		return errors.Wrap(err, "could not delete file")
	}
	return nil
}

func (b *fs) Cleanup() error {
	// Find empty directories.
	//
	stats := map[string]int{}
	err := filepath.Walk(b.workspace, func(path string, info fspkg.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path == b.workspace {
				return nil
			}
			stats[path] = 0
			return nil
		}

		if strings.HasSuffix(path, ".DS_Store") {
			return nil
		}

		trimmedpath := strings.Replace(path, b.workspace, "", 1)
		base := b.workspace

		for _, segment := range strings.Split(filepath.Dir(trimmedpath), string(os.PathSeparator)) {
			base = filepath.Join(base, segment)
			if !strings.HasPrefix(base, b.workspace) {
				continue
			}
			stats[base]++
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "cleanup")
	}

	// Remove empty directories.
	//
	for dirname, count := range stats {
		if count == 0 {
			os.RemoveAll(dirname)
		}
	}
	return nil
}

func (b *fs) mkdirAllWithFilename(container, object string) {
	b.mkdirAll(container, filepath.Dir(object))
}

func (b *fs) mkdirAll(container, object string) {
	if !b.Exist(container, object) {
		os.MkdirAll(filepath.Join(b.workspace, container, object), 0755)
	}
}
