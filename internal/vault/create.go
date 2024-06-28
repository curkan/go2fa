package vault

import "go2fa/internal/crypto"

func Create() bool {
	crypto.CreateDirs()
	vault := crypto.GetEmptyVault()
	crypto.GeneratePublicPrivateKeys()
	crypto.SetEmptyVault(vault)

	return true
}
