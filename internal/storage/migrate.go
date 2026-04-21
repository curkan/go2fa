package storage

import (
	"encoding/json"
	"fmt"
	"strings"

	"go2fa/internal/structure"
)

// normalizeStore parses a decrypted db JSON string into a Store, performing
// lazy v1 -> v2 migration. It returns the store, a flag indicating whether
// migration from v1 happened, and an error on corrupt input.
func normalizeStore(dbJSON string) (Store, bool, error) {
	trimmed := strings.TrimLeft(dbJSON, " \t\r\n")

	if trimmed == "" {
		return newEmptyStore(), false, nil
	}

	first := trimmed[0]

	switch first {
	case '[':
		var legacy []structure.TwoFactorItem
		if err := json.Unmarshal([]byte(trimmed), &legacy); err != nil {
			return Store{}, false, fmt.Errorf("parse v1 vault: %w", err)
		}
		for i := range legacy {
			legacy[i].FolderID = DefaultFolderID
		}
		s := newEmptyStore()
		s.Items = legacy
		return s, true, nil

	case '{':
		var s Store
		if err := json.Unmarshal([]byte(trimmed), &s); err != nil {
			return Store{}, false, fmt.Errorf("parse v2 vault: %w", err)
		}
		ensureDefaults(&s)
		return s, false, nil

	default:
		return Store{}, false, fmt.Errorf("corrupt vault: unexpected first byte %q", first)
	}
}

// ensureDefaults guarantees structural invariants on a Store:
//   - version is set to the current one;
//   - a Default folder exists;
//   - every item points at an existing folder (falling back to Default).
func ensureDefaults(s *Store) {
	s.Version = StoreVersion

	if s.Folders == nil {
		s.Folders = []Folder{}
	}

	hasDefault := false
	known := make(map[string]struct{}, len(s.Folders))
	for _, f := range s.Folders {
		known[f.ID] = struct{}{}
		if f.ID == DefaultFolderID {
			hasDefault = true
		}
	}
	if !hasDefault {
		s.Folders = append([]Folder{{ID: DefaultFolderID, Name: DefaultFolderName}}, s.Folders...)
		known[DefaultFolderID] = struct{}{}
	}

	if s.Items == nil {
		s.Items = []structure.TwoFactorItem{}
	}
	for i := range s.Items {
		fid := s.Items[i].FolderID
		if fid == "" {
			s.Items[i].FolderID = DefaultFolderID
			continue
		}
		if _, ok := known[fid]; !ok {
			s.Items[i].FolderID = DefaultFolderID
		}
	}
}

// serializeStore marshals a Store into the v2 JSON form expected inside vault.Db.
func serializeStore(s Store) (string, error) {
	ensureDefaults(&s)
	data, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
