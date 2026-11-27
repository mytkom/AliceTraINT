package service

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"go.abhg.dev/goldmark/mermaid"
)

// DocMeta represents a single documentation page in the index.
type DocMeta struct {
	Slug  string
	Title string
}

// DocGroup represents a numbered folder of documentation pages, shown as one sidebar group.
type DocGroup struct {
	Title string
	Pages []DocMeta
}

// DocSection represents a single section (heading) within a doc, used for sidebars/TOCs.
type DocSection struct {
	ID    string
	Title string
	Level int
}

// Doc represents a fully rendered documentation page.
type Doc struct {
	Slug       string
	Title      string
	HTML       string
	Sections   []DocSection
	HasMermaid bool
}

type IDocsService interface {
	ListGroups() ([]DocGroup, error)
	GetDoc(slug string) (*Doc, error)
}

type docsService struct {
	docsDir string
	md      goldmark.Markdown
	docs    map[string]*Doc
	groups  []DocGroup
}

// numberPrefix matches the "<number>-" prefix of group folders and page files.
var numberPrefix = regexp.MustCompile(`^(\d+)-`)

func NewDocsService(docsDir string) IDocsService {
	s := &docsService{
		docsDir: docsDir,
		md:      newDocsMarkdown(),
		docs:    make(map[string]*Doc),
		groups:  []DocGroup{},
	}

	s.loadAllDocs()

	return s
}

func newDocsMarkdown() goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			// Client mode only: never render on the server, even when mmdc is on PATH.
			// NoScript: the page loads the vendored Mermaid itself.
			&mermaid.Extender{RenderMode: mermaid.RenderModeClient, NoScript: true},
			highlighting.NewHighlighting(
				highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
			),
		),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	)
}

// numberedEntry is a directory entry whose name starts with a number prefix.
type numberedEntry struct {
	number int
	name   string
}

// readNumbered returns the entries of dir accepted by keep, ordered by their
// number prefix (as integers) and then by name. Entries with the same number
// are logged as an error.
func readNumbered(dir string, keep func(os.DirEntry) bool, onUnnumbered func(os.DirEntry)) ([]numberedEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var out []numberedEntry
	for _, entry := range entries {
		if !keep(entry) {
			continue
		}

		m := numberPrefix.FindStringSubmatch(entry.Name())
		if m == nil {
			if onUnnumbered != nil {
				onUnnumbered(entry)
			}
			continue
		}

		number, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}

		out = append(out, numberedEntry{number: number, name: entry.Name()})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].number != out[j].number {
			return out[i].number < out[j].number
		}
		return out[i].name < out[j].name
	})

	return out, nil
}

func (s *docsService) loadAllDocs() {
	groupDirs, err := readNumbered(s.docsDir, func(e os.DirEntry) bool { return e.IsDir() }, nil)
	if err != nil {
		// No docs directory yet, or unreadable – leave docs empty (best-effort).
		return
	}

	// File each loaded slug came from, for duplicate-slug errors.
	loadedFrom := make(map[string]string)

	for _, groupDir := range groupDirs {
		groupPath := filepath.Join(s.docsDir, groupDir.name)

		pageFiles, err := readNumbered(groupPath,
			func(e os.DirEntry) bool { return !e.IsDir() && filepath.Ext(e.Name()) == ".md" },
			func(e os.DirEntry) {
				log.Printf("docs: %s has no number prefix and is not loaded", filepath.Join(groupPath, e.Name()))
			},
		)
		if err != nil {
			log.Printf("docs: cannot read group %s: %v", groupPath, err)
			continue
		}

		group := DocGroup{Title: humanize(numberPrefix.ReplaceAllString(groupDir.name, ""))}

		for i, pageFile := range pageFiles {
			path := filepath.Join(groupPath, pageFile.name)

			if i > 0 && pageFiles[i-1].number == pageFile.number {
				log.Printf("docs: error: %s and %s have the same number; ordered by file name",
					filepath.Join(groupPath, pageFiles[i-1].name), path)
			}

			slug := strings.TrimSuffix(numberPrefix.ReplaceAllString(pageFile.name, ""), ".md")
			slug = strings.TrimSpace(slug)
			if slug == "" {
				continue
			}

			if existing, ok := loadedFrom[slug]; ok {
				log.Printf("docs: error: %s has the same slug %q as %s and is not loaded", path, slug, existing)
				continue
			}

			doc, err := s.renderDoc(path, slug)
			if err != nil {
				log.Printf("docs: cannot render %s: %v", path, err)
				continue
			}

			s.docs[slug] = doc
			loadedFrom[slug] = path
			group.Pages = append(group.Pages, DocMeta{Slug: slug, Title: doc.Title})
		}

		if len(group.Pages) > 0 {
			s.groups = append(s.groups, group)
		}
	}
}

// prefixIDs prefixes goldmark's heading IDs with "sec-", so they never clash with
// ids in the page layout. It wraps goldmark's own generator, which keeps the
// duplicate-suffix state, so a new one is needed per document.
type prefixIDs struct{ inner parser.IDs }

func (p *prefixIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	return append([]byte("sec-"), p.inner.Generate(value, kind)...)
}

func (p *prefixIDs) Put(value []byte) { p.inner.Put(value) }

func (s *docsService) renderDoc(path, slug string) (*Doc, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ctx := parser.NewContext(parser.WithIDs(&prefixIDs{inner: parser.NewContext().IDs()}))
	root := s.md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))

	doc := &Doc{Slug: slug}
	rawHTML := false

	err = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Heading:
			title := plainText(n, src)
			if n.Level == 1 {
				if doc.Title == "" {
					doc.Title = title
				}
			} else {
				id, _ := n.AttributeString("id")
				idBytes, _ := id.([]byte)
				doc.Sections = append(doc.Sections, DocSection{
					ID:    string(idBytes),
					Title: title,
					Level: n.Level,
				})
			}
		case *ast.HTMLBlock:
			// HTML comments are allowed and dropped silently.
			if n.HTMLBlockType != ast.HTMLBlockType2 {
				rawHTML = true
			}
		case *ast.RawHTML:
			if !bytes.HasPrefix(n.Segments.Value(src), []byte("<!--")) {
				rawHTML = true
			}
		case *mermaid.Block:
			doc.HasMermaid = true
		}

		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}

	if rawHTML {
		log.Printf("docs: warning: %s contains raw HTML, which is not rendered", path)
	}

	if doc.Title == "" {
		doc.Title = humanize(slug)
	}

	var buf bytes.Buffer
	if err := s.md.Renderer().Render(&buf, src, root); err != nil {
		return nil, err
	}
	doc.HTML = buf.String()

	return doc, nil
}

// plainText concatenates the text of all descendants of n, including the text
// inside code spans.
func plainText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch c := c.(type) {
		case *ast.Text:
			b.Write(c.Value(src))
		case *ast.String:
			b.Write(c.Value)
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(b.String())
}

func (s *docsService) ListGroups() ([]DocGroup, error) {
	// Return a copy to avoid accidental external modification.
	out := make([]DocGroup, len(s.groups))
	for i, group := range s.groups {
		out[i] = DocGroup{
			Title: group.Title,
			Pages: append([]DocMeta(nil), group.Pages...),
		}
	}
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

// humanize turns a slug or folder name into a title, e.g. "user-guide" -> "User guide".
func humanize(slug string) string {
	slug = strings.TrimSpace(strings.ReplaceAll(slug, "-", " "))
	if slug == "" {
		return "Documentation"
	}

	return strings.ToUpper(slug[:1]) + slug[1:]
}
