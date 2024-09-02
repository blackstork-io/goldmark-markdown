package markdown

import (
	"github.com/blackstork-io/goldmark-markdown/internal/mdrenderer"
)

// NewRenderer returns a new renderer. Use [goldmark.WithRenderer] to add it to a goldmark instance.
func NewRenderer() *mdrenderer.Renderer {
	return mdrenderer.NewRenderer()
}
