# gosyncit

let's go sync it.

- mirror A --> B, loose and strict
- sync A <--> B

## TODO

- package structure
- tests
- multiple comparison methods (timestamp, md5)
- mirror loose vs. strict
- sync (two-way mirror)
- follow symlinks ? --> os.Readlink
- flags
- config file
- consider sftp integration...
- test if cobra / viper can help to standardize flags and config
- add force write to disk option (`os.File.Sync`)
