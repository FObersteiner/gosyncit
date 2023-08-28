package copy_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"syscall"
	"testing"

	"gosyncit/lib/config"
	cp "gosyncit/lib/copy"
)

var (
	funcMirrorbasic cp.CopyFunc = cp.MirrorBasic
	funcMirror      cp.CopyFunc = cp.Mirror
	funcSync        cp.CopyFunc = cp.Sync
)

func TestSelect(t *testing.T) {
	c := &config.Config{}
	err := c.Load([]string{"", "-sync=false"})
	if err != nil {
		log.Fatal(err)
	}
	f := cp.SelectCopyFunc("/tmp/nonexisting", c)
	if reflect.ValueOf(f).Pointer() != reflect.ValueOf(funcMirrorbasic).Pointer() {
		t.Log("non-existing dst should yield mirror basic func")
		t.Fail()
	}

	dirA, err := os.MkdirTemp("", "dirA")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dirA)
	f = cp.SelectCopyFunc(dirA, c)
	if reflect.ValueOf(f).Pointer() != reflect.ValueOf(funcMirrorbasic).Pointer() {
		t.Log("existing but empty dst should yield mirror basic func")
		t.Fail()
	}

	file := filepath.Join(dirA, "tmpfileA")
	if err := os.WriteFile(file, []byte("content"), 0655); err != nil {
		log.Fatal(err)
	}
	f = cp.SelectCopyFunc(dirA, c)
	if reflect.ValueOf(f).Pointer() != reflect.ValueOf(funcMirror).Pointer() {
		t.Log("existing, not empty dst should yield mirror func")
		t.Fail()
	}

	err = c.Load([]string{"", "-sync=true"})
	if err != nil {
		log.Fatal(err)
	}
	f = cp.SelectCopyFunc(dirA, c)
	if reflect.ValueOf(f).Pointer() != reflect.ValueOf(funcSync).Pointer() {
		t.Log("existing, not empty dst with sync=true should yield sync func")
		t.Fail()
	}
}

func TestCopyFile(t *testing.T) {
	// remove umask so that arbitrary permissions can be specified
	// TODO : does this work on !linux ?!
	oldumask := syscall.Umask(0000)
	defer syscall.Umask(oldumask)
	// could also use os.Chmod after write, to set permissions on src file

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
