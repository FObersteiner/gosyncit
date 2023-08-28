package main

import (
	"os"
	"path/filepath"

	"log/slog"

	"gosyncit/lib/config"
	"gosyncit/lib/copy"
	"gosyncit/lib/pathlib"
)

func handleErrFatal(err error) {
	if err != nil {
		slog.Error("fatal", err)
		os.Exit(1)
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
	handleErrFatal(err)

	wd, _ := os.Getwd()
	src := filepath.Join(wd, "testdata/dirA/")
	handleErrFatal(err)
	dst := filepath.Join(wd, "testdata/dirB/")
	handleErrFatal(err)
	src, dst, err = pathlib.CheckSrcDst(src, dst)
	handleErrFatal(err)

	c.Log.Info().Msgf("config : %s\n---", c)
	c.Log.Info().Msgf("src:\n%v\ndst:\n%v\n", src, dst)
	f := copy.SelectCopyFunc(dst, c)
	err = f(src, dst, c)
	handleErrFatal(err)
}
