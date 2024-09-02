package markdown

import (
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"

	"github.com/blackstork-io/goldmark-markdown/internal/options"
)

var (
	IncorrectOptionError   = options.ErrIncorrectOption
	UnsupportedOptionError = options.ErrUnsupportedOption
)

// WithThematicBreaks sets the thematic break tags to use, in the order of preference
func WithThematicBreaks(breaks ...string) renderer.Option {
	return renderer.WithOption(
		options.OptThematicBreaks,
		options.ThematicBreaksOpt(breaks),
	)
}

// WithParser sets the markdown parser used to verify other options's validity/applicability
func WithParser(p parser.Parser) renderer.Option {
	return renderer.WithOption(options.OptParser, options.ParserOpt{
		Parser: p,
	})
}

// TODO: Create more options
