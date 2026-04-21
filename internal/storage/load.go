package storage

import (
	"sync"

	"go2fa/internal/crypto"
)

// v1BackupOnce guards the proactive backup that runs exactly once per process
// when the loader first detects a legacy v1 vault. SetDataVault already writes
// a backup on every save, so this extra snapshot only matters as a safety net
// between the first launch with the new binary and the user's first mutation.
var v1BackupOnce sync.Once

// LoadStore reads the vault, decrypts it and returns a normalized v2 Store.
// A v1 (array) payload is transparently migrated in memory; the file on disk
// stays v1 until the next SaveStore call.
func LoadStore() (Store, error) {
	vault := crypto.GetDataVault()
	s, migrated, err := normalizeStore(vault.Db)
	if err != nil {
		return Store{}, err
	}
	if migrated {
		v1BackupOnce.Do(func() { _ = crypto.BackupVault() })
	}
	return s, nil
}
