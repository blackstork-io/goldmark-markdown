package options

var defaultHeadings = Headings{
	Atx: [7]Atx{
		{},
		{Open: []byte("#")},
		{Open: []byte("##")},
		{Open: []byte("###")},
		{Open: []byte("####")},
		{Open: []byte("#####")},
		{Open: []byte("######")},
	},
	Setext: [3]Setext{
		{},
		{Style: SetextStyleLongestLine},
		{Style: SetextStyleLongestLine},
	},
	PreferSetext: [3]bool{false, false, false},
}

type Headings struct {
	Atx          [7]Atx
	Setext       [3]Setext
	PreferSetext [3]bool
}

type Atx struct {
	Open  []byte // '#' (one space is added after this if content is not empty)
	Close []byte // ''; ' #' (directly follows content)
}

type Setext struct {
	Underline []byte // exact underline sequence or nil if dynamic
	Style     setextStyle
}

type setextStyle int

const (
	// SetextStyleNone indicates that the underline length is not dynamic.
	SetextStyleNone setextStyle = iota
	// SetextStyleLongestLine indicates that the underline length is the longest line.
	SetextStyleLongestLine
	// SetextStyleLastLine indicates that the underline length is the last line.
	SetextStyleLastLine
)
