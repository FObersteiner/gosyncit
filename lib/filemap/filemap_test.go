package filemap_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	fm "gosyncit/lib/filemap"
)

func TestContains(t *testing.T) {
	data := []byte("content")
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	file := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(file, data, 0644); err != nil {
		log.Fatal(err)
	}

	info, err := os.Stat(file)
	if err != nil {
		log.Fatal(err)
	}
	m := make(fm.Filemap)
	m[file] = info

	if !m.Contains(file) {
		t.Log("Expected map to contain file")
		t.Fail()
	}

	storedInfo := m[file]
	if storedInfo.Size() != int64(len(data)) {
		t.Logf("Expected size %v, got %v", len(data), storedInfo.Size())
		t.Fail()
	}
}
