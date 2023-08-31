package config_test

import (
	"testing"

	"gosyncit/lib/config"
)

func TestLoad(t *testing.T) {
	c := &config.Config{}
	err := c.Load([]string{"", "-dryRun=false"})
	if err != nil {
		t.Fatal(err)
	}
	if c.DryRun {
		t.Log("dry run should be false")
		t.Fail()
	}
	if c.Log == nil {
		t.Log("logger must be defined")
		t.Fail()
	}
}
