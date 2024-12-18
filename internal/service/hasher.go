package service

import "github.com/mytkom/AliceTraINT/internal/hash"

// Hasher defines the interface for key hashing and verification.
type Hasher interface {
	GenerateKey() (string, error)
	HashKey(key string) (string, error)
	VerifyKey(key, encodedHash string) (bool, error)
}

// Argon2Hasher implements the Hasher interface using the hash package.
type Argon2Hasher struct{}

func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{}
}

func (a *Argon2Hasher) GenerateKey() (string, error) {
	return hash.GenerateKey(32)
}

func (a *Argon2Hasher) HashKey(key string) (string, error) {
	return hash.HashKey(key)
}

func (a *Argon2Hasher) VerifyKey(key, encodedHash string) (bool, error) {
	return hash.VerifyKey(key, encodedHash)
}
