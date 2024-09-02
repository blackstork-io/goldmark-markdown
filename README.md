# goldmark-markdown

[Goldmark](https://github.com/yuin/goldmark) renderer, that renders goldmark AST back into markdown. Can be useful for programmatic markdown editing or formatting.

## Goal:

Rendered markdown should be semantically equivalent to the original markdown parsed by `goldmark`.

In practice, it means that
`markdown input` -> `goldmark parser AST` -> `goldmark HTML render`
and
`markdown input` -> `goldmark parser AST` -> `goldmark-markdown rendered markdown` -> `goldmark parser AST` -> `goldmark HTML render`
should be identical, since the canonical markdown target is HTML.

Features:
- [x] Correctly renders all examples in [commonmark 0.31.2 spec](https://spec.commonmark.org/0.31.2).
- [x] Correctly renders all examples in [GitHub Flavored Markdown 0.29 spec](https://github.github.com/gfm)
- [x] Correctly renders [a wide variety](https://github.com/blackstork-io/goldmark-markdown/tree/main/pkg/mdexamples/testdata/documents) of markdown documents.
- [ ] Supports rendering all markdown elements
  - [ ] TODO: indented code blocks are replaced by fenced code blocks. It's hard to calculate appropriate padding that doesn't conflict with lazy list continuations.
- [x] Supports rendering all GitHub Flavored Markdown elements

- [ ] Handles all edge cases.
  - [ ] Known issue: not all emphasis rules are followed, [some unnatural nested emphasis sequences](https://github.com/blackstork-io/goldmark-markdown/blob/e3da4ace8762bd353736fce10b3391326074a2ae/markdown_test.go#L117) change their meaning.
  - [ ] Currently, the escaping is overly eager
- [ ] In-depth customization. Choose your preferred heading, hr, code block styles, emphasis characters, etc.


## Example

```bash
go build ./examples/mdformat
./mdformat -i README.md -o README2.md
```