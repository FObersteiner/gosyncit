package cmd_test

import (
	"testing"

	"github.com/FObersteiner/gosyncit/cmd"
)

func TestMirror(t *testing.T) {
	dry := false
	clean := false
	ignorehidden := false
	err := cmd.Mirror("A", "B", dry, clean, ignorehidden)
	if err == nil {
		t.Fail()
		t.Log("sync must fail with invalid src/dst input")
	}
}
