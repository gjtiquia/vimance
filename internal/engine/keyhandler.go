package engine

// ParseResult is the outcome of feeding one key in normal mode.
type ParseResult int

const (
	// ParseInvalid means the key was not handled (browser may use default behavior).
	ParseInvalid ParseResult = iota
	// ParseIncomplete means more keys are needed to complete a sequence (e.g. first "g" of "gg").
	ParseIncomplete
	// ParseExecuted means a command or motion ran.
	ParseExecuted
)

// normalKeyHandler implements normal-mode key parsing: simple commands, motions, and multi-key motions.
type normalKeyHandler struct {
	motions  *MotionRegistry
	commands *SimpleCommandRegistry
	pending  []string
}

func newNormalKeyHandler() *normalKeyHandler {
	return &normalKeyHandler{
		motions:  NewMotionRegistry(),
		commands: NewSimpleCommandRegistry(),
	}
}

// Feed handles one key in normal mode. It mutates eng (cursor, mode) when a command or motion runs.
func (kh *normalKeyHandler) Feed(eng *Engine, key string) ParseResult {
	if len(kh.pending) > 0 {
		return kh.feedWithPending(eng, key)
	}

	if fn, ok := kh.commands.Get(key); ok {
		fn(eng)
		return ParseExecuted
	}

	mr, mfn := kh.motions.Lookup([]string{key})
	switch mr {
	case MatchExact:
		mfn(eng, 1)
		return ParseExecuted
	case MatchPrefix:
		kh.pending = append(kh.pending[:0], key)
		return ParseIncomplete
	case MatchNone:
		return ParseInvalid
	default:
		return ParseInvalid
	}
}

func (kh *normalKeyHandler) feedWithPending(eng *Engine, key string) ParseResult {
	kh.pending = append(kh.pending, key)
	mr, mfn := kh.motions.Lookup(kh.pending)
	switch mr {
	case MatchExact:
		mfn(eng, 1)
		kh.pending = kh.pending[:0]
		return ParseExecuted
	case MatchPrefix:
		return ParseIncomplete
	case MatchNone:
		kh.pending = kh.pending[:0]
		return kh.Feed(eng, key)
	default:
		kh.pending = kh.pending[:0]
		return ParseInvalid
	}
}

// ResetPending clears an incomplete multi-key sequence (e.g. after mode switch).
func (kh *normalKeyHandler) ResetPending() {
	kh.pending = kh.pending[:0]
}
