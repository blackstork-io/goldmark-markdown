package mdrenderer

import (
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) renderTaskCheckBox(n ast.Node, entering bool) (err error) {
	if entering {
		n := n.(*east.TaskCheckBox)
		if n.IsChecked {
			r.appendData(r.config.CheckedCheckbox)
		} else {
			r.appendData(r.config.UncheckedCheckbox)
		}

	}
	return nil
}
