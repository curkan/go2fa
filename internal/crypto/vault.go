package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type vault struct {
	Iterator int64 `json:"iterator"`
	Db string `json:"db"`
}

func toData() vault {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")
	jsonFile, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()
	var vault vault

	jsonData, err := io.ReadAll(jsonFile)

	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(jsonData, &vault)

	if err != nil {
		fmt.Println(err)
	}

	return vault
}

func GetDataVault() vault {
	vault := toData()
	db, _ := base64.StdEncoding.DecodeString(vault.Db)
	vault.Db = string(db)

	return vault
}

func SetDataVault(vault vault) bool {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")

	db := base64.StdEncoding.EncodeToString([]byte(vault.Db))
	vault.Db = string(db)
	vault.Iterator = vault.Iterator + 1
	data, _ := json.Marshal(vault)

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}
	
	return true
}
