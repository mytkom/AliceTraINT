package hash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash(t *testing.T) {
	key, err := GenerateKey(32)
	assert.NoError(t, err)
	hashed, err := HashKey(key)
	assert.NoError(t, err)
	ok, err := VerifyKey(key, hashed)
	assert.NoError(t, err)
	assert.True(t, ok)
}
