package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	r "math/rand/v2"
)

var (
	b64encode = base64.RawURLEncoding.EncodeToString
	b64decode = base64.RawURLEncoding.DecodeString
)

func GenSalt() []byte {
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

// sha2 is unsafe for password hashing by itself, for demo purposes only
func HashPassword(pass string, salt []byte, count uint32) string {
	if count < 1 {
		panic("hash iteration count must be >= 1")
	}

	h := sha256.New()

	_, err := h.Write(append([]byte(pass), salt...))
	if err != nil {
		panic(err)
	}

	for range count {
		h.Write(h.Sum(nil))
	}

	return b64encode(h.Sum(nil))
}

func GenCreds(pass string) (salt string, hash string, count uint32) {
	count = 100 + uint32(r.UintN(256))
	s := GenSalt()
	salt = b64encode(s)
	hash = HashPassword(pass, s, count)

	return salt, hash, count
}
