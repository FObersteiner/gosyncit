# gosyncit

let's go sync it.

- copy `A --> B`: copy everything from source to destination, overwriting existing content
- mirror `A --> B`: only write files if source is newer, only keep files that exist in source
- sync `A <--> B`: ensure source and destination have the same content, keep only the newest version of all the files in source and destination

## TODO

- add github action to run tests
- add integration tests ?
