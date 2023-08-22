package filemap

import "os"

// Filemap stores a set of file paths and maps them to according os.FileInfo
type Filemap map[string]os.FileInfo

// Contains is a helper to test set membership
func (fm Filemap) Contains(path string) bool {
	if _, ok := fm[path]; ok {
		return true
	}
	return false
}
