package markdown

import (
	"github.com/blackstork-io/goldmark-markdown/internal/options"
	"github.com/yuin/goldmark/ast"
)

var (
	IncorrectOptionError   = options.ErrIncorrectOption
	UnsupportedOptionError = options.ErrUnsupportedOption
)

// WithThematicBreaks sets the thematic break tags to use, in the order of preference
func WithThematicBreaks(breaks ...string) options.Option {
	return options.ThematicBreaksOpt(breaks)
}

// WithIgnoredNodes sets up the renderer to ignore a node, proceeding to render its children
func WithIgnoredNodes(kind ...ast.NodeKind) options.Option {
	return options.IgnoreOpt{
		Kinds: kind,
	}
}

// TODO: Create more options
