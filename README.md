# gosyncit

let's go sync it.

#### copy `A --> B`

Copy everything from source to destination, overwriting existing content.

#### mirror `A --> B`

Only write files if source is newer, only keep files that exist in source.

#### sync `A <--> B`

Ensure source and destination have the same content, keep only the newest version of any file.

## TODO

- add integration tests ? sync at least ...
- usage documentation
- extend commands: SFTP
