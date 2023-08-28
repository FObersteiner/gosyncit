package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const TimeFormat = "2006-01-02 15:04:05.000"

type Config struct {
	Log *zerolog.Logger
	// config string // path to a config file (toml)
	DryRun          bool // only print to console
	IgnoreHidden    bool // ignore ".*"
	ForceWrite      bool // overwrite files in dst even if newer
	CleanDst        bool // remove files from dst that do not exist in src
	Sync            bool // 2-way mirror, keep only newer files
	KeepPermissions bool // set original permissions for updated files
	DeepCompare     bool // compare files on a byte level if basic comparison (mtime, size) returns true
}

func (c *Config) Load(args []string) error {
	zerolog.TimeFieldFormat = TimeFormat
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: TimeFormat,
	})
	c.Log = &log.Logger

	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		dryRun       = flags.Bool("dryRun", true, "dry run: only tell what you will do but don't actually do it")
		sync         = flags.Bool("sync", false, "synchronize src and dst, keeping only the newer files")
		ignoreHidden = flags.Bool("ignoreHidden", true, "ignore hidden files")
		forceWrite   = flags.Bool("forceWrite", false, "force write from source to destination, regardless of file timestamp etc.")
		cleanDst     = flags.Bool("cleanDst", false, "strict mirror, only files from source must exist in destination")
		deepCompare  = flags.Bool("deepCompare", false, "compare files on a byte level if basic comparison (mtime, size) returns true")
	)

	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	c.DryRun = *dryRun
	c.IgnoreHidden = *ignoreHidden // TODO : add include / exclude filters, --> config file ?!
	c.ForceWrite = *forceWrite
	c.CleanDst = *cleanDst
	c.Sync = *sync
	c.KeepPermissions = true // TODO : this does not work on windows ?!
	c.DeepCompare = *deepCompare

	return nil
}

func (c *Config) String() string {
	repr := fmt.Sprintf("dryRun: %v", c.DryRun)
	repr += fmt.Sprintf(", ignoreHidden: %v", c.IgnoreHidden)
	repr += fmt.Sprintf(", forceWrite: %v", c.ForceWrite)
	repr += fmt.Sprintf(", cleanDst: %v", c.CleanDst)
	repr += fmt.Sprintf(", keepPermissions: %v", c.KeepPermissions)
	repr += fmt.Sprintf(", deepCompare: %v", c.DeepCompare)
	return repr
}
