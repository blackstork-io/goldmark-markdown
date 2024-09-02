package mdrenderer

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/util"
)

var referenceRe = regexp.MustCompile(`^(?:([A-Za-z][A-Za-z0-9]{1,50})|(#[xX][0-9A-Fa-f]{1,6})|(#\d{1,7}));`)

func refLen(text []byte) int {
	idxs := referenceRe.FindSubmatch(text)
	if idxs == nil {
		return 0
	}
	if len(idxs[1]) != 0 {
		_, found := util.LookUpHTML5EntityByName(string(idxs[1]))
		if !found {
			return 0
		}
	}
	return len(idxs[0])
}

func (r *Renderer) escapeText(text []byte) {
	// Any punctuation character may be backslash-escaped
	// !"#$%&'()*+,-./:;<=>?@[\]^_`{|}~
	// Trying to escape only the necessary characters
	start := 0
	afterBackslash := false
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\\':
			afterBackslash = !afterBackslash
			continue
		case '&':
			if afterBackslash {
				afterBackslash = false
				continue
			}
			length := refLen(text[i+1:])
			if length != 0 {
				// skip over the reference
				i += length
				continue
			}
		case '*', '#', '-', '=', '`', '_', '+', '>', '|', '~':
			if afterBackslash {
				afterBackslash = false
				continue
			}
			if i > start {
				r.appendData(text[start:i], backslash)
				start = i
			} else {
				r.appendData(backslash)
			}
		default:
			afterBackslash = false
		}
	}
	r.appendData(text[start:])
	if afterBackslash {
		r.appendData(backslash)
	}
}

func (r *Renderer) renderText(node ast.Node, entering bool) error {
	if !entering {
		n := node.(*ast.Text)
		if n.IsRaw() {
			text := n.Text(r.source)
			if r.table == nil {
				r.appendData(text)
			} else {
				// we need to escape pipes in table cells, even in raw text
				for {
					idx := bytes.IndexByte(text, '|')
					if idx == -1 {
						r.appendData(text)
						break
					}
					if idx != 0 {
						r.appendData(text[:idx])
					}
					r.appendData(backslash, pipe)
					text = text[idx+1:]
				}
			}
		} else {
			r.escapeText(n.Text(r.source))
		}
		if n.HardLineBreak() {
			r.appendData(r.config.HardLineBreak)
			r.endLine()
		}
		if n.SoftLineBreak() {
			r.endLine()
		}
	}
	return nil
}

func (r *Renderer) renderString(n ast.Node, entering bool) error {
	node := n.(*ast.String)
	if entering {
		if node.IsCode() || node.IsRaw() {
			r.appendData(node.Value)
		} else {
			r.escapeText(node.Value)
		}
	}
	return nil
}
