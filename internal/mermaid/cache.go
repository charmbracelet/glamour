package mermaid

import (
	"container/list"
	"sync"
)

// cacheCapacity bounds how many rendered diagrams are kept around, so
// repeated renders of the same block at the same width (resizes, reloads)
// are cheap while memory stays capped.
const cacheCapacity = 64

type cacheKey struct {
	src   string
	limit int
	g     glyphSet
}

type cacheEntry struct {
	key   cacheKey
	lines []string
	err   error
}

var diagramCache = struct {
	sync.Mutex
	entries map[cacheKey]*list.Element
	order   *list.List
}{
	entries: make(map[cacheKey]*list.Element),
	order:   list.New(),
}

// renderCached renders src through the diagram cache.
func renderCached(src string, limit int, g glyphSet) ([]string, error) {
	key := cacheKey{src: src, limit: limit, g: g}

	diagramCache.Lock()
	if el, ok := diagramCache.entries[key]; ok {
		diagramCache.order.MoveToFront(el)
		entry := el.Value.(*cacheEntry)
		diagramCache.Unlock()
		return entry.lines, entry.err
	}
	diagramCache.Unlock()

	lines, err := render(src, limit, g)

	diagramCache.Lock()
	el := diagramCache.order.PushFront(&cacheEntry{key: key, lines: lines, err: err})
	diagramCache.entries[key] = el
	for len(diagramCache.entries) > cacheCapacity {
		oldest := diagramCache.order.Back()
		if oldest == nil {
			break
		}
		diagramCache.order.Remove(oldest)
		delete(diagramCache.entries, oldest.Value.(*cacheEntry).key)
	}
	diagramCache.Unlock()

	return lines, err
}
