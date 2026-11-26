package ansi

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestNewElementCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		kind     ast.NodeKind
		code     string
		language string
	}{
		{
			name:     "fenced",
			markdown: "```go title=example\nfirst\n\nlast\n```\n",
			kind:     ast.KindFencedCodeBlock,
			code:     "first\n\nlast\n",
			language: "go",
		},
		{
			name:     "empty fenced",
			markdown: "```\n```\n",
			kind:     ast.KindFencedCodeBlock,
		},
		{
			name:     "fenced unicode and tabs",
			markdown: "~~~text\n\t你好 🌍\n\tend\n~~~\n",
			kind:     ast.KindFencedCodeBlock,
			code:     "\t你好 🌍\n\tend\n",
			language: "text",
		},
		{
			name:     "fenced CRLF",
			markdown: "```go\r\nfirst\r\nlast\r\n```\r\n",
			kind:     ast.KindFencedCodeBlock,
			code:     "first\r\nlast\r\n",
			language: "go",
		},
		{
			name:     "unclosed fence without final newline",
			markdown: "```go\nfirst\nlast",
			kind:     ast.KindFencedCodeBlock,
			code:     "first\nlast\n",
			language: "go",
		},
		{
			name:     "fence in list",
			markdown: "- item\n\n  ```go\n  first\n  \tlast\n  ```\n",
			kind:     ast.KindFencedCodeBlock,
			code:     "first\n\tlast\n",
			language: "go",
		},
		{
			name:     "indented",
			markdown: "    first\n\n    last\n",
			kind:     ast.KindCodeBlock,
			code:     "first\n\nlast\n",
		},
		{
			name:     "indented unicode and tabs",
			markdown: "\t你好 🌍\n\t\tend\n",
			kind:     ast.KindCodeBlock,
			code:     "你好 🌍\n\tend\n",
		},
		{
			name:     "indented CRLF",
			markdown: "    first\r\n    last\r\n",
			kind:     ast.KindCodeBlock,
			code:     "first\r\nlast\r\n",
		},
		{
			name:     "indented without final newline",
			markdown: "    first\n    last",
			kind:     ast.KindCodeBlock,
			code:     "first\nlast\n",
		},
		{
			name:     "indented in list",
			markdown: "- item\n\n      first\n      last\n",
			kind:     ast.KindCodeBlock,
			code:     "first\nlast\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := []byte(tt.markdown)
			node := findCodeBlock(t, source, tt.kind)
			element := (&ANSIRenderer{}).NewElement(node, source)
			code, ok := element.Renderer.(*CodeBlockElement)
			if !ok {
				t.Fatalf("renderer = %T, want *CodeBlockElement", element.Renderer)
			}
			if code.Code != tt.code {
				t.Errorf("Code = %q, want %q", code.Code, tt.code)
			}
			if code.Language != tt.language {
				t.Errorf("Language = %q, want %q", code.Language, tt.language)
			}
			if element.Entering != "\n" || element.Exiting != "" || element.Finisher != nil {
				t.Errorf("unexpected element boundaries: %#v", element)
			}
		})
	}
}

func TestNewElementEmptyIndentedCodeBlock(t *testing.T) {
	// A parser normally omits an empty indented block, but NewElement also
	// accepts AST nodes constructed by callers.
	element := (&ANSIRenderer{}).NewElement(ast.NewCodeBlock(), nil)
	code, ok := element.Renderer.(*CodeBlockElement)
	if !ok {
		t.Fatalf("renderer = %T, want *CodeBlockElement", element.Renderer)
	}
	if code.Code != "" || code.Language != "" {
		t.Errorf("empty block = %#v", code)
	}
}

func findCodeBlock(tb testing.TB, source []byte, kind ast.NodeKind) ast.Node {
	tb.Helper()
	document := goldmark.New().Parser().Parse(text.NewReader(source))
	var found ast.Node
	err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && node.Kind() == kind {
			found = node
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		tb.Fatal(err)
	}
	if found == nil {
		tb.Fatalf("no %s in markdown", kind)
	}
	return found
}
