package compare_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FObersteiner/gosyncit/lib/compare"
)

func TestBasic(t *testing.T) {
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	dirB, err := os.MkdirTemp("", "dirB")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirB)

	fileA := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(fileA, []byte("file A"), 0644); err != nil {
		t.Fatal(err)
	}
	fileB := filepath.Join(dirB, "tmpfileA")
	if err := os.WriteFile(fileB, []byte("file B"), 0644); err != nil {
		t.Fatal(err)
	}

	fileInfoA, _ := os.Stat(fileA)
	fileInfoB, _ := os.Stat(fileB)

	// mtime is equal or B is younger. basic comparison therefore must return unequal == false
	unequal := compare.BasicUnequal(fileInfoB, fileInfoA)
	if unequal {
		t.Log("expected basic comparison to return false, got true")
		t.Fail()
	}
	unequal = compare.SrcYounger(fileInfoA, fileInfoB)
	if unequal {
		t.Log("expected src younger comparison to return false, got true")
		t.Fail()
	}
}

func TestDeep(t *testing.T) {
	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirA)

	dirB, err := os.MkdirTemp("", "dirB")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dirB)

	fileA := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(fileA, []byte("file A"), 0644); err != nil {
		t.Fatal(err)
	}
	fileB := filepath.Join(dirB, "tmpfileA")
	if err := os.WriteFile(fileB, []byte("file B"), 0644); err != nil {
		t.Fatal(err)
	}

	equal, err := compare.DeepEqual(fileA, fileB)
	if err != nil {
		t.Fatal(err)
	}
	if equal {
		t.Log("expected deep comparison to return false, got true")
		t.Fail()
	}
}
