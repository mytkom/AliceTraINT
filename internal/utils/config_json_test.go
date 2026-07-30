package utils_test

import (
	"testing"

	"github.com/mytkom/AliceTraINT/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfigJSON(t *testing.T) {
	assert.Equal(t, `{"seed":1}`, string(utils.ConfigJSON(map[string]any{"seed": float64(1)})))
	assert.Equal(t, "null", string(utils.ConfigJSON(nil)))
	assert.Equal(t, "null", string(utils.ConfigJSON("")))
}

func TestFormatConfigPretty(t *testing.T) {
	pretty := utils.FormatConfigPretty(map[string]any{
		"seed": float64(1),
		"model": map[string]any{
			"architecture": "attention",
		},
	})
	assert.Contains(t, pretty, "\n")
	assert.Contains(t, pretty, `"seed": 1`)
	assert.Contains(t, pretty, `"architecture": "attention"`)
	assert.Equal(t, "", utils.FormatConfigPretty(nil))
}

func TestHasConfiguration(t *testing.T) {
	assert.False(t, utils.HasConfiguration(nil))
	assert.False(t, utils.HasConfiguration(""))
	assert.True(t, utils.HasConfiguration(map[string]any{"bs": float64(1)}))
	assert.True(t, utils.HasConfiguration(`{"bs":1}`))
}
