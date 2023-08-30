# gosyncit

let's go sync it.

- mirror A --> B, loose and strict
- sync A <--> B

## TODO

- sync (two-way mirror)
- package structure : separate command functions ?
- tests --> integration tests with command functions (mirror, sync)
- follow symlinks ? --> os.Readlink
- add force write to disk option (`os.File.Sync`)
- config file, flags (flags override config file)
- test if cobra / viper can help to standardize flags and config
- SFTP integration...
