package mdrenderer

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func (r *Renderer) defaultRenderer(n ast.Node, entering bool) error {
	return fmt.Errorf("unsupported node type %q", n.Kind())
}

func (r *Renderer) renderBaseBlock(n ast.Node, entering bool) error {
	if entering {
		if n.HasBlankPreviousLines() {
			for !r.previousLineBlank {
				r.endLine()
			}
		}
		return nil
	} else {
		r.endNonemptyLine()
		return r.flush()
	}
}

func (r *Renderer) renderSegment(segment text.Segment) {
	if segment.IsEmpty() {
		r.appendData(nil)
		return
	}
	if segment.Padding > 0 {
		r.appendData(spaces.get(segment.Padding))
	}
	data := r.source[segment.Start:segment.Stop]

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		if idx == -1 {
			r.appendData(data)
			return
		}
		r.appendData(data[:idx])
		r.endLine()
		data = data[idx+1:]
	}
}

func (r *Renderer) renderSegments(segments *text.Segments, newLines bool) {
	for i := 0; i < segments.Len(); i++ {
		seg := segments.At(i)
		r.renderSegment(seg)
		if newLines {
			r.endNonemptyLine()
		}
	}
}

func (r *Renderer) renderBlockquote(n ast.Node, entering bool) error {
	if entering {
		r.renderBaseBlock(n, entering)
		lastBuf := r.last()
		if !lastBuf.endOfLine && !lastBuf.isPrefix {
			// append prefix to the same line
			r.endLine()
		}
		// preserve empty blockquote by adding nil
		r.appendData(r.config.BlockquotePrefix, nil)
		r.bufs[len(r.bufs)-2].isPrefix = true
		r.pushPrefix(r.config.BlockquotePrefix)
		return nil
	} else {
		r.endNonemptyLine()
		err := r.popPrefixAndFlush(len(r.config.BlockquotePrefix))
		if err != nil {
			return err
		}
		return r.renderBaseBlock(n, entering)
	}
}

func (r *Renderer) renderParagraph(n ast.Node, entering bool) (err error) {
	if entering {
		err = r.renderBaseBlock(n, entering)
		if !n.HasBlankPreviousLines() {
			// enforce paragraph separation, for example in blockquotes
			// HasBlankPreviousLines is false
			prev := n.PreviousSibling()
			for ; prev != nil; prev = prev.LastChild() {
				if prev.Kind() == ast.KindParagraph {
					r.endLine()
					break
				}
			}
		}
	} else {
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) renderThematicBreak(n ast.Node, entering bool) error {
	if entering {
		return r.renderBaseBlock(n, entering)
	} else {
	nextTB:
		for _, tb := range r.config.ThematicBreaks {
			if tb.Fits(n) {
				for _, list := range r.listStack {
					if list.marker[0] == tb.Char() {
						continue nextTB
					}
				}
				r.appendData(tb.Src)
				return r.renderBaseBlock(n, entering)
			}
		}
		return fmt.Errorf("tb not found")
	}
}
