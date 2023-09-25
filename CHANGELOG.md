# CHANGELOG

## 2023-09-25 (v0.0.9)

- add verbose (print) option to `sync`

## 2023-09-20 (v0.0.8)

- bug fix sync: only compare mtime

## 2023-09-19 (v0.0.7)

- add flag 'skiphidden' for `mirror` to ignore hidden files

## 2023-09-18 (v0.0.6)

- add flag 'skiphidden' for `sync` to ignore hidden files
- file timestamp comparison granularity to microseconds

## 2023-09-17 (v0.0.5)

- add `sync` integration test

## 2023-09-13 (v0.0.4)

- add sync
- update copy function; needs to set mtime

## 2023-09-12

- add mirror

## 2023-09-10

- copy is working

## 2023-09-09 (v0.0.1)

- switch to using cobra/viper

## 2023-08-30

- mirror basically working, integration test missing

## 2023-08-24

- got very basics working, add deep compare function
