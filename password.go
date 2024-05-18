package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSalt() []byte {
	data := make([]byte, sha256.Size)
	n, err := rand.Read(data)
	if err != nil {
		panic(err)
	}
	if n != sha256.Size {
		panic("salt is not correct length")
	}

	return data
}

func HashPassword(pass string, salt string) string {
	h := sha256.New()
	_, err := h.Write([]byte(pass))
	if err != nil {
		panic(err)
	}

	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		panic(err)
	}

	_, err = h.Write(saltBytes)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
