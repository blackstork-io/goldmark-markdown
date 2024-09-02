package mdrenderer

import (
	"bytes"
	"fmt"
	"unicode/utf8"

	"github.com/blackstork-io/goldmark-markdown/internal/options"
	"github.com/yuin/goldmark/ast"
)

var setextChars = [3][]byte{nil, {'='}, {'-'}}

func (r *Renderer) renderHeading(n ast.Node, entering bool) error {
	if entering {
		r.renderBaseBlock(n, entering)
		r.appendData(nil, nil) // reserve space for the ATX heading and space
		r.pushPos(len(r.bufs))
		return nil
	} else {
		n := n.(*ast.Heading)
		if n.Level < 1 || n.Level > 6 {
			return fmt.Errorf("invalid heading level %d", n.Level)
		}
		start := r.popPos()
		lines := 1
		curLineLen := 0
		maxLineLen := 0
		for _, sub := range r.bufs[start:] {
			curLineLen += utf8.RuneCount(sub.data)
			if sub.endOfLine {
				maxLineLen = max(maxLineLen, curLineLen)
				curLineLen = 0
				lines++
				continue
			}
		}
		maxLineLen = max(maxLineLen, curLineLen)

		var isSetext bool
		if maxLineLen == 0 || n.Level > 2 {
			// ATX
			isSetext = false
		} else if lines > 1 {
			// Setext
			isSetext = true
		} else {
			// eiter ATX or Setext
			isSetext = r.config.Headings.PreferSetext[n.Level]
		}

		if isSetext {
			// Setext
			r.endLine()
			switch r.config.Headings.Setext[n.Level].Style {
			case options.SetextStyleNone:
				r.appendData(r.config.Headings.Setext[n.Level].Underline)
			case options.SetextStyleLongestLine:
				r.appendData(bytes.Repeat(setextChars[n.Level], max(1, maxLineLen)))
			case options.SetextStyleLastLine:
				r.appendData(bytes.Repeat(setextChars[n.Level], max(1, curLineLen)))
			}
		} else {
			// ATX
			r.bufs[start-2].data = r.config.Headings.Atx[n.Level].Open
			if maxLineLen > 0 {
				r.bufs[start-1].data = space
			}
			closer := r.config.Headings.Atx[n.Level].Close
			if len(closer) > 0 {
				r.appendData(closer)
			}
		}
		return r.renderBaseBlock(n, entering)
	}
}
