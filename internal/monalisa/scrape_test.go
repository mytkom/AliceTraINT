package monalisa

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHyperloopRunListUnmarshalJSON(t *testing.T) {
	raw := []byte(`[
		{"type":"MC","runlist":"101,102,101","period":"LHC24m"},
		{"type":"DATA","runlist":"102,103","period":"LHC24data"}
	]`)

	var got HyperloopRunList
	if err := got.UnmarshalJSON(raw); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v, want nil", err)
	}

	if len(got.TagToRuns["LHC24m"]) != 2 {
		t.Fatalf("TagToRuns[LHC24m] len = %d, want 2", len(got.TagToRuns["LHC24m"]))
	}
	if len(got.RunToTags[102]) != 2 {
		t.Fatalf("RunToTags[102] len = %d, want 2", len(got.RunToTags[102]))
	}
	if !got.TagToIsMC["LHC24m"] {
		t.Fatalf("TagToIsMC[LHC24m] = false, want true")
	}
	if got.TagToIsMC["LHC24data"] {
		t.Fatalf("TagToIsMC[LHC24data] = true, want false")
	}
}

func TestHyperloopRunListUnmarshalJSONBadRunNumber(t *testing.T) {
	raw := []byte(`[{"type":"MC","runlist":"123,nope","period":"LHC24m"}]`)

	var got HyperloopRunList
	if err := got.UnmarshalJSON(raw); err == nil {
		t.Fatalf("UnmarshalJSON() error = nil, want non-nil")
	}
}

func TestGetRunListCachesResult(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/alihyperloop-data/runlist/list-runlist.jsp" {
			t.Fatalf("path = %q, want runlist endpoint", r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"type":"MC","runlist":"2001,2002","period":"LHC24m"}]`))
	}))
	defer server.Close()

	c := &MonalisaClient{
		baseURL:                   server.URL,
		cert:                      tls.Certificate{},
		runListExpireAfterMinutes: 60,
	}

	first, err := c.GetRunList()
	if err != nil {
		t.Fatalf("GetRunList() first error = %v, want nil", err)
	}
	second, err := c.GetRunList()
	if err != nil {
		t.Fatalf("GetRunList() second error = %v, want nil", err)
	}

	if first != second {
		t.Fatalf("cached pointer mismatch: first=%p second=%p", first, second)
	}
	if requestCount != 1 {
		t.Fatalf("requestCount = %d, want 1", requestCount)
	}
}

func TestGetRunListRefreshesWhenExpired(t *testing.T) {
	requestCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = w.Write([]byte(fmt.Sprintf(`[{"type":"MC","runlist":"%d","period":"LHC24m"}]`, 3000+requestCount)))
	}))
	defer server.Close()

	c := &MonalisaClient{
		baseURL:                   server.URL,
		cert:                      tls.Certificate{},
		runListExpireAfterMinutes: -1,
		runListObtainedAt:         time.Now(),
		runList: &HyperloopRunList{
			RunToTags: map[uint64][]string{1: {"old"}},
			TagToRuns: map[string][]uint64{"old": {1}},
			TagToIsMC: map[string]bool{"old": true},
		},
	}

	first, err := c.GetRunList()
	if err != nil {
		t.Fatalf("GetRunList() first error = %v, want nil", err)
	}
	second, err := c.GetRunList()
	if err != nil {
		t.Fatalf("GetRunList() second error = %v, want nil", err)
	}

	if requestCount != 2 {
		t.Fatalf("requestCount = %d, want 2", requestCount)
	}
	if first == second {
		t.Fatalf("expected refreshed pointer, got same pointer %p", first)
	}
}

func TestGetMCRowParsesHTML(t *testing.T) {
	const html = `
<html><body>
  <a class="link" target="alicemctag">LHC24m</a>
  <a class="link" target="alicemcaprod">LHC24anchor</a>
  <table>
    <tbody>
      <tr class="table_row">
        <td>c0</td><td>c1</td><td>c2</td><td>c3</td><td>c4</td><td> passAOD </td>
      </tr>
    </tbody>
  </table>
</body></html>`

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	c := &MonalisaClient{
		baseURL: server.URL,
		cert:    tls.Certificate{},
	}

	row, err := c.GetMCRow("LHC24m")
	if err != nil {
		t.Fatalf("GetMCRow() error = %v, want nil", err)
	}
	if row.MCTag != "LHC24m" {
		t.Fatalf("MCTag = %q, want %q", row.MCTag, "LHC24m")
	}
	if row.AnchorProdTag != "LHC24anchor" {
		t.Fatalf("AnchorProdTag = %q, want %q", row.AnchorProdTag, "LHC24anchor")
	}
	if row.PassName != "passAOD" {
		t.Fatalf("PassName = %q, want %q", row.PassName, "passAOD")
	}
}

func TestGetMCRowNon200(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	c := &MonalisaClient{
		baseURL: server.URL,
		cert:    tls.Certificate{},
	}

	if _, err := c.GetMCRow("LHC24m"); err == nil {
		t.Fatalf("GetMCRow() error = nil, want non-nil")
	}
}
