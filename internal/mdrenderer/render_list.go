package mdrenderer

import (
	"strconv"

	"github.com/yuin/goldmark/ast"
)

type listCtx struct {
	// exclusive to ordered lists
	num       int64
	prefixLen int
	marker    [2]byte
	isOrdered bool
}

func (r *Renderer) renderList(n ast.Node, entering bool) (err error) {
	if entering {
		err = r.renderBaseBlock(n, entering)
		list := n.(*ast.List)
		ctx := listCtx{
			marker:    [2]byte{list.Marker, ' '},
			num:       int64(list.Start),
			isOrdered: list.IsOrdered(),
			prefixLen: 2,
		}
		if list.Marker == '*' && len(r.listStack) >= 2 && r.listStack[len(r.listStack)-2].marker[0] == '*' && r.listStack[len(r.listStack)-1].marker[0] == '*' {
			// avoid * * * in a row if all lists were empty
			ctx.marker[0] = '+'
		}

		r.listStack = append(r.listStack, ctx)
	} else {
		r.listStack = r.listStack[:len(r.listStack)-1]
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) renderListItem(n ast.Node, entering bool) (err error) {
	list := &r.listStack[len(r.listStack)-1]
	if entering {
		err = r.renderBaseBlock(n, entering)
		var pref []byte
		if list.isOrdered {
			pref = strconv.AppendInt(nil, list.num, 10)
			pref = append(pref, list.marker[0], list.marker[1])
			list.num++
			list.prefixLen = len(pref)
		} else {
			pref = list.marker[:]
		}
		lastBuf := r.last()
		if !lastBuf.endOfLine && !lastBuf.isPrefix {
			// append prefix to the same line
			lastBuf.endOfLine = true
		}
		r.appendData(pref)
		r.bufs[len(r.bufs)-1].isPrefix = true
		r.pushPrefix(spaces.get(list.prefixLen))
	} else {
		err = r.popPrefixAndFlush(list.prefixLen)
		if err != nil {
			return
		}
		err = r.renderBaseBlock(n, entering)
	}
	return
}
