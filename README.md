# gosyncit

The aim of this project is to explore file handling in go, alongside a cobra/viper-based CLI. This could also be extended by a platform-independent tool to sync files via SFTP (e.g. no rsync available on the SFTP server). Anyways, let's go sync it.

## installation

This assumes the [go toolchain](https://go.dev) to be installed:

```sh
go install github.com/FObersteiner/gosyncit@latest
```

## usage

#### mirror A &#8594; B

Write files to destination if source is newer or file size doesn't match, only keep files that exist in source (see flags).

```
Usage:
  gosyncit mirror [flags]

Aliases:
  mirror, mi

Flags:
  -n, --dryrun       show what will be done
  -x, --dirty        do not remove anything from dst that is not found in source
  -s, --skiphidden   skip hidden files
  -v, --verbose      verbose output to the command line
  -h, --help         help for mirror

Global Flags:
      --config string   config file (default is $HOME/.gosyncit.toml)
```

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

```
Usage:
  gosyncit sync [flags]

Aliases:
  sync, sy

Flags:
  -n, --dryrun       show what will be done
  -s, --skiphidden   skip hidden files
  -v, --verbose      verbose output to the command line
  -h, --help         help for sync

Global Flags:
      --config string   config file (default is $HOME/.gosyncit.toml)
```

---

## Notes

Directory tree traversal is always recursive. There is no option to just copy/mirror/sync the top-level directory.

### file comparison quirks

- Test for equality is currently (v0.0.6) only done by comparing modification timestamp (mtime) and size (in bytes). Theoretically, if two files have the same name, mtime and size, they will be considered 'identical' although their _content_ could be different. To prevent this incorrect result, a byte-wise comparison ('deep-equal') would be needed if the basic comparison says 'equal'
- timestamp comparison granularity is _microseconds_ at the moment (see `lib/compare/compare.go`, `BasicUnequal`). Nanosecond granularity was causing issues if a file was copied to a remote server.
- `sync`, `mirror`: if two files with the same name and path also have the same mtime in source and destination, then the content of the source will take prevalence (i.e. will copied to destination)

### not implemented

- follow symlinks
- inclusion / exclusion lists (regex)
- see also the [issues on github](https://github.com/FObersteiner/gosyncit/issues)
