// Command chroma-css writes the stylesheet for syntax-highlighted code blocks in the docs.
package main

import (
	"flag"
	"log"
	"os"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

const style = "github-dark"

func main() {
	out := flag.String("o", "static/css/chroma.css", "output file")
	flag.Parse()

	f, err := os.Create(*out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := chromahtml.New(chromahtml.WithClasses(true)).WriteCSS(f, styles.Get(style)); err != nil {
		log.Fatal(err)
	}
}
