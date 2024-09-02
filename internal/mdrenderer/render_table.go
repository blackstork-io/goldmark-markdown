package mdrenderer

import (
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

var (
	pipe  = []byte{'|'}
	colon = []byte{':'}
)

type tableCtx struct {
	// indexes of the "---" in delimiter row
	delimRow []int
	// max length of the content in each column
	maxLen []int
	// cells, including header cells
	cells [][]cellInfo
}

type cellInfo struct {
	paddingIdx int
	contentLen int
}

var dashes = padding("--------")

func (r *Renderer) renderTable(n ast.Node, entering bool) (err error) {
	// do not call r.renderBaseBlock for table cells/rows/header
	// even though they are block elements. They are non-standard and
	// handle their own line breaks.
	if entering {
		// force a blank line before a table
		// n.HasBlankPreviousLines() is always false, irregarless of the actual content
		for !r.previousLineBlank {
			r.endLine()
		}
		r.table = &tableCtx{}
	} else {
		tbl := r.table
		r.table = nil
		for i, maxLen := range tbl.maxLen {
			cellPadding := max(maxLen+1, 2) // except the first space, it is already there
			for _, row := range tbl.cells {
				if i < len(row) {
					r.bufs[row[i].paddingIdx].data = spaces.get(cellPadding - row[i].contentLen)
				}
			}
			if i < len(tbl.delimRow) {
				// leading and trailing spaces/colons are already there
				r.bufs[tbl.delimRow[i]].data = dashes.get(max(maxLen, 1))
			}
		}
		r.endLine()
		err = r.renderBaseBlock(n, entering)
	}
	return
}

func (r *Renderer) renderTableHeader(n ast.Node, entering bool) error {
	err := r.renderTableRow(n, entering)
	if err != nil {
		return err
	}
	if !entering {
		n := n.Parent().(*east.Table)
		r.appendData(pipe)

		for i := range r.table.cells[0] {
			align := east.AlignNone
			if i < len(n.Alignments) {
				align = n.Alignments[i]
			}
			if align == east.AlignLeft || align == east.AlignCenter {
				r.appendData(colon)
			} else {
				r.appendData(space)
			}
			r.appendData(nil)
			r.table.delimRow = append(r.table.delimRow, len(r.bufs)-1)

			if align == east.AlignRight || align == east.AlignCenter {
				r.appendData(colon, pipe)
			} else {
				r.appendData(space, pipe)
			}
		}
		r.endLine()
	}
	return nil
}

func (r *Renderer) renderTableRow(_ ast.Node, entering bool) error {
	if entering {
		r.table.cells = append(r.table.cells, nil)
		r.appendData(pipe)
	} else {
		r.endLine()
	}
	return nil
}

func (r *Renderer) renderTableCell(_ ast.Node, entering bool) error {
	if entering {
		r.appendData(space)
		r.pushPos(len(r.bufs))
	} else {
		start := r.popPos()
		contentLen := 0
		for _, sub := range r.bufs[start:] {
			contentLen += utf8.RuneCount(sub.data)
		}
		row := &r.table.cells[len(r.table.cells)-1]
		curCol := len(*row)
		if len(r.table.maxLen) == curCol {
			r.table.maxLen = append(r.table.maxLen, contentLen)
		} else {
			r.table.maxLen[curCol] = max(r.table.maxLen[curCol], contentLen)
		}
		r.appendData(nil, pipe)
		*row = append(*row, cellInfo{
			paddingIdx: len(r.bufs) - 2,
			contentLen: contentLen,
		})
	}
	return nil
}
