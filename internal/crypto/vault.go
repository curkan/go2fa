package crypto

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
    "os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
    "github.com/spf13/afero"
)

type vault struct {
	Iterator int64 `json:"iterator"`
	Db string `json:"db"`
}

func toData() vault {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")
    var vault vault
    jsonData, err := afero.ReadFile(FS, filePath)

	if err != nil {
		fmt.Println(err)
        return vault
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

    backupsDir := filepath.Join(homeDir, ".local", "share", "go2fa", "backups")

    // ensure backups dir exists
    if err := FS.MkdirAll(backupsDir, 0755); err != nil {
        fmt.Println(err)
        return false
    }

    // read source vault file
    data, err := afero.ReadFile(FS, filePath)
    if err != nil {
        fmt.Println(err)
        return false
    }

    // write backup content
    if err := afero.WriteFile(FS, filepath.Join(backupsDir, backupFile), data, 0644); err != nil {
        fmt.Println(err)
        return false
    }

    return true

}

// BackupVault produces a timestamped copy of the current vault.json under
// ~/.local/share/go2fa/backups/. Returns false if the backup failed.
func BackupVault() bool {
	return backupVault()
}

func CreateDirs() {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa")
    err := FS.MkdirAll(filepath.Join(filePath, "stores"), os.ModePerm)

	if err != nil {
		logrus.Fatal(err)
	}

    err = FS.MkdirAll(filepath.Join(filePath, "backups"), os.ModePerm)

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

    afero.WriteFile(FS, filePath, data, 0644)
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

    err := afero.WriteFile(FS, filePath, data, 0644)
	if err != nil {
		fmt.Println(err)
		return false
	}
	
	return true
}
