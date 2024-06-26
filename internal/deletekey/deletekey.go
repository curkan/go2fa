package deletekey

import (
	"encoding/json"
	"go2fa/internal/crypto"
	"go2fa/internal/structure"
)

func DeleteKey(itemKeysList *[]structure.TwoFactorItem, TwoFactorItem structure.TwoFactorItem ) bool {
	*itemKeysList = deleteItem(*itemKeysList, TwoFactorItem)

	vault := crypto.GetDataVault()

	data, _ := json.Marshal(itemKeysList)
	vault.Db = string(data)

	crypto.SetDataVault(vault)

	return true
}


func deleteItem(items []structure.TwoFactorItem, itemToDelete structure.TwoFactorItem) []structure.TwoFactorItem {
    for i, item := range items {
        if item.Title == itemToDelete.Title && item.Desc == itemToDelete.Desc && item.Secret == itemToDelete.Secret {
            return append(items[:i], items[i+1:]...)
        }
    }

    return items
}
