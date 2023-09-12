package fileset

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidBasepath = errors.New("basepath must be an existing directory")

// Fileset stores a set of file paths and maps them to according os.FileInfo
type Fileset struct {
	Paths    map[string]os.FileInfo
	Basepath string
}

// New returns a new Fileset with only the basepath specified
func New(basepath string) (*Fileset, error) {
	if basepath == "" {
		return nil, ErrInvalidBasepath
	}
	info, err := os.Stat(basepath)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(basepath, 0755)
		info, err = os.Stat(basepath)
	}
	if err != nil {
		return nil, ErrInvalidBasepath
	}
	if !info.IsDir() {
		return nil, ErrInvalidBasepath
	}
	if !strings.HasSuffix(basepath, string(os.PathSeparator)) {
		basepath += string(os.PathSeparator)
	}
	m := make(map[string]os.FileInfo)
	return &Fileset{Basepath: basepath, Paths: m}, nil
}

// Populate walks the basepath of the Fileset to populate the paths map
func (fs *Fileset) Populate() error {
	err := filepath.Walk(fs.Basepath,
		func(path string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fs.Paths[strings.TrimPrefix(path, fs.Basepath)] = finfo
			return nil
		})
	if err != nil {
		return err
	}

	delete(fs.Paths, "") // removes basepath from map

	return nil
}

// Contains is a helper to test set membership
func (fm *Fileset) Contains(path string) bool {
	if _, ok := fm.Paths[path]; ok {
		return true
	}
	return false
}

func (fs *Fileset) String() string {
	repr := fmt.Sprintf("Fileset,\n  basepath: %v\n  Content:", fs.Basepath)
	for k := range fs.Paths {
		repr += fmt.Sprintf("\n    %v", k)
	}
	return repr
}
