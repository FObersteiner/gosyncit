# gosyncit

The aim of this project is to explore file handling in go, alongside a cobra/viper-based CLI. This could also be extended by a platform-independent tool to sync files via SFTP (e.g. no rsync available on the SFTP server). Anyways, let's go sync it.

#### copy A &#8594; B

Copy everything from source to destination, overwrite existing content. Optionally clean the destination before.

#### mirror A &#8594; B

Like copy, but only write files if source is newer or file size doesn't match, and only keep files that exist in source.

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

## Notes

Directory tree traversal is always recursive. There is no option to just copy/mirror/sync the top-level directory.

### file comparison quirks

- Test for equality is currently (v0.0.5) only done by comparing modification timestamp (mtime) and size (in bytes). Theoretically, if two files have the same name, mtime and size, they will be considered 'identical' although their _content_ could be different. To prevent this incorrect result, a byte-wise comparison ('deep-equal') would be needed if the basic comparison says 'equal'
- timestamp comparison granularity is _microseconds_ at the moment (see `lib/compare/compare.go`, `BasicUnequal`). Nanosecond granularity was causing issues if a file was copied to a remote server.
- `sync`, `mirror`: if two files with the same name and path also have the same mtime in source and destination, then the content of the source will take prevalence (i.e. copied to destination)

### not implemented

- follow symlinks
- inclusion / exclusion lists (regex)

## usage

TODO

---

## TODO

- add option to skip hidden files
- add 'quiet' option a la rsync
- extend commands: SFTP
- make use of viper cfg or remove it
- sync, mirror: 'deep-equal' comparison option?
