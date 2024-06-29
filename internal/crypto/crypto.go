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

func Encrypt(publicKeyPEM []byte, encryptData []byte) ([]byte) {
    publicKeyBlock, _ := pem.Decode(publicKeyPEM)
    publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
    if err != nil {
        panic(err)
    }

    rsaPublicKey := publicKey.(*rsa.PublicKey)
    chunkSize := rsaPublicKey.Size() - 11 // 11 is the overhead for PKCS#1 v1.5 padding

    var ciphertext []byte
    for len(encryptData) > 0 {
        chunk := encryptData
        if len(chunk) > chunkSize {
            chunk = encryptData[:chunkSize]
            encryptData = encryptData[chunkSize:]
        } else {
            encryptData = nil
        }

        encryptedChunk, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPublicKey, chunk)
        if err != nil {
			panic(err)
        }

        ciphertext = append(ciphertext, encryptedChunk...)
    }

    return ciphertext
}

func Decrypt(privateKeyPEM []byte, cipherData []byte) ([]byte) {
    privateKeyBlock, _ := pem.Decode(privateKeyPEM)
    privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
    if err != nil {
		panic(err)
    }

    rsaPrivateKey := privateKey
    chunkSize := rsaPrivateKey.Size()

    var plaintext []byte
    for len(cipherData) > 0 {
        chunk := cipherData
        if len(chunk) > chunkSize {
            chunk = cipherData[:chunkSize]
            cipherData = cipherData[chunkSize:]
        } else {
            cipherData = nil
        }

        decryptedChunk, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivateKey, chunk)
        if err != nil {
			panic(err)
        }

        plaintext = append(plaintext, decryptedChunk...)
    }

    return plaintext
}
