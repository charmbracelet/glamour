package mermaid

import (
	"container/list"
	"strings"
	"sync"
	"testing"
)

func resetCache() {
	diagramCache.Lock()
	diagramCache.entries = make(map[cacheKey]*list.Element)
	diagramCache.order.Init()
	diagramCache.Unlock()
}

func TestCacheHits(t *testing.T) {
	resetCache()
	src := "flowchart LR\nA --> B"

	first, err := renderCached(src, 0, unicodeGlyphs)
	if err != nil {
		t.Fatal(err)
	}

	diagramCache.Lock()
	size := len(diagramCache.entries)
	diagramCache.Unlock()
	if size != 1 {
		t.Fatalf("cache has %d entries, want 1", size)
	}

	second, err := renderCached(src, 0, unicodeGlyphs)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(first, "\n") != strings.Join(second, "\n") {
		t.Error("cached render differs from the first")
	}

	diagramCache.Lock()
	size = len(diagramCache.entries)
	diagramCache.Unlock()
	if size != 1 {
		t.Errorf("cache has %d entries after a hit, want 1", size)
	}

	// A different width or glyph set is a different key.
	if _, err := renderCached(src, 40, unicodeGlyphs); err != nil {
		t.Fatal(err)
	}
	if _, err := renderCached(src, 0, asciiGlyphs); err != nil {
		t.Fatal(err)
	}
	diagramCache.Lock()
	size = len(diagramCache.entries)
	diagramCache.Unlock()
	if size != 3 {
		t.Errorf("cache has %d entries, want 3", size)
	}
}

func TestCacheErrors(t *testing.T) {
	resetCache()
	src := "sankey-beta\na,b,10"

	if _, err := renderCached(src, 0, unicodeGlyphs); err == nil {
		t.Fatal("expected an error")
	}
	diagramCache.Lock()
	size := len(diagramCache.entries)
	diagramCache.Unlock()
	if size != 1 {
		t.Fatalf("failures are not cached: %d entries", size)
	}
	if _, err := renderCached(src, 0, unicodeGlyphs); err == nil {
		t.Fatal("expected the cached error")
	}
}

func TestCacheEviction(t *testing.T) {
	resetCache()
	for i := 0; i < cacheCapacity+10; i++ {
		src := "flowchart LR\nA" + string(rune('a'+i%26)) + " --> B"
		if _, err := renderCached(src, 0, unicodeGlyphs); err != nil {
			t.Fatal(err)
		}
	}
	diagramCache.Lock()
	size := len(diagramCache.entries)
	diagramCache.Unlock()
	if size > cacheCapacity {
		t.Errorf("cache grew to %d entries, capacity is %d", size, cacheCapacity)
	}
}

func TestCacheConcurrent(t *testing.T) {
	resetCache()
	srcs := []string{
		"flowchart LR\nA --> B",
		"flowchart TD\nA --> B --> C",
		"sequenceDiagram\nA->>B",
	}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_, _ = renderCached(srcs[(i+j)%len(srcs)], 40, unicodeGlyphs) //nolint:errcheck
			}
		}(i)
	}
	wg.Wait()
}

func TestDetectASCII(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want bool
	}{
		{map[string]string{"LC_ALL": "en_US.UTF-8"}, false},
		{map[string]string{"LC_ALL": "C"}, true},
		{map[string]string{"LC_ALL": "POSIX"}, true},
		{map[string]string{"LANG": "C.UTF-8"}, false},
		{map[string]string{"LANG": "de_DE.UTF-8"}, false},
		{map[string]string{"LANG": "C"}, true},
		{map[string]string{}, false},
	}
	for _, c := range cases {
		for _, k := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
			t.Setenv(k, "")
		}
		for k, v := range c.env {
			t.Setenv(k, v)
		}
		if got := detectASCII(); got != c.want {
			t.Errorf("detectASCII() with %v = %v, want %v", c.env, got, c.want)
		}
	}
}
