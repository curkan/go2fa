package storage

import (
	"encoding/json"
	"sync"
	"testing"

	"go2fa/internal/crypto"
	"go2fa/internal/structure"

	"github.com/spf13/afero"
)

func setupFS(t *testing.T) {
	t.Helper()
	crypto.FS = afero.NewMemMapFs()
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()
	// Reset the process-wide once so each test gets a clean backup latch.
	v1BackupOnce = sync.Once{}
}

func TestIntegration_LazyMigrationAndFolderFlow(t *testing.T) {
	setupFS(t)

	// Seed a legacy v1 vault (array payload) via the existing crypto API.
	legacy := []structure.TwoFactorItem{
		{Title: "Old", Desc: "pre-folders", Secret: "MFRGGZDFMZTWQ2LK"},
	}
	raw, _ := json.Marshal(legacy)
	v := crypto.GetEmptyVault()
	v.Db = string(raw)
	if !crypto.SetDataVault(v) {
		t.Fatal("seed SetDataVault failed")
	}

	// First read must migrate transparently.
	s, err := LoadStore()
	if err != nil {
		t.Fatalf("LoadStore: %v", err)
	}
	if len(s.Items) != 1 || s.Items[0].FolderID != DefaultFolderID {
		t.Fatalf("migrated item should live in Default, got %+v", s.Items)
	}

	// Create a new folder, add a key, move it to another folder,
	// then delete that folder and expect the key to land in Default.
	work, err := NewFolder(&s, "Work")
	if err != nil {
		t.Fatal(err)
	}
	home, err := NewFolder(&s, "Home")
	if err != nil {
		t.Fatal(err)
	}
	AddItem(&s, structure.TwoFactorItem{Title: "GitHub", Secret: "JBSWY3DPEBLW64TMMQ2HY2LOM4======", FolderID: work.ID})
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	s, err = LoadStore()
	if err != nil {
		t.Fatal(err)
	}
	// Move GitHub from Work -> Home. Locate index by title.
	idx := -1
	for i, it := range s.Items {
		if it.Title == "GitHub" {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatal("GitHub item not persisted")
	}
	if err := MoveItem(&s, idx, home.ID); err != nil {
		t.Fatal(err)
	}
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	// Delete "Home" — GitHub must fall back to Default.
	s, _ = LoadStore()
	if err := DeleteFolder(&s, home.ID, ""); err != nil {
		t.Fatal(err)
	}
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	// Final state check.
	s, _ = LoadStore()
	if _, ok := FindFolderByID(s, home.ID); ok {
		t.Error("Home folder still present after delete")
	}
	for _, it := range s.Items {
		if it.Title == "GitHub" && it.FolderID != DefaultFolderID {
			t.Errorf("GitHub should have landed in Default after folder delete, got %q", it.FolderID)
		}
		if it.Title == "Old" && it.FolderID != DefaultFolderID {
			t.Errorf("Legacy item must stay in Default, got %q", it.FolderID)
		}
	}
	if len(s.Items) != 2 {
		t.Errorf("expected 2 items after the flow, got %d", len(s.Items))
	}
}

func TestIntegration_V2RoundTripPersistsFolders(t *testing.T) {
	setupFS(t)

	s := newEmptyStore()
	f, err := NewFolder(&s, "Personal")
	if err != nil {
		t.Fatal(err)
	}
	AddItem(&s, structure.TwoFactorItem{Title: "Bank", Secret: "AAAA", FolderID: f.ID})
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	// Reload from disk.
	reloaded, err := LoadStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := FindFolderByID(reloaded, f.ID); !ok {
		t.Fatal("custom folder lost on round-trip")
	}
	if reloaded.Items[0].FolderID != f.ID {
		t.Errorf("folder_id not persisted, got %q", reloaded.Items[0].FolderID)
	}
}

func TestIntegration_RenamePreservesItems(t *testing.T) {
	setupFS(t)

	s := newEmptyStore()
	f, _ := NewFolder(&s, "Work")
	AddItem(&s, structure.TwoFactorItem{Title: "X", Secret: "AA", FolderID: f.ID})
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	s, _ = LoadStore()
	if err := RenameFolder(&s, f.ID, "Office"); err != nil {
		t.Fatal(err)
	}
	if err := SaveStore(s); err != nil {
		t.Fatal(err)
	}

	reloaded, _ := LoadStore()
	got, ok := FindFolderByID(reloaded, f.ID)
	if !ok || got.Name != "Office" {
		t.Fatalf("rename didn't persist: %+v ok=%v", got, ok)
	}
	if reloaded.Items[0].FolderID != f.ID {
		t.Errorf("items lost their folder association after rename: %q", reloaded.Items[0].FolderID)
	}
}
