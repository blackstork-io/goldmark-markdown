package options

import (
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
)

const OptParser renderer.OptionName = "Parser"

type ParserOpt struct {
	Parser parser.Parser
}

// apply implements Option.
func (t ParserOpt) Apply(c *Config) {
	if t.Parser != nil {
		c.parser = t.Parser
	}
}

var _ Option = ParserOpt{}
