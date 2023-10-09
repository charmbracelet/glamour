package ansi

import (
	"bytes"
)

// BlockStack is a stack of block elements, used to calculate the current
// indentation & margin level during the rendering process.
type BlockStack []BlockElement

// Len returns the length of the stack.
func (s *BlockStack) Len() int {
	return len(*s)
}

// Push appends an item to the stack.
func (s *BlockStack) Push(e BlockElement) {
	*s = append(*s, e)
}

// Pop removes the last item on the stack.
func (s *BlockStack) Pop() {
	stack := *s
	if len(stack) == 0 {
		return
	}

	stack = stack[0 : len(stack)-1]
	*s = stack
}

// TODO do I need to return the available rendering width.

// Parent returns the current BlockElement's parent.
func (s BlockStack) Parent() BlockElement {
	if len(s) == 1 {
		return BlockElement{
			Block: &bytes.Buffer{},
		}
	}

	return s[len(s)-2]
}

// Current returns the current BlockElement.
func (s BlockStack) Current() BlockElement {
	if len(s) == 0 {
		return BlockElement{
			Block: &bytes.Buffer{},
		}
	}

	return s[len(s)-1]
}
