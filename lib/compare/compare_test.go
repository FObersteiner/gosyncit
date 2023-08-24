package compare_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"gosyncit/lib/compare"
)

func TestBasic(t *testing.T) {
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	dirB, err := os.MkdirTemp("", "dirB")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirB)

	fileA := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(fileA, []byte("file A"), 0644); err != nil {
		log.Fatal(err)
	}
	fileB := filepath.Join(dirB, "tmpfileA")
	if err := os.WriteFile(fileB, []byte("file B"), 0644); err != nil {
		log.Fatal(err)
	}

	fileInfoA, _ := os.Stat(fileA)
	fileInfoB, _ := os.Stat(fileB)

	// mtime is equal or B is younger. basic comparison therefore returns true:
	equal := compare.BasicEqual(fileInfoB, fileInfoA)
	if !equal {
		t.Log("expected basic comparison to return true, got false")
		t.Fail()
	}
}

func TestDeep(t *testing.T) {
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	dirB, err := os.MkdirTemp("", "dirB")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirB)

	fileA := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(fileA, []byte("file A"), 0644); err != nil {
		log.Fatal(err)
	}
	fileB := filepath.Join(dirB, "tmpfileA")
	if err := os.WriteFile(fileB, []byte("file B"), 0644); err != nil {
		log.Fatal(err)
	}

	equal, err := compare.DeepEqual(fileA, fileB)
	if err != nil {
		log.Fatal(err)
	}
	if equal {
		t.Log("expected deep comparison to return false, got true")
		t.Fail()
	}
}
