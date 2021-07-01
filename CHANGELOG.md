# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Switch all legacy `coreos` flags to `flatcar` (drops CoreOS support).

## [1.2.0] - 2021-06-09

### Added

- Add join function for use in ignition templates.

## [1.1.1] - 2020-07-09

### Changed

- Fix killing `dnsmasq` process if it is running.

## [1.1.0] - 2020-06-30

### Added

- Add github workflows.

### Changed

- Switch from `dep` to go modules.
- Use `architect-orb` `0.9.0`.

[Unreleased]: https://github.com/giantswarm/mayu/compare/v1.2.0...HEAD
[1.2.0]: https://github.com/giantswarm/mayu/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/giantswarm/mayu/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/giantswarm/mayu/releases/tag/v1.1.0
