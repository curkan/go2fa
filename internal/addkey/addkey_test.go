package addkey

import (
	"encoding/json"
	"testing"

	"go2fa/internal/crypto"

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

	ok := AddKey(inputs)
	if !ok {
		t.Fatalf("AddKey returned false")
	}

	v := crypto.GetDataVault()
	var arr []map[string]string
	if err := json.Unmarshal([]byte(v.Db), &arr); err != nil {
		t.Fatalf("vault JSON invalid: %v", err)
	}
	if len(arr) != 1 || arr[0]["title"] != "GitHub" {
		t.Fatalf("unexpected vault content: %+v", arr)
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

	if ok := AddKey(inputs); ok {
		t.Fatalf("AddKey should fail on invalid secret")
	}
}
