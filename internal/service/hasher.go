package service

import (
	"github.com/mytkom/AliceTraINT/internal/hash"
	"github.com/stretchr/testify/mock"
)

type Hasher interface {
	GenerateKey() (string, error)
	HashKey(key string) (string, error)
	VerifyKey(key, encodedHash string) (bool, error)
}

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

type MockHasher struct {
	mock.Mock
}

func (m *MockHasher) GenerateKey() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockHasher) HashKey(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *MockHasher) VerifyKey(key, encodedHash string) (bool, error) {
	args := m.Called(key, encodedHash)
	return args.Bool(0), args.Error(1)
}
