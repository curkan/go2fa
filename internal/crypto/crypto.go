package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

func GetPublicKey() []byte {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "keys", "public.pem")

    publicKeyPEM, err := os.ReadFile(filePath)
    if err != nil {
        panic(err)
    }

	return publicKeyPEM
}

func GetPrivateKey() []byte {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "keys", "private.pem")

    privateKey, err := os.ReadFile(filePath)
    if err != nil {
        panic(err)
    }

	return privateKey
}

func GeneratePublicPrivateKeys() {
	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "keys")
	err := os.MkdirAll(filePath, os.ModePerm)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        panic(err)
    }

    publicKey := &privateKey.PublicKey

    privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
    privateKeyPEM := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PRIVATE KEY",
        Bytes: privateKeyBytes,
    })
    err = os.WriteFile(filepath.Join(filePath, "private.pem"), privateKeyPEM, 0644)
    if err != nil {
        panic(err)
    }

    publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
    if err != nil {
        panic(err)
    }
    publicKeyPEM := pem.EncodeToMemory(&pem.Block{
        Type:  "RSA PUBLIC KEY",
        Bytes: publicKeyBytes,
    })

    err = os.WriteFile(filepath.Join(filePath, "public.pem"), publicKeyPEM, 0644)
    if err != nil {
        panic(err)
    }
}

func Encrypt(publicKeyPEM []byte, encryptData []byte) []byte {
    publicKeyBlock, _ := pem.Decode(publicKeyPEM)
    publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
    if err != nil {
        panic(err)
    }

    ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey.(*rsa.PublicKey), encryptData)
    if err != nil {
        panic(err)
    }

	return ciphertext
}

func Decrypt(privateKeyPEM []byte, cipherData []byte) []byte {
    privateKeyBlock, _ := pem.Decode(privateKeyPEM)
    privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
    if err != nil {
        panic(err)
    }

    plaintext, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, cipherData)
    if err != nil {
        panic(err)
    }

	return plaintext
}
