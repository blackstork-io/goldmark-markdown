package mdrenderer

import "github.com/yuin/goldmark/ast"

var (
	imagePrefix       = []byte("!")
	linkOpen          = []byte("[")
	linkMiddle        = []byte("](")
	linkClose         = []byte(")")
	titleClose        = []byte{'"'}
	bracketOpen       = []byte("<")
	bracketClose      = []byte(">")
	protocolSeparator = []byte("://")
)

func (r *Renderer) renderImage(n ast.Node, entering bool) error {
	node := n.(*ast.Image)
	if entering {
		r.appendData(imagePrefix)
	}
	return r.renderBaseLink(node.Destination, node.Title, entering)
}

func (r *Renderer) renderLink(n ast.Node, entering bool) error {
	node := n.(*ast.Link)
	return r.renderBaseLink(node.Destination, node.Title, entering)
}

func validBracketless(dest []byte) bool {
	if len(dest) == 0 {
		return true
	}
	// does not start with <
	if dest[0] == '<' {
		return false
	}
	// does not include ASCII control characters or space character
	// U+0000–1F (both including) or U+007F.
	parenCount := 0
	backslashEscaped := false
	for _, c := range dest {
		if c <= 0x1f || c == 0x7f {
			return false
		}
		// and includes parentheses only if (a) they are backslash-escaped or (b) they are part of a balanced pair of unescaped parentheses
		switch c {
		case ' ':
			return false
		case '\\':
			backslashEscaped = !backslashEscaped
		case '(':
			if !backslashEscaped {
				parenCount++
			}
			backslashEscaped = false
		case ')':
			if !backslashEscaped {
				parenCount--
				if parenCount < 0 {
					return false
				}
			}
			backslashEscaped = false
		default:
			backslashEscaped = false
		}
	}
	return parenCount == 0 && !backslashEscaped
}

func escapeBracketedIfNeeded(dest []byte) (res []byte) {
	start := 0
	backslashEscaped := false
	for i, c := range dest {
		switch c {
		case '\n':
			if res == nil {
				res = make([]byte, i, len(dest)+10)
				copy(res, dest[:i])
			} else {
				res = append(res, dest[start:i]...)
			}
			if backslashEscaped {
				// Can't escape newlines, treat backslash as a literal character
				res = append(res, '\\')
				backslashEscaped = false
			}
			start = i + 1
			res = append(res, "%0A"...)
		case '\\':
			backslashEscaped = !backslashEscaped
		case '<', '>':
			if !backslashEscaped {
				if res == nil {
					res = make([]byte, i, len(dest)+10)
					copy(res, dest[:i])
				} else {
					res = append(res, dest[start:i]...)
				}
				start = i + 1
				res = append(res, '\\', c)
			}
			backslashEscaped = false
		default:
			backslashEscaped = false
		}
	}
	if res == nil {
		// no need to escape anything
		res = dest[:len(dest):len(dest)]
	} else {
		res = append(res, dest[start:]...)
	}
	if backslashEscaped {
		res = append(res, '\\')
	}
	return
}

func escapeTitleIfNeeded(dest []byte, quote byte) (res []byte) {
	start := 0
	backslashEscaped := false
	for i, c := range dest {
		switch c {
		case '\\':
			backslashEscaped = !backslashEscaped
		case quote:
			if !backslashEscaped {
				if res == nil {
					res = make([]byte, i, len(dest)+10)
					copy(res, dest[:i])
				} else {
					res = append(res, dest[start:i]...)
				}
				start = i + 1
				res = append(res, '\\', quote)
			}
			backslashEscaped = false
		default:
			backslashEscaped = false
		}
	}
	if res == nil {
		// no need to escape anything
		res = dest[:len(dest):len(dest)]
	} else {
		res = append(res, dest[start:]...)
	}
	if backslashEscaped {
		res = append(res, '\\')
	}
	return
}

func (r *Renderer) renderBaseLink(dest, title []byte, entering bool) error {
	if entering {
		r.appendData(linkOpen)
	} else {
		bracketless := validBracketless(dest)
		if bracketless {
			if len(dest) == 0 && title != nil {
				bracketless = false
			}
		}
		if bracketless {
			r.appendData(linkMiddle, dest)
		} else {
			r.appendData(linkMiddle, bracketOpen, escapeBracketedIfNeeded(dest), bracketClose)
		}
		if title != nil {
			r.appendData(space, r.config.LinkTitleQuote, escapeTitleIfNeeded(title, r.config.LinkTitleQuote[0]), r.config.LinkTitleQuote)
		}
		r.appendData(linkClose)
	}
	return nil
}

func (r *Renderer) renderAutoLink(n ast.Node, entering bool) error {
	if entering {
		n := n.(*ast.AutoLink)
		r.appendData(bracketOpen)
		if n.Protocol != nil {
			r.appendData(n.Protocol, protocolSeparator)
		}
		r.appendData(n.Label(r.source))
	} else {
		r.appendData(bracketClose)
	}
	return nil
}
