package service_test

import (
	"bytes"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/mytkom/AliceTraINT/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeDocs creates files under a temporary docs directory; an empty content creates a directory.
func writeDocs(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, name)
		if content == "" {
			require.NoError(t, os.MkdirAll(path, 0o755))
			continue
		}
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
	return dir
}

// loadDocs loads the docs directory and returns the service and everything it logged.
func loadDocs(t *testing.T, dir string) (service.IDocsService, string) {
	t.Helper()
	var logs bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(prev)

	return service.NewDocsService(dir), logs.String()
}

func listGroups(t *testing.T, svc service.IDocsService) []service.DocGroup {
	t.Helper()
	groups, err := svc.ListGroups()
	require.NoError(t, err)
	return groups
}

func getDoc(t *testing.T, svc service.IDocsService, slug string) *service.Doc {
	t.Helper()
	doc, err := svc.GetDoc(slug)
	require.NoError(t, err)
	return doc
}

func TestDocsService_GroupAndPageOrder(t *testing.T) {
	dir := writeDocs(t, map[string]string{
		"2-b/1-b-one.md":       "# B one\n",
		"10-c/1-c-one.md":      "# C one\n",
		"1-a-pages/10-ten.md":  "# Ten\n",
		"1-a-pages/2-two.md":   "# Two\n",
		"1-a-pages/1-one.md":   "# One\n",
		"1-a-pages/3-no-h1.md": "no heading\n",
	})

	svc, logs := loadDocs(t, dir)
	assert.Empty(t, logs)

	groups := listGroups(t, svc)
	require.Len(t, groups, 3)
	assert.Equal(t, "A pages", groups[0].Title)
	assert.Equal(t, "B", groups[1].Title)
	assert.Equal(t, "C", groups[2].Title)

	assert.Equal(t, []service.DocMeta{
		{Slug: "one", Title: "One"},
		{Slug: "two", Title: "Two"},
		{Slug: "no-h1", Title: "No h1"},
		{Slug: "ten", Title: "Ten"},
	}, groups[0].Pages)
}

func TestDocsService_IgnoredFoldersAndFiles(t *testing.T) {
	dir := writeDocs(t, map[string]string{
		"adr/0001-decision.md":                "# Decision\n",
		"x.md":                                "# Top level\n",
		"1-guide/1-page.md":                   "# Page\n",
		"1-guide/1-page/image.png":            "png",
		"1-guide/1-page/2-nested.md":          "# Nested\n",
		"1-guide/notes.md":                    "# Notes\n",
		"2-empty":                             "",
		"3-images-only/1-page/screenshot.png": "png",
	})

	svc, logs := loadDocs(t, dir)

	groups := listGroups(t, svc)
	require.Len(t, groups, 1)
	assert.Equal(t, []service.DocMeta{{Slug: "page", Title: "Page"}}, groups[0].Pages)

	for _, slug := range []string{"decision", "0001-decision", "x", "nested", "notes"} {
		_, err := svc.GetDoc(slug)
		assert.ErrorIs(t, err, os.ErrNotExist, slug)
	}

	assert.Contains(t, logs, filepath.Join("1-guide", "notes.md"))
}

func TestDocsService_DuplicateSlugsAndNumbers(t *testing.T) {
	dir := writeDocs(t, map[string]string{
		"1-a/1-same.md":  "# First\n",
		"2-b/1-same.md":  "# Second\n",
		"2-b/2-beta.md":  "# Beta\n",
		"2-b/2-alpha.md": "# Alpha\n",
	})

	svc, logs := loadDocs(t, dir)

	assert.Equal(t, "First", getDoc(t, svc, "same").Title)
	assert.Contains(t, logs, filepath.Join("1-a", "1-same.md"))
	assert.Contains(t, logs, filepath.Join("2-b", "1-same.md"))

	groups := listGroups(t, svc)
	require.Len(t, groups, 2)
	assert.Equal(t, []service.DocMeta{
		{Slug: "alpha", Title: "Alpha"},
		{Slug: "beta", Title: "Beta"},
	}, groups[1].Pages)
	assert.Contains(t, logs, filepath.Join("2-b", "2-alpha.md"))
	assert.Contains(t, logs, filepath.Join("2-b", "2-beta.md"))
}

func TestDocsService_Rendering(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		check    func(t *testing.T, doc *service.Doc, logs string)
	}{
		{
			name:     "sections from headings with code",
			markdown: "# Title with `code`\n\n## Using `make run`\n\n### Tasks & results\n\n#### Deep\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.Equal(t, "Title with code", doc.Title)
				assert.Equal(t, []service.DocSection{
					{ID: "sec-using-make-run", Title: "Using make run", Level: 2},
					{ID: "sec-tasks--results", Title: "Tasks & results", Level: 3},
					{ID: "sec-deep", Title: "Deep", Level: 4},
				}, doc.Sections)
				assert.Contains(t, doc.HTML, `<h2 id="sec-using-make-run">`)
			},
		},
		{
			name:     "duplicate headings",
			markdown: "# T\n\n## Setup\n\n## Setup\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				require.Len(t, doc.Sections, 2)
				assert.Equal(t, "sec-setup", doc.Sections[0].ID)
				assert.Equal(t, "sec-setup-1", doc.Sections[1].ID)
			},
		},
		{
			name:     "mermaid block",
			markdown: "# T\n\n```mermaid\ngraph TD; A-->B;\n```\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.True(t, doc.HasMermaid)
				assert.Contains(t, doc.HTML, `<pre class="mermaid">`)
				assert.NotContains(t, doc.HTML, "<script")
			},
		},
		{
			name:     "no mermaid block",
			markdown: "# T\n\n```text\ngraph TD; A-->B;\n```\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.False(t, doc.HasMermaid)
			},
		},
		{
			name:     "highlighted block",
			markdown: "# T\n\n```go\nfunc main() {}\n```\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.Contains(t, doc.HTML, `class="chroma"`)
			},
		},
		{
			name:     "nested list",
			markdown: "# T\n\n- outer\n  - inner\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.Contains(t, doc.HTML, "<li>outer\n<ul>\n<li>inner</li>")
			},
		},
		{
			name:     "raw HTML",
			markdown: "# T\n\n<div class=\"x\">block</div>\n\nInline <span>tag</span>.\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.Contains(t, logs, "page.md")
				assert.NotContains(t, doc.HTML, "<div")
				assert.NotContains(t, doc.HTML, "<span")
			},
		},
		{
			name:     "HTML comment",
			markdown: "# T\n\n<!-- a comment -->\n\nText <!-- inline --> here.\n",
			check: func(t *testing.T, doc *service.Doc, logs string) {
				assert.Empty(t, logs)
				assert.NotContains(t, doc.HTML, "comment")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := writeDocs(t, map[string]string{"1-g/1-page.md": tt.markdown})
			svc, logs := loadDocs(t, dir)
			tt.check(t, getDoc(t, svc, "page"), logs)
		})
	}
}

// TestDocsLinks checks every internal link, anchor and image in the real docs.
func TestDocsLinks(t *testing.T) {
	const docsDir = "../../docs"

	svc, logs := loadDocs(t, docsDir)
	assert.Empty(t, logs)

	groups := listGroups(t, svc)
	require.NotEmpty(t, groups)

	pages := make(map[string]*goquery.Document)
	for _, group := range groups {
		for _, page := range group.Pages {
			doc := getDoc(t, svc, page.Slug)
			html, err := goquery.NewDocumentFromReader(strings.NewReader(doc.HTML))
			require.NoError(t, err)
			pages[page.Slug] = html
		}
	}

	hasAnchor := func(slug, id string) bool {
		return pages[slug].Find(`[id="`+id+`"]`).Length() > 0
	}

	for slug, html := range pages {
		check := func(ref string) {
			u, err := url.Parse(ref)
			if !assert.NoError(t, err, "%s: %s", slug, ref) || u.Scheme != "" || u.Host != "" {
				return
			}

			switch {
			case strings.HasPrefix(u.Path, "/docs/static/"):
				path := filepath.Join(docsDir, filepath.FromSlash(strings.TrimPrefix(u.Path, "/docs/static/")))
				assert.FileExists(t, path, "%s: %s", slug, ref)
			case strings.HasPrefix(u.Path, "/docs/"):
				target := strings.TrimPrefix(u.Path, "/docs/")
				if !assert.Contains(t, pages, target, "%s: %s names no page", slug, ref) {
					return
				}
				if u.Fragment != "" {
					assert.True(t, hasAnchor(target, u.Fragment), "%s: %s names no heading", slug, ref)
				}
			case u.Path == "" && u.Fragment != "":
				assert.True(t, hasAnchor(slug, u.Fragment), "%s: %s names no heading", slug, ref)
			}
		}

		html.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
			check(s.AttrOr("href", ""))
		})
		html.Find("img[src]").Each(func(_ int, s *goquery.Selection) {
			check(s.AttrOr("src", ""))
		})
	}
}
