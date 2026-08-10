package gridtls

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadCertPoolWithOptionalFile returns the system cert pool with optional
// extra CA material appended. certPath may be a single PEM/DER file or a
// directory of .crt/.pem files. Empty path leaves the system pool unchanged;
// a non-nil error is a soft warning for the caller.
func LoadCertPoolWithOptionalFile(certPath string) (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
		if err != nil {
			err = fmt.Errorf("system cert pool unavailable: %w", err)
		}
	}

	if certPath == "" {
		return pool, err
	}

	if appendErr := appendCertPath(pool, certPath); appendErr != nil {
		if err != nil {
			return pool, fmt.Errorf("%v; also failed to load %s: %w", err, certPath, appendErr)
		}
		return pool, fmt.Errorf("failed to load extra CA %s: %w", certPath, appendErr)
	}

	return pool, nil
}

func appendCertPath(pool *x509.CertPool, certPath string) error {
	info, err := os.Stat(certPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return appendCertFile(pool, certPath)
	}

	entries, err := os.ReadDir(certPath)
	if err != nil {
		return err
	}

	loaded := 0
	var firstErr error
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".crt") && !strings.HasSuffix(lower, ".pem") && !strings.HasSuffix(lower, ".cer") {
			continue
		}
		if fileErr := appendCertFile(pool, filepath.Join(certPath, name)); fileErr != nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s: %w", name, fileErr)
			}
			continue
		}
		loaded++
	}

	if loaded == 0 {
		if firstErr != nil {
			return firstErr
		}
		return fmt.Errorf("no CA certificates found in directory")
	}
	return nil
}

func appendCertFile(pool *x509.CertPool, certPath string) error {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return fmt.Errorf("file is empty")
	}

	if pool.AppendCertsFromPEM(data) {
		return nil
	}

	// CERN Root CA 2 is published as DER (.crt).
	cert, err := x509.ParseCertificate(data)
	if err != nil {
		block, _ := pem.Decode(data)
		if block != nil {
			cert, err = x509.ParseCertificate(block.Bytes)
		}
		if err != nil {
			return fmt.Errorf("not valid PEM or DER certificate: %w", err)
		}
	}

	pool.AddCert(cert)
	return nil
}
