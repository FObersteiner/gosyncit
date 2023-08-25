package fileset

import "os"

// Fileset stores a set of file paths and maps them to according os.FileInfo
type Fileset struct {
	Paths    map[string]os.FileInfo // TODO : is fileinfo useful here or could it be just a struct ?
	Basepath string
}

func New() Fileset {
	m := make(map[string]os.FileInfo)
	return Fileset{Paths: m}
}

// Contains is a helper to test set membership
func (fm Fileset) Contains(path string) bool {
	if _, ok := fm.Paths[path]; ok {
		return true
	}
	return false
}
