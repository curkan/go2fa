package deletekey

import (
	"fmt"
	"go2fa/internal/storage"
	"go2fa/internal/structure"
)

// DeleteKey removes the first item matching the target by (title, desc, secret, folder_id)
// — folder_id is included in the match to prevent cross-folder collateral deletes.
func DeleteKey(target structure.TwoFactorItem) bool {
	store, err := storage.LoadStore()
	if err != nil {
		fmt.Println(err)
		return false
	}

	matcher := func(it structure.TwoFactorItem) bool {
		if it.Title != target.Title || it.Desc != target.Desc || it.Secret != target.Secret {
			return false
		}
		if target.FolderID == "" {
			return true
		}
		return it.FolderID == target.FolderID
	}

	if !storage.DeleteItem(&store, matcher) {
		return false
	}

	if err := storage.SaveStore(store); err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
