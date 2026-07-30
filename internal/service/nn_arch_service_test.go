package service_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNNArchService_LoadProposed(t *testing.T) {
	arch := service.NewNNArchService("../../web/nn_architectures/proposed.json")
	require.NotNil(t, arch)

	groups := arch.GetFieldGroups()
	assert.NotEmpty(t, groups)

	names := make([]string, 0, len(groups))
	for _, group := range groups {
		names = append(names, group.Name)
		assert.NotEmpty(t, group.FullName)
	}
	assert.Contains(t, names, "data")
	assert.Contains(t, names, "model")
	assert.Contains(t, names, "training")
	assert.Contains(t, names, "validation")

	results := arch.GetExpectedResults()
	assert.Contains(t, results.Onnx, "kaon.onnx")
}
