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

func TestAddKey_NormalizesSecretWithSpacesAndLowerCase(t *testing.T) {
	crypto.FS = afero.NewMemMapFs()
	t.Setenv("HOME", "/home/test")
	crypto.CreateDirs()
	crypto.GeneratePublicPrivateKeys()

	inputs := make([]textinput.Model, 3)
	for i := range inputs {
		inputs[i] = textinput.New()
	}
	inputs[0].SetValue("Discord")
	inputs[1].SetValue("main")
	// Discord-style: lowercased, space-separated groups.
	inputs[2].SetValue("5hjm tgku sjdp tkzr p2zd va5c d7yu 3del")

	if ok := AddKey(inputs, ""); !ok {
		t.Fatalf("AddKey should accept space-separated lowercase secret")
	}

	v := crypto.GetDataVault()
	var s storage.Store
	if err := json.Unmarshal([]byte(v.Db), &s); err != nil {
		t.Fatalf("vault JSON invalid: %v", err)
	}
	if len(s.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(s.Items))
	}
	const want = "5HJMTGKUSJDPTKZRP2ZDVA5CD7YU3DEL"
	if s.Items[0].Secret != want {
		t.Fatalf("secret not normalized: got %q want %q", s.Items[0].Secret, want)
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
