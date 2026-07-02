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
	root, err := b.containerRoot(container, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	defer root.Close()

	rc, err := root.Open(object)
	if err != nil {
		return nil, errors.Wrap(err, "could not open file")
	}
	return rc, nil
}

func (b *fs) Writer(container, object string) (io.WriteCloser, error) {
	root, err := b.containerRoot(container, true)
	if err != nil {
		return nil, errors.Wrap(err, "could not create file")
	}
	defer root.Close()

	if dir := filepath.Dir(object); dir != "." {
		if err := root.MkdirAll(dir, 0755); err != nil {
			return nil, errors.Wrap(err, "could not create directory")
		}
	}

	wc, err := root.Create(object)
	if err != nil {
		return nil, errors.Wrap(err, "could not create file")
	}
	return wc, nil
}

func (b *fs) Copy(sc, so, dc, do string) error {
	srcRoot, err := b.containerRoot(sc, false)
	if err != nil {
		return errors.Wrap(err, "copy: source")
	}
	defer srcRoot.Close()

	src, err := srcRoot.Open(so)
	if err != nil {
		return errors.Wrap(err, "copy: source")
	}
	defer src.Close()

	//

	dstRoot, err := b.containerRoot(dc, true)
	if err != nil {
		return errors.Wrap(err, "copy: destination")
	}
	defer dstRoot.Close()

	if dir := filepath.Dir(do); dir != "." {
		if err := dstRoot.MkdirAll(dir, 0755); err != nil {
			return errors.Wrap(err, "copy: destination")
		}
	}

	dst, err := dstRoot.Create(do)
	if err != nil {
		return errors.Wrap(err, "copy: destination")
	}
	defer dst.Close()

	//

	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrap(err, "copy")
	}

	return errors.Wrap(dst.Sync(), "copy: destination")
}

func (b *fs) FilenamesFrom(prefix string) ([]string, error) {
	root, err := os.OpenRoot(b.workspace)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	name := prefix
	if name == "" {
		name = "."
	}
	dir, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	entries, err := dir.ReadDir(-1)
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

func (b *fs) Remove(container, object string) error {
	root, err := os.OpenRoot(b.workspace)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to delete
		}
		return errors.Wrap(err, "could not delete file")
	}
	defer root.Close()

	// RemoveAll(container) routes here with an empty object and drops the whole
	// container directory.
	if object == "" {
		if err := root.RemoveAll(container); err != nil {
			return errors.Wrap(err, "could not delete file")
		}
		return nil
	}

	// Confine the object beneath its container so "../" cannot delete outside it.
	croot, err := root.OpenRoot(container)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // container already gone
		}
		return errors.Wrap(err, "could not delete file")
	}
	defer croot.Close()

	if err := croot.RemoveAll(object); err != nil {
		return errors.Wrap(err, "could not delete file")
	}
	return nil
}

func (b *fs) RemoveAll(path string) error {
	return b.Remove(path, "")
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

// containerRoot returns an os.Root confined to the container's directory inside
// the workspace.  Object names may legitimately contain "/" (pseudo-directories)
// but cannot escape the container: os.Root resolves every path beneath the
// workspace and then the container at the syscall level, rejecting "../"
// traversal and symlinks rather than relying on string comparisons.  When
// create is set the workspace and container directories are created first.
func (b *fs) containerRoot(container string, create bool) (*os.Root, error) {
	if create {
		if err := os.MkdirAll(b.workspace, 0755); err != nil {
			return nil, err
		}
	}

	root, err := os.OpenRoot(b.workspace)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	if create {
		if err := root.MkdirAll(container, 0755); err != nil {
			return nil, err
		}
	}

	return root.OpenRoot(container)
}
