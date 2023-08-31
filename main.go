package main

import (
	"os"

	"gosyncit/lib/config"
	"gosyncit/lib/copy"
	"gosyncit/lib/pathlib"
)

func handleErrFatal(err error, c *config.Config) {
	if err != nil {
		c.Log.Fatal().Err(err).Msg("need to stop.")
	}
}

//-----------------------------------------------------------------------------
//
// TODO :
// func mirror (dst exists and is not empty)
// func sync
//
//-----------------------------------------------------------------------------

func main() {
	c := &config.Config{}
	err := c.Load(os.Args)
	handleErrFatal(err, c)

	// wd, _ := os.Getwd()
	// src := filepath.Join(wd, "testdata/dirA/")
	// handleErrFatal(err)
	// dst := filepath.Join(wd, "testdata/dirB/")
	// handleErrFatal(err)

	src, err := pathlib.CheckDirPath("/home/va6504/Downloads/A/")
	handleErrFatal(err, c)
	dst, err := pathlib.CheckDirPath("/home/va6504/Downloads/B/")
	handleErrFatal(err, c)

	c.Log.Info().Msgf("config : %s\n---", c)
	c.Log.Info().Msgf("src:\n%v\ndst:\n%v\n", src, dst)

	f := copy.SelectCopyFunc(dst, c)
	err = f(src, dst, c)
	handleErrFatal(err, c)
}
