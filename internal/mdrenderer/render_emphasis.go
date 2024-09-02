package mdrenderer

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var (
	starEmph       = [3][]byte{{}, {'*'}, {'*', '*'}}
	underscoreEmph = [3][]byte{{}, {'_'}, {'_', '_'}}
)

type emphCtx struct {
	level    int
	openIdx  int
	closeIdx int
}

func (r *Renderer) renderEmphasis(n ast.Node, entering bool) error {
	node := n.(*ast.Emphasis)
	r.appendData(nil)
	if entering {
		r.emphStack = append(r.emphStack, emphCtx{
			level:   node.Level,
			openIdx: len(r.bufs) - 1,
		})
		r.emphCloseStack = append(r.emphCloseStack, len(r.emphStack)-1)
	} else {
		closeIdx := pop(&r.emphCloseStack)
		r.emphStack[closeIdx].closeIdx = len(r.bufs) - 1
		if len(r.emphCloseStack) == 0 {
			// find all required stars
			for _, ctx := range r.emphStack {
				prev := ctx.openIdx - 1
				next := ctx.closeIdx + 1
				notAnUnderscore := false
				if prev >= 0 && len(r.bufs[prev].data) != 0 {
					prevChar := r.bufs[prev].data[len(r.bufs[prev].data)-1]
					notAnUnderscore = !(util.IsSpace(prevChar) || util.IsPunct(prevChar))
				} else if next < len(r.bufs) && len(r.bufs[next].data) != 0 {
					nextChar := r.bufs[next].data[0]
					notAnUnderscore = !(util.IsSpace(nextChar) || util.IsPunct(nextChar))
				}
				if notAnUnderscore {
					r.bufs[ctx.openIdx].data = starEmph[ctx.level]
					r.bufs[ctx.closeIdx].data = starEmph[ctx.level]
				}
			}
			opposites := make([]int, len(r.emphStack))
			for i := range opposites {
				opposites[i] = -1
			}

			for i := 1; i < len(r.emphStack); i++ {
				ctx := r.emphStack[i]
				prevCtx := r.emphStack[i-1]
				if prevCtx.openIdx+1 == ctx.openIdx || ctx.closeIdx+1 == prevCtx.closeIdx {
					// tightly nested emphasis
					if ctx.level == 1 {
						// emphasis inside strong
						// or emphasis inside emphasis
						// => need to differentiate characters
						opposites[i] = i - 1
						opposites[i-1] = i
					}
				}
			}
			set := 1
			for set != 0 {
				set = 0
				for i1, i2 := range opposites {
					if i2 == -1 {
						continue
					}
					ctx1 := r.emphStack[i1]
					ctx2 := r.emphStack[i2]
					if r.bufs[ctx1.openIdx].data != nil && r.bufs[ctx2.openIdx].data == nil {
						set++
						if r.bufs[ctx1.openIdx].data[0] == '*' {
							r.bufs[ctx2.openIdx].data = underscoreEmph[ctx2.level]
							r.bufs[ctx2.closeIdx].data = underscoreEmph[ctx2.level]
						} else {
							r.bufs[ctx2.openIdx].data = starEmph[ctx2.level]
							r.bufs[ctx2.closeIdx].data = starEmph[ctx2.level]
						}
					}
				}
			}
			// required emphasis all set
			for i, ctx := range r.emphStack {
				if r.bufs[ctx.openIdx].data == nil {
					r.bufs[ctx.openIdx].data = starEmph[ctx.level]
					r.bufs[ctx.closeIdx].data = starEmph[ctx.level]
				}
				if opposites[i] != -1 {
					ctx2 := r.emphStack[opposites[i]]
					if r.bufs[ctx.openIdx].data[0] == '*' {
						r.bufs[ctx2.openIdx].data = underscoreEmph[ctx2.level]
						r.bufs[ctx2.closeIdx].data = underscoreEmph[ctx2.level]
					} else {
						r.bufs[ctx2.openIdx].data = starEmph[ctx2.level]
						r.bufs[ctx2.closeIdx].data = starEmph[ctx2.level]
					}
				}
			}

			r.emphStack = r.emphStack[:0]
		}
	}

	return nil
}
