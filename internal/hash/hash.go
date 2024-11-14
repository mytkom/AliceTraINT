package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength  = 16
	memory      = 1 * 1024 * 1024 // 1 GB
	iterations  = 2               // number of iterations
	paralellism = 4               // number of threads
	keyLength   = 32              // size of the derived key
)

func GenerateKey(keyLength uint) (string, error) {
	key := make([]byte, keyLength)
	_, err := rand.Read(key)
	return base64.RawStdEncoding.EncodeToString(key), err
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	return salt, err
}

func HashKey(key string) (string, error) {
	salt, err := generateSalt()
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(key), salt, iterations, memory, paralellism, keyLength)

	// Encode salt and hash to base64 for storage
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("%s:%s", saltBase64, hashBase64), nil
}

func VerifyKey(key, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, ":")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid hash format")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}

	computedHash := argon2.IDKey([]byte(key), salt, iterations, memory, paralellism, uint32(len(hash)))

	return subtle.ConstantTimeCompare(hash, computedHash) == 1, nil
}
