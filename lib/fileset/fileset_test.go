package fileset_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	fm "gosyncit/lib/fileset"
)

func TestNewFileset(t *testing.T) {
	_, err := fm.New("")
	if err != fm.ErrInvalidBasepath {
		log.Fatal(err)
	}

	data := []byte("content")
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	fname := "tmpfileA"
	file := filepath.Join(dirA, fname)
	if err := os.WriteFile(file, data, 0644); err != nil {
		log.Fatal(err)
	}

	m, err := fm.New(dirA)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Populate()
	if err != nil {
		log.Fatal(err)
	}

	//	log.Println(m.String())

	if !m.Contains(fname) {
		t.Log("Expected map to contain file")
		t.Fail()
	}

	_, ok := m.Paths[file]
	if ok {
		t.Log("found unexpected filename")
		t.Fail()
	}

	storedInfo, ok := m.Paths[fname]
	if !ok {
		t.Log("expected to find fname")
		t.Fail()
	}

	if storedInfo.Size() != int64(len(data)) {
		t.Logf("Expected size %v, got %v", len(data), storedInfo.Size())
		t.Fail()
	}
}