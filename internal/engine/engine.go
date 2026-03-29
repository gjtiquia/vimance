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

func (eng *Engine) KeyPress(key string) {
	switch eng.mode {

	case ModeNormal:
		switch key {

		case "i":
			eng.setMode(ModeInsert, InsertPositionBefore)

		case "a":
			eng.setMode(ModeInsert, InsertPositionAfter)

		case "v":
			eng.setMode(ModeVisual, InsertPositionNone)

		case "h", "b":
			if eng.cursorX > 0 {
				eng.cursorX--
				eng.notifyCursorMoved()
			}

		case "l", "w", "e":
			if eng.cursorX < eng.cols-1 {
				eng.cursorX++
				eng.notifyCursorMoved()
			}

		case "j":
			if eng.cursorY < eng.rows-1 {
				eng.cursorY++
				eng.notifyCursorMoved()
			}

		case "k":
			if eng.cursorY > 0 {
				eng.cursorY--
				eng.notifyCursorMoved()
			}
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
