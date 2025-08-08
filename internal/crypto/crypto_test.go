package crypto

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
)

func setupTestFS(t *testing.T) {
	t.Helper()
	FS = afero.NewMemMapFs()
	os.Setenv("HOME", "/home/test")
	CreateDirs()
	GeneratePublicPrivateKeys()
}

func TestGenerateKeysAndRead(t *testing.T) {
	setupTestFS(t)

	homeDir := os.Getenv("HOME")
	keysDir := filepath.Join(homeDir, ".local", "share", "go2fa", "keys")
	if _, err := FS.Stat(filepath.Join(keysDir, "private.pem")); err != nil {
		t.Fatalf("private.pem not created: %v", err)
	}
	if _, err := FS.Stat(filepath.Join(keysDir, "public.pem")); err != nil {
		t.Fatalf("public.pem not created: %v", err)
	}

	priv := GetPrivateKey()
	pub := GetPublicKey()
	if len(priv) == 0 || len(pub) == 0 {
		t.Fatalf("keys should not be empty")
	}
}

func TestVaultSetAndGet(t *testing.T) {
	setupTestFS(t)

	// Start from empty vault
	v := GetEmptyVault()
	if ok := SetEmptyVault(v); !ok {
		t.Fatalf("SetEmptyVault failed")
	}

	// Prepare data
	type item struct{ Title, Desc, Secret string }
	items := []item{{Title: "GitHub", Desc: "main", Secret: "MFRGGZDFMZTWQ2LK"}}
	dataBytes, _ := json.Marshal(items)
	v.Db = string(dataBytes)

	if ok := SetDataVault(v); !ok {
		t.Fatalf("SetDataVault failed")
	}

	// backups dir should contain at least one file
	homeDir := os.Getenv("HOME")
	backupsDir := filepath.Join(homeDir, ".local", "share", "go2fa", "backups")
	entries, err := afero.ReadDir(FS, backupsDir)
	if err != nil || len(entries) == 0 {
		t.Fatalf("backup file not created: %v, entries=%d", err, len(entries))
	}

	// Read and decrypt
	got := GetDataVault()
	if got.Db != string(dataBytes) {
		t.Fatalf("decrypted vault mismatch: got=%s want=%s", got.Db, string(dataBytes))
	}
}
