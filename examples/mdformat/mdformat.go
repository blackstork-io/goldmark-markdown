// A simple markdown formatter that reads markdown from a file and writes it back to another file.
package main

import (
	"flag"
	"io"
	"log"
	"os"

	markdown "github.com/blackstork-io/goldmark-markdown"
	"github.com/davecgh/go-spew/spew"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

func withStdoutCapture(fn func()) ([]byte, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	go func() {
		old := os.Stdout
		defer func() {
			os.Stdout = old
			w.Close()
		}()
		os.Stdout = w
		fn()
	}()
	return io.ReadAll(r)
}

type args struct {
	input  io.Reader
	output io.Writer
	dump   bool
}

func parseArgs() (res args) {
	var err error

	inputFile := flag.String("i", "-", "input file, use - for stdin")
	outputFile := flag.String("o", "-", "output file, use - for stdout")
	dump := flag.Bool("dump", false, "output dump of the parsed markdown")
	flag.Parse()
	res.dump = *dump
	if *inputFile == "-" {
		res.input = os.Stdin
	} else {
		res.input, err = os.Open(*inputFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *outputFile == "-" {
		res.output = os.Stdout
	} else {
		res.output, err = os.Create(*outputFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	return
}

func main() {
	args := parseArgs()

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
			extension.Strikethrough,
			extension.TaskList,
		),
		goldmark.WithRenderer(markdown.NewRenderer()),
	)

	src, err := io.ReadAll(args.input)
	if err != nil {
		log.Fatal(err)
	}

	tree := md.Parser().Parse(text.NewReader(src))
	if args.dump {
		err = os.WriteFile("input_spew.txt", []byte(spew.Sdump(tree)), 0o600)
		if err != nil {
			log.Fatal(err)
		}
		dump, err := withStdoutCapture(func() {
			tree.Dump(src, 0)
		})
		if err != nil {
			log.Fatal(err)
		}
		err = os.WriteFile("input_dump.txt", dump, 0o600)
		if err != nil {
			log.Fatal(err)
		}
	}

	// "Convert" markdown to markdown
	err = md.Renderer().Render(args.output, src, tree)
	if err != nil {
		log.Fatal(err)
	}
}
