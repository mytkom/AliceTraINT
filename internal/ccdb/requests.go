package ccdb

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Inspired by retrieveHeaders method of O2 CcdbApi
func doRemoteHeaderCall(url, uniqueAgentID string, timestamp int64, tlsConfig *tls.Config) (map[string]string, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("If-None-Match", fmt.Sprintf("%d", timestamp))
	req.Header.Set("User-Agent", uniqueAgentID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("CCDB: run information not found")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("CCDB: unexpected status %d", resp.StatusCode)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) == 1 {
			headers[key] = values[0]
		} else {
			headers[key] = fmt.Sprintf("%s", values)
		}
	}

	return headers, nil
}

func removeExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

func uploadFile(filename, url string, fileReader io.Reader, val, valEnd uint64, ssl *tls.Config) error {
	uploadPath := fmt.Sprintf("%s/%s/%d/%d", url, removeExtension(filename), val, valEnd)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("blob", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %v", err)
	}

	_, err = io.Copy(part, fileReader)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %v", err)
	}

	request, err := http.NewRequest("POST", uploadPath, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: ssl,
		},
	}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	//nolint:errcheck
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status code %d", resp.StatusCode)
	}

	return nil
}
