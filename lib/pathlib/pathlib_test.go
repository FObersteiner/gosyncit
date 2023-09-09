package pathlib_test

import (
	"os"
	"testing"

	"gosyncit/lib/pathlib"
)

func TestCheckPath(t *testing.T) {
	tmp, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	cleaned, err := pathlib.CheckDirPath(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if tmp != cleaned {
		t.Fatal("auto-generated directory path must be clean")
	}

	f, err := os.CreateTemp("", "tempfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, err = pathlib.CheckDirPath(f.Name())
	if err == nil {
		t.Fatal("path to a file must give an error; only paths allowed")
	}
}

func TestSrcDstCheck(t *testing.T) {
	_, _, err := pathlib.CheckSrcDst("nonexisting", "")
	if err == nil {
		t.Fatalf("expected non-existing dir error, got %v", err)
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = pathlib.CheckSrcDst(tmp, "")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp("", "tempfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _, err = pathlib.CheckSrcDst(tmp, f.Name())
	if err == nil {
		t.Fatal("path to a file must give an error; only paths allowed")
	}
}
