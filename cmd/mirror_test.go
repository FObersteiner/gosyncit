package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/FObersteiner/gosyncit/cmd"
	"github.com/FObersteiner/gosyncit/lib/fileset"
)

func TestMirror(t *testing.T) {
	dry := false
	clean := false
	ignorehidden := false

	err := cmd.Mirror("A", "B", dry, clean, ignorehidden)
	if err == nil {
		t.Fail()
		t.Log("mirror must fail with invalid src/dst input")
	}

	src, err := os.MkdirTemp("", "src")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(src)

	dst, err := os.MkdirTemp("", "dst")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst)

	// hidden file in root
	fname := filepath.Join(src, ".hidden.file")
	if err := os.WriteFile(fname, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// hidden file in subdirectory
	fname = filepath.Join(src, "subdir", ".hidden.file")
	_ = os.MkdirAll(filepath.Dir(fname), 0755)
	if err := os.WriteFile(fname, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// hidden directory
	fname = filepath.Join(src, ".hidden", "a.file")
	_ = os.MkdirAll(filepath.Dir(fname), 0755)
	if err := os.WriteFile(fname, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Round 1: ignore
	clean = false
	ignorehidden = true
	if err := cmd.Mirror(src, dst, dry, clean, ignorehidden); err != nil {
		t.Logf("mirror (ignore hidden: %v) failed with %v", ignorehidden, err)
		t.Fail()
	}
	fs_dst, _ := fileset.New(dst)
	_ = fs_dst.Populate()
	have, want := len(fs_dst.Paths), 0
	if have != want {
		t.Logf("ignore hidden: want %v entries in dst, have %v", want, have)
	}

	// Round 2: do not ignore
	clean = false
	ignorehidden = false
	if err := cmd.Mirror(src, dst, dry, clean, ignorehidden); err != nil {
		t.Logf("mirror (ignore hidden: %v) failed with %v", ignorehidden, err)
		t.Fail()
	}
	fs_dst, _ = fileset.New(dst)
	_ = fs_dst.Populate()
	have, want = len(fs_dst.Paths), 5
	if have != want {
		t.Logf("copy hidden: want %v entries in dst, have %v", want, have)
	}

	// Round 3: clean ??
	// if hidden stuff is ignored in src, it should not exist in dst
	clean = true
	ignorehidden = true
	if err := cmd.Mirror(src, dst, dry, clean, ignorehidden); err != nil {
		t.Logf("mirror (ignore hidden: %v) failed with %v", ignorehidden, err)
		t.Fail()
	}
	fs_dst, _ = fileset.New(dst)
	_ = fs_dst.Populate()
	have, want = len(fs_dst.Paths), 0
	if have != want {
		t.Logf("clean hidden: want %v entries in dst, have %v", want, have)
	}
}
