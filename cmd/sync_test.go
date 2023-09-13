package cmd_test

import (
	"log"
	"testing"

	"gosyncit/cmd"
)

func TestSync(t *testing.T) {
	log.Println("integration test for sync")
	err := cmd.Sync("A", "B", true)
	if err == nil {
		t.Fail()
		t.Log("sync must fail with invalid src/dst input")
	}
}
