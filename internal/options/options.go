package options

import (
	"errors"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	IncorrectOptionError   = errors.New("incorrect option value")
	UnsupportedOptionError = errors.New("unsupported option")
)

type Config struct {
	Errs             []error
	parser           parser.Parser
	ThematicBreaks   []ThematicBreak
	Headings         Headings
	HardLineBreak    []byte
	CodeBlockIndent  []byte
	BlockquotePrefix []byte
	// TODO: configurable emphasis
	// Emphasis            [3][]byte
	FencedCodeBlockChar byte
	LinkTitleQuote      []byte
	CheckedCheckbox     []byte
	UncheckedCheckbox   []byte
}

func DefaultConfig() Config {
	return Config{
		parser:              goldmark.DefaultParser(),
		ThematicBreaks:      defaultThematicBreaks,
		Headings:            defaultHeadings,
		HardLineBreak:       []byte{'\\'},   // alternative - two+ spaces
		CodeBlockIndent:     []byte("    "), // exactly 4 spaces or 1 tab
		BlockquotePrefix:    []byte("> "),   // alternatives - up to 3 spaces, ">", optional space
		FencedCodeBlockChar: '`',            // alternative - '~'
		LinkTitleQuote:      []byte{'"'},    // alternative - "'"
		CheckedCheckbox:     []byte("[x] "), // alternative - "[X] "
		UncheckedCheckbox:   []byte("[ ] "),
	}
}

func (c *Config) parse(src []byte) ast.Node {
	return c.parser.Parse(text.NewReader(src))
}

// Option for markdown renderer
type Option interface {
	Apply(*Config)
}
