package copy_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"

	cp "gosyncit/lib/copy"
)

func TestCopyFile(t *testing.T) {
	// remove umask so that arbitrary permissions can be specified
	// TODO : does this work on !linux ?!
	oldumask := syscall.Umask(0000)
	defer syscall.Umask(oldumask)

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
	if err := os.WriteFile(file, []byte("content"), 0655); err != nil {
		log.Fatal(err)
	}

	fstSrc, _ := os.Stat(file)
	// log.Println(fstSrc.Mode())

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

	// fstDst, _ := os.Stat(dst)
	// log.Printf("%o\n", fstDst.Mode())

	if runtime.GOOS != "windows" {
		err = cp.CopyPerm(file, dst)
		if err != nil {
			log.Fatal(err)
		}
		fstDst2, _ := os.Stat(dst)
		// log.Println(fstDst2.Mode())

		if fstSrc.Mode() != fstDst2.Mode() {
			t.Logf("Expected permission %o, got %o", fstSrc.Mode(), fstDst2.Mode())
			t.Fail()
		}
	}
}
