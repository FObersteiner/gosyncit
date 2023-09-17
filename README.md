# gosyncit

The aim of this project is to explore file handling in go, alongside a cobra/viper-based CLI. Anyways, let's go sync it.

#### copy A &#8594; B

Copy everything from source to destination, overwriting existing content.

#### mirror A &#8594; B

Only write files if source is newer, only keep files that exist in source.

#### sync A &#8596; B

Ensure source and destination have the same content, keep only the newest version of any file.

## Notes

### file comparison quirks

- Test for equality is currently (v0.0.5) only done by comparing modification timestamp (mtime) and size (in bytes). Theoretically, if two files have the same name, mtime and size, they will be considered 'identical' although their _content_ could be different. To prevent this incorrect result, a byte-wise comparison ('deep-equal') would be needed if the basic comparison says 'equal'
- `sync`, `mirror`: if two files with the same name and path also have the same mtime in source and destination, then the content of the source will take prevalence (i.e. copied to destination)

## TODO

- usage documentation
- extend commands: SFTP
- make use of viper cfg or remove it
- sync, mirror: 'deep-equal' comparison option?
