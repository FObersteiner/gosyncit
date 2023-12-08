# gosyncit

A project to explore directory mirror / sync in `go`, alongside a cobra/viper-based CLI. Since `v0.0.13`, this is extended by a platform-independent tool to mirror / sync files via SFTP. This can be useful for instance if there is no `rsync` available on the SFTP server.

## installation

Assumes the [go toolchain](https://go.dev) to be installed:

```sh
go install github.com/FObersteiner/gosyncit@latest
```

## usage

### local storage

#### mirror A &#8594; B

Write files to destination if source is newer or file size doesn't match, only keep files that exist in source (see flags).

```text
Usage:
  gosyncit mirror 'src' 'dst' [flags]

Aliases:
  mirror, mi

Flags:
  -n, --dryrun       show what will be done
  -x, --dirty        do not remove anything from destination that is not found in source
  -s, --skiphidden   skip hidden files
  -v, --verbose      verbose output to the command line
  -h, --help         help for mirror

Global Flags:
      --config string   config file (default is $HOME/.gosyncit.toml)
```

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

```text
Usage:
  gosyncit sync 'src' 'dst' [flags]

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

### local storage to SFTP and vice versa

The direction can either be "local --> remote" or "remote --> local". "local" in this context means local file system, remote means file system of the SFTP server.

```text
Usage:
  gosyncit sftpmirror 'local-path' 'remote-path' 'remote-url' 'username' [flags]

Aliases:
  sftpmirror, smir

Flags:
  -p, --port int     ssh port number (default 22)
  -r, --reverse      reverse mirror: remote to local instead of local to remote
  -n, --dryrun       show what will be done
  -s, --skiphidden   skip hidden files
  -x, --dirty        do not remove anything from dst that is not found in source
  -v, --verbose      verbose output to the command line
  -h, --help         help for sftpmirror

Global Flags:
      --config string   config file (default is $HOME/.gosyncit.toml)
```

## Notes

- Directory tree traversal is always recursive. There is no option to just copy/mirror/sync the top-level directory

### file comparison quirks

- Test for equality is only done by comparing modification timestamp (`mtime`) and size (n bytes). Theoretically, if two files have the same name, `mtime` and size, they will be considered 'identical' although their _content_ could be different. To prevent this incorrect result, a byte-wise comparison ('deep-equal') would be needed if the basic comparison says 'equal'
- timestamp comparison granularity is _microseconds_ at the moment (see `lib/compare/compare.go`, `BasicUnequal`). Nanosecond granularity was causing issues if a file was copied to a remote server. Windows only supports precision down to a period of 100 ns
- `sync`, `mirror`: if two files with unequal size but the same name and path also have the same mtime in source and destination, then the content of the source will take prevalence (i.e. will copied to destination)

### open issues

- see [issues](https://github.com/FObersteiner/gosyncit/issues)

### not planned

- follow symlinks
- inclusion / exclusion lists, regex patterns etc.
