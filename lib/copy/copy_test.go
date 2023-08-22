package copy_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	cp "gosyncit/lib/copy"
)

func TestCopyFile(t *testing.T) {
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

	file := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(file, []byte("content"), 0644); err != nil {
		log.Fatal(err)
	}

	dst := filepath.Join(dirB, "tmpfileA")
	err = cp.CopyFile(file, dst)
	if err != nil {
		log.Fatal(err)
	}

	f1, err1 := os.ReadFile(file)
	if err1 != nil {
		log.Fatal(err1)
	}

	f2, err2 := os.ReadFile(dst)
	if err2 != nil {
		log.Fatal(err2)
	}

	if !bytes.Equal(f1, f2) {
		t.Logf("Expected %x, got %x", f1, f2)
		t.Fail()
	}
}
