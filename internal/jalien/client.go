package jalien

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

const (
	defaultJAlienHost   = "alice-jcentral.cern.ch"
	defaultJAlienPort   = "8097"
	defaultJAlienWSPath = "/websocket/json"
)

// Client provides minimal JAliEn websocket functionality needed by this project.
type Client struct {
	url        string
	httpClient *http.Client
}

// jalienResponse mirrors the generic structure returned by JAliEn websocket API.
type jalienResponse struct {
	Metadata map[string]any   `json:"metadata"`
	Results  []map[string]any `json:"results"`
}

// NewClient creates a JAliEn client using the provided host, port and
// client certificate/key paths.
func NewClient(host, port, certPath, keyPath, certDir string) (*Client, error) {
	if host == "" {
		host = defaultJAlienHost
	}
	if port == "" {
		port = defaultJAlienPort
	}

	clientCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	rootCAs, err := loadRootCAs(certDir)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      rootCAs,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	url := fmt.Sprintf("wss://%s:%s%s", host, port, defaultJAlienWSPath)

	return &Client{
		url:        url,
		httpClient: &http.Client{Transport: transport},
	}, nil
}

// loadRootCAs attempts to build a CertPool based on well-known grid CA
// locations. If none are available, the system cert pool is used.
func loadRootCAs(certDir string) (*x509.CertPool, error) {
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
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

// send issues a single JAliEn command and returns the decoded response.
func (c *Client) send(ctx context.Context, cmd string, options []string) (*jalienResponse, error) {
	if c == nil {
		return nil, errors.New("jalien: client is nil")
	}

	payload := map[string]any{
		"command": cmd,
		"options": options,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	dialCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(dialCtx, c.url, &websocket.DialOptions{
		HTTPClient: c.httpClient,
	})
	if err != nil {
		return nil, err
	}
	// Many JAliEn commands can return responses larger than the default 32 KiB
	// read limit. Raise the limit to 16 MiB to accommodate typical payloads.
	conn.SetReadLimit(16 * 1024 * 1024)
	defer conn.Close(websocket.StatusNormalClosure, "")

	if err = conn.Write(ctx, websocket.MessageText, data); err != nil {
		return nil, err
	}

	_, respBytes, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	var resp jalienResponse
	if err = json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	exitCode := 0
	if meta := resp.Metadata; meta != nil {
		if raw, ok := meta["exitcode"]; ok {
			switch v := raw.(type) {
			case float64:
				exitCode = int(v)
			case string:
				if n, parseErr := strconv.Atoi(v); parseErr == nil {
					exitCode = n
				}
			}
		}
	}

	if exitCode != 0 {
		errMsg := "unknown JAliEn error"
		if meta := resp.Metadata; meta != nil {
			if raw, ok := meta["error"]; ok {
				if s, ok2 := raw.(string); ok2 && s != "" {
					errMsg = s
				}
			}
		}
		return nil, fmt.Errorf("jalien: command %q failed with exit code %d: %s", cmd, exitCode, errMsg)
	}

	return &resp, nil
}
