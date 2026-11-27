package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsHandler_IndexRedirectsToFirstPageOfFirstGroup(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"2-second/1-other.md": "# Other\n",
		"1-first/2-later.md":  "# Later\n",
		"1-first/1-start.md":  "# Start\n",
	} {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}

	h := &DocsHandler{Service: service.NewDocsService(dir)}

	rec := httptest.NewRecorder()
	h.Index(rec, httptest.NewRequest(http.MethodGet, "/docs", nil))

	assert.Equal(t, http.StatusTemporaryRedirect, rec.Code)
	assert.Equal(t, "/docs/start", rec.Header().Get("Location"))
}
