package deletekey

import (
	"encoding/json"
	"testing"

	"go2fa/internal/crypto"
	"go2fa/internal/storage"
	"go2fa/internal/structure"

	"github.com/spf13/afero"
)

func TestDeleteKey_RemovesAndPersists(t *testing.T) {
	crypto.FS = afero.NewMemMapFs()
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()

	// seed vault in legacy v1 format to also exercise migration
	items := []structure.TwoFactorItem{
		{Title: "GitHub", Desc: "main", Secret: "MFRGGZDFMZTWQ2LK"},
		{Title: "AWS", Desc: "root", Secret: "MZXW6YTBOI======"},
	}
	b, _ := json.Marshal(items)
	v := crypto.GetEmptyVault()
	v.Db = string(b)
	if !crypto.SetDataVault(v) {
		t.Fatal("seed SetDataVault failed")
	}

	ok := DeleteKey(items[0])
	if !ok {
		t.Fatal("DeleteKey returned false")
	}

	got := crypto.GetDataVault()
	var s storage.Store
	if err := json.Unmarshal([]byte(got.Db), &s); err != nil {
		t.Fatalf("persisted json invalid: %v", err)
	}
	if len(s.Items) != 1 || s.Items[0].Title != "AWS" {
		t.Fatalf("unexpected persisted content: %+v", s.Items)
	}
}
