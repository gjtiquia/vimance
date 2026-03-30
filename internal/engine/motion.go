package engine

// MotionFunc applies a cursor motion. count is reserved for repeat (Phase 2); use 1 for now.
type MotionFunc func(eng *Engine, count int)

// MotionRegistry maps key sequences (one or more KeyPress strings) to motions.
type MotionRegistry struct {
	trie Trie
}

func NewMotionRegistry() *MotionRegistry {
	r := &MotionRegistry{}
	r.registerBuiltins()
	return r
}

func (r *MotionRegistry) Insert(keys []string, fn MotionFunc) {
	r.trie.Insert(keys, fn)
}

// Lookup returns how keys match and the motion when Exact.
func (r *MotionRegistry) Lookup(keys []string) (MatchResult, MotionFunc) {
	mr, v := r.trie.Match(keys)
	if mr != MatchExact {
		return mr, nil
	}
	fn, _ := v.(MotionFunc)
	return MatchExact, fn
}

func (r *MotionRegistry) registerBuiltins() {
	left := func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX-1, eng.cursorY)
	}
	right := func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX+1, eng.cursorY)
	}
	down := func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX, eng.cursorY+1)
	}
	up := func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX, eng.cursorY-1)
	}

	r.Insert([]string{"h"}, left)
	r.Insert([]string{"b"}, left)
	r.Insert([]string{"ArrowLeft"}, left)

	r.Insert([]string{"l"}, right)
	r.Insert([]string{"w"}, right)
	r.Insert([]string{"e"}, right)
	r.Insert([]string{"ArrowRight"}, right)

	r.Insert([]string{"j"}, down)
	r.Insert([]string{"ArrowDown"}, down)

	r.Insert([]string{"k"}, up)
	r.Insert([]string{"ArrowUp"}, up)

	r.Insert([]string{"g", "g"}, func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX, 0)
	})

	r.Insert([]string{"G"}, func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cursorX, eng.rows-1)
	})

	r.Insert([]string{"0"}, func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(0, eng.cursorY)
	})

	r.Insert([]string{"$"}, func(eng *Engine, count int) {
		_ = count
		eng.moveCursorTo(eng.cols-1, eng.cursorY)
	})
}
