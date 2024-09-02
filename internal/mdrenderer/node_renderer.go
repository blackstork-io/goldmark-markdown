package mdrenderer

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer"

	"github.com/blackstork-io/goldmark-markdown/internal/noderenderer"
	"github.com/blackstork-io/goldmark-markdown/internal/options"
)

// TODO: add the ability to register new node renderers
// var _ renderer.NodeRenderer = nodeRenderer{}

var errSkipChildren = errors.New("skip children")

func renderHelper(n ast.Node, fn noderenderer.RenderFunc) (err error) {
	err = fn(n, true)
	if err != nil {
		if !errors.Is(err, errSkipChildren) {
			return
		}
		// this error gets overwritten by next fn call
	} else {
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if err = renderHelper(c, fn); err != nil {
				return
			}
		}
	}
	err = fn(n, false)
	if errors.Is(err, errSkipChildren) {
		// children were already rendered, nothing to skip
		err = nil
	}
	return
}

// NewRenderer returns a new renderer. Use [goldmark.WithRenderer] to add it to a goldmark instance.
func NewRenderer() *Renderer {
	return &Renderer{
		config:            options.DefaultConfig(),
		previousLineBlank: true, // document (conceptually) starts with a blank line
		wasAtNL:           true,
	}
}

type buf struct {
	data      []byte
	isPrefix  bool
	endOfLine bool
}

type bufs []buf

// Renderer renders markdown nodes to a writer.
type Renderer struct {
	config            options.Config
	initOnce          sync.Once
	nodeRendererFuncs []noderenderer.RenderFunc

	source []byte

	// stack of offsets in bufs that are inspected (halts flushing)
	positionStack []int
	listStack     []listCtx
	emphStack     []emphCtx
	// tracks which emphStack will be closed next
	emphCloseStack    []int
	table             *tableCtx
	bufs              bufs
	prefix            []byte
	previousLineBlank bool
	wasAtNL           bool

	w io.Writer
}

var _ renderer.Renderer = &Renderer{}

func (r *Renderer) last() *buf {
	if len(r.bufs) == 0 {
		return &buf{endOfLine: r.wasAtNL}
	}
	return &r.bufs[len(r.bufs)-1]
}

func (r *Renderer) atEmptyLine() bool {
	return r.last().endOfLine
}

func (r *Renderer) setRenderer(kind ast.NodeKind, fn noderenderer.RenderFunc) {
	r.nodeRendererFuncs = setAt(r.nodeRendererFuncs, int(kind), fn)
}

func (r *Renderer) getRenderer(kind ast.NodeKind) noderenderer.RenderFunc {
	idx := int(kind)
	var fn noderenderer.RenderFunc
	if idx < len(r.nodeRendererFuncs) {
		fn = r.nodeRendererFuncs[idx]
	}
	if fn == nil {
		return r.defaultRenderer
	}
	return fn
}

// AddOptions adds options to the renderer.
//
// Only options defined in this package are supported (functions starting with "With").
//
// You can call this method directly or use [goldmark.WithRendererOptions].
func (r *Renderer) AddOptions(opts ...renderer.Option) {
	r.initOnce.Do(r.init)

	config := renderer.NewConfig()
	for _, opt := range opts {
		opt.SetConfig(config)
	}
	if opt, found := config.Options[options.OptParser]; found {
		delete(config.Options, options.OptParser)
		opt.(options.Option).Apply(&r.config)
	}
	for name, val := range config.Options {
		if opt, ok := val.(options.Option); ok {
			opt.Apply(&r.config)
		} else {
			r.config.Errs = append(r.config.Errs, fmt.Errorf("%w: %s", options.ErrUnsupportedOption, name))
		}
	}

	config.NodeRenderers.Sort()
	l := len(config.NodeRenderers)
	for i := l - 1; i >= 0; i-- {
		v := config.NodeRenderers[i]
		nr, ok := v.Value.(noderenderer.NodeRenderer)
		if !ok {
			continue
		}
		r.setRenderer(nr.Kind, nr.Fn)
	}
}

// Render renders the given AST node to the given writer.
func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.initOnce.Do(r.init)
	err := errors.Join(r.config.Errs...)
	if err != nil {
		return err
	}
	r.w = w
	r.source = source
	r.bufs = r.bufs[:0]

	err = renderHelper(n, func(n ast.Node, entering bool) (err error) {
		err = r.getRenderer(n.Kind())(n, entering)
		if err != nil {
			return
		}
		if !entering {
			return r.flush()
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(r.positionStack) != 0 {
		return fmt.Errorf("leftover state at the end of the document: %+v", r.positionStack)
	}
	r.endNonemptyLine()
	return r.flush()
}

func (r *Renderer) init() {
	r.setRenderer(ast.KindDocument, r.renderBaseBlock)
	r.setRenderer(ast.KindHeading, r.renderHeading)
	r.setRenderer(ast.KindBlockquote, r.renderBlockquote)
	r.setRenderer(ast.KindCodeBlock, r.renderCodeBlock)
	r.setRenderer(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	r.setRenderer(ast.KindHTMLBlock, r.renderHTMLBlock)
	r.setRenderer(ast.KindList, r.renderList)
	r.setRenderer(ast.KindListItem, r.renderListItem)
	r.setRenderer(ast.KindParagraph, r.renderParagraph)
	r.setRenderer(ast.KindTextBlock, r.renderBaseBlock)
	r.setRenderer(ast.KindThematicBreak, r.renderThematicBreak)

	// inlines
	r.setRenderer(ast.KindAutoLink, r.renderAutoLink)
	r.setRenderer(ast.KindCodeSpan, r.renderCodeSpan)
	r.setRenderer(ast.KindEmphasis, r.renderEmphasis)
	r.setRenderer(ast.KindImage, r.renderImage)
	r.setRenderer(ast.KindLink, r.renderLink)
	r.setRenderer(ast.KindRawHTML, r.renderRawHTML)
	r.setRenderer(ast.KindText, r.renderText)
	r.setRenderer(ast.KindString, r.renderString)

	// GFM
	r.setRenderer(east.KindTable, r.renderTable)
	r.setRenderer(east.KindTableHeader, r.renderTableHeader)
	r.setRenderer(east.KindTableRow, r.renderTableRow)
	r.setRenderer(east.KindTableCell, r.renderTableCell)
	r.setRenderer(east.KindStrikethrough, r.renderStrikethrough)
	r.setRenderer(east.KindTaskCheckBox, r.renderTaskCheckBox)
}

// appends new line if the previous line is not empty
func (r *Renderer) endNonemptyLine() {
	if !r.last().endOfLine {
		r.endLine()
	}
}

func (r *Renderer) endLine() {
	r.previousLineBlank = r.atEmptyLine()
	if r.previousLineBlank {
		r.bufs = append(r.bufs, buf{
			data:      r.prefix,
			isPrefix:  true,
			endOfLine: true,
		})
	} else {
		if len(r.bufs) == 0 || r.bufs[len(r.bufs)-1].endOfLine {
			r.appendData(nil)
		}
		r.bufs[len(r.bufs)-1].endOfLine = true
	}
}

func (r *Renderer) flush() error {
	if len(r.positionStack) != 0 || len(r.emphStack) != 0 || len(r.bufs) == 0 || r.table != nil {
		return nil
	}

	for _, b := range r.bufs {
		if len(b.data) != 0 {
			_, err := r.w.Write(b.data)
			if err != nil {
				return err
			}
		}
		if b.endOfLine {
			_, err := r.w.Write(newLine)
			if err != nil {
				return err
			}
		}
	}
	r.wasAtNL = r.bufs[len(r.bufs)-1].endOfLine
	r.bufs = r.bufs[:0]
	return nil
}

func (r *Renderer) appendData(bufs ...[]byte) {
	if r.atEmptyLine() && len(r.prefix) > 0 {
		r.bufs = append(r.bufs, buf{
			data:     r.prefix,
			isPrefix: true,
		})
	}
	for _, b := range bufs {
		r.bufs = append(r.bufs, buf{
			data: b,
		})
	}
}

var (
	spaces    = padding("        ")
	backticks = padding("````````")
	tildas    = padding("~~~~~~~~")
	backslash = []byte{'\\'}

	newLine = []byte{'\n'}
	space   = spaces.get(1)
)

func (r *Renderer) pushPos(state int) {
	r.positionStack = append(r.positionStack, state)
}

func (r *Renderer) popPos() int {
	idx := len(r.positionStack) - 1
	state := r.positionStack[idx]
	r.positionStack = r.positionStack[:idx]
	return state
}

func (r *Renderer) pushPrefix(prefix []byte) {
	r.prefix = append(r.prefix, prefix...)
}

func (r *Renderer) popPrefixAndFlush(prefixLength int) error {
	newLen := len(r.prefix) - prefixLength
	if len(r.positionStack) == 0 {
		r.prefix = r.prefix[:newLen]
		return r.flush()
	} else {
		// can't flush now, so we lower the capacity
		// this results in a new allocation on the next append, and the old prefix is not overwritten
		r.prefix = r.prefix[:newLen:newLen]
		return nil
	}
}
