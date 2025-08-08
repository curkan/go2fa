package deletekey

import (
	"encoding/json"
	"testing"

	"go2fa/internal/crypto"
	"go2fa/internal/structure"

	"github.com/spf13/afero"
)

func TestDeleteKey_RemovesAndPersists(t *testing.T) {
	crypto.FS = afero.NewMemMapFs()
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()

	// seed vault
	items := []structure.TwoFactorItem{{Title: "GitHub", Desc: "main", Secret: "MFRGGZDFMZTWQ2LK"}, {Title: "AWS", Desc: "root", Secret: "MZXW6YTBOI======"}}
	b, _ := json.Marshal(items)
	v := crypto.GetEmptyVault()
	v.Db = string(b)
	if !crypto.SetDataVault(v) {
		t.Fatal("seed SetDataVault failed")
	}

	// delete GitHub
	list := append([]structure.TwoFactorItem(nil), items...)
	ok := DeleteKey(&list, items[0])
	if !ok {
		t.Fatal("DeleteKey returned false")
	}
	if len(list) != 1 || list[0].Title != "AWS" {
		t.Fatalf("delete not applied in memory: %+v", list)
	}

	// persisted state
	got := crypto.GetDataVault()
	var persisted []structure.TwoFactorItem
	if err := json.Unmarshal([]byte(got.Db), &persisted); err != nil {
		t.Fatalf("persisted json invalid: %v", err)
	}
	if len(persisted) != 1 || persisted[0].Title != "AWS" {
		t.Fatalf("unexpected persisted content: %+v", persisted)
	}
}
