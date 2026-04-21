package addkey

import (
	"encoding/json"
	"testing"

	"go2fa/internal/crypto"
	"go2fa/internal/storage"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/spf13/afero"
)

func TestAddKey_ValidatesAndPersists(t *testing.T) {
	// in-memory FS
	crypto.FS = afero.NewMemMapFs()
	// HOME for deterministic paths
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()

	// build inputs
	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}
	inputs[0].SetValue("GitHub")
	inputs[1].SetValue("main")
	inputs[2].SetValue("MFRGGZDFMZTWQ2LK") // base32

	ok := AddKey(inputs, "")
	if !ok {
		t.Fatalf("AddKey returned false")
	}

	v := crypto.GetDataVault()
	var s storage.Store
	if err := json.Unmarshal([]byte(v.Db), &s); err != nil {
		t.Fatalf("vault JSON invalid: %v", err)
	}
	if len(s.Items) != 1 || s.Items[0].Title != "GitHub" {
		t.Fatalf("unexpected vault content: %+v", s)
	}
	if s.Items[0].FolderID != storage.DefaultFolderID {
		t.Fatalf("new item should fall back to Default, got %q", s.Items[0].FolderID)
	}
}

func TestAddKey_InvalidSecret(t *testing.T) {
	crypto.FS = afero.NewMemMapFs()
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()

	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}
	inputs[0].SetValue("GitHub")
	inputs[1].SetValue("main")
	inputs[2].SetValue("not-base32!!")

	if ok := AddKey(inputs, ""); ok {
		t.Fatalf("AddKey should fail on invalid secret")
	}
}
