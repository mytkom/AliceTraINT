package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// DocMeta represents a single documentation page in the index.
type DocMeta struct {
	Slug  string
	Title string
}

// DocSection represents a single section (heading) within a doc, used for sidebars/TOCs.
type DocSection struct {
	ID    string
	Title string
	Level int
}

// Doc represents a fully rendered documentation page.
type Doc struct {
	Slug     string
	Title    string
	HTML     string
	Sections []DocSection
}

type IDocsService interface {
	ListDocs() ([]DocMeta, error)
	GetDoc(slug string) (*Doc, error)
}

type docsService struct {
	docsDir string
	docs    map[string]*Doc
	ordered []DocMeta
}

func NewDocsService(docsDir string) IDocsService {
	s := &docsService{
		docsDir: docsDir,
		docs:    make(map[string]*Doc),
		ordered: []DocMeta{},
	}

	s.loadAllDocs()

	return s
}

func (s *docsService) loadAllDocs() {
	entries, err := os.ReadDir(s.docsDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No docs directory yet – leave docs empty.
			return
		}
		// On any other error, also leave docs empty (best-effort).
		return
	}

	var docs []DocMeta

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		slug := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		slug = regexp.MustCompile(`^\d+-`).ReplaceAllString(slug, "")
		slug = strings.TrimSpace(slug)
		if slug == "" {
			continue
		}

		content, err := os.ReadFile(filepath.Join(s.docsDir, entry.Name()))
		if err != nil {
			continue
		}

		markdown := string(content)

		title := extractTitle(markdown, slug)

		renderer := blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			HeadingIDPrefix: "sec-",
		})

		htmlBytes := blackfriday.Run(
			content,
			blackfriday.WithRenderer(renderer),
			blackfriday.WithExtensions(
				blackfriday.CommonExtensions|
					blackfriday.AutoHeadingIDs,
			),
		)

		html := string(htmlBytes)
		sections := extractSectionsFromHTML(html)

		s.docs[slug] = &Doc{
			Slug:     slug,
			Title:    title,
			HTML:     html,
			Sections: sections,
		}

		docs = append(docs, DocMeta{
			Slug:  slug,
			Title: title,
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Title < docs[j].Title
	})

	s.ordered = docs
}

func (s *docsService) ListDocs() ([]DocMeta, error) {
	// Return a copy to avoid accidental external modification.
	out := make([]DocMeta, len(s.ordered))
	copy(out, s.ordered)
	return out, nil
}

func (s *docsService) GetDoc(slug string) (*Doc, error) {
	if slug == "" {
		return nil, fmt.Errorf("empty slug")
	}

	doc, ok := s.docs[slug]
	if !ok {
		return nil, os.ErrNotExist
	}

	return doc, nil
}

// extractTitle tries to use the first markdown H1 ("# Title") as the page title.
// If no such heading is found, it falls back to a humanized slug.
func extractTitle(markdown, slug string) string {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}

	// Fallback – humanize the slug, e.g. "getting-started" -> "Getting Started".
	slug = strings.ReplaceAll(slug, "-", " ")
	if slug == "" {
		return "Documentation"
	}

	return strings.ToUpper(slug[:1]) + slug[1:]
}

// extractSectionsFromHTML parses the rendered HTML and returns all H2+ headings
// as sections that can be used for a sidebar/table of contents.
func extractSectionsFromHTML(html string) []DocSection {
	// Matches: <h2 id="sec-foo">Title</h2>, capturing level, id and title.
	re := regexp.MustCompile(`<h([1-6]) id="([^"]+)">([^<]+)</h[1-6]>`)

	matches := re.FindAllStringSubmatch(html, -1)
	sections := make([]DocSection, 0, len(matches))

	for _, m := range matches {
		if len(m) != 4 {
			continue
		}

		levelStr := m[1]
		id := m[2]
		title := m[3]

		// Skip H1 (usually the page title); keep H2+ as "sections".
		if levelStr == "1" {
			continue
		}

		level := int(levelStr[0] - '0')

		sections = append(sections, DocSection{
			ID:    id,
			Title: title,
			Level: level,
		})
	}

	return sections
}
