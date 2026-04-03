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
	visualKH        *visualKeyHandler
	lastKeyCaptured bool

	undoStack UndoStack

	visualAnchorX int
	visualAnchorY int

	lastVisualAnchorX int
	lastVisualAnchorY int
	lastVisualCursorX int
	lastVisualCursorY int
	lastVisualMode    Mode

	normalKeymap  KeymapTable
	insertKeymap  KeymapTable
	visualKeymap  KeymapTable
	keymapPending []string
	keymapDepth   int
}

type EventListener interface {
	OnModeChanged(mode Mode, insertPosition InsertPosition)
	OnCursorMoved(x, y int)
	OnBufferChanged()
	OnClipboardWrite(text string)
	OnSelectionChanged(startX, startY, endX, endY, cursorX, cursorY int)
}

type Mode string

const (
	ModeNormal      Mode = "n"
	ModeInsert      Mode = "i"
	ModeVisual      Mode = "v"
	ModeVisualLine  Mode = "V"
	ModeVisualBlock Mode = "vb"
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

const (
	KeyEsc        = "Escape"
	KeyShiftTab   = "Shift+Tab"
	KeyShiftEnter = "Shift+Enter"
)

// New builds an engine from DataSource.Load(). The grid must be rectangular and non-empty.
func New(dataSource DataSource) Engine {
	cells := cloneCells(dataSource.Load())
	cols, rows, ok := validateRectangularGrid(cells)
	if !ok {
		panic("engine.New: DataSource.Load must return a non-empty rectangular grid")
	}
	nh := newNormalKeyHandler()
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
		normalKH:             nh,
		visualKH:             newVisualKeyHandler(nh.motions, nh.operators),
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

// SetCellValueUndoable sets a cell and records an undo step when the value changes (e.g. insert-mode saves from WASM).
func (eng *Engine) SetCellValueUndoable(x, y int, value string) bool {
	old, ok := eng.CellValue(x, y)
	if !ok {
		return false
	}
	if old == value {
		return true
	}
	eng.pushUndoCheckpoint()
	return eng.SetCellValue(x, y, value)
}

func (eng *Engine) snapshotCurrent() UndoEntry {
	return UndoEntry{
		Cells:   cloneCells(eng.cells),
		CursorX: eng.cursorX,
		CursorY: eng.cursorY,
	}
}

func (eng *Engine) pushUndoCheckpoint() {
	eng.undoStack.PushUndo(eng.snapshotCurrent())
	eng.undoStack.ClearRedo()
}

func (eng *Engine) restoreFromEntry(e UndoEntry) {
	eng.cells = cloneCells(e.Cells)
	eng.rows = len(eng.cells)
	if eng.rows == 0 {
		panic("engine: undo restore produced empty grid")
	}
	eng.cols = len(eng.cells[0])
	cx := e.CursorX
	cy := e.CursorY
	if cx < 0 {
		cx = 0
	}
	if cy < 0 {
		cy = 0
	}
	if cx >= eng.cols {
		cx = eng.cols - 1
	}
	if cy >= eng.rows {
		cy = eng.rows - 1
	}
	eng.cursorX = cx
	eng.cursorY = cy
	eng.notifyBufferChanged()
	eng.notifyCursorMoved()
}

// Undo restores the previous grid + cursor snapshot. No-op if the undo stack is empty.
func (eng *Engine) Undo() {
	if eng.undoStack.UndoLen() == 0 {
		return
	}
	eng.undoStack.PushRedo(eng.snapshotCurrent())
	e, ok := eng.undoStack.PopUndo()
	if !ok {
		return
	}
	eng.restoreFromEntry(e)
}

// Redo reapplies the last undone mutation. No-op if the redo stack is empty.
func (eng *Engine) Redo() {
	if eng.undoStack.RedoLen() == 0 {
		return
	}
	eng.undoStack.PushUndo(eng.snapshotCurrent())
	e, ok := eng.undoStack.PopRedo()
	if !ok {
		return
	}
	eng.restoreFromEntry(e)
}

// UndoDepth returns the number of undo steps available (for tests).
func (eng *Engine) UndoDepth() int {
	return eng.undoStack.UndoLen()
}

// RedoDepth returns the number of redo steps available (for tests).
func (eng *Engine) RedoDepth() int {
	return eng.undoStack.RedoLen()
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

func (eng *Engine) isVisualMode() bool {
	switch eng.mode {
	case ModeVisual, ModeVisualLine, ModeVisualBlock:
		return true
	default:
		return false
	}
}

func (eng *Engine) isVisualLinewise() bool {
	return eng.mode == ModeVisualLine
}

// snapshotLastVisualForGV stores anchor/cursor/mode for gv re-selection.
func (eng *Engine) snapshotLastVisualForGV() {
	eng.lastVisualAnchorX = eng.visualAnchorX
	eng.lastVisualAnchorY = eng.visualAnchorY
	eng.lastVisualCursorX = eng.cursorX
	eng.lastVisualCursorY = eng.cursorY
	eng.lastVisualMode = eng.mode
}

// enterVisualMode sets anchor at the cursor and enters the given visual sub-mode.
func (eng *Engine) enterVisualMode(mode Mode) {
	eng.visualAnchorX = eng.cursorX
	eng.visualAnchorY = eng.cursorY
	eng.setMode(mode, InsertPositionNone)
	eng.notifySelectionChanged()
}

// exitVisualMode saves last-visual state for gv and returns to normal mode.
func (eng *Engine) exitVisualMode() {
	eng.snapshotLastVisualForGV()
	eng.setMode(ModeNormal, InsertPositionNone)
}

// RestoreLastVisualSelection re-enters visual mode with the last gv snapshot (no-op if none).
func (eng *Engine) RestoreLastVisualSelection() {
	if eng.lastVisualMode == "" {
		return
	}
	eng.visualAnchorX = eng.lastVisualAnchorX
	eng.visualAnchorY = eng.lastVisualAnchorY
	eng.moveCursorTo(eng.lastVisualCursorX, eng.lastVisualCursorY)
	eng.setMode(eng.lastVisualMode, InsertPositionNone)
	eng.notifySelectionChanged()
}

// GetVisualSelection returns inclusive bounds of the current visual selection.
func (eng *Engine) GetVisualSelection() (startX, startY, endX, endY int) {
	if eng.mode == ModeVisualLine {
		startX = 0
		endX = eng.cols - 1
		minY, maxY := eng.visualAnchorY, eng.cursorY
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		return startX, minY, endX, maxY
	}
	ax, ay := eng.visualAnchorX, eng.visualAnchorY
	cx, cy := eng.cursorX, eng.cursorY
	minX, maxX := ax, cx
	if minX > maxX {
		minX, maxX = maxX, minX
	}
	minY, maxY := ay, cy
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	return minX, minY, maxX, maxY
}

func (eng *Engine) notifySelectionChanged() {
	if !eng.isVisualMode() {
		return
	}
	sx, sy, ex, ey := eng.GetVisualSelection()
	for _, listener := range eng.listeners {
		listener.OnSelectionChanged(sx, sy, ex, ey, eng.cursorX, eng.cursorY)
	}
}

// ExecuteVisualOperator runs d/y/c (or x as d) on the current visual selection, then leaves visual mode
// (or enters insert after change).
func (eng *Engine) ExecuteVisualOperator(op string) {
	if op == "x" {
		op = "d"
	}
	eng.snapshotLastVisualForGV()
	var wentInsert bool
	if eng.isVisualLinewise() {
		wentInsert = eng.executeVisualLinewiseOp(op)
	} else {
		wentInsert = eng.executeVisualCellwiseOp(op)
	}
	if wentInsert {
		eng.visualKH.ResetPending()
		return
	}
	eng.setMode(ModeNormal, InsertPositionNone)
}

func (eng *Engine) executeVisualCellwiseOp(op string) bool {
	sx, sy, ex, ey := eng.GetVisualSelection()
	switch op {
	case "d":
		eng.deleteRect(sx, sy, ex, ey)
		return false
	case "y":
		eng.yankRect(sx, sy, ex, ey)
		return false
	case "c":
		return eng.changeRect(sx, sy, ex, ey)
	default:
		return false
	}
}

func (eng *Engine) executeVisualLinewiseOp(op string) bool {
	_, minY, _, maxY := eng.GetVisualSelection()
	switch op {
	case "d":
		startY := minY
		if startY < 1 {
			startY = 1
		}
		if startY > maxY {
			return false
		}
		eng.deleteRowsRange(startY, maxY+1)
		return false
	case "y":
		eng.yankRowsRange(minY, maxY+1)
		return false
	case "c":
		startY := minY
		if startY < 1 {
			startY = 1
		}
		if startY > maxY {
			return false
		}
		eng.changeRowsRange(startY, maxY+1)
		return true
	default:
		return false
	}
}

func (eng *Engine) deleteRect(startX, startY, endX, endY int) {
	if startX > endX || startY > endY {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startY < 0 {
		startY = 0
	}
	if endY >= eng.rows {
		endY = eng.rows - 1
	}
	if startX > endX || startY > endY {
		return
	}
	mutStartY := startY
	if mutStartY < 1 {
		mutStartY = 1
	}
	if mutStartY > endY {
		return
	}
	eng.pushUndoCheckpoint()
	rows := make([][]string, 0, endY-mutStartY+1)
	for y := mutStartY; y <= endY; y++ {
		row := make([]string, endX-startX+1)
		for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
			row[i] = eng.cells[y][x]
		}
		rows = append(rows, row)
	}
	eng.register = Register{Cells: rows, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
	for y := mutStartY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			eng.cells[y][x] = ""
		}
	}
	eng.moveCursorTo(startX, mutStartY)
	eng.notifyBufferChanged()
}

func (eng *Engine) yankRect(startX, startY, endX, endY int) {
	if startX > endX || startY > endY {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startY < 0 {
		startY = 0
	}
	if endY >= eng.rows {
		endY = eng.rows - 1
	}
	if startX > endX || startY > endY {
		return
	}
	rows := make([][]string, 0, endY-startY+1)
	for y := startY; y <= endY; y++ {
		row := make([]string, endX-startX+1)
		for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
			row[i] = eng.cells[y][x]
		}
		rows = append(rows, row)
	}
	eng.register = Register{Cells: rows, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
}

// changeRect clears non-header cells in the rectangle and enters insert mode.
// Returns false if the selection was header-only (no mutation, no mode change).
func (eng *Engine) changeRect(startX, startY, endX, endY int) bool {
	if startX > endX || startY > endY {
		return false
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startY < 0 {
		startY = 0
	}
	if endY >= eng.rows {
		endY = eng.rows - 1
	}
	if startX > endX || startY > endY {
		return false
	}
	mutStartY := startY
	if mutStartY < 1 {
		mutStartY = 1
	}
	if mutStartY > endY {
		return false
	}
	eng.pushUndoCheckpoint()
	rows := make([][]string, 0, endY-mutStartY+1)
	for y := mutStartY; y <= endY; y++ {
		row := make([]string, endX-startX+1)
		for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
			row[i] = eng.cells[y][x]
		}
		rows = append(rows, row)
	}
	eng.register = Register{Cells: rows, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
	for y := mutStartY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			eng.cells[y][x] = ""
		}
	}
	eng.notifyBufferChanged()
	eng.moveCursorTo(startX, mutStartY)
	eng.setMode(ModeInsert, InsertPositionHighlight)
	return true
}

// ExecuteLinewiseDoubled runs dd, yy, or cc (second key matched the operator).
func (eng *Engine) ExecuteLinewiseDoubled(op string, ctx OperatorContext) {
	n := ctx.Count
	if n < 1 {
		n = 1
	}
	start := eng.cursorY
	end := start + n
	if end > eng.rows {
		end = eng.rows
	}
	switch op {
	case "d":
		if start == 0 {
			return
		}
		eng.deleteRowsRange(start, end)
	case "y":
		eng.yankRowsRange(start, end)
	case "c":
		if start == 0 {
			return
		}
		eng.changeRowsRange(start, end)
	}
}

// ExecuteOperatorWithMotion runs d/y/c after a motion (e.g. dj, d$, ygg).
func (eng *Engine) ExecuteOperatorWithMotion(op string, minfo *MotionInfo, ctx MotionContext) {
	if minfo == nil {
		return
	}
	startX, startY := eng.cursorX, eng.cursorY
	res := minfo.Fn(eng, ctx)
	if minfo.Linewise {
		eng.applyLinewiseOperatorMotion(op, startY, res.TargetY)
		return
	}
	eng.applyNonLinewiseOperatorMotion(op, minfo, startY, startX, res.TargetX)
}

// ExecuteOperatorWithTextObject runs d/y/c after a text object (e.g. diw, ciw, yiw).
func (eng *Engine) ExecuteOperatorWithTextObject(op string, r TextObjectRange) {
	switch op {
	case "d":
		eng.deleteCellsInRowRange(r.Y, r.StartX, r.EndX)
		eng.moveCursorTo(r.StartX, r.Y)
	case "y":
		eng.yankCellsInRowRange(r.Y, r.StartX, r.EndX)
	case "c":
		eng.changeCellsInRowRange(r.Y, r.StartX, r.EndX)
		eng.moveCursorTo(r.StartX, r.Y)
	}
}

func (eng *Engine) applyLinewiseOperatorMotion(op string, startY, targetY int) {
	minY := startY
	maxY := targetY
	if minY > maxY {
		minY, maxY = maxY, minY
	}
	endExclusive := maxY + 1
	switch op {
	case "d":
		if minY < 1 {
			minY = 1
		}
		if minY >= endExclusive || minY >= eng.rows {
			return
		}
		eng.deleteRowsRange(minY, endExclusive)
	case "y":
		if minY >= endExclusive || minY >= eng.rows {
			return
		}
		eng.yankRowsRange(minY, endExclusive)
	case "c":
		if minY < 1 {
			minY = 1
		}
		if minY >= endExclusive || minY >= eng.rows {
			return
		}
		eng.changeRowsRange(minY, endExclusive)
	}
}

func (eng *Engine) applyNonLinewiseOperatorMotion(op string, minfo *MotionInfo, y, startX, targetX int) {
	if y < 0 || y >= eng.rows {
		return
	}
	inclusive := true
	if minfo != nil {
		inclusive = minfo.Inclusive
	}
	lo, hi := cellRangeNonLinewise(startX, targetX, inclusive)
	switch op {
	case "d":
		eng.deleteCellsInRowRange(y, lo, hi)
	case "y":
		eng.yankCellsInRowRange(y, lo, hi)
	case "c":
		eng.changeCellsInRowRange(y, lo, hi)
	}
	eng.moveCursorTo(lo, y)
}

func cellRangeNonLinewise(startX, targetX int, inclusive bool) (lo, hi int) {
	if inclusive {
		if startX <= targetX {
			return startX, targetX
		}
		return targetX, startX
	}
	if targetX > startX {
		lo = startX
		hi = targetX - 1
	} else if targetX < startX {
		lo = targetX + 1
		hi = startX
	} else {
		return startX, startX
	}
	if hi < lo {
		return startX, startX
	}
	return lo, hi
}

// deleteRowsRange removes rows [startY, endY) (endY exclusive). startY must be >= 1 for header safety at call sites.
func (eng *Engine) deleteRowsRange(startY, endY int) {
	if startY < 1 || startY >= eng.rows || startY >= endY {
		return
	}
	if endY > eng.rows {
		endY = eng.rows
	}
	eng.pushUndoCheckpoint()
	toCopy := eng.cells[startY:endY]
	eng.register = Register{Cells: cloneCells(toCopy), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())

	eng.cells = append(eng.cells[:startY], eng.cells[endY:]...)
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

func (eng *Engine) yankRowsRange(startY, endY int) {
	if startY >= eng.rows || startY >= endY {
		return
	}
	if endY > eng.rows {
		endY = eng.rows
	}
	slice := eng.cells[startY:endY]
	eng.register = Register{Cells: cloneCells(slice), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
}

func (eng *Engine) changeRowsRange(startY, endY int) {
	if startY >= eng.rows || startY >= endY {
		return
	}
	if endY > eng.rows {
		endY = eng.rows
	}
	eng.pushUndoCheckpoint()
	slice := eng.cells[startY:endY]
	eng.register = Register{Cells: cloneCells(slice), Linewise: true}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())

	for y := startY; y < endY; y++ {
		for x := 0; x < eng.cols; x++ {
			eng.cells[y][x] = ""
		}
	}
	eng.notifyBufferChanged()
	eng.setMode(ModeInsert, InsertPositionHighlight)
}

func (eng *Engine) deleteCellsInRowRange(y, startX, endX int) {
	if y < 0 || y >= eng.rows {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startX > endX {
		return
	}
	eng.pushUndoCheckpoint()
	row := make([]string, endX-startX+1)
	for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
		row[i] = eng.cells[y][x]
	}
	eng.register = Register{Cells: [][]string{row}, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
	for x := startX; x <= endX; x++ {
		eng.cells[y][x] = ""
	}
	eng.notifyBufferChanged()
}

func (eng *Engine) yankCellsInRowRange(y, startX, endX int) {
	if y < 0 || y >= eng.rows {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startX > endX {
		return
	}
	row := make([]string, endX-startX+1)
	for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
		row[i] = eng.cells[y][x]
	}
	eng.register = Register{Cells: [][]string{row}, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
}

func (eng *Engine) changeCellsInRowRange(y, startX, endX int) {
	if y < 0 || y >= eng.rows {
		return
	}
	if startX < 0 {
		startX = 0
	}
	if endX >= eng.cols {
		endX = eng.cols - 1
	}
	if startX > endX {
		return
	}
	eng.pushUndoCheckpoint()
	row := make([]string, endX-startX+1)
	for i, x := 0, startX; x <= endX; i, x = i+1, x+1 {
		row[i] = eng.cells[y][x]
	}
	eng.register = Register{Cells: [][]string{row}, Linewise: false}
	eng.notifyClipboardWrite(eng.formatRegisterClipboard())
	for x := startX; x <= endX; x++ {
		eng.cells[y][x] = ""
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
	eng.pushUndoCheckpoint()
	eng.register = Register{Cells: [][]string{{v}}, Linewise: false}
	eng.SetCellValue(eng.cursorX, eng.cursorY, "")
	eng.notifyBufferChanged()
}

// pasteRect overlays non-linewise register cells onto the grid at (startX, startY).
// Rows are clamped to the grid; column index may be negative (cells skipped). Header row 0 is never
// written: the overlay is anchored at startY, but if startY < 1 the first pasted row maps to grid row 1.
func (eng *Engine) pasteRect(startX, startY int) {
	reg := eng.register.Cells
	if len(reg) == 0 {
		return
	}
	eng.pushUndoCheckpoint()
	anchorY := startY
	if anchorY < 1 {
		anchorY = 1
	}
	for ry, row := range reg {
		ty := anchorY + ry
		if ty >= eng.rows {
			break
		}
		for rx, val := range row {
			tx := startX + rx
			if tx < 0 {
				continue
			}
			if tx >= eng.cols {
				break
			}
			eng.cells[ty][tx] = val
		}
	}
	cy := startY
	if cy < 1 {
		cy = 1
	}
	if startX < 0 {
		startX = 0
	} else if startX >= eng.cols {
		startX = eng.cols - 1
	}
	eng.moveCursorTo(startX, cy)
	eng.notifyBufferChanged()
}

// PasteAfter puts register contents at or below the cursor (linewise or character).
func (eng *Engine) PasteAfter() {
	if eng.register.Linewise {
		if len(eng.register.Cells) == 0 {
			return
		}
		eng.pushUndoCheckpoint()
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
	if len(eng.register.Cells) == 0 {
		return
	}
	eng.pasteRect(eng.cursorX, eng.cursorY)
}

// PasteBefore puts register contents above the cursor (linewise) or replaces current cell (non-linewise).
func (eng *Engine) PasteBefore() {
	if eng.register.Linewise {
		if len(eng.register.Cells) == 0 {
			return
		}
		eng.pushUndoCheckpoint()
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
	if len(eng.register.Cells) == 0 {
		return
	}
	eng.pasteRect(eng.cursorX, eng.cursorY)
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
		if eng.visualKH != nil {
			eng.visualKH.ResetPending()
		}
		eng.keymapPending = eng.keymapPending[:0]
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
	eng.keymapDepth = 0
	eng.feedKey(key, true)
}

func (eng *Engine) feedKey(key string, checkKeymap bool) {
	if len(eng.keymapPending) > 0 {
		table := eng.keymapTableForMode()
		eng.keymapPending = append(eng.keymapPending, key)
		mr, entry := table.Lookup(eng.keymapPending)
		switch mr {
		case MatchExact:
			eng.keymapPending = eng.keymapPending[:0]
			eng.expandMapping(entry)
			return
		case MatchPrefix:
			eng.lastKeyCaptured = true
			return
		case MatchNone:
			pending := append([]string(nil), eng.keymapPending...)
			eng.keymapPending = eng.keymapPending[:0]
			for i := 0; i < len(pending); i++ {
				eng.feedKey(pending[i], false)
			}
			return
		default:
			eng.keymapPending = eng.keymapPending[:0]
			return
		}
	}

	if checkKeymap {
		table := eng.keymapTableForMode()
		mr, entry := table.Lookup([]string{key})
		switch mr {
		case MatchExact:
			eng.expandMapping(entry)
			return
		case MatchPrefix:
			eng.keymapPending = append(eng.keymapPending[:0], key)
			eng.lastKeyCaptured = true
			return
		case MatchNone:
			// fall through to mode handler
		}
	}

	eng.feedKeyMode(key)
}

func (eng *Engine) keymapTableForMode() *KeymapTable {
	switch eng.mode {
	case ModeNormal:
		return &eng.normalKeymap
	case ModeInsert:
		return &eng.insertKeymap
	case ModeVisual, ModeVisualLine, ModeVisualBlock:
		return &eng.visualKeymap
	default:
		return &eng.normalKeymap
	}
}

func (eng *Engine) expandMapping(entry *KeymapEntry) {
	if entry == nil {
		return
	}
	if eng.keymapDepth >= maxKeymapDepth {
		for _, k := range entry.RHS {
			eng.feedKey(k, false)
		}
		return
	}
	eng.keymapDepth++
	for _, k := range entry.RHS {
		eng.feedKey(k, entry.Recursive)
	}
	eng.keymapDepth--
}

func (eng *Engine) feedKeyMode(key string) {
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
		case KeyShiftTab:
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.moveCursorTo(eng.cursorX-1, eng.cursorY)
			eng.setMode(ModeInsert, InsertPositionHighlight)
			eng.lastKeyCaptured = true
		case "Enter":
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.moveCursorTo(eng.cursorX, eng.cursorY+1)
			eng.setMode(ModeInsert, InsertPositionHighlight)
			eng.lastKeyCaptured = true
		case KeyShiftEnter:
			eng.setMode(ModeNormal, InsertPositionNone)
			eng.moveCursorTo(eng.cursorX, eng.cursorY-1)
			eng.setMode(ModeInsert, InsertPositionHighlight)
			eng.lastKeyCaptured = true
		}

	case ModeVisual, ModeVisualLine, ModeVisualBlock:
		r := eng.visualKH.Feed(eng, key)
		if r != ParseInvalid {
			eng.lastKeyCaptured = true
		}
	}
}

// Nmap registers a recursive normal-mode mapping (RHS may trigger other maps).
func (eng *Engine) Nmap(lhs, rhs string) {
	eng.mapKeys(&eng.normalKeymap, lhs, rhs, true)
}

// Nnoremap registers a non-recursive normal-mode mapping.
func (eng *Engine) Nnoremap(lhs, rhs string) {
	eng.mapKeys(&eng.normalKeymap, lhs, rhs, false)
}

// Imap registers a recursive insert-mode mapping.
func (eng *Engine) Imap(lhs, rhs string) {
	eng.mapKeys(&eng.insertKeymap, lhs, rhs, true)
}

// Inoremap registers a non-recursive insert-mode mapping.
func (eng *Engine) Inoremap(lhs, rhs string) {
	eng.mapKeys(&eng.insertKeymap, lhs, rhs, false)
}

// Vmap registers a recursive visual-mode mapping.
func (eng *Engine) Vmap(lhs, rhs string) {
	eng.mapKeys(&eng.visualKeymap, lhs, rhs, true)
}

// Vnoremap registers a non-recursive visual-mode mapping.
func (eng *Engine) Vnoremap(lhs, rhs string) {
	eng.mapKeys(&eng.visualKeymap, lhs, rhs, false)
}

func (eng *Engine) mapKeys(table *KeymapTable, lhs, rhs string, recursive bool) {
	lk := ParseKeys(lhs)
	rk := ParseKeys(rhs)
	if len(lk) == 0 {
		return
	}
	table.Set(lk, rk, recursive)
}

// Unmap removes a mapping for the given mode. Returns false if no mapping existed.
func (eng *Engine) Unmap(mode Mode, lhs string) bool {
	lk := ParseKeys(lhs)
	if len(lk) == 0 {
		return false
	}
	switch mode {
	case ModeNormal:
		return eng.normalKeymap.Delete(lk)
	case ModeInsert:
		return eng.insertKeymap.Delete(lk)
	case ModeVisual, ModeVisualLine, ModeVisualBlock:
		return eng.visualKeymap.Delete(lk)
	default:
		return false
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
	if eng.isVisualMode() {
		eng.exitVisualMode()
	}
	eng.normalKH.ResetPending()
	eng.keymapPending = eng.keymapPending[:0]
	eng.moveCursorTo(x, y)
}

// SetCursorAndEdit moves the cursor to (x, y) and enters insert mode with the full cell selected for overwrite (same as Enter).
// Exits insert mode first if needed. Out-of-bounds coordinates are ignored (insert mode is still exited if active).
func (eng *Engine) SetCursorAndEdit(x, y int) {
	if eng.mode == ModeInsert {
		eng.setMode(ModeNormal, InsertPositionNone)
	}
	if eng.isVisualMode() {
		eng.exitVisualMode()
	}
	eng.normalKH.ResetPending()
	eng.keymapPending = eng.keymapPending[:0]
	if x < 0 || x >= eng.cols || y < 0 || y >= eng.rows {
		return
	}
	eng.moveCursorTo(x, y)
	eng.setMode(ModeInsert, InsertPositionHighlight)
}
