package copy_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	cp "github.com/FObersteiner/gosyncit/lib/copy"
)

func TestCopyFile(t *testing.T) {
	// remove umask so that arbitrary permissions can be specified
	// TODO : does this work on non-linux platform ?!
	oldumask := syscall.Umask(0000)
	defer syscall.Umask(oldumask)
	// could also use os.Chmod after write, to set permissions on src file

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

	file := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(file, []byte("content"), 0655); err != nil {
		t.Fatal(err)
	}

	fstSrc, _ := os.Stat(file)

	dst := filepath.Join(dirB, "tmpfileA")
	err = cp.CopyFile(file, dst, fstSrc, false)
	if err != nil {
		t.Fatal(err)
	}

	f1, err1 := os.ReadFile(file)
	if err1 != nil {
		t.Fatal(err1)
	}

	f2, err2 := os.ReadFile(dst)
	if err2 != nil {
		t.Fatal(err2)
	}

	if !bytes.Equal(f1, f2) {
		t.Logf("Expected %x, got %x", f1, f2)
		t.Fail()
	}

	fstDst, _ := os.Stat(dst)
	if fstSrc.ModTime() != fstDst.ModTime() {
		t.Logf("Expected mtime to be equal, got %v (src) and %v (dst)", fstSrc.ModTime(), fstDst.ModTime())
		t.Fail()
	}

	if runtime.GOOS != "windows" {
		err = cp.CopyPerm(file, dst)
		if err != nil {
			t.Fatal(err)
		}

		fstDst2, _ := os.Stat(dst)

		if fstSrc.Mode() != fstDst2.Mode() {
			t.Logf("Expected permission %o, got %o", fstSrc.Mode(), fstDst2.Mode())
			t.Fail()
		}
	}
}
