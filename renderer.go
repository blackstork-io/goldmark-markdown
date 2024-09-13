package markdown

import (
	"github.com/blackstork-io/goldmark-markdown/internal/mdrenderer"
	"github.com/blackstork-io/goldmark-markdown/internal/options"
)

// NewRenderer returns a new renderer. Use [goldmark.WithExtensions] to add it to a goldmark instance.
func NewRenderer(opts ...options.Option) *mdrenderer.Renderer {
	return mdrenderer.NewRenderer(opts...)
}
