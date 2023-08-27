package pathlib_test

import (
	"log"
	"os"
	"testing"

	"gosyncit/lib/pathlib"
)

func TestCheckPath(t *testing.T) {
	tmp, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		log.Fatal(err)
	}
	cleaned, err := pathlib.CheckDirPath(tmp)
	if err != nil {
		log.Fatal(err)
	}
	if tmp != cleaned {
		t.Log("auto-generated directory path must be clean")
		t.Fail()
	}

	f, err := os.CreateTemp("", "tempfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, err = pathlib.CheckDirPath(f.Name())
	if err == nil {
		t.Log("path to a file must give an error; only paths allowed")
		t.Fail()
	}
}

func TestSrcDstCheck(t *testing.T) {
	_, _, err := pathlib.CheckSrcDst("nonexisting", "")
	if !os.IsNotExist(err) {
		t.Logf("expected non-existing dir error, got %v", err)
		t.Fail()
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		log.Fatal(err)
	}
	_, _, err = pathlib.CheckSrcDst(tmp, "")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.CreateTemp("", "tempfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _, err = pathlib.CheckSrcDst(tmp, f.Name())
	if err == nil {
		t.Log("path to a file must give an error; only paths allowed")
		t.Fail()
	}
}
