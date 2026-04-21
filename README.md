<p align="center">
    <img src="docs/logo.png" width="250" alt="go2fa totp manager" />
    <h5 align="center">Store and use your TOTP keys right in the terminal</h5>
</p>

> English | [Русский](README_RU.md)

---

I got tired of constantly using Google Authenticator and switching between my PC and phone to confirm two-factor authentication. So I built this TUI that lets you store, manage, view, and copy 2FA codes in just a couple of keystrokes.

<p align="center">
  <img src="docs/present.gif" alt="animated" />
</p>

---

# Description

Go2FA TOTP is a lightweight terminal application designed for secure storage and management of your Time-Based One-Time Password (TOTP) keys.

The TUI is built on [bubbletea](https://github.com/charmbracelet/bubbletea)

## Features
- **Secure Storage**: The TOTP vault stores your secrets in encrypted form, ensuring the safety of your sensitive information.
- **Quick Access**: Easily copy TOTP codes with a single command, eliminating the need for manual code entry or switching between apps.
- **Filtering**: Organize your TOTP codes using custom names and descriptions, making it easy to find and access the codes you need.
- **Lightweight**: The TOTP vault is a terminal application that requires minimal system resources and has no dependencies, written in Go.

## Installation


### [Homebrew](https://brew.sh) (Linux/MacOS)
> Make sure **Xclip** or **Xsel** is installed. Otherwise, copying to clipboard will not work.

```shell
brew install curkan/public/go2fa
```

### From Source

```shell
go install github.com/curkan/go2fa@latest
```

### Manual Installation

Download the [latest release](https://github.com/curkan/go2fa/releases/latest) and add the binary to your PATH.

Run with the command `go2fa`

### Viewing Keys
On the key viewing screen, you can filter, add, edit, delete, move, reorder, and copy the desired TOTP key.

- `1`…`9` - quick-pick and copy the Nth key in the current view (slot `0` = 10th, like tab/weapon shortcuts)
- `a` - add a new key (the current folder is preselected in the picker)
- `e` - edit the selected key (title/description)
- `d` - trigger deletion (Enter - confirm, Esc - go back)
- `m` - move the selected key to another folder
- `shift+↑` / `shift+↓` or `K` / `J` - reorder the selected key within the current folder (persists in the vault)
- `enter` - copy to clipboard. When copied, the left border becomes thicker.
- `/` - filter by name

Each row is prefixed with its slot label (`[1]`, `[2]`, …, `[0]` for the 10th).

### Folders
Keys can be organized into folders (Work / Home / Personal / …). The folders list is the landing screen; on top you'll see a synthetic **All keys** entry that behaves like the old unscoped list.

Inside the folders screen:

- `0`…`9` - quick-open the folder at slot N (slot `0` is always **All keys**)
- `a` - add a new key (the highlighted folder — or Default on *All keys* — is preselected)
- `n` - create a new folder
- `e` - rename the highlighted folder
- `d` - delete the highlighted folder (its keys are moved to **Default**; the `Default` folder itself cannot be deleted)
- `Enter` - open the folder
- `Esc` / `q` - quit

Each row is prefixed with its slot label (`[0]` for All keys, `[1]`, `[2]`, …).

### Adding Keys
To add a new key, enter the **Name** and **SecretKey**; Description is optional.\
SecretKey must be in base32 format (spaces and lowercase are tolerated and normalized automatically), otherwise an error will be returned.\
The fourth field is a folder picker — use `←` / `→` (or `h` / `l`) to cycle through available folders while it has focus. When opened from a folder or from the folders list, that folder is preselected.

## Vault
A JSON-based vault is used for storing additional information in `vault.json`.\
On the first launch, the application will create a *publicKey* and *privateKey* to encrypt your Vault.

```json
{
  "iterator": 4,
  "db": "CtSRXlMkbXrMmLh/IeMiJCzRbzJkTMagWGVwnvaOkqroDUViVJaBaMbih258o..."
}
```
`db` - an encrypted field that stores the structure of name, description, secretKey\
`iterator` - an additional field that increments with each vault modification. The iterator allows you to quickly identify the previous version and restore it from a backup.

The open JSON format was chosen for convenient application extension. Not all additional fields need to be encrypted.


## File Structure
All used files are stored at: `$HOME/.local/share/go2fa`

```shell
go2fa
├── backups
├── keys
└── stores
```


`backups` - when adding/deleting keys, backups are created with the timestamp of the change. This allows you to restore the desired version. Backup files are encrypted, just like the main `vault.json` file.

`keys` - stores privateKey and publicKey

```shell
└── keys
    ├── private.pem
    └── public.pem
```

`stores` - vaults, currently only vault.json

### Vault format v2 & rollback
Starting from the folders release the decrypted `db` payload is a JSON object with explicit `folders` and `items` arrays (version `2`). Older vaults (a bare JSON array) are migrated **lazily in memory on read** — the file on disk stays v1 until your first mutation (add / delete / move / create-folder / rename / delete-folder). On first load of a v1 vault a one-off backup is produced in `backups/` as an extra safety net.

**Downgrading back to a pre-folders binary?** An old binary cannot parse the v2 object and will error out on start. Recover by replacing `stores/vault.json` with the most recent `backups/backup_*_vault.json` snapshot (those are still v1 if they were written before your first v2 save).

## Testing

- **Run all tests**:
  - `go test ./...`
  - with coverage: `go test ./... -cover`

- **Where to write tests**:
  - Next to the code, in `*_test.go` files within the corresponding packages, for example: `internal/crypto/crypto_test.go`, `internal/addkey/addkey_test.go`, `internal/deletekey/deletekey_test.go`, `internal/twofactor/generate_test.go`.

- **Isolation from the real environment**:
  - Tests use an in-memory filesystem via `afero`. This prevents any changes to the real `$HOME/.local/share/go2fa`.
  - Basic test template:

```go
import (
    "testing"
    "go2fa/internal/crypto"
    "github.com/spf13/afero"
)

func TestSomething(t *testing.T) {
    crypto.FS = afero.NewMemMapFs()   // isolated FS
    t.Setenv("HOME", "/home/test")  // deterministic paths
    crypto.CreateDirs()
    crypto.GeneratePublicPrivateKeys()
    // ... test logic ...
}
```

## TODO:
- Add synchronization with a Git repository
- Add short commands for quickly copying the desired TOTP to clipboard
- Backup restoration screen


## Copyright and License

GO2FA is licensed under the terms of the MIT License. The full license text can be found in the [`LICENSE`](https://github.com/curkan/go2fa/blob/master/LICENSE) file.
