package ansi

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestMarginWriterEmptyContext(t *testing.T) {
	w := NewMarginWriter(NewRenderContext(Options{}), io.Discard, StyleBlock{})
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}
func TestMarginWriterLiveStyles(t *testing.T) {
	ctx := NewRenderContext(Options{WordWrap: 8})
	indent := uint(1)
	bold := false
	rules := StyleBlock{StylePrimitive: StylePrimitive{Bold: &bold}, Indent: &indent}
	ctx.blockStack.Push(BlockElement{Style: StyleBlock{}})
	ctx.blockStack.Push(BlockElement{Style: rules})
	var got, want bytes.Buffer
	w := NewMarginWriter(ctx, &got, rules)
	// Parent changes after construction, before the first indent callback.
	parentBold := true
	(*ctx.blockStack)[0].Style.Bold = &parentBold
	_, _ = io.WriteString(w, "x\n")
	_, _ = renderText(&want, (*ctx.blockStack)[0].Style.StylePrimitive, " ")
	want.WriteString("x")
	for range 6 {
		_, _ = renderText(&want, rules.StylePrimitive, " ")
	}
	want.WriteByte('\n')
	// Pointer targets and parent context remain live between writes.
	bold = true
	parentBold = false
	_, _ = io.WriteString(w, "y\n")
	_, _ = renderText(&want, (*ctx.blockStack)[0].Style.StylePrimitive, " ")
	want.WriteString("y")
	for range 6 {
		_, _ = renderText(&want, rules.StylePrimitive, " ")
	}
	want.WriteByte('\n')
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if got.String() != want.String() {
		t.Fatalf("live styles: got %q; want %q", got.String(), want.String())
	}
}
func TestStyledCellCacheValues(t *testing.T) {
	data, err := os.ReadFile("../styles/dark.json")
	if err != nil {
		t.Fatal(err)
	}
	var styles StyleConfig
	if err = json.Unmarshal(data, &styles); err != nil {
		t.Fatal(err)
	}
	rules := StylePrimitive{}
	var cache styledCellCache
	check := func() {
		t.Helper()
		var want bytes.Buffer
		_, _ = renderText(&want, rules, "token")
		if got := cache.text(rules, "token"); got != want.String() {
			t.Fatalf("cache: got %q; want %q", got, want.String())
		}
	}
	check()
	empty := ""
	rules.Color = &empty
	check()
	rules.BackgroundColor = &empty
	check()
	rules.Color = styles.Document.Color
	check()
	rules.Color = styles.Heading.Color
	check()
	rules.BackgroundColor = styles.H1.BackgroundColor
	check()
	rules.Color = nil
	rules.BackgroundColor = nil
	check()
	value := reflect.ValueOf(&rules).Elem()
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if field.Kind() != reflect.Pointer || field.Type().Elem().Kind() != reflect.Bool {
			continue
		}
		flag := reflect.New(field.Type().Elem())
		flag.Elem().SetBool(true)
		field.Set(flag)
		check()
		flag.Elem().SetBool(false)
		check()
		field.Set(reflect.Zero(field.Type()))
		check()
	}
}
