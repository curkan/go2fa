package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
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

func backupVault() bool {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")

	prefix := fmt.Sprintf("backup_%v_", time.Now().Format("2006-01-02_15-04-05"))

	backupFile := prefix + filepath.Base(filePath) 

	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	backupsDir := filepath.Join(homeDir, ".local", "share", "go2fa", "backups")

	// создаем директорию для бэкапов, если она не существует
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		err := os.MkdirAll(backupsDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	backup, err := os.Create(filepath.Join(backupsDir, backupFile))

	if err != nil {
		log.Fatal(err)
	}
	defer backup.Close()

	_, err = io.Copy(backup, f)
	if err != nil {
		log.Fatal(err)
	}

	return true

}

func CreateDirs() {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa")
	err := os.MkdirAll(filepath.Join(filePath, "stores"), os.ModePerm)

	if err != nil {
		logrus.Fatal(err)
	}

	err = os.MkdirAll(filepath.Join(filePath, "backups"), os.ModePerm)

	if err != nil {
		logrus.Fatal(err)
	}
}

func GetEmptyVault() vault {
	return vault{}
}

func GetDataVault() vault {
	vault := toData()
	db, _ := base64.StdEncoding.DecodeString(vault.Db)
	if len(db) != 0 {
		db = Decrypt(GetPrivateKey(), db)
	}
	vault.Db = string(db)

	return vault
}

func SetEmptyVault(vault vault) bool {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")

	vault.Db = ""
	vault.Iterator = vault.Iterator + 1
	data, _ := json.Marshal(vault)

	os.WriteFile(filePath, data, 0644)
	return true
}

func SetDataVault(vault vault) bool {
	backupVault()
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")

	db := string(Encrypt(GetPublicKey(), []byte(vault.Db)))
	db = base64.StdEncoding.EncodeToString([]byte(db))
	vault.Db = db
	vault.Iterator = vault.Iterator + 1
	data, _ := json.Marshal(vault)

	err := os.WriteFile(filePath, data, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}
	
	return true
}
