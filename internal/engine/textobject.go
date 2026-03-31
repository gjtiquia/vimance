package engine

// TextObjectRange is a contiguous horizontal range of cells on one row (spreadsheet "word" = one cell).
type TextObjectRange struct {
	Y      int
	StartX int
	EndX   int
}

// TextObjectFunc computes the cell range for a text object without mutating the engine.
type TextObjectFunc func(eng *Engine) TextObjectRange

// TextObjectInfo describes a registered text object.
type TextObjectInfo struct {
	Fn TextObjectFunc
}

// TextObjectRegistry maps key sequences (e.g. i then w) to text objects.
type TextObjectRegistry struct {
	trie Trie
}

func NewTextObjectRegistry() *TextObjectRegistry {
	r := &TextObjectRegistry{}
	r.registerBuiltins()
	return r
}

func (r *TextObjectRegistry) Insert(keys []string, fn TextObjectFunc) {
	r.trie.Insert(keys, TextObjectInfo{Fn: fn})
}

// Lookup returns how keys match and the text object when Exact.
func (r *TextObjectRegistry) Lookup(keys []string) (MatchResult, *TextObjectInfo) {
	mr, v := r.trie.Match(keys)
	if mr != MatchExact {
		return mr, nil
	}
	info, ok := v.(TextObjectInfo)
	if !ok {
		return MatchNone, nil
	}
	return MatchExact, &info
}

func currentCellRange(eng *Engine) TextObjectRange {
	return TextObjectRange{
		Y:      eng.cursorY,
		StartX: eng.cursorX,
		EndX:   eng.cursorX,
	}
}

func (r *TextObjectRegistry) registerBuiltins() {
	// iw / aw: inner word and around word both map to the current cell (no intra-cell words in grid).
	r.Insert([]string{"i", "w"}, currentCellRange)
	r.Insert([]string{"a", "w"}, currentCellRange)
}
