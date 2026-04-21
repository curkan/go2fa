# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-04-21

### Added
- Folders: group TOTP keys into folders with create, rename, delete, and move operations.
- Folders landing screen as the new main screen, including a synthetic "All keys" view.
- Inline edit of a key's title and description (`e`).
- Reorder keys inside a folder with `shift+↑/↓` and `J/K` (counter-strike mode).
- Quick-pick by digit (`0`–`9`) on both folders and keys screens.
- Preselect current folder when adding a key from a scoped list.
- Lazy, non-destructive v1 → v2 vault migration with an extra one-shot backup on first upgrade.

### Fixed
- RSA keypair is never regenerated if one already exists, so existing backups stay decryptable.
- Secrets are stripped of whitespace and upper-cased before base32 validation (Discord-style `xxxx xxxx xxxx` now accepted).
- Delete/edit/move match items by the full tuple (title, desc, secret, folder_id) to prevent cross-folder collateral changes.

### Changed
- Vault on-disk format bumped to v2 (`{version, folders, items}`); v1 array payloads are still read transparently.

## [0.1.7] - 2024-07-01

### Changed
- Documented `xclip` as a runtime requirement on Linux for clipboard support.

## [0.1.6] - 2024-06-29

### Fixed
- RSA encrypt/decrypt now processes data in chunks, so long secrets no longer fail to save or load.

## [0.1.5] - 2024-06-28

### Fixed
- Create the required data directories on first run instead of panicking on a missing path.

## [0.1.1] - 2024-06-28

### Added
- GoReleaser pipeline with Homebrew tap for binary distribution.

## [0.1.0] - 2024-06-28

### Added
- Initial release.
- TUI with a list of TOTP keys, live code generation, and expiry countdown.
- Add, delete, and copy-to-clipboard for keys.
- RSA-encrypted vault stored under `~/.local/share/go2fa/` with automatic backups on each save.

[0.2.0]: https://github.com/curkan/go2fa/compare/v0.1.7...v0.2.0
[0.1.7]: https://github.com/curkan/go2fa/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/curkan/go2fa/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/curkan/go2fa/compare/v0.1.4...v0.1.5
[0.1.1]: https://github.com/curkan/go2fa/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/curkan/go2fa/releases/tag/v0.1.0
