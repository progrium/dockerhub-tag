## [Unreleased][unreleased]

### Fixed

- pagination parsing added. The new api returns tags by groups of 10, providing a **next**/previous links
- automatic build trigger added after successfull add/set

### Added

- `list` command: diplays all automated builds in a table format
- `add` command: simple adds a new git tag based automated build, without deleting any other automated build

### Removed

### Changed

- `create` command is renamed to `set`

## [v0.1.6] - 2015-06-05

First usable version. With mocked browser working old DockerHub design (before 2015 jukly)
