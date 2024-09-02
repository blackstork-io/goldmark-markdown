// Package mdexamples contains utilities for reading test data.
package mdexamples

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type SpecExample struct {
	Markdown []byte
	HTML     []byte
	Link     string
	ID       int
	Section  string
}

type spec struct {
	ExampleLinkFormat string        `json:"exampleLinkFormat"`
	Examples          []specExample `json:"examples"`
}

type specExample struct {
	Markdown string `json:"markdown"`
	HTML     string `json:"html"`
	ID       int    `json:"id"`
	Section  string `json:"section,omitempty"`
}

func (s *spec) toExamples() []SpecExample {
	res := make([]SpecExample, len(s.Examples))
	for i, ex := range s.Examples {
		res[i] = SpecExample{
			Markdown: []byte(ex.Markdown),
			HTML:     []byte(ex.HTML),
			Link:     fmt.Sprintf(s.ExampleLinkFormat, ex.ID),
			ID:       ex.ID,
			Section:  ex.Section,
		}
	}
	return res
}

type FileExample struct {
	Name string
	Data []byte
}

func ReadDocumentExample(name string) FileExample {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", "documents", name)

	f, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return FileExample{
		Name: name,
		Data: f,
	}
}

func ReadAllDocumentExamples() (tests []FileExample) {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", "documents", "*.md")
	matches, err := filepath.Glob(path)
	if err != nil {
		panic(err)
	}
	for _, match := range matches {
		tests = append(tests, ReadDocumentExample(filepath.Base(match)))
	}
	if len(tests) == 0 {
		panic("no document examples found")
	}
	return
}

type SpecExampleFile struct {
	Name     string
	Examples []SpecExample
}

func ReadAllSpecExamples() (tests []SpecExampleFile) {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", "specs", "*.json")
	matches, err := filepath.Glob(path)
	if err != nil {
		panic(err)
	}
	if len(matches) == 0 {
		panic("no spec examples found")
	}
	for _, match := range matches {
		name := filepath.Base(match)
		tests = append(tests, SpecExampleFile{
			Name:     name,
			Examples: ReadSpecExamples(name),
		})
	}
	return
}

func ReadSpecExamples(name string) []SpecExample {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "testdata", "specs", name)
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	var specfile spec

	err = dec.Decode(&specfile)
	if err != nil {
		panic(err)
	}
	return specfile.toExamples()
}

// SKIP: [Autolinks (extension)] [Disallowed Raw HTML (extension)]
// {
// 	 "markdown": "foo <!-- not a comment -- two hyphens -->\n",
// 	 "html": "<p>foo &lt;!-- not a comment -- two hyphens --&gt;</p>\n",
// 	 "id": 649,
// 	 "comment": "skip: This is incorrect: https://html.spec.whatwg.org/#comments"
// },
// {
// 	 "markdown": "foo <!--> foo -->\n\nfoo <!-- foo--->\n",
// 	 "html": "<p>foo &lt;!--&gt; foo --&gt;</p>\n<p>foo &lt;!-- foo---&gt;</p>\n",
// 	 "id": 650,
// 	 "comment": "skip: Probably incorrect too (see above)"
// },
