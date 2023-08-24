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
	cleaned, err := pathlib.CheckPath(tmp)
	if err != nil {
		log.Fatal(err)
	}
	if tmp != cleaned {
		t.Log("auto-generated directory path must be clean")
		t.Fail()
	}
}
