package engine

import "strings"

type Engine struct {
	listeners []EventListener
	mode      Mode
	cursorX   int
	cursorY   int
	cols      int
	rows      int

	cells      [][]string
	dataSource DataSource

	register           Register
	clipboardDelimiter string

	normalKH        *normalKeyHandler
	lastKeyCaptured bool
}

type EventListener interface {
	OnModeChanged(mode Mode, insertPosition InsertPosition)
	OnCursorMoved(x, y int)
	OnBufferChanged()
	OnClipboardWrite(text string)
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

// New builds an engine from DataSource.Load(). The grid must be rectangular and non-empty.
func New(dataSource DataSource) Engine {
	cells := cloneCells(dataSource.Load())
	cols, rows, ok := validateRectangularGrid(cells)
	if !ok {
		panic("engine.New: DataSource.Load must return a non-empty rectangular grid")
	}
	return Engine{
		listeners:            []EventListener{},
		mode:                 ModeNormal,
		cursorX:              0,
		cursorY:              0,
		cols:                 cols,
		rows:                 rows,
		cells:                cells,
		dataSource:           dataSource,
		clipboardDelimiter:   " ",
		normalKH:             newNormalKeyHandler(),
		lastKeyCaptured:      false,
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

// Cols returns the number of columns in the grid.
func (eng *Engine) Cols() int {
	return eng.cols
}

// Rows returns the number of rows in the grid.
func (eng *Engine) Rows() int {
	return eng.rows
}

// CellValue returns the text in cell (x, y).
func (eng *Engine) CellValue(x, y int) (string, bool) {
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return "", false
	}
	return eng.cells[y][x], true
}

// SetCellValue sets cell (x, y). Returns false if out of bounds.
func (eng *Engine) SetCellValue(x, y int, value string) bool {
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return false
	}
	eng.cells[y][x] = value
	return true
}

// CellsSnapshot returns a deep copy of the grid (for RPC / tests).
func (eng *Engine) CellsSnapshot() [][]string {
	return cloneCells(eng.cells)
}

// SaveBuffer persists the current grid via DataSource.Save (no-op for StubDataSource until you add storage).
func (eng *Engine) SaveBuffer() error {
	if eng.dataSource == nil {
		return nil
	}
	return eng.dataSource.Save(cloneCells(eng.cells))
}

// LastKeyCaptured reports whether the last KeyPress consumed the key (motion/command, incomplete prefix, or Escape in insert/visual).
func (eng *Engine) LastKeyCaptured() bool {
	return eng.lastKeyCaptured
}

// RegisterSnapshot returns a copy of the unnamed register (for tests).
func (eng *Engine) RegisterSnapshot() Register {
	return Register{
		Cells:    cloneCells(eng.register.Cells),
		Linewise: eng.register.Linewise,
	}
}

func (eng *Engine) notifyBufferChanged() {
	for _, listener := range eng.listeners {
		listener.OnBufferChanged()
	}
}

func (eng *Engine) notifyClipboardWrite(text string) {
	for _, listener := range eng.listeners {
		listener.OnClipboardWrite(text)
	}
}

func (eng *Engine) formatRegisterClipboard() string {
	if len(eng.register.Cells) == 0 {
		return ""
	}
	d := eng.clipboardDelimiter
	if d == "" {
		d = " "
	}
	lines := make([]string, len(eng.register.Cells))
	for i, row := range eng.register.Cells {
		lines[i] = strings.Join(row, d)
	}
	return strings.Join(lines, "\n")
}

// ExecuteLinewiseDoubled runs dd, yy, or cc (second key matched the operator).
func (eng *Engine) ExecuteLinewiseDoubled(op string, ctx OperatorContext) {
	n := ctx.Count
	if n < 1 {
		n = 1
	}
	switch op {
	case "d":
		eng.linewiseDelete(n)
	case "y":
		eng.linewiseYank(n)
	case "c":
		eng.linewiseChange(n)
	}
}

func (eng *Engine) linewiseDelete(n int) {
	start := eng.cursorY
	if start == 0 {
		return // header row protected
	}
	end := start + n
	if end > eng.rows {
		end = eng.rows
	}
	if start >= eng.rows {
		return
	}
	toCopy := eng.cells[start:end]
	eng.register = Register{Cells: cloneCells(toCopy), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())

	eng.cells = append(eng.cells[:start], eng.cells[end:]...)
	eng.rows = len(eng.cells)
	if eng.rows == 0 {
		panic("engine: grid became empty after delete")
	}
	eng.cols = len(eng.cells[0])

	cx := eng.cursorX
	if cx >= eng.cols {
		cx = eng.cols - 1
	}
	cy := eng.cursorY
	if cy >= eng.rows {
		cy = eng.rows - 1
	}
	eng.moveCursorTo(cx, cy)
	eng.notifyBufferChanged()
}

func (eng *Engine) linewiseYank(n int) {
	start := eng.cursorY
	end := start + n
	if end > eng.rows {
		end = eng.rows
	}
	if start >= eng.rows {
		return
	}
	slice := eng.cells[start:end]
	eng.register = Register{Cells: cloneCells(slice), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
}

func (eng *Engine) linewiseChange(n int) {
	start := eng.cursorY
	if start == 0 {
		return // header protected
	}
	end := start + n
	if end > eng.rows {
		end = eng.rows
	}
	if start >= eng.rows {
		return
	}
	slice := eng.cells[start:end]
	eng.register = Register{Cells: cloneCells(slice), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())

	for y := start; y < end; y++ {
		for x := 0; x < eng.cols; x++ {
			eng.cells[y][x] = ""
		}
	}
	eng.notifyBufferChanged()
	eng.setMode(ModeInsert, InsertPositionHighlight)
}

// DeleteCharUnderCursor clears the current cell and stores the old value in the register (non-linewise).
func (eng *Engine) DeleteCharUnderCursor() {
	v, ok := eng.CellValue(eng.cursorX, eng.cursorY)
	if !ok {
		return
	}
	eng.register = Register{Cells: [][]string{{v}}, Linewise: false}
	eng.SetCellValue(eng.cursorX, eng.cursorY, "")
	eng.notifyBufferChanged()
}

// PasteAfter puts register contents at or below the cursor (linewise or character).
func (eng *Engine) PasteAfter() {
	if eng.register.Linewise {
		if len(eng.register.Cells) == 0 {
			return
		}
		newRows := cloneCells(eng.register.Cells)
		insertAt := eng.cursorY + 1
		if insertAt > eng.rows {
			insertAt = eng.rows
		}
		eng.cells = spliceRowsAt(eng.cells, insertAt, newRows)
		eng.rows = len(eng.cells)
		eng.cols = len(eng.cells[0])
		lastY := insertAt + len(newRows) - 1
		if lastY >= eng.rows {
			lastY = eng.rows - 1
		}
		x := eng.cursorX
		if x >= eng.cols {
			x = eng.cols - 1
		}
		eng.moveCursorTo(x, lastY)
		eng.notifyBufferChanged()
		return
	}
	if len(eng.register.Cells) != 1 || len(eng.register.Cells[0]) != 1 {
		return
	}
	v := eng.register.Cells[0][0]
	eng.SetCellValue(eng.cursorX, eng.cursorY, v)
	eng.notifyBufferChanged()
}

// PasteBefore puts register contents above the cursor (linewise) or replaces current cell (non-linewise).
func (eng *Engine) PasteBefore() {
	if eng.register.Linewise {
		if len(eng.register.Cells) == 0 {
			return
		}
		newRows := cloneCells(eng.register.Cells)
		insertAt := eng.cursorY
		if insertAt < 1 {
			insertAt = 1 // never above header row
		}
		eng.cells = spliceRowsAt(eng.cells, insertAt, newRows)
		eng.rows = len(eng.cells)
		eng.cols = len(eng.cells[0])
		lastY := insertAt + len(newRows) - 1
		if lastY >= eng.rows {
			lastY = eng.rows - 1
		}
		x := eng.cursorX
		if x >= eng.cols {
			x = eng.cols - 1
		}
		eng.moveCursorTo(x, lastY)
		eng.notifyBufferChanged()
		return
	}
	if len(eng.register.Cells) != 1 || len(eng.register.Cells[0]) != 1 {
		return
	}
	v := eng.register.Cells[0][0]
	eng.SetCellValue(eng.cursorX, eng.cursorY, v)
	eng.notifyBufferChanged()
}

func spliceRowsAt(cells [][]string, at int, insert [][]string) [][]string {
	if at < 0 {
		at = 0
	}
	if at > len(cells) {
		at = len(cells)
	}
	out := make([][]string, 0, len(cells)+len(insert))
	out = append(out, cells[:at]...)
	for _, row := range insert {
		out = append(out, append([]string(nil), row...))
	}
	out = append(out, cells[at:]...)
	return out
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
		switch key {
		case KeyEsc:
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.lastKeyCaptured = true
		case "Tab":
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.moveCursorTo(eng.cursorX+1, eng.cursorY)
			eng.setMode(ModeInsert, InsertPositionHighlight)
			eng.lastKeyCaptured = true
		case "Enter":
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.moveCursorTo(eng.cursorX, eng.cursorY+1)
			eng.setMode(ModeInsert, InsertPositionHighlight)
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
