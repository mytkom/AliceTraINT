package gridtls

import (
	"crypto/x509"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadRootCAs builds a CertPool from well-known grid CA locations.
// If none are available, the system cert pool is used.
func LoadRootCAs(certDir string) (*x509.CertPool, error) {
	// 0) If CERT dir is explicitly set, use it
	if certDir != "" {
		if pool, ok := loadCertPoolFromDir(certDir); ok {
			return pool, nil
		}
	}

	homeDir, _ := os.UserHomeDir()

	// Try to populate a local CA bundle similar to `alien.py getCAcerts`.
	if homeDir != "" {
		_ = ensureLocalGridCAs(homeDir) // best-effort; ignore error and fall back to other locations
	}

	// 1) Local getCAcerts default location: ~/.globus/certificates
	if homeDir != "" {
		localDir := filepath.Join(homeDir, ".globus", "certificates")
		if pool, ok := loadCertPoolFromDir(localDir); ok {
			return pool, nil
		}
	}

	// 2) Common grid CA locations (CVMFS and system-wide)
	gridDirs := []string{
		"/cvmfs/alice.cern.ch/etc/grid-security/certificates",
		"/Users/Shared/cvmfs/alice.cern.ch/etc/grid-security/certificates",
		"/etc/grid-security/certificates",
	}
	for _, d := range gridDirs {
		if pool, ok := loadCertPoolFromDir(d); ok {
			return pool, nil
		}
	}

	// 3) Fallback to system roots.
	sysPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	return sysPool, nil
}

func loadCertPoolFromDir(dir string) (*x509.CertPool, bool) {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil, false
	}

	pool := x509.NewCertPool()
	found := false

	err = filepath.WalkDir(dir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".pem") && !strings.HasSuffix(d.Name(), ".crt") {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if pool.AppendCertsFromPEM(data) {
			found = true
		}
		return nil
	})

	if err != nil || !found {
		return nil, false
	}
	return pool, true
}

// ensureLocalGridCAs roughly replicates `alien.py getCAcerts`:
// it clones the ALICE CA bundle from GitHub into ~/.globus/certificates if
// that directory doesn't already contain any PEM/CRT files.
func ensureLocalGridCAs(homeDir string) error {
	baseDir := filepath.Join(homeDir, ".globus")
	certDir := filepath.Join(baseDir, "certificates")

	// If we already have some CA files, do nothing.
	if _, ok := loadCertPoolFromDir(certDir); ok {
		return nil
	}

	// Require git; if it's not available, just return silently.
	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}

	tempDir := filepath.Join(baseDir, "aliencas_temp")

	_ = os.RemoveAll(tempDir)
	_ = os.RemoveAll(certDir)

	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return err
	}

	cloneCmd := exec.Command(
		"git", "clone",
		"--single-branch", "--branch", "master", "--depth=1",
		"https://github.com/alisw/alien-cas.git",
		tempDir,
	)
	cloneCmd.Stdout = io.Discard
	cloneCmd.Stderr = io.Discard
	if err := cloneCmd.Run(); err != nil {
		_ = os.RemoveAll(tempDir)
		return err
	}

	// Copy all files from tempDir into certDir, excluding any .git directory.
	if err := copyTree(tempDir, certDir); err != nil {
		_ = os.RemoveAll(tempDir)
		return err
	}

	_ = os.RemoveAll(tempDir)

	// Best-effort rehash, as in alien.py (not required for Go, but harmless).
	if _, err := exec.LookPath("openssl"); err == nil {
		_ = exec.Command("openssl", "rehash", certDir).Run()
	} else if _, err := exec.LookPath("c_rehash"); err == nil {
		_ = exec.Command("c_rehash", certDir).Run()
	}

	return nil
}

// copyTree recursively copies all files and subdirectories from src to dst.
// Existing files are overwritten.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(src, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}

		target := filepath.Join(dst, rel)

		if d.IsDir() {
			// Skip VCS metadata.
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := in.Close(); err != nil {
			log.Printf("error closing file: %v\n", err)
		}
	}()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("error closing file: %v\n", err)
		}
	}()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
