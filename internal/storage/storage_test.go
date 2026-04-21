package storage

import (
	"encoding/json"
	"strings"
	"testing"

	"go2fa/internal/structure"
)

func TestNormalizeStore_Empty(t *testing.T) {
	s, migrated, err := normalizeStore("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if migrated {
		t.Fatal("empty should not be flagged as migrated")
	}
	if s.Version != StoreVersion {
		t.Errorf("version: got %d want %d", s.Version, StoreVersion)
	}
	if len(s.Folders) != 1 || s.Folders[0].ID != DefaultFolderID {
		t.Errorf("default folder missing: %+v", s.Folders)
	}
	if len(s.Items) != 0 {
		t.Errorf("items should be empty: %+v", s.Items)
	}
}

func TestNormalizeStore_V1Array(t *testing.T) {
	v1 := `[{"title":"GitHub","desc":"x","secret":"JBSWY3DPEBLW64TMMQ2HY2LOM4======"}]`

	s, migrated, err := normalizeStore(v1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !migrated {
		t.Fatal("v1 array must be flagged migrated")
	}
	if len(s.Folders) != 1 || s.Folders[0].ID != DefaultFolderID {
		t.Fatalf("expected only Default folder, got %+v", s.Folders)
	}
	if len(s.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(s.Items))
	}
	if s.Items[0].FolderID != DefaultFolderID {
		t.Errorf("v1 item must be assigned to default: got %q", s.Items[0].FolderID)
	}
	if s.Items[0].Title != "GitHub" {
		t.Errorf("title lost in migration: %q", s.Items[0].Title)
	}
}

func TestNormalizeStore_V2Object(t *testing.T) {
	v2 := `{
	  "version": 2,
	  "folders": [
	    {"id":"fld_default","name":"Default"},
	    {"id":"fld_work","name":"Work"}
	  ],
	  "items": [
	    {"title":"Bank","desc":"","secret":"AAA","folder_id":"fld_work"}
	  ]
	}`

	s, migrated, err := normalizeStore(v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if migrated {
		t.Fatal("v2 must not be flagged as migrated")
	}
	if len(s.Folders) != 2 {
		t.Fatalf("folders: %+v", s.Folders)
	}
	if s.Items[0].FolderID != "fld_work" {
		t.Errorf("folder_id lost: %q", s.Items[0].FolderID)
	}
}

func TestNormalizeStore_UnknownFolderFallsBackToDefault(t *testing.T) {
	v2 := `{"version":2,"folders":[{"id":"fld_default","name":"Default"}],"items":[{"title":"x","desc":"","secret":"y","folder_id":"fld_ghost"}]}`
	s, _, err := normalizeStore(v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Items[0].FolderID != DefaultFolderID {
		t.Errorf("unknown folder_id must be rewritten to default, got %q", s.Items[0].FolderID)
	}
}

func TestNormalizeStore_MissingDefaultInjected(t *testing.T) {
	v2 := `{"version":2,"folders":[{"id":"fld_work","name":"Work"}],"items":[]}`
	s, _, err := normalizeStore(v2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := FindFolderByID(s, DefaultFolderID); !ok {
		t.Fatal("Default folder must be auto-inserted")
	}
}

func TestNormalizeStore_Corrupt(t *testing.T) {
	if _, _, err := normalizeStore("garbage"); err == nil {
		t.Fatal("expected error on corrupt input")
	}
}

func TestSerializeStore_RoundTrip(t *testing.T) {
	s := newEmptyStore()
	f, err := NewFolder(&s, "Work")
	if err != nil {
		t.Fatal(err)
	}
	AddItem(&s, structure.TwoFactorItem{Title: "a", Secret: "b", FolderID: f.ID})

	payload, err := serializeStore(s)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(payload, "{") {
		t.Fatalf("v2 serialization must start with '{', got: %q", payload)
	}

	back, migrated, err := normalizeStore(payload)
	if err != nil {
		t.Fatal(err)
	}
	if migrated {
		t.Fatal("round-trip must not be flagged as migrated")
	}
	if len(back.Items) != 1 || back.Items[0].FolderID != f.ID {
		t.Errorf("round-trip lost item folder_id: %+v", back.Items)
	}
}

func TestNewFolder_EmptyRejected(t *testing.T) {
	s := newEmptyStore()
	if _, err := NewFolder(&s, "   "); err == nil {
		t.Fatal("empty name must be rejected")
	}
}

func TestNewFolder_DuplicateCaseInsensitive(t *testing.T) {
	s := newEmptyStore()
	if _, err := NewFolder(&s, "Work"); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFolder(&s, "work"); err == nil {
		t.Fatal("case-insensitive duplicate must be rejected")
	}
}

func TestRenameFolder_DuplicateRejected(t *testing.T) {
	s := newEmptyStore()
	a, _ := NewFolder(&s, "Work")
	_, _ = NewFolder(&s, "Home")
	if err := RenameFolder(&s, a.ID, "Home"); err == nil {
		t.Fatal("rename into duplicate name must be rejected")
	}
}

func TestRenameFolder_DefaultAllowed(t *testing.T) {
	s := newEmptyStore()
	if err := RenameFolder(&s, DefaultFolderID, "Inbox"); err != nil {
		t.Fatalf("renaming default should be allowed: %v", err)
	}
	f, _ := FindFolderByID(s, DefaultFolderID)
	if f.Name != "Inbox" {
		t.Errorf("rename didn't persist: %+v", f)
	}
}

func TestDeleteFolder_DefaultRefused(t *testing.T) {
	s := newEmptyStore()
	if err := DeleteFolder(&s, DefaultFolderID, ""); err == nil {
		t.Fatal("deleting default must be refused")
	}
}

func TestDeleteFolder_ItemsReassigned(t *testing.T) {
	s := newEmptyStore()
	work, _ := NewFolder(&s, "Work")
	AddItem(&s, structure.TwoFactorItem{Title: "a", Secret: "b", FolderID: work.ID})
	AddItem(&s, structure.TwoFactorItem{Title: "c", Secret: "d", FolderID: work.ID})

	if err := DeleteFolder(&s, work.ID, ""); err != nil {
		t.Fatal(err)
	}
	if len(s.Folders) != 1 {
		t.Errorf("folder not removed: %+v", s.Folders)
	}
	for _, it := range s.Items {
		if it.FolderID != DefaultFolderID {
			t.Errorf("item %q kept a stale folder id: %q", it.Title, it.FolderID)
		}
	}
}

func TestMoveItem(t *testing.T) {
	s := newEmptyStore()
	work, _ := NewFolder(&s, "Work")
	AddItem(&s, structure.TwoFactorItem{Title: "a", Secret: "b"})

	if err := MoveItem(&s, 0, work.ID); err != nil {
		t.Fatal(err)
	}
	if s.Items[0].FolderID != work.ID {
		t.Errorf("item not moved, folder_id=%q", s.Items[0].FolderID)
	}
}

func TestItemsInFolder(t *testing.T) {
	s := newEmptyStore()
	work, _ := NewFolder(&s, "Work")
	AddItem(&s, structure.TwoFactorItem{Title: "a", Secret: "b", FolderID: DefaultFolderID})
	AddItem(&s, structure.TwoFactorItem{Title: "c", Secret: "d", FolderID: work.ID})

	if got := len(ItemsInFolder(s, "")); got != 2 {
		t.Errorf("all keys: got %d, want 2", got)
	}
	if got := len(ItemsInFolder(s, work.ID)); got != 1 {
		t.Errorf("scoped: got %d, want 1", got)
	}
}

func TestAddItem_UnknownFolderFallsBack(t *testing.T) {
	s := newEmptyStore()
	AddItem(&s, structure.TwoFactorItem{Title: "a", Secret: "b", FolderID: "fld_ghost"})
	if s.Items[0].FolderID != DefaultFolderID {
		t.Errorf("unknown folder should have fallen back to Default, got %q", s.Items[0].FolderID)
	}
}

func TestV1ItemJSONCompatibility(t *testing.T) {
	// An old-schema JSON without folder_id must still deserialize cleanly.
	raw := `{"title":"A","desc":"D","secret":"S"}`
	var it structure.TwoFactorItem
	if err := json.Unmarshal([]byte(raw), &it); err != nil {
		t.Fatalf("v1 item must remain parseable: %v", err)
	}
	if it.FolderID != "" {
		t.Errorf("absent folder_id should default to empty, got %q", it.FolderID)
	}
}
