package jalien

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/mytkom/AliceTraINT/internal/gridtls"
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
	timeout    time.Duration
}

// jalienResponse mirrors the generic structure returned by JAliEn websocket API.
type jalienResponse struct {
	Metadata map[string]any   `json:"metadata"`
	Results  []map[string]any `json:"results"`
}

// NewClient creates a JAliEn client using the provided host, port and
// client certificate/key paths.
func NewClient(host, port, certPath, keyPath, certDir string, timeoutSeconds uint) (*Client, error) {
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

	rootCAs, err := gridtls.LoadRootCAs(certDir)
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
		timeout:    time.Duration(timeoutSeconds) * time.Second,
	}, nil
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

	dialCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	conn, _, err := websocket.Dial(dialCtx, c.url, &websocket.DialOptions{
		HTTPClient: c.httpClient,
	})
	if err != nil {
		return nil, err
	}
	// Many JAliEn commands can return responses larger than the default 32 KiB
	// read limit. Raise the limit to 128 MiB to accommodate typical payloads.
	conn.SetReadLimit(128 * 1024 * 1024)
	defer func() {
		if err := conn.Close(websocket.StatusNormalClosure, ""); err != nil {
			log.Printf("error closing websocket: %v\n", err)
		}
	}()

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
