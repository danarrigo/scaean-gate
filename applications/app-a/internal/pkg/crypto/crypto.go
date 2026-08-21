package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

func GenerateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func GeneratePKCE() (verifier, challenge, state string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)

	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", "", err
	}
	state = hex.EncodeToString(stateBytes)

	return verifier, challenge, state, nil
}

func HashSHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}
