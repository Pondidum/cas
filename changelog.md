# Changelog

## [0.0.8] - 2022-12-20

## Added

- Implement generic environment variable fallback mechanism for flags.
- add `CAS_VERBOSE` flag for `fetch`
- add `--debug` to `fetch`, writes a metadata file to the backend with intermediate hashes, useful for debugging why a hash has changed
- `fetch` will only download an artifact if the local version doesn't exist, or doesn't match the stored version

## Fixed

- `cas artifact` will create the `.archive` file locally when storing a `**/.archive` file, so that subsequent `make` invocations will find the file

## [0.0.7] - 2022-12-18

## Added

- `hash` command which prints the hash that `fetch` uses for restoring
- `example/store-directory` to show how to restore a directory with the `.archive` suffix

## Fixed

- `fetch` restores the marker file for `.archive` usage

## [0.0.6] - 2022-12-14

## Added

- `artifact` command, replaces the `store` command

## Changed

- `fetch` command now does hashing and is designed to work within a `makefile`.  See the added `example/basic-usage` for details.
- Refactored localstorage to also modify file timestamps

## Removed

- `store` command
- `read` command
- `write` command

## [0.0.5] - 2022-12-12

### Added

- add `--debug` flag to `hash` command, so that intermediate hashes can be viewed, helping debug when the hash changes and you think it should not have

### Fixed

- Handle that `dirFS` doesn't cope with file paths starting with `./`
- Update dependencies

## [0.0.4] - 2022-12-10

### Added

- `fetch` and `store` support archiving an entire directory by naming it `.archive` e.g. `cas store .dist/bin/.archive`

## [0.0.3] - 2022-11-26

### Fixed

- update direct dependencies to resolve minor security warning in `github.com/Masterminds/goutils`
- make the build process publish release binaries and notes correctly

## [0.0.2] - 2022-11-08

### Added

- New command `hash` to generate hashes of files and their content

## [0.0.1] - 2022-11-02

### Added

- Support S3 backend
- Read and Write metadata
- Fetch and Store artifacts

## [0.0.0] - 2022-10-27

### Added

- Initial Version
