package addkey

import (
	"encoding/base32"
	"encoding/json"
	"fmt"
	"go2fa/internal/crypto"
	"go2fa/internal/structure"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

func AddKey(inputs []textinput.Model) bool {
	var twoFactorItems []structure.TwoFactorItem
	var twoFactorItem structure.TwoFactorItem

	_, err := base32.StdEncoding.DecodeString(inputs[2].Value())

	if err != nil {
		return false
	}

	secret := strings.ToUpper(inputs[2].Value())

	twoFactorItem.Title = inputs[0].Value()
	twoFactorItem.Desc = inputs[1].Value()
	twoFactorItem.Secret = secret


	if twoFactorItem.Title == "" {
		return false
	}

	if twoFactorItem.Secret == "" {
		return false
	}

	vault := crypto.GetDataVault()

	err = json.Unmarshal([]byte(vault.Db), &twoFactorItems)
	if err != nil {
		fmt.Println(err)
	}

	twoFactorItems = append(twoFactorItems, twoFactorItem)
	data, _ := json.Marshal(twoFactorItems)
	vault.Db = string(data)

	crypto.SetDataVault(vault)

	return true
}
