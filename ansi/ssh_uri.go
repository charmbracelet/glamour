package ansi

import (
	"strings"

	"github.com/yuin/goldmark/ast"
)

// isSSHURISuffix reports whether text is the path portion of an scp-style URI
// (for example ":gitea/tea" following "git@gitea.com").
func isSSHURISuffix(text string) bool {
	return len(text) > 1 && text[0] == ':' && !strings.ContainsAny(text[1:], " \t\n\r")
}

// sshURIFromAutolink checks whether an email autolink is the user@host prefix of
// an scp-style URI whose path segment was parsed as a separate text node.
func sshURIFromAutolink(n *ast.AutoLink, source []byte) (string, bool) {
	if n.AutoLinkType != ast.AutoLinkEmail {
		return "", false
	}

	next := n.NextSibling()
	if next == nil || next.Kind() != ast.KindText {
		return "", false
	}

	suffix := string(next.(*ast.Text).Segment.Value(source))
	if !isSSHURISuffix(suffix) {
		return "", false
	}

	return string(n.URL(source)) + suffix, true
}

// isSSHURISuffixNode reports whether a text node is the path suffix of a
// preceding scp-style URI autolink that was already rendered as plain text.
func isSSHURISuffixNode(node ast.Node, source []byte) bool {
	prev := node.PreviousSibling()
	if prev == nil || prev.Kind() != ast.KindAutoLink {
		return false
	}

	_, ok := sshURIFromAutolink(prev.(*ast.AutoLink), source)
	return ok
}
