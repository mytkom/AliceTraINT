package service

import "github.com/stretchr/testify/mock"

// MockHasher is a mock implementation of the Hasher interface.
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
