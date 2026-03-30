package engine

type Engine struct {
	listeners []EventListener
	mode      Mode
	cursorX   int
	cursorY   int
	cols      int
	rows      int
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
)

const KeyEsc string = "Escape"

// New creates an engine for a grid with cols columns and rows rows (0-based indices up to cols-1, rows-1).
func New(cols, rows int) Engine {
	return Engine{
		listeners: []EventListener{},
		mode:      ModeNormal,
		cursorX:   0,
		cursorY:   0,
		cols:      cols,
		rows:      rows,
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

func (eng *Engine) setMode(mode Mode, insertPosition InsertPosition) {
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
	switch eng.mode {

	case ModeNormal:
		switch key {

		case "i":
			eng.setMode(ModeInsert, InsertPositionBefore)

		case "a", "Enter":
			eng.setMode(ModeInsert, InsertPositionAfter)

		case "v":
			eng.setMode(ModeVisual, InsertPositionNone)

		case "h", "b", "ArrowLeft":
			eng.moveCursorTo(eng.cursorX-1, eng.cursorY)

		case "l", "w", "e", "ArrowRight":
			eng.moveCursorTo(eng.cursorX+1, eng.cursorY)

		case "j", "ArrowDown":
			eng.moveCursorTo(eng.cursorX, eng.cursorY+1)

		case "k", "ArrowUp":
			eng.moveCursorTo(eng.cursorX, eng.cursorY-1)
		}

	case ModeInsert:
		if key == KeyEsc {
			eng.setMode(ModeNormal, InsertPositionNone)
		}

	case ModeVisual:
		if key == KeyEsc {
			eng.setMode(ModeNormal, InsertPositionNone)
		}
	}
}

// SetCursor moves the cursor to (x, y). If the engine is in insert mode, it switches to normal first.
// Out-of-bounds coordinates are ignored. Cursor listeners are notified only when the position changes.
func (eng *Engine) SetCursor(x, y int) {
	if eng.mode == ModeInsert {
		eng.setMode(ModeNormal, InsertPositionNone)
	}
	eng.moveCursorTo(x, y)
}

// SetCursorAndEdit moves the cursor to (x, y) and enters insert mode with the caret after the cell text.
// Exits insert mode first if needed. Out-of-bounds coordinates are ignored (insert mode is still exited if active).
func (eng *Engine) SetCursorAndEdit(x, y int) {
	if eng.mode == ModeInsert {
		eng.setMode(ModeNormal, InsertPositionNone)
	}
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return
	}
	eng.moveCursorTo(x, y)
	eng.setMode(ModeInsert, InsertPositionAfter)
}
