package mdrenderer

import "github.com/yuin/goldmark/ast"

func (r *Renderer) renderHTMLBlock(n ast.Node, entering bool) (err error) {
	if entering {
		err = r.renderBaseBlock(n, entering)
		r.renderSegments(n.Lines(), true)
	} else {
		n := n.(*ast.HTMLBlock)
		if n.HasClosure() {
			r.renderSegment(n.ClosureLine)
			r.endLine()
		}
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) renderRawHTML(n ast.Node, entering bool) error {
	if entering {
		r.renderSegments(n.(*ast.RawHTML).Segments, false)
	}
	return nil
}
