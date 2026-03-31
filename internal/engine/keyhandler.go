package engine

import "strconv"

// ParseResult is the outcome of feeding one key in normal mode.
type ParseResult int

const (
	// ParseInvalid means the key was not handled (browser may use default behavior).
	ParseInvalid ParseResult = iota
	// ParseIncomplete means more keys are needed to complete a sequence (e.g. first "g" of "gg", or digits in a count).
	ParseIncomplete
	// ParseExecuted means a command or motion ran.
	ParseExecuted
)

// normalKeyHandler implements normal-mode key parsing: counts, simple commands, motions, and multi-key motions.
type normalKeyHandler struct {
	motions   *MotionRegistry
	commands  *SimpleCommandRegistry
	operators *OperatorRegistry

	countDigits string

	pending             []string
	pendingMotionCtx    MotionContext
	hasPendingMotionCtx bool

	pendingOp             string
	pendingOpCount        int
	pendingOpCountGiven   bool

	pendingOpMotion []string
	opCount2Digits  string
}

func newNormalKeyHandler() *normalKeyHandler {
	return &normalKeyHandler{
		motions:        NewMotionRegistry(),
		commands:       NewSimpleCommandRegistry(),
		operators:      NewOperatorRegistry(),
		pendingOpCount: 1,
	}
}

func isDigitKey(key string) bool {
	if len(key) != 1 {
		return false
	}
	c := key[0]
	return c >= '0' && c <= '9'
}

func (kh *normalKeyHandler) takeCount() (n int, countGiven bool) {
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

// Feed handles one key in normal mode. It mutates eng (cursor, mode) when a command or motion runs.
func (kh *normalKeyHandler) Feed(eng *Engine, key string) ParseResult {
	if len(kh.pending) > 0 {
		return kh.feedWithPending(eng, key)
	}
	if kh.pendingOp != "" {
		return kh.feedWithOperatorPending(eng, key)
	}

	if isDigitKey(key) {
		if key == "0" && kh.countDigits == "" {
			return kh.execLoneZeroMotion(eng)
		}
		kh.countDigits += key
		return ParseIncomplete
	}

	n, countGiven := kh.takeCount()

	if fn, ok := kh.commands.Get(key); ok {
		fn(eng, CommandContext{Count: n, CountGiven: countGiven})
		return ParseExecuted
	}

	if kh.operators.IsOperator(key) {
		kh.pendingOp = key
		kh.pendingOpCount = n
		kh.pendingOpCountGiven = countGiven
		return ParseIncomplete
	}

	mr, minfo := kh.motions.Lookup([]string{key})
	switch mr {
	case MatchExact:
		res := minfo.Fn(eng, MotionContext{Count: n, CountGiven: countGiven})
		eng.moveCursorTo(res.TargetX, res.TargetY)
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

func (kh *normalKeyHandler) execLoneZeroMotion(eng *Engine) ParseResult {
	mr, minfo := kh.motions.Lookup([]string{"0"})
	if mr != MatchExact {
		return ParseInvalid
	}
	res := minfo.Fn(eng, MotionContext{Count: 1, CountGiven: false})
	eng.moveCursorTo(res.TargetX, res.TargetY)
	return ParseExecuted
}

func (kh *normalKeyHandler) feedWithPending(eng *Engine, key string) ParseResult {
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

func (kh *normalKeyHandler) feedWithOperatorPending(eng *Engine, key string) ParseResult {
	if key == KeyEsc {
		kh.clearOperatorPending()
		return ParseExecuted
	}
	if len(kh.pendingOpMotion) > 0 {
		return kh.feedWithOperatorPendingMotion(eng, key)
	}

	if isDigitKey(key) {
		if key == "0" && kh.opCount2Digits == "" {
			mr, minfo := kh.motions.Lookup([]string{"0"})
			if mr == MatchExact {
				mctx := kh.motionContextForOperator()
				op := kh.pendingOp
				kh.clearOperatorPending()
				eng.ExecuteOperatorWithMotion(op, minfo, mctx)
				return ParseExecuted
			}
		}
		kh.opCount2Digits += key
		return ParseIncomplete
	}

	if key == kh.pendingOp {
		op := kh.pendingOp
		ctx := OperatorContext{
			Count:      kh.pendingOpCount,
			CountGiven: kh.pendingOpCountGiven,
		}
		kh.clearOperatorPending()
		eng.ExecuteLinewiseDoubled(op, ctx)
		return ParseExecuted
	}

	mr, minfo := kh.motions.Lookup([]string{key})
	switch mr {
	case MatchExact:
		mctx := kh.motionContextForOperator()
		op := kh.pendingOp
		kh.clearOperatorPending()
		eng.ExecuteOperatorWithMotion(op, minfo, mctx)
		return ParseExecuted
	case MatchPrefix:
		kh.pendingOpMotion = append(kh.pendingOpMotion[:0], key)
		return ParseIncomplete
	case MatchNone:
		kh.clearOperatorPending()
		return kh.Feed(eng, key)
	default:
		kh.clearOperatorPending()
		return ParseInvalid
	}
}

func (kh *normalKeyHandler) feedWithOperatorPendingMotion(eng *Engine, key string) ParseResult {
	if key == KeyEsc {
		kh.pendingOpMotion = kh.pendingOpMotion[:0]
		kh.clearOperatorPending()
		return ParseExecuted
	}
	kh.pendingOpMotion = append(kh.pendingOpMotion, key)
	mr, minfo := kh.motions.Lookup(kh.pendingOpMotion)
	switch mr {
	case MatchExact:
		ctx := kh.motionContextForOperator()
		op := kh.pendingOp
		kh.pendingOpMotion = kh.pendingOpMotion[:0]
		kh.clearOperatorPending()
		eng.ExecuteOperatorWithMotion(op, minfo, ctx)
		return ParseExecuted
	case MatchPrefix:
		return ParseIncomplete
	case MatchNone:
		kh.pendingOpMotion = kh.pendingOpMotion[:0]
		kh.clearOperatorPending()
		return kh.Feed(eng, key)
	default:
		kh.pendingOpMotion = kh.pendingOpMotion[:0]
		kh.clearOperatorPending()
		return ParseInvalid
	}
}

func (kh *normalKeyHandler) clearOperatorPending() {
	kh.pendingOp = ""
	kh.pendingOpCount = 1
	kh.pendingOpCountGiven = false
	kh.pendingOpMotion = kh.pendingOpMotion[:0]
	kh.opCount2Digits = ""
}

func (kh *normalKeyHandler) motionContextForOperator() MotionContext {
	motionBase := 1
	if kh.opCount2Digits != "" {
		n, err := strconv.Atoi(kh.opCount2Digits)
		kh.opCount2Digits = ""
		if err == nil && n >= 1 {
			motionBase = n
		}
	}
	effective := kh.pendingOpCount * motionBase
	if effective < 1 {
		effective = 1
	}
	countGiven := motionBase > 1 || (kh.pendingOpCount > 1 && kh.pendingOpCountGiven)
	return MotionContext{
		Count:      effective,
		CountGiven: countGiven,
	}
}

// ResetPending clears an incomplete multi-key sequence and count buffer (e.g. after mode switch).
func (kh *normalKeyHandler) ResetPending() {
	kh.pending = kh.pending[:0]
	kh.countDigits = ""
	kh.pendingMotionCtx = MotionContext{}
	kh.hasPendingMotionCtx = false
	kh.clearOperatorPending()
}
