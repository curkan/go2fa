package addkey

import (
	"encoding/base32"
	"fmt"
	"go2fa/internal/storage"
	"go2fa/internal/structure"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

// AddKey inserts a new TOTP item into the vault. folderID selects the target
// folder; an empty string means "use the Default folder".
func AddKey(inputs []textinput.Model, folderID string) bool {
	if _, err := base32.StdEncoding.DecodeString(inputs[2].Value()); err != nil {
		return false
	}

	item := structure.TwoFactorItem{
		Title:    inputs[0].Value(),
		Desc:     inputs[1].Value(),
		Secret:   strings.ToUpper(inputs[2].Value()),
		FolderID: folderID,
	}

	if item.Title == "" || item.Secret == "" {
		return false
	}

	store, err := storage.LoadStore()
	if err != nil {
		fmt.Println(err)
		return false
	}

	storage.AddItem(&store, item)

	if err := storage.SaveStore(store); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
