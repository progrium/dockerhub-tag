## [Unreleased][unreleased]

### Fixed

- pagination parsing added. The new api returns tags by groups of 10, providing a **next**/previous links.

### Added

- single command: Keep one single **Tag** while leaving all  **Branch** untouched.
- list command: diplays all automated build in a table format

### Removed

### Changed

- create command: doesn’t deletes other tags (use `single` for that behaviour)

## [v0.1.6] - 2015-06-05

First usable version. With mocked browser working old DockerHub design (before 2015 jukly)
