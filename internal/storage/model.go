package storage

import "go2fa/internal/structure"

const (
	StoreVersion      = 2
	DefaultFolderID   = "fld_default"
	DefaultFolderName = "Default"
)

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Store struct {
	Version int                       `json:"version"`
	Folders []Folder                  `json:"folders"`
	Items   []structure.TwoFactorItem `json:"items"`
}

func newEmptyStore() Store {
	return Store{
		Version: StoreVersion,
		Folders: []Folder{{ID: DefaultFolderID, Name: DefaultFolderName}},
		Items:   []structure.TwoFactorItem{},
	}
}
