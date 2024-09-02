package options

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

var defaultThematicBreaks = []ThematicBreak{
	{Src: []byte("***"), char: '*'},
	{Src: []byte("---"), char: '-', potentialSetext: true},
	{Src: []byte("___"), char: '_'},
}

const OptThematicBreaks renderer.OptionName = "ThematicBreaks"

type ThematicBreaksOpt []string

// apply implements Option.
func (t ThematicBreaksOpt) Apply(c *Config) {
	c.ThematicBreaks = make([]ThematicBreak, 0, len(t)+len(defaultThematicBreaks))
	for _, tb := range t {
		tbb := []byte(tb)
		n := c.parse(tbb)
		if n.ChildCount() != 1 || n.FirstChild().Kind() != ast.KindThematicBreak {
			c.Errs = append(c.Errs, fmt.Errorf("%w: %q is not a thematic break", ErrIncorrectOption, tb))
			continue
		}
		res := ThematicBreak{
			Src: tbb,
		}
		for _, c := range tbb {
			if c != ' ' {
				res.char = c
				break
			}
		}
		if res.char == '-' {
			res.potentialSetext = c.parse(append([]byte("t\n"), tbb...)).
				FirstChild().Kind() == ast.KindHeading
		}
		c.ThematicBreaks = append(c.ThematicBreaks, res)
	}
	c.ThematicBreaks = append(c.ThematicBreaks, defaultThematicBreaks...)
}

var _ Option = ThematicBreaksOpt{}

type ThematicBreak struct {
	Src             []byte
	potentialSetext bool
	char            byte
}

func (tb *ThematicBreak) Char() byte {
	return tb.char
}

func (tb *ThematicBreak) Fits(self ast.Node) bool {
	c := tb.Char()
	switch c {
	case '-':
		// check if setext
		if tb.potentialSetext && !self.HasBlankPreviousLines() {
			sib := self.PreviousSibling()
			if sib != nil && sib.Kind() == ast.KindParagraph {
				return false
			}
		}
		fallthrough
	case '*':
		// check marker
		p := self.Parent()
		if p.Kind() == ast.KindListItem {
			if l, ok := p.Parent().(*ast.List); ok {
				if l.Marker == c {
					return false
				}
			}
		}
	}
	return true
}
