package storage

import (
	"crypto/rand"
	"encoding/hex"
)

// NewFolderID generates a stable random id of the form "fld_<16 hex chars>".
func NewFolderID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic(err)
	}
	return "fld_" + hex.EncodeToString(buf[:])
}
