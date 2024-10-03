package mdrenderer

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func (r *Renderer) renderCodeSpan(n ast.Node, entering bool) error {
	if entering {
		// reserving space for optional space and backticks
		r.appendData(nil, nil)
		r.pushPos(len(r.bufs))
		return nil
	} else {
		start := r.popPos()
		counts := map[int]struct{}{}
		seqLen := 0
		needsSpaces := false
		startsWithSpace := false
		firstByte := true
		isEntirelySpace := true
		var c byte
		for _, buf := range r.bufs[start:] {
			for _, c = range buf.data {
				isEntirelySpace = isEntirelySpace && c == ' ' || c == '\t' || c == '\n'
				if firstByte {
					firstByte = false
					if c == '`' {
						// starts with '`'
						needsSpaces = true
					} else if c == ' ' {
						startsWithSpace = true
					}
				}
				if c == '`' {
					seqLen++
				} else if seqLen > 0 {
					counts[seqLen] = struct{}{}
					seqLen = 0
				}
			}
		}
		if seqLen > 0 {
			// ends with '`'
			counts[seqLen] = struct{}{}
			needsSpaces = true
		} else if !isEntirelySpace && startsWithSpace && c == ' ' {
			// starts and ends with space, but not entirely spaces
			needsSpaces = true
		}
		// find shortest available sequence of backticks
		seqLen = 1
		for {
			if _, found := counts[seqLen]; !found {
				break
			}
			seqLen++
		}
		backticks := backticks.get(seqLen)
		r.bufs[start-2].data = backticks
		if needsSpaces {
			r.bufs[start-1].data = space
			r.appendData(space, backticks)
		} else {
			r.appendData(backticks)
		}
		return r.flush()
	}
}

func (r *Renderer) renderCodeBlock(n ast.Node, entering bool) (err error) {
	// render code block as fenced code blocks to avoid the pain
	// with list continuation vs code block
	if entering {
		err = r.renderBaseBlock(n, entering)
		r.enterCodeBlock(nil, n.Lines())
	} else {
		r.exitCodeBlock()
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) renderFencedCodeBlock(n ast.Node, entering bool) (err error) {
	if entering {
		err = r.renderBaseBlock(n, entering)
		n := n.(*ast.FencedCodeBlock)
		r.enterCodeBlock(n.Info, n.Lines())
	} else {
		r.exitCodeBlock()
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) enterCodeBlock(info *ast.Text, lines *text.Segments) {
	r.appendData(nil)
	r.pushPos(len(r.bufs))
	if info != nil {
		r.renderSegment(info.Segment)
	}
	r.endLine()
	r.renderSegments(lines, true)
}

func (r *Renderer) exitCodeBlock() {
	start := r.popPos()
	fenceChar := r.config.FencedCodeBlockChar
	if fenceChar == '`' {
		// Backtick fence info strings can't contain backticks
		for _, sub := range r.bufs[start:] {
			if bytes.IndexByte(sub.data, '`') != -1 {
				fenceChar = '~'
				break
			}
			if sub.endOfLine {
				break
			}
		}
	}

	counts := map[int]struct{}{}
	spaceCount := 0
	fenceCharCount := 0
	skipLine := true // skip the info line
	inPrefix := true
	for _, sub := range r.bufs[start:] {
		if sub.isPrefix {
			continue
		}
		if skipLine {
			if sub.endOfLine {
				skipLine = false
			}
			continue
		}
	loop:
		for _, c := range sub.data {
			switch c {
			case '\t':
				if inPrefix {
					// not a fence candidate
					spaceCount = 0
					fenceCharCount = 0
					inPrefix = true
					skipLine = true
					break
				}
			case ' ':
				if !inPrefix {
					continue
				}
				if spaceCount == 3 {
					// not a fence candidate
					spaceCount = 0
					skipLine = true
					break
				}
				spaceCount++
			case fenceChar:
				fenceCharCount++
				inPrefix = false
				continue
			default:
				// not a fence candidate
				spaceCount = 0
				fenceCharCount = 0
				inPrefix = true
				skipLine = true
				break loop
			}
		}
		if sub.endOfLine {
			if fenceCharCount >= 3 {
				counts[fenceCharCount] = struct{}{}
			}
			spaceCount = 0
			fenceCharCount = 0
			inPrefix = true
			skipLine = false
		}
	}
	fenceCharCount = 3
	for {
		if _, found := counts[fenceCharCount]; !found {
			break
		}
		fenceCharCount++
	}
	var fence []byte
	if fenceChar == '`' {
		fence = backticks.get(fenceCharCount)
	} else {
		fence = tildas.get(fenceCharCount)
	}
	r.bufs[start-1].data = fence
	r.endNonemptyLine()
	r.appendData(fence)
	r.endLine()
}
