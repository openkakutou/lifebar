# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-21

### Added
- `Round.Win`: each side's win call (`p1.win.*` / `p2.win.*` over `win.*`), as Ikemen GO reads it.
- `Copy Action N` blocks take action N's frames (through up to 8 copies), dropped when they cannot, as Ikemen GO's resolveCopyAction does.

### Changed
- character v0.23.0: a `.air` line that is not a frame is skipped, as Ikemen GO does, instead of dropping its action with a warning.

### Added
- Read a MUGEN or Ikemen GO lifebar (fight.def) the way Ikemen GO reads it: screen size, referenced files, life and power bars, faces, names, timer, win icons, combo counter, round announcements, action messages and their animations, with Ikemen GO's defaults and its tolerance of damaged files. Sections Ikemen GO ignores, and animations that cannot be read, are reported with their line number.
- Work out what a lifebar shows the way Ikemen GO does: where each element lands on the screen (mirroring and sprite offsets included), which animation frame is showing at a given moment, how much of a life or power bar is filled for a given life, the screen scale and offset for any game size, and the fading trailing part of the life bar after a hit, following each lifebar's own settings.
