package engine

// MotionContext is passed to motions after optional count digits (e.g. 3j, 5G).
// Count is always at least 1. CountGiven is true when the user typed one or more digits before the motion.
type MotionContext struct {
	Count      int
	CountGiven bool
}

// MotionResult is the target cell after a motion (pure, no side effects).
type MotionResult struct {
	TargetX int
	TargetY int
}

// MotionFunc computes where the cursor would move without mutating the engine.
type MotionFunc func(eng *Engine, ctx MotionContext) MotionResult

// MotionInfo describes a registered motion.
type MotionInfo struct {
	Fn       MotionFunc
	Linewise bool
	// Inclusive applies to non-linewise horizontal ranges: true for $ and 0; false for h/l/w/e/b (exclusive like vim).
	Inclusive bool
}

// MotionRegistry maps key sequences (one or more KeyPress strings) to motions.
type MotionRegistry struct {
	trie Trie
}

func NewMotionRegistry() *MotionRegistry {
	r := &MotionRegistry{}
	r.registerBuiltins()
	return r
}

func (r *MotionRegistry) Insert(keys []string, fn MotionFunc, linewise, inclusive bool) {
	r.trie.Insert(keys, MotionInfo{Fn: fn, Linewise: linewise, Inclusive: inclusive})
}

// Lookup returns how keys match and the motion info when Exact.
func (r *MotionRegistry) Lookup(keys []string) (MatchResult, *MotionInfo) {
	mr, v := r.trie.Match(keys)
	if mr != MatchExact {
		return mr, nil
	}
	info, _ := v.(MotionInfo)
	return MatchExact, &info
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
	left := func(eng *Engine, ctx MotionContext) MotionResult {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		x := eng.cursorX - n
		if x < 0 {
			x = 0
		}
		return MotionResult{TargetX: x, TargetY: eng.cursorY}
	}
	right := func(eng *Engine, ctx MotionContext) MotionResult {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		x := eng.cursorX + n
		if x >= eng.cols {
			x = eng.cols - 1
		}
		return MotionResult{TargetX: x, TargetY: eng.cursorY}
	}
	down := func(eng *Engine, ctx MotionContext) MotionResult {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		y := eng.cursorY + n
		if y >= eng.rows {
			y = eng.rows - 1
		}
		return MotionResult{TargetX: eng.cursorX, TargetY: y}
	}
	up := func(eng *Engine, ctx MotionContext) MotionResult {
		n := ctx.Count
		if n < 1 {
			n = 1
		}
		y := eng.cursorY - n
		if y < 0 {
			y = 0
		}
		return MotionResult{TargetX: eng.cursorX, TargetY: y}
	}

	r.Insert([]string{"h"}, left, false, false)
	r.Insert([]string{"b"}, left, false, false)
	r.Insert([]string{"ArrowLeft"}, left, false, false)

	r.Insert([]string{"l"}, right, false, false)
	r.Insert([]string{"w"}, right, false, false)
	r.Insert([]string{"e"}, right, false, false)
	r.Insert([]string{"ArrowRight"}, right, false, false)

	r.Insert([]string{"j"}, down, true, true)
	r.Insert([]string{"ArrowDown"}, down, true, true)

	r.Insert([]string{"k"}, up, true, true)
	r.Insert([]string{"ArrowUp"}, up, true, true)

	r.Insert([]string{"g", "g"}, func(eng *Engine, ctx MotionContext) MotionResult {
		var target int
		if !ctx.CountGiven {
			target = 0
		} else {
			target = ctx.Count - 1
		}
		target = clampInt(target, 0, eng.rows-1)
		return MotionResult{TargetX: eng.cursorX, TargetY: target}
	}, true, true)

	r.Insert([]string{"G"}, func(eng *Engine, ctx MotionContext) MotionResult {
		var target int
		if !ctx.CountGiven {
			target = eng.rows - 1
		} else {
			target = ctx.Count - 1
		}
		target = clampInt(target, 0, eng.rows-1)
		return MotionResult{TargetX: eng.cursorX, TargetY: target}
	}, true, true)

	r.Insert([]string{"0"}, func(eng *Engine, ctx MotionContext) MotionResult {
		_ = ctx
		return MotionResult{TargetX: 0, TargetY: eng.cursorY}
	}, false, true)

	r.Insert([]string{"$"}, func(eng *Engine, ctx MotionContext) MotionResult {
		_ = ctx
		return MotionResult{TargetX: eng.cols - 1, TargetY: eng.cursorY}
	}, false, true)
}
