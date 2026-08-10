package gridtls

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAppendCertFile_DER(t *testing.T) {
	path := "/tmp/cern-root-ca-2.crt"
	if _, err := os.Stat(path); err != nil {
		t.Skip("CERN Root CA fixture missing at", path)
	}

	pool, err := LoadCertPoolWithOptionalFile(path)
	if err != nil {
		t.Fatalf("LoadCertPoolWithOptionalFile: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
}

func TestLoadCertPoolWithOptionalFile_Missing(t *testing.T) {
	pool, err := LoadCertPoolWithOptionalFile("/nonexistent/cern-root-ca-2.crt")
	if pool == nil {
		t.Fatal("expected fallback pool")
	}
	if err == nil {
		t.Fatal("expected soft error for missing file")
	}
}

func TestLoadCertPoolWithOptionalFile_Directory(t *testing.T) {
	rootPath := "/tmp/cern-root-ca-2.crt"
	gridPath := "/tmp/cern-grid-ca-1.crt"
	if _, err := os.Stat(rootPath); err != nil {
		t.Skip("CERN Root CA fixture missing")
	}
	if _, err := os.Stat(gridPath); err != nil {
		t.Skip("CERN Grid CA fixture missing")
	}

	dir := t.TempDir()
	mustCopy := func(src, name string) {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustCopy(rootPath, "cern-root-ca-2.crt")
	mustCopy(gridPath, "cern-grid-ca-1.crt")

	pool, err := LoadCertPoolWithOptionalFile(dir)
	if err != nil {
		t.Fatalf("load dir: %v", err)
	}
	if pool == nil {
		t.Fatal("expected non-nil pool")
	}
}
