[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![CI Status](https://github.com/FObersteiner/gosyncit/tree/master/.github/workflows/go.yml/badge.svg)

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
<!--[[[cog
   import subprocess
   import cog
   text = subprocess.check_output("gosyncit mirror --help", shell=True)
   cog.out("""```text
   >>> gosyncit mirror --help

   """, dedent=True)
   cog.out(text.decode('utf-8'))
   cog.out("```")
]]]-->
<!--[[[end]]]-->

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

<!--[[[cog
   import subprocess
   import cog
   text = subprocess.check_output("gosyncit sync --help", shell=True)
   cog.out("""```text
   >>> gosyncit sync --help

   """, dedent=True)
   cog.out(text.decode('utf-8'))
   cog.out("```")
]]]-->
<!--[[[end]]]-->

### local storage to SFTP and vice versa

The direction can either be "local --> remote" or "remote --> local". "local" in this context means local file system, remote means file system of the SFTP server.

<!--[[[cog
   import subprocess
   import cog
   text = subprocess.check_output("gosyncit sftpmirror --help", shell=True)
   cog.out("""```text
   >>> gosyncit sftpmirror --help

   """, dedent=True)
   cog.out(text.decode('utf-8'))
   cog.out("```")
]]]-->
<!--[[[end]]]-->

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
