package noderenderer

import "github.com/yuin/goldmark/ast"

// TODO: allow to set custom renderers for nodes

type RenderFunc func(n ast.Node, entering bool) error

type NodeRenderer struct {
	Kind ast.NodeKind
	Fn   RenderFunc
}

func New(kind ast.NodeKind, fn RenderFunc) NodeRenderer {
	return NodeRenderer{
		Kind: kind,
		Fn:   fn,
	}
}
