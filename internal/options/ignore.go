package options

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

const OptIgnore renderer.OptionName = "IgnoreNodeKinds"

type IgnoreOpt struct {
	Kinds []ast.NodeKind
}

// Apply implements Option.
func (t IgnoreOpt) Apply(c *Config) {
	c.IgnoredNodes = append(c.IgnoredNodes, t.Kinds...)
}

var _ Option = IgnoreOpt{}
