# gosyncit

The aim of this project is to explore file handling in go, alongside a cobra/viper-based CLI. Anyways, let's go sync it.

#### copy A &#8594; B

Copy everything from source to destination, overwriting existing content.

#### mirror A &#8594; B

Only write files if source is newer, only keep files that exist in source.

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

## TODO

- usage documentation
- extend commands: SFTP
- make use of viper cfg or remove it
