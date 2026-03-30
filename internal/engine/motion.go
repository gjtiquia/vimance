package engine

// MotionContext is passed to motions after optional count digits (e.g. 3j, 5G).
// Count is always at least 1. CountGiven is true when the user typed one or more digits before the motion.
type MotionContext struct {
	Count      int
	CountGiven bool
}

// MotionFunc applies a cursor motion.
type MotionFunc func(eng *Engine, ctx MotionContext)

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

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func (r *MotionRegistry) registerBuiltins() {
	left := func(eng *Engine, ctx MotionContext) {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		x := eng.cursorX - n
		if x < 0 {
			x = 0
		}
		eng.moveCursorTo(x, eng.cursorY)
	}
	right := func(eng *Engine, ctx MotionContext) {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		x := eng.cursorX + n
		if x >= eng.cols {
			x = eng.cols - 1
		}
		eng.moveCursorTo(x, eng.cursorY)
	}
	down := func(eng *Engine, ctx MotionContext) {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		y := eng.cursorY + n
		if y >= eng.rows {
			y = eng.rows - 1
		}
		eng.moveCursorTo(eng.cursorX, y)
	}
	up := func(eng *Engine, ctx MotionContext) {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		y := eng.cursorY - n
		if y < 0 {
			y = 0
		}
		eng.moveCursorTo(eng.cursorX, y)
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

	r.Insert([]string{"g", "g"}, func(eng *Engine, ctx MotionContext) {
		var target int
		if !ctx.CountGiven {
			target = 0
		} else {
			target = ctx.Count - 1
		}
		target = clampInt(target, 0, eng.rows-1)
		eng.moveCursorTo(eng.cursorX, target)
	})

	r.Insert([]string{"G"}, func(eng *Engine, ctx MotionContext) {
		var target int
		if !ctx.CountGiven {
			target = eng.rows - 1
		} else {
			target = ctx.Count - 1
		}
		target = clampInt(target, 0, eng.rows-1)
		eng.moveCursorTo(eng.cursorX, target)
	})

	r.Insert([]string{"0"}, func(eng *Engine, ctx MotionContext) {
		_ = ctx
		eng.moveCursorTo(0, eng.cursorY)
	})

	r.Insert([]string{"$"}, func(eng *Engine, ctx MotionContext) {
		_ = ctx
		eng.moveCursorTo(eng.cols-1, eng.cursorY)
	})
}
