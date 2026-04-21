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
	secret := normalizeSecret(inputs[2].Value())
	if _, err := base32.StdEncoding.DecodeString(secret); err != nil {
		return false
	}

	item := structure.TwoFactorItem{
		Title:    inputs[0].Value(),
		Desc:     inputs[1].Value(),
		Secret:   secret,
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

// normalizeSecret strips whitespace (spaces, tabs, newlines) and uppercases
// the secret so formats like "5hjm tgku sjdp ..." (Discord) are accepted.
func normalizeSecret(s string) string {
	s = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r':
			return -1
		}
		return r
	}, s)
	return strings.ToUpper(s)
}
