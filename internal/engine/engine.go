package engine

type Engine struct {
	listeners []EventListener
	mode      Mode
	cursorX   int
	cursorY   int
	cols      int
	rows      int

	normalKH        *normalKeyHandler
	lastKeyCaptured bool
}

type EventListener interface {
	OnModeChanged(mode Mode, insertPosition InsertPosition)
	OnCursorMoved(x, y int)
}

type Mode string

const (
	ModeNormal Mode = "n"
	ModeInsert Mode = "i"
	ModeVisual Mode = "v"
)

// InsertPosition is a hint for where the insert caret should start when entering insert mode.
type InsertPosition string

const (
	InsertPositionNone   InsertPosition = ""
	InsertPositionBefore InsertPosition = "before"
	InsertPositionAfter  InsertPosition = "after"
	// InsertPositionHighlight selects the whole cell text so typing overwrites (Enter/double-click); the value is not cleared until the user types.
	InsertPositionHighlight InsertPosition = "highlight"
)

const KeyEsc string = "Escape"

// New creates an engine for a grid with cols columns and rows rows (0-based indices up to cols-1, rows-1).
func New(cols, rows int) Engine {
	return Engine{
		listeners:       []EventListener{},
		mode:            ModeNormal,
		cursorX:         0,
		cursorY:         0,
		cols:            cols,
		rows:            rows,
		normalKH:        newNormalKeyHandler(),
		lastKeyCaptured: false,
	}
}

func (eng *Engine) AddListener(listener EventListener) {
	eng.listeners = append(eng.listeners, listener)
}

func (eng *Engine) Mode() Mode {
	return eng.mode
}

func (eng *Engine) CursorX() int {
	return eng.cursorX
}

func (eng *Engine) CursorY() int {
	return eng.cursorY
}

// LastKeyCaptured reports whether the last KeyPress consumed the key (motion/command, incomplete prefix, or Escape in insert/visual).
func (eng *Engine) LastKeyCaptured() bool {
	return eng.lastKeyCaptured
}

func (eng *Engine) setMode(mode Mode, insertPosition InsertPosition) {
	if eng.mode != mode {
		eng.normalKH.ResetPending()
	}
	eng.mode = mode

	for _, listener := range eng.listeners {
		listener.OnModeChanged(mode, insertPosition)
	}
}

func (eng *Engine) notifyCursorMoved() {
	for _, listener := range eng.listeners {
		listener.OnCursorMoved(eng.cursorX, eng.cursorY)
	}
}

// moveCursorTo sets the cursor to (x, y) if in bounds and different from the current cell.
// All grid cursor movement should go through this so listeners are notified consistently.
func (eng *Engine) moveCursorTo(x, y int) bool {
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return false
	}
	if eng.cursorX == x && eng.cursorY == y {
		return false
	}
	eng.cursorX = x
	eng.cursorY = y
	eng.notifyCursorMoved()
	return true
}

func (eng *Engine) KeyPress(key string) {
	eng.lastKeyCaptured = false

	switch eng.mode {

	case ModeNormal:
		r := eng.normalKH.Feed(eng, key)
		if r != ParseInvalid {
			eng.lastKeyCaptured = true
		}

	case ModeInsert:
		if key == KeyEsc {
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.lastKeyCaptured = true
		}

	case ModeVisual:
		if key == KeyEsc {
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.lastKeyCaptured = true
		}
	}
}

// SetCursor moves the cursor to (x, y). If the engine is in insert mode, it switches to normal first
// (unless the target cell is already the current cell — then the call is a no-op so redundant pointer
// events do not exit insert mode). Out-of-bounds coordinates are ignored.
func (eng *Engine) SetCursor(x, y int) {
	if eng.mode == ModeInsert {
		if eng.cursorX == x && eng.cursorY == y {
			return
		}
		eng.setMode(ModeNormal, InsertPositionNone)
	}
	eng.normalKH.ResetPending()
	eng.moveCursorTo(x, y)
}

// SetCursorAndEdit moves the cursor to (x, y) and enters insert mode with the full cell selected for overwrite (same as Enter).
// Exits insert mode first if needed. Out-of-bounds coordinates are ignored (insert mode is still exited if active).
func (eng *Engine) SetCursorAndEdit(x, y int) {
	if eng.mode == ModeInsert {
		eng.setMode(ModeNormal, InsertPositionNone)
	}
	eng.normalKH.ResetPending()
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return
	}
	eng.moveCursorTo(x, y)
	eng.setMode(ModeInsert, InsertPositionHighlight)
}
