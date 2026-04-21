package storage

import (
	"errors"
	"fmt"
	"strings"

	"go2fa/internal/crypto"
	"go2fa/internal/structure"
)

// SaveStore encrypts and writes the store as v2 JSON.
func SaveStore(s Store) error {
	ensureDefaults(&s)
	payload, err := serializeStore(s)
	if err != nil {
		return err
	}
	vault := crypto.GetDataVault()
	vault.Db = payload
	if ok := crypto.SetDataVault(vault); !ok {
		return errors.New("failed to write vault")
	}
	return nil
}

// FindFolderByID returns the folder with the given id (or false if missing).
func FindFolderByID(s Store, id string) (Folder, bool) {
	for _, f := range s.Folders {
		if f.ID == id {
			return f, true
		}
	}
	return Folder{}, false
}

// ItemsInFolder returns the items that belong to folderID. An empty folderID
// means "all items" (the All Keys synthetic scope).
func ItemsInFolder(s Store, folderID string) []structure.TwoFactorItem {
	if folderID == "" {
		out := make([]structure.TwoFactorItem, len(s.Items))
		copy(out, s.Items)
		return out
	}
	out := make([]structure.TwoFactorItem, 0, len(s.Items))
	for _, it := range s.Items {
		if it.FolderID == folderID {
			out = append(out, it)
		}
	}
	return out
}

// CountByFolder returns a map folderID -> number of items. It only counts
// items whose folder currently exists; unknown ids are bucketed into Default
// by ensureDefaults before this runs.
func CountByFolder(s Store) map[string]int {
	counts := make(map[string]int, len(s.Folders))
	for _, f := range s.Folders {
		counts[f.ID] = 0
	}
	for _, it := range s.Items {
		counts[it.FolderID]++
	}
	return counts
}

// NewFolder appends a new folder with a generated id. Name is trimmed;
// empty names and case-insensitive duplicates are rejected.
func NewFolder(s *Store, name string) (Folder, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return Folder{}, errors.New("folder name must not be empty")
	}
	lower := strings.ToLower(name)
	for _, f := range s.Folders {
		if strings.ToLower(f.Name) == lower {
			return Folder{}, fmt.Errorf("folder %q already exists", name)
		}
	}
	f := Folder{ID: NewFolderID(), Name: name}
	s.Folders = append(s.Folders, f)
	return f, nil
}

// RenameFolder changes the display name of a folder. Uniqueness rules from
// NewFolder apply. Renaming Default is allowed (its id stays stable).
func RenameFolder(s *Store, id, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return errors.New("folder name must not be empty")
	}
	lower := strings.ToLower(newName)
	idx := -1
	for i, f := range s.Folders {
		if f.ID == id {
			idx = i
			continue
		}
		if strings.ToLower(f.Name) == lower {
			return fmt.Errorf("folder %q already exists", newName)
		}
	}
	if idx == -1 {
		return fmt.Errorf("folder %q not found", id)
	}
	s.Folders[idx].Name = newName
	return nil
}

// DeleteFolder removes a folder and reassigns its items to moveTo
// (or Default when moveTo is empty). Deleting Default is forbidden.
func DeleteFolder(s *Store, id, moveTo string) error {
	if id == DefaultFolderID {
		return errors.New("default folder cannot be deleted")
	}
	idx := -1
	for i, f := range s.Folders {
		if f.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("folder %q not found", id)
	}

	target := moveTo
	if target == "" || target == id {
		target = DefaultFolderID
	}
	if _, ok := FindFolderByID(*s, target); !ok {
		target = DefaultFolderID
	}

	for i := range s.Items {
		if s.Items[i].FolderID == id {
			s.Items[i].FolderID = target
		}
	}
	s.Folders = append(s.Folders[:idx], s.Folders[idx+1:]...)
	return nil
}

// AddItem appends an item, falling back to Default if FolderID is missing or unknown.
func AddItem(s *Store, item structure.TwoFactorItem) {
	if item.FolderID == "" {
		item.FolderID = DefaultFolderID
	} else if _, ok := FindFolderByID(*s, item.FolderID); !ok {
		item.FolderID = DefaultFolderID
	}
	s.Items = append(s.Items, item)
}

// DeleteItem removes the first item for which match returns true.
func DeleteItem(s *Store, match func(structure.TwoFactorItem) bool) bool {
	for i, it := range s.Items {
		if match(it) {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

// MoveItem reassigns item at index to folderID. The target folder must exist.
func MoveItem(s *Store, index int, folderID string) error {
	if index < 0 || index >= len(s.Items) {
		return fmt.Errorf("item index %d out of range", index)
	}
	if _, ok := FindFolderByID(*s, folderID); !ok {
		return fmt.Errorf("folder %q not found", folderID)
	}
	s.Items[index].FolderID = folderID
	return nil
}
