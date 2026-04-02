package engine

import "strconv"

// visualKeyHandler handles keys in visual sub-modes (v, V, Ctrl+v): motions extend selection,
// operators act on the selection immediately, and v/V/Ctrl+v toggle or exit.
type visualKeyHandler struct {
	motions   *MotionRegistry
	operators *OperatorRegistry

	countDigits string

	pending             []string
	pendingMotionCtx    MotionContext
	hasPendingMotionCtx bool
}

func newVisualKeyHandler(motions *MotionRegistry, operators *OperatorRegistry) *visualKeyHandler {
	return &visualKeyHandler{
		motions:   motions,
		operators: operators,
	}
}

func (kh *visualKeyHandler) takeCount() (n int, countGiven bool) {
	if kh.countDigits == "" {
		return 1, false
	}
	n, err := strconv.Atoi(kh.countDigits)
	kh.countDigits = ""
	if err != nil || n < 1 {
		return 1, true
	}
	return n, true
}

// Feed handles one key in visual mode (any of ModeVisual, ModeVisualLine, ModeVisualBlock).
func (kh *visualKeyHandler) Feed(eng *Engine, key string) ParseResult {
	if len(kh.pending) > 0 {
		return kh.feedWithPending(eng, key)
	}
	if key == KeyEsc {
		eng.exitVisualMode()
		return ParseExecuted
	}
	if kh.handleVisualModeToggle(eng, key) {
		return ParseExecuted
	}
	if key == "o" {
		kh.swapAnchorAndCursor(eng)
		return ParseExecuted
	}
	if isDigitKey(key) {
		if key == "0" && kh.countDigits == "" {
			return kh.execLoneZeroMotion(eng)
		}
		kh.countDigits += key
		return ParseIncomplete
	}

	n, countGiven := kh.takeCount()

	if kh.operators.IsOperator(key) {
		eng.ExecuteVisualOperator(key)
		return ParseExecuted
	}
	if key == "x" {
		eng.ExecuteVisualOperator("x")
		return ParseExecuted
	}

	mr, minfo := kh.motions.Lookup([]string{key})
	switch mr {
	case MatchExact:
		res := minfo.Fn(eng, MotionContext{Count: n, CountGiven: countGiven})
		eng.moveCursorTo(res.TargetX, res.TargetY)
		eng.notifySelectionChanged()
		return ParseExecuted
	case MatchPrefix:
		kh.pending = append(kh.pending[:0], key)
		kh.pendingMotionCtx = MotionContext{Count: n, CountGiven: countGiven}
		kh.hasPendingMotionCtx = true
		return ParseIncomplete
	case MatchNone:
		return ParseInvalid
	default:
		return ParseInvalid
	}
}

func (kh *visualKeyHandler) handleVisualModeToggle(eng *Engine, key string) bool {
	m := eng.Mode()
	switch key {
	case "v":
		switch m {
		case ModeVisual:
			eng.exitVisualMode()
			return true
		case ModeVisualLine, ModeVisualBlock:
			eng.setMode(ModeVisual, InsertPositionNone)
			eng.notifySelectionChanged()
			return true
		}
	case "V":
		switch m {
		case ModeVisualLine:
			eng.exitVisualMode()
			return true
		case ModeVisual, ModeVisualBlock:
			eng.setMode(ModeVisualLine, InsertPositionNone)
			eng.notifySelectionChanged()
			return true
		}
	case "Ctrl+v":
		switch m {
		case ModeVisualBlock:
			eng.exitVisualMode()
			return true
		case ModeVisual, ModeVisualLine:
			eng.setMode(ModeVisualBlock, InsertPositionNone)
			eng.notifySelectionChanged()
			return true
		}
	}
	return false
}

func (kh *visualKeyHandler) swapAnchorAndCursor(eng *Engine) {
	ax, ay := eng.visualAnchorX, eng.visualAnchorY
	cx, cy := eng.cursorX, eng.cursorY
	if ax == cx && ay == cy {
		return
	}
	eng.visualAnchorX, eng.visualAnchorY = cx, cy
	eng.moveCursorTo(ax, ay)
	eng.notifySelectionChanged()
}

func (kh *visualKeyHandler) execLoneZeroMotion(eng *Engine) ParseResult {
	mr, minfo := kh.motions.Lookup([]string{"0"})
	if mr != MatchExact {
		return ParseInvalid
	}
	res := minfo.Fn(eng, MotionContext{Count: 1, CountGiven: false})
	eng.moveCursorTo(res.TargetX, res.TargetY)
	eng.notifySelectionChanged()
	return ParseExecuted
}

func (kh *visualKeyHandler) feedWithPending(eng *Engine, key string) ParseResult {
	kh.pending = append(kh.pending, key)
	mr, minfo := kh.motions.Lookup(kh.pending)
	switch mr {
	case MatchExact:
		ctx := MotionContext{Count: 1, CountGiven: false}
		if kh.hasPendingMotionCtx {
			ctx = kh.pendingMotionCtx
			kh.hasPendingMotionCtx = false
		}
		res := minfo.Fn(eng, ctx)
		eng.moveCursorTo(res.TargetX, res.TargetY)
		eng.notifySelectionChanged()
		kh.pending = kh.pending[:0]
		kh.pendingMotionCtx = MotionContext{}
		return ParseExecuted
	case MatchPrefix:
		return ParseIncomplete
	case MatchNone:
		kh.pending = kh.pending[:0]
		kh.pendingMotionCtx = MotionContext{}
		kh.hasPendingMotionCtx = false
		return kh.Feed(eng, key)
	default:
		kh.pending = kh.pending[:0]
		kh.pendingMotionCtx = MotionContext{}
		kh.hasPendingMotionCtx = false
		return ParseInvalid
	}
}

// ResetPending clears incomplete motion / count state (e.g. on mode switch).
func (kh *visualKeyHandler) ResetPending() {
	kh.pending = kh.pending[:0]
	kh.countDigits = ""
	kh.pendingMotionCtx = MotionContext{}
	kh.hasPendingMotionCtx = false
}
