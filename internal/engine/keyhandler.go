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

	mr, mfn := kh.motions.Lookup([]string{key})
	switch mr {
	case MatchExact:
		mfn(eng, MotionContext{Count: n, CountGiven: countGiven})
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
	mr, mfn := kh.motions.Lookup([]string{"0"})
	if mr != MatchExact {
		return ParseInvalid
	}
	mfn(eng, MotionContext{Count: 1, CountGiven: false})
	return ParseExecuted
}

func (kh *normalKeyHandler) feedWithPending(eng *Engine, key string) ParseResult {
	kh.pending = append(kh.pending, key)
	mr, mfn := kh.motions.Lookup(kh.pending)
	switch mr {
	case MatchExact:
		ctx := MotionContext{Count: 1, CountGiven: false}
		if kh.hasPendingMotionCtx {
			ctx = kh.pendingMotionCtx
			kh.hasPendingMotionCtx = false
		}
		mfn(eng, ctx)
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
	kh.clearOperatorPending()
	return kh.Feed(eng, key)
}

func (kh *normalKeyHandler) clearOperatorPending() {
	kh.pendingOp = ""
	kh.pendingOpCount = 1
	kh.pendingOpCountGiven = false
}

// ResetPending clears an incomplete multi-key sequence and count buffer (e.g. after mode switch).
func (kh *normalKeyHandler) ResetPending() {
	kh.pending = kh.pending[:0]
	kh.countDigits = ""
	kh.pendingMotionCtx = MotionContext{}
	kh.hasPendingMotionCtx = false
	kh.clearOperatorPending()
}
