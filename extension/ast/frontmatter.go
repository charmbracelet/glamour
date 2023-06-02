package ast

import (
	"fmt"
	"github.com/yuin/goldmark/ast"
)

// Frontmatter AST Node holding parsed YAML Frontmatter.
// Status Parse Error or empty.
type Frontmatter struct {
	ast.BaseBlock
	MetaData map[interface{}]interface{}
	Status   string
}

var KindFrontmatter = ast.NewNodeKind("Frontmatter")

func (f Frontmatter) Kind() ast.NodeKind {
	return KindFrontmatter
}

func (f *Frontmatter) Dump(source []byte, level int) {
	m := map[string]string{}
	m["_status"] = fmt.Sprintf("%v", f.Status)
	for key, value := range f.MetaData {
		m[fmt.Sprintf("%v", key)] = fmt.Sprintf("%v", value)
	}
	ast.DumpHelper(f, source, level, m, nil)
}
