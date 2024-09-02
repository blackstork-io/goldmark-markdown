package mdrenderer

import (
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderStrikethrough(n ast.Node, entering bool) (err error) {
	r.appendData(tildas.get(2))
	return nil
}
