package engine_test

import (
	"testing"

	"github.com/gjtiquia/vimance/internal/engine"
)

const testCols = 6
const testRows = 5

func newTestEngine(cols, rows int) engine.Engine {
	cells := make([][]string, rows)
	for y := range cells {
		cells[y] = make([]string, cols)
	}
	return engine.New(&engine.StaticDataSource{Cells: cells})
}

type TestEngineEventListener struct {
	OnModeChangedCounter      int
	OnCursorMovedCounter      int
	OnBufferChangedCounter    int
	OnClipboardWriteCounter   int
	LastMode                  engine.Mode
	LastInsertPosition        engine.InsertPosition
	LastCursorX, LastCursorY  int
	LastClipboardText         string
}

func (l *TestEngineEventListener) OnModeChanged(mode engine.Mode, insertPosition engine.InsertPosition) {
	l.OnModeChangedCounter++
	l.LastMode = mode
	l.LastInsertPosition = insertPosition
}

func (l *TestEngineEventListener) OnCursorMoved(x, y int) {
	l.OnCursorMovedCounter++
	l.LastCursorX = x
	l.LastCursorY = y
}

func (l *TestEngineEventListener) OnBufferChanged() {
	l.OnBufferChangedCounter++
}

func (l *TestEngineEventListener) OnClipboardWrite(text string) {
	l.OnClipboardWriteCounter++
	l.LastClipboardText = text
}

func TestNewWithStubDataSourceDimensions(t *testing.T) {
	eng := engine.New(&engine.StubDataSource{})
	if eng.Cols() != 6 || eng.Rows() != 5 {
		t.Fatalf("stub grid 6x5, got %dx%d", eng.Cols(), eng.Rows())
	}
	v, ok := eng.CellValue(0, 0)
	if !ok || v != "Name" {
		t.Fatalf("cell (0,0) want Name, got %q ok=%v", v, ok)
	}
}

func TestInitialState(t *testing.T) {
	eng := newTestEngine(testCols, testRows)

	if eng.Mode() != engine.ModeNormal {
		t.Errorf("expected initial mode to be normal, got %v", eng.Mode())
	}
	if eng.CursorX() != 0 || eng.CursorY() != 0 {
		t.Errorf("expected initial cursor (0,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestModeSwitching(t *testing.T) {

	eng := newTestEngine(testCols, testRows)

	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	// Switch to insert mode
	eng.KeyPress("i")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected mode to be insert, got %v", eng.Mode())
	}
	if listener.LastInsertPosition != engine.InsertPositionBefore {
		t.Errorf("expected insert position before for i, got %v", listener.LastInsertPosition)
	}

	// Switch back to normal mode
	eng.KeyPress(engine.KeyEsc) // Escape key
	if eng.Mode() != engine.ModeNormal {
		t.Fatalf("expected mode to be normal, got %v", eng.Mode())
	}

	// Switch to visual mode
	eng.KeyPress("v")
	if eng.Mode() != engine.ModeVisual {
		t.Fatalf("expected mode to be visual, got %v", eng.Mode())
	}

	// Switch back to normal mode
	eng.KeyPress(engine.KeyEsc) // Escape key
	if eng.Mode() != engine.ModeNormal {
		t.Fatalf("expected mode to be normal, got %v", eng.Mode())
	}

	if listener.OnModeChangedCounter != 4 {
		t.Errorf("expected OnModeChanged to be called 4 times, got %v", listener.OnModeChangedCounter)
	}
}

func TestInsertPositionForA(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("a")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected mode to be insert, got %v", eng.Mode())
	}
	if listener.LastInsertPosition != engine.InsertPositionAfter {
		t.Errorf("expected insert position after for a, got %v", listener.LastInsertPosition)
	}
}

func TestHjklMovesCursor(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("l")
	if eng.CursorX() != 1 || eng.CursorY() != 0 {
		t.Errorf("after l, expected cursor (1,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastCursorX != 1 || listener.LastCursorY != 0 {
		t.Errorf("OnCursorMoved expected (1,0), got (%d,%d)", listener.LastCursorX, listener.LastCursorY)
	}

	eng.KeyPress("j")
	if eng.CursorX() != 1 || eng.CursorY() != 1 {
		t.Errorf("after j, expected cursor (1,1), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("h")
	if eng.CursorX() != 0 || eng.CursorY() != 1 {
		t.Errorf("after h, expected cursor (0,1), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("k")
	if eng.CursorX() != 0 || eng.CursorY() != 0 {
		t.Errorf("after k, expected cursor (0,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	if listener.OnCursorMovedCounter != 4 {
		t.Errorf("expected OnCursorMoved 4 times, got %v", listener.OnCursorMovedCounter)
	}
}

func TestHjklDoesNotFireInInsertMode(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("i")
	listener.OnCursorMovedCounter = 0

	eng.KeyPress("l")
	if eng.CursorX() != 0 {
		t.Errorf("cursor should not move in insert mode, got x=%d", eng.CursorX())
	}
	if listener.OnCursorMovedCounter != 0 {
		t.Errorf("OnCursorMoved should not fire in insert mode")
	}
}

func TestHjklBounds(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	// At origin, h and k do nothing and emit no event
	eng.KeyPress("h")
	eng.KeyPress("k")
	if eng.CursorX() != 0 || eng.CursorY() != 0 {
		t.Errorf("expected cursor still (0,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.OnCursorMovedCounter != 0 {
		t.Errorf("expected no OnCursorMoved at origin for h/k, got %v", listener.OnCursorMovedCounter)
	}

	// Move to bottom-right (5, 4)
	for range testCols - 1 {
		eng.KeyPress("l")
	}
	for range testRows - 1 {
		eng.KeyPress("j")
	}
	if eng.CursorX() != testCols-1 || eng.CursorY() != testRows-1 {
		t.Fatalf("expected cursor (%d,%d), got (%d,%d)", testCols-1, testRows-1, eng.CursorX(), eng.CursorY())
	}

	listener.OnCursorMovedCounter = 0
	eng.KeyPress("l")
	eng.KeyPress("j")
	if eng.CursorX() != testCols-1 || eng.CursorY() != testRows-1 {
		t.Errorf("expected cursor still at max, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.OnCursorMovedCounter != 0 {
		t.Errorf("expected no OnCursorMoved at max edge for l/j, got %v", listener.OnCursorMovedCounter)
	}
}

func TestWebEAndBMoveHorizontally(t *testing.T) {
	eng := newTestEngine(testCols, testRows)

	eng.KeyPress("w")
	if eng.CursorX() != 1 || eng.CursorY() != 0 {
		t.Errorf("after w, expected (1,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("e")
	if eng.CursorX() != 2 || eng.CursorY() != 0 {
		t.Errorf("after e, expected (2,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("b")
	if eng.CursorX() != 1 || eng.CursorY() != 0 {
		t.Errorf("after b, expected (1,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestArrowKeysMoveCursorLikeHjkl(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("ArrowRight")
	if eng.CursorX() != 1 || eng.CursorY() != 0 {
		t.Errorf("after ArrowRight, expected (1,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("ArrowDown")
	if eng.CursorX() != 1 || eng.CursorY() != 1 {
		t.Errorf("after ArrowDown, expected (1,1), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("ArrowLeft")
	if eng.CursorX() != 0 || eng.CursorY() != 1 {
		t.Errorf("after ArrowLeft, expected (0,1), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	eng.KeyPress("ArrowUp")
	if eng.CursorX() != 0 || eng.CursorY() != 0 {
		t.Errorf("after ArrowUp, expected (0,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}

	if listener.OnCursorMovedCounter != 4 {
		t.Errorf("expected OnCursorMoved 4 times, got %v", listener.OnCursorMovedCounter)
	}
}

func TestEnterEntersInsertWithHighlightSelection(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("Enter")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected mode to be insert, got %v", eng.Mode())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected insert position highlight for Enter, got %v", listener.LastInsertPosition)
	}
}

func TestSetCursorMovesAndRespectsBounds(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.SetCursor(3, 2)
	if eng.CursorX() != 3 || eng.CursorY() != 2 {
		t.Errorf("expected cursor (3,2), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.OnCursorMovedCounter != 1 {
		t.Errorf("expected OnCursorMoved once, got %v", listener.OnCursorMovedCounter)
	}

	listener.OnCursorMovedCounter = 0
	eng.SetCursor(-1, 0)
	if eng.CursorX() != 3 || eng.CursorY() != 2 {
		t.Errorf("invalid x should not move cursor, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.OnCursorMovedCounter != 0 {
		t.Errorf("expected no OnCursorMoved for out-of-bounds, got %v", listener.OnCursorMovedCounter)
	}

	eng.SetCursor(testCols, 0)
	if eng.CursorX() != 3 || eng.CursorY() != 2 {
		t.Errorf("invalid x should not move cursor, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestSetCursorExitsInsertModeFirst(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("i")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode")
	}
	listener.OnModeChangedCounter = 0

	eng.SetCursor(2, 1)
	if eng.Mode() != engine.ModeNormal {
		t.Errorf("expected normal mode after SetCursor, got %v", eng.Mode())
	}
	if eng.CursorX() != 2 || eng.CursorY() != 1 {
		t.Errorf("expected cursor (2,1), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.OnModeChangedCounter < 1 {
		t.Errorf("expected OnModeChanged when leaving insert, got %v", listener.OnModeChangedCounter)
	}
}

// Redundant SetCursor (same cell) while in insert must not exit insert — avoids jank when e.g. touch
// is followed by a synthetic click with the same coordinates.
func TestSetCursorSameCellWhileInsertIsNoOp(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.SetCursorAndEdit(2, 2)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	listener.OnModeChangedCounter = 0

	eng.SetCursor(2, 2)
	if eng.Mode() != engine.ModeInsert {
		t.Errorf("expected insert mode after redundant SetCursor, got %v", eng.Mode())
	}
	if listener.OnModeChangedCounter != 0 {
		t.Errorf("expected no OnModeChanged for redundant SetCursor, got %v", listener.OnModeChangedCounter)
	}
}

func TestSetCursorAndEditMovesAndEntersInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.SetCursorAndEdit(4, 3)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorX() != 4 || eng.CursorY() != 3 {
		t.Errorf("expected cursor (4,3), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected insert position highlight, got %v", listener.LastInsertPosition)
	}
}

func TestSetCursorAndEditFromInsertMode(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("i")
	eng.SetCursorAndEdit(1, 2)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorX() != 1 || eng.CursorY() != 2 {
		t.Errorf("expected cursor (1,2), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected insert position highlight, got %v", listener.LastInsertPosition)
	}
}

func TestGGMovesToFirstRowKeepingColumn(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	for range 3 {
		eng.KeyPress("l")
	}
	for range 2 {
		eng.KeyPress("j")
	}
	if eng.CursorX() != 3 || eng.CursorY() != 2 {
		t.Fatalf("setup cursor, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	eng.KeyPress("g")
	eng.KeyPress("g")
	if eng.CursorX() != 3 || eng.CursorY() != 0 {
		t.Errorf("after gg, expected (3,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestGMovesToLastRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("G")
	if eng.CursorY() != testRows-1 {
		t.Errorf("after G, expected y=%d, got %d", testRows-1, eng.CursorY())
	}
}

func TestZeroAndDollarMoveColumn(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	for range 4 {
		eng.KeyPress("l")
	}
	eng.KeyPress("0")
	if eng.CursorX() != 0 {
		t.Errorf("after 0, expected x=0, got %d", eng.CursorX())
	}
	eng.KeyPress("$")
	if eng.CursorX() != testCols-1 {
		t.Errorf("after $, expected x=%d, got %d", testCols-1, eng.CursorX())
	}
}

func TestDDOnHeaderIsNoOp(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != testRows {
		t.Fatalf("dd on header: expected %d rows, got %d", testRows, eng.Rows())
	}
	if listener.OnBufferChangedCounter != 0 {
		t.Errorf("expected no OnBufferChanged, got %d", listener.OnBufferChangedCounter)
	}
}

func TestDDDeletesRowAndShrinksGrid(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.KeyPress("j") // row 1
	before := eng.CellsSnapshot()
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != testRows-1 {
		t.Fatalf("expected %d rows after dd, got %d", testRows-1, eng.Rows())
	}
	reg := eng.RegisterSnapshot()
	if !reg.Linewise || len(reg.Cells) != 1 {
		t.Fatalf("register: want 1 linewise row, got %+v", reg)
	}
	for x := 0; x < testCols; x++ {
		if reg.Cells[0][x] != before[1][x] {
			t.Errorf("register cell x=%d mismatch", x)
		}
	}
	if listener.OnBufferChangedCounter < 1 {
		t.Errorf("expected OnBufferChanged after dd")
	}
	if listener.OnClipboardWriteCounter < 1 {
		t.Errorf("expected OnClipboardWrite after dd")
	}
}

func TestThreeDDDeletesThreeRows(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("j")
	eng.KeyPress("3")
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != testRows-3 {
		t.Fatalf("3dd: want %d rows, got %d", testRows-3, eng.Rows())
	}
	reg := eng.RegisterSnapshot()
	if len(reg.Cells) != 3 {
		t.Fatalf("register: want 3 rows, got %d", len(reg.Cells))
	}
}

func TestDDWhenOnlyHeaderAndOneDataRow(t *testing.T) {
	eng := newTestEngine(testCols, 2)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.KeyPress("j")
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != 1 {
		t.Fatalf("expected 1 row (header only), got %d", eng.Rows())
	}
	if eng.CursorY() != 0 {
		t.Errorf("cursor should be on header, got y=%d", eng.CursorY())
	}
	if listener.OnBufferChangedCounter < 1 {
		t.Error("expected OnBufferChanged")
	}
}

func TestYYFillsRegisterAndClipboard(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.KeyPress("j")
	eng.KeyPress("y")
	eng.KeyPress("y")
	reg := eng.RegisterSnapshot()
	if !reg.Linewise || len(reg.Cells) != 1 {
		t.Fatalf("yy register: %+v", reg)
	}
	if eng.Rows() != testRows {
		t.Error("yy should not change row count")
	}
	if listener.OnClipboardWriteCounter != 1 {
		t.Errorf("OnClipboardWrite want 1, got %d", listener.OnClipboardWriteCounter)
	}
	if listener.LastClipboardText == "" {
		t.Error("clipboard text should be non-empty")
	}
	if listener.OnBufferChangedCounter != 0 {
		t.Error("yy should not fire OnBufferChanged")
	}
}

func TestCCClearsRowAndEntersInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("j")
	eng.KeyPress("c")
	eng.KeyPress("c")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("cc should enter insert, got %v", eng.Mode())
	}
	v, _ := eng.CellValue(0, 1)
	if v != "" {
		t.Errorf("row should be cleared, cell (0,1)=%q", v)
	}
}

func TestPasteAfterLinewiseYY(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("j")
	eng.KeyPress("y")
	eng.KeyPress("y")
	before := eng.Rows()
	eng.KeyPress("p")
	if eng.Rows() != before+1 {
		t.Fatalf("p after yy: want %d rows, got %d", before+1, eng.Rows())
	}
}

func TestPasteBeforeLinewiseYY(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("j")
	eng.KeyPress("j")
	eng.KeyPress("y")
	eng.KeyPress("y")
	eng.KeyPress("k")
	before := eng.Rows()
	eng.KeyPress("P")
	if eng.Rows() != before+1 {
		t.Fatalf("P: want %d rows, got %d", before+1, eng.Rows())
	}
}

func TestPasteBeforeOnHeaderInsertsAtRow1(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(0, 1, "pasted-marker")
	eng.KeyPress("j")
	eng.KeyPress("y")
	eng.KeyPress("y")
	eng.KeyPress("g")
	eng.KeyPress("g")
	if eng.CursorY() != 0 {
		t.Fatalf("cursor on header")
	}
	before := eng.Rows()
	eng.KeyPress("P")
	if eng.Rows() != before+1 {
		t.Fatalf("P on header: want %d rows", before+1)
	}
	v, ok := eng.CellValue(0, 1)
	if !ok || v != "pasted-marker" {
		t.Fatalf("expected pasted row at (0,1), got ok=%v v=%q", ok, v)
	}
}

func TestXClearsCellAndSetsRegister(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "hello")
	eng.SetCursor(2, 2)
	eng.KeyPress("x")
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Errorf("cell should be empty, got %q", v)
	}
	reg := eng.RegisterSnapshot()
	if reg.Linewise || len(reg.Cells) != 1 || len(reg.Cells[0]) != 1 || reg.Cells[0][0] != "hello" {
		t.Fatalf("register: %+v", reg)
	}
}

func TestPasteAfterXReplacesCell(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(1, 1, "a")
	eng.SetCursor(1, 1)
	eng.KeyPress("x")
	eng.SetCursor(2, 2)
	eng.KeyPress("p")
	v, _ := eng.CellValue(2, 2)
	if v != "a" {
		t.Errorf("p after x: want a, got %q", v)
	}
}

func TestTabInInsertModeMovesRight(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.SetCursor(1, 1)
	eng.KeyPress("i")
	listener.OnModeChangedCounter = 0
	listener.OnCursorMovedCounter = 0

	eng.KeyPress("Tab")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode after Tab, got %v", eng.Mode())
	}
	if eng.CursorX() != 2 || eng.CursorY() != 1 {
		t.Errorf("expected cursor (2,1) after Tab, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected highlight insert position, got %v", listener.LastInsertPosition)
	}
	if !eng.LastKeyCaptured() {
		t.Error("Tab should be captured")
	}
}

func TestTabAtLastColumnStaysInInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(testCols-1, 1)
	eng.KeyPress("i")
	eng.KeyPress("Tab")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorX() != testCols-1 {
		t.Errorf("cursor should stay at last col, got x=%d", eng.CursorX())
	}
}

func TestEnterInInsertModeMovesDown(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.SetCursor(2, 1)
	eng.KeyPress("i")
	listener.OnModeChangedCounter = 0
	listener.OnCursorMovedCounter = 0

	eng.KeyPress("Enter")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode after Enter, got %v", eng.Mode())
	}
	if eng.CursorX() != 2 || eng.CursorY() != 2 {
		t.Errorf("expected cursor (2,2) after Enter, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected highlight insert position, got %v", listener.LastInsertPosition)
	}
	if !eng.LastKeyCaptured() {
		t.Error("Enter should be captured")
	}
}

func TestEnterAtLastRowStaysInInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(1, testRows-1)
	eng.KeyPress("i")
	eng.KeyPress("Enter")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorY() != testRows-1 {
		t.Errorf("cursor should stay at last row, got y=%d", eng.CursorY())
	}
}

func TestShiftTabInInsertModeMovesLeft(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.SetCursor(2, 1)
	eng.KeyPress("i")
	listener.OnModeChangedCounter = 0
	listener.OnCursorMovedCounter = 0

	eng.KeyPress(engine.KeyShiftTab)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode after Shift+Tab, got %v", eng.Mode())
	}
	if eng.CursorX() != 1 || eng.CursorY() != 1 {
		t.Errorf("expected cursor (1,1) after Shift+Tab, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected highlight insert position, got %v", listener.LastInsertPosition)
	}
	if !eng.LastKeyCaptured() {
		t.Error("Shift+Tab should be captured")
	}
}

func TestShiftTabAtFirstColumnStaysInInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("i")
	eng.KeyPress(engine.KeyShiftTab)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorX() != 0 {
		t.Errorf("cursor should stay at first col, got x=%d", eng.CursorX())
	}
}

func TestShiftEnterInInsertModeMovesUp(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.SetCursor(2, 2)
	eng.KeyPress("i")
	listener.OnModeChangedCounter = 0
	listener.OnCursorMovedCounter = 0

	eng.KeyPress(engine.KeyShiftEnter)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode after Shift+Enter, got %v", eng.Mode())
	}
	if eng.CursorX() != 2 || eng.CursorY() != 1 {
		t.Errorf("expected cursor (2,1) after Shift+Enter, got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionHighlight {
		t.Errorf("expected highlight insert position, got %v", listener.LastInsertPosition)
	}
	if !eng.LastKeyCaptured() {
		t.Error("Shift+Enter should be captured")
	}
}

func TestShiftEnterAtFirstRowStaysInInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(1, 0)
	eng.KeyPress("i")
	eng.KeyPress(engine.KeyShiftEnter)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorY() != 0 {
		t.Errorf("cursor should stay at first row, got y=%d", eng.CursorY())
	}
}

func TestDJDeletesCurrentAndRowBelow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 2)
	eng.KeyPress("d")
	eng.KeyPress("j")
	if eng.Rows() != testRows-2 {
		t.Fatalf("dj: want %d rows, got %d", testRows-2, eng.Rows())
	}
}

func TestDKDeletesCurrentAndRowAbove(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 2)
	eng.KeyPress("d")
	eng.KeyPress("k")
	if eng.Rows() != testRows-2 {
		t.Fatalf("dk: want %d rows, got %d", testRows-2, eng.Rows())
	}
}

func TestD3JDeletesFourRows(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("d")
	eng.KeyPress("3")
	eng.KeyPress("j")
	if eng.Rows() != testRows-4 {
		t.Fatalf("d3j: want %d rows, got %d", testRows-4, eng.Rows())
	}
}

func TestDGGDeletesFromCursorToRow1ExcludingHeader(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 4)
	eng.KeyPress("d")
	eng.KeyPress("g")
	eng.KeyPress("g")
	if eng.Rows() != 1 {
		t.Fatalf("dgg: want header row only, got %d rows", eng.Rows())
	}
}

func TestDGDeletesToLastRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("d")
	eng.KeyPress("G")
	if eng.Rows() != 1 {
		t.Fatalf("dG from row 1: want header only, got %d rows", eng.Rows())
	}
}

func TestDollarClearsToEndOfRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "a")
	eng.SetCellValue(3, 2, "b")
	eng.SetCellValue(4, 2, "c")
	eng.SetCursor(2, 2)
	eng.KeyPress("d")
	eng.KeyPress("$")
	v2, _ := eng.CellValue(2, 2)
	v3, _ := eng.CellValue(3, 2)
	v4, _ := eng.CellValue(4, 2)
	if v2 != "" || v3 != "" || v4 != "" {
		t.Fatalf("d$: cells should be empty, got %q %q %q", v2, v3, v4)
	}
	if eng.CursorX() != 2 {
		t.Errorf("cursor x should be 2, got %d", eng.CursorX())
	}
}

func TestD0ClearsToStartOfRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(1, 2, "a")
	eng.SetCellValue(2, 2, "b")
	eng.SetCursor(2, 2)
	eng.KeyPress("d")
	eng.KeyPress("0")
	v1, _ := eng.CellValue(1, 2)
	v2, _ := eng.CellValue(2, 2)
	if v1 != "" || v2 != "" {
		t.Fatalf("d0: want empty, got %q %q", v1, v2)
	}
}

func TestDLClearsCurrentCell(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "x")
	eng.SetCursor(2, 2)
	eng.KeyPress("d")
	eng.KeyPress("l")
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Errorf("dl: want empty, got %q", v)
	}
}

func TestYJTwoRowYank(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 2)
	eng.KeyPress("y")
	eng.KeyPress("j")
	reg := eng.RegisterSnapshot()
	if !reg.Linewise || len(reg.Cells) != 2 {
		t.Fatalf("yj: want 2 linewise rows, got %+v", reg)
	}
}

func TestCJClearsTwoRowsAndInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 2)
	eng.KeyPress("c")
	eng.KeyPress("j")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("cj: insert mode, got %v", eng.Mode())
	}
	v, _ := eng.CellValue(0, 2)
	if v != "" {
		t.Errorf("row 2 col 0 should be cleared")
	}
}

func TestOperatorEscapeCancelsPending(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("d")
	eng.KeyPress(engine.KeyEsc)
	if eng.Rows() != testRows {
		t.Error("d then Escape should not delete rows")
	}
	if !eng.LastKeyCaptured() {
		t.Error("Escape after d should capture")
	}
}

func TestYankDollarToEndOfRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "a")
	eng.SetCellValue(3, 2, "b")
	eng.SetCursor(2, 2)
	eng.KeyPress("y")
	eng.KeyPress("$")
	reg := eng.RegisterSnapshot()
	if reg.Linewise || len(reg.Cells) != 1 || len(reg.Cells[0]) != testCols-2 {
		t.Fatalf("y$: register shape wrong: %+v", reg)
	}
}

func TestChangeDollarToEndInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "a")
	eng.SetCursor(2, 2)
	eng.KeyPress("c")
	eng.KeyPress("$")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("c$: insert, got %v", eng.Mode())
	}
}

// --- Phase 3c: undo/redo

func TestUndoEmptyStackNoOp(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	before := eng.CellsSnapshot()
	eng.KeyPress("u")
	if eng.UndoDepth() != 0 || eng.RedoDepth() != 0 {
		t.Fatalf("empty undo: depth undo=%d redo=%d", eng.UndoDepth(), eng.RedoDepth())
	}
	if len(eng.CellsSnapshot()) != len(before) {
		t.Error("u on empty stack should not change grid")
	}
}

func TestRedoEmptyStackNoOp(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("Ctrl+r")
	if eng.RedoDepth() != 0 {
		t.Errorf("redo depth want 0, got %d", eng.RedoDepth())
	}
}

func TestDDUndoRestoresRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(0, 1, "row1")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != r0-1 {
		t.Fatalf("after dd want %d rows, got %d", r0-1, eng.Rows())
	}
	eng.KeyPress("u")
	if eng.Rows() != r0 {
		t.Fatalf("after u want %d rows, got %d", r0, eng.Rows())
	}
	v, _ := eng.CellValue(0, 1)
	if v != "row1" {
		t.Errorf("cell restored want row1, got %q", v)
	}
}

func TestDDUndoRedo(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("d")
	eng.KeyPress("d")
	eng.KeyPress("u")
	eng.KeyPress("Ctrl+r")
	if eng.Rows() != r0-1 {
		t.Fatalf("after redo want %d rows, got %d", r0-1, eng.Rows())
	}
	if eng.RedoDepth() != 0 {
		t.Errorf("redo stack should be empty after redo, got %d", eng.RedoDepth())
	}
}

func TestXUndoRestoresCell(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "hello")
	eng.SetCursor(2, 2)
	eng.KeyPress("x")
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Fatalf("after x want empty, got %q", v)
	}
	eng.KeyPress("u")
	v, _ = eng.CellValue(2, 2)
	if v != "hello" {
		t.Errorf("after u want hello, got %q", v)
	}
}

func TestPUndoRemovesPastedRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(0, 1, "a")
	eng.SetCursor(0, 1)
	eng.KeyPress("y")
	eng.KeyPress("y")
	rAfterYank := eng.Rows()
	eng.KeyPress("p")
	if eng.Rows() != rAfterYank+1 {
		t.Fatalf("paste should add row, got rows=%d", eng.Rows())
	}
	eng.KeyPress("u")
	if eng.Rows() != rAfterYank {
		t.Fatalf("u should remove pasted row, got rows=%d", eng.Rows())
	}
}

func Test3DDUndoRestoresThreeRows(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("3")
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != r0-3 {
		t.Fatalf("3dd want %d rows, got %d", r0-3, eng.Rows())
	}
	eng.KeyPress("u")
	if eng.Rows() != r0 {
		t.Fatalf("u want %d rows, got %d", r0, eng.Rows())
	}
}

func TestDDTwiceUndoOnce(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("d")
	eng.KeyPress("d")
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.Rows() != r0-2 {
		t.Fatalf("two dd want %d rows", r0-2)
	}
	eng.KeyPress("u")
	if eng.Rows() != r0-1 {
		t.Fatalf("one u want %d rows, got %d", r0-1, eng.Rows())
	}
}

func TestUndoThenMutationClearsRedo(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("d")
	eng.KeyPress("d")
	eng.KeyPress("u")
	if eng.RedoDepth() != 1 {
		t.Fatalf("want redo=1, got %d", eng.RedoDepth())
	}
	eng.KeyPress("d")
	eng.KeyPress("d")
	if eng.RedoDepth() != 0 {
		t.Errorf("new dd should clear redo, got redo=%d", eng.RedoDepth())
	}
}

func TestYYDoesNotPushUndo(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("y")
	eng.KeyPress("y")
	if eng.UndoDepth() != 0 {
		t.Errorf("yy should not add undo, got depth=%d", eng.UndoDepth())
	}
}

func TestSetCellValueUndoableThenUndo(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(1, 1, "old")
	if !eng.SetCellValueUndoable(1, 1, "new") {
		t.Fatal("SetCellValueUndoable failed")
	}
	v, _ := eng.CellValue(1, 1)
	if v != "new" {
		t.Fatalf("want new, got %q", v)
	}
	eng.KeyPress("u")
	v, _ = eng.CellValue(1, 1)
	if v != "old" {
		t.Errorf("u want old, got %q", v)
	}
}

func TestUndoRedoFiresOnBufferChanged(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)
	eng.KeyPress("d")
	eng.KeyPress("d")
	n1 := listener.OnBufferChangedCounter
	eng.KeyPress("u")
	if listener.OnBufferChangedCounter <= n1 {
		t.Error("u should fire OnBufferChanged")
	}
	n2 := listener.OnBufferChangedCounter
	eng.KeyPress("Ctrl+r")
	if listener.OnBufferChangedCounter <= n2 {
		t.Error("redo should fire OnBufferChanged")
	}
}

// --- Phase 4: text objects (iw / aw on current cell)

func TestCIWClearsCellAndInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "cell")
	eng.SetCursor(2, 2)
	eng.KeyPress("c")
	eng.KeyPress("i")
	eng.KeyPress("w")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("ciw want insert, got %v", eng.Mode())
	}
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Errorf("cell should be cleared, got %q", v)
	}
	reg := eng.RegisterSnapshot()
	if reg.Linewise || len(reg.Cells) != 1 || len(reg.Cells[0]) != 1 || reg.Cells[0][0] != "cell" {
		t.Fatalf("register want non-linewise [cell], got %+v", reg)
	}
}

func TestDIWClearsCellNormal(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 2, "x")
	eng.SetCursor(2, 2)
	eng.KeyPress("d")
	eng.KeyPress("i")
	eng.KeyPress("w")
	if eng.Mode() != engine.ModeNormal {
		t.Fatalf("diw want normal, got %v", eng.Mode())
	}
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Errorf("want empty, got %q", v)
	}
}

func TestYIWYanksCellUnchanged(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(1, 2, "keep")
	eng.SetCursor(1, 2)
	eng.KeyPress("y")
	eng.KeyPress("i")
	eng.KeyPress("w")
	v, _ := eng.CellValue(1, 2)
	if v != "keep" {
		t.Errorf("cell unchanged want keep, got %q", v)
	}
	reg := eng.RegisterSnapshot()
	if reg.Linewise || len(reg.Cells) != 1 || len(reg.Cells[0]) != 1 || reg.Cells[0][0] != "keep" {
		t.Fatalf("yiw register wrong: %+v", reg)
	}
}

func TestCAWSameAsCIW(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(0, 1, "a")
	eng.SetCursor(0, 1)
	eng.KeyPress("c")
	eng.KeyPress("a")
	eng.KeyPress("w")
	if eng.Mode() != engine.ModeInsert {
		t.Fatal("caw should enter insert")
	}
	v, _ := eng.CellValue(0, 1)
	if v != "" {
		t.Errorf("caw should clear cell, got %q", v)
	}
}

func TestDAWSameAsDIW(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(3, 3, "b")
	eng.SetCursor(3, 3)
	eng.KeyPress("d")
	eng.KeyPress("a")
	eng.KeyPress("w")
	v, _ := eng.CellValue(3, 3)
	if v != "" {
		t.Errorf("daw want empty, got %q", v)
	}
}

func TestYAWSameAsYIW(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(2, 1, "z")
	eng.SetCursor(2, 1)
	eng.KeyPress("y")
	eng.KeyPress("a")
	eng.KeyPress("w")
	v, _ := eng.CellValue(2, 1)
	if v != "z" {
		t.Errorf("yaw cell unchanged want z, got %q", v)
	}
}

func TestDIWEmptyCell(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(2, 2)
	eng.KeyPress("d")
	eng.KeyPress("i")
	eng.KeyPress("w")
	reg := eng.RegisterSnapshot()
	if !reg.Linewise && len(reg.Cells) == 1 && len(reg.Cells[0]) == 1 && reg.Cells[0][0] == "" {
		// ok
	} else {
		t.Fatalf("diw empty cell register: %+v", reg)
	}
}

func TestCIWUndo(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(1, 1, "undo-me")
	eng.SetCursor(1, 1)
	eng.KeyPress("c")
	eng.KeyPress("i")
	eng.KeyPress("w")
	eng.KeyPress(engine.KeyEsc)
	eng.KeyPress("u")
	v, _ := eng.CellValue(1, 1)
	if v != "undo-me" {
		t.Errorf("u after ciw want undo-me, got %q", v)
	}
}

func TestDIThenEscapeCancelsWithoutMutation(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCellValue(0, 2, "stay")
	eng.SetCursor(0, 2)
	eng.KeyPress("d")
	eng.KeyPress("i")
	eng.KeyPress(engine.KeyEsc)
	v, _ := eng.CellValue(0, 2)
	if v != "stay" {
		t.Errorf("cell should stay, got %q", v)
	}
	if eng.Mode() != engine.ModeNormal {
		t.Errorf("want normal after Escape, got %v", eng.Mode())
	}
}

func TestStandaloneIEntersInsert(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.KeyPress("i")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("i alone want insert, got %v", eng.Mode())
	}
}

// --- Phase 5: keymap / remapping

func TestNnoremapXDeletesRow(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nnoremap("X", "dd")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("X")
	if eng.Rows() != r0-1 {
		t.Fatalf("X want %d rows, got %d", r0-1, eng.Rows())
	}
}

func TestNmapQDIW(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nmap("Q", "diw")
	eng.SetCellValue(2, 2, "gone")
	eng.SetCursor(2, 2)
	eng.KeyPress("Q")
	v, _ := eng.CellValue(2, 2)
	if v != "" {
		t.Errorf("Q diw want empty, got %q", v)
	}
}

func TestNnoremapHL(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nnoremap("H", "0")
	eng.Nnoremap("L", "$")
	eng.SetCursor(3, 2)
	eng.KeyPress("H")
	if eng.CursorX() != 0 {
		t.Errorf("H want col 0, got %d", eng.CursorX())
	}
	eng.KeyPress("L")
	if eng.CursorX() != testCols-1 {
		t.Errorf("L want last col, got %d", eng.CursorX())
	}
}

func TestNmapRecursiveABToDD(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nmap("A", "B")
	eng.Nmap("B", "dd")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("A")
	if eng.Rows() != r0-1 {
		t.Errorf("recursive A want delete row, rows=%d", eng.Rows())
	}
}

func TestNnoremapADoesNotExpandB(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nnoremap("A", "B")
	eng.Nmap("B", "dd")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("A")
	if eng.Rows() != r0 {
		t.Errorf("nnoremap A to B should not run B's map, rows=%d want %d", eng.Rows(), r0)
	}
}

func TestNmapZZTerminates(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nmap("z", "z")
	eng.KeyPress("z")
	// If this returns, recursion guard worked.
}

func TestNnoremapGXDD(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nnoremap("gx", "dd")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	eng.KeyPress("g")
	eng.KeyPress("x")
	if eng.Rows() != r0-1 {
		t.Errorf("gx want delete row, got rows=%d", eng.Rows())
	}
}

func TestGGMotionAfterGXKeymapDrain(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nnoremap("gx", "dd")
	eng.SetCursor(0, 2)
	eng.KeyPress("g")
	eng.KeyPress("g")
	if eng.CursorY() != 0 {
		t.Errorf("gg want row 0, got %d", eng.CursorY())
	}
}

func TestInoremapJKToEscape(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Inoremap("jk", "<Escape>")
	eng.KeyPress("i")
	if eng.Mode() != engine.ModeInsert {
		t.Fatal("want insert")
	}
	eng.KeyPress("j")
	eng.KeyPress("k")
	if eng.Mode() != engine.ModeNormal {
		t.Errorf("jk want normal, got %v", eng.Mode())
	}
}

func TestInoremapTabToEscape(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Inoremap("<Tab>", "<Escape>")
	eng.KeyPress("i")
	eng.KeyPress("Tab")
	if eng.Mode() != engine.ModeNormal {
		t.Errorf("Tab imap want normal, got %v", eng.Mode())
	}
}

func TestVmapQuitsVisual(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Vmap("q", "<Escape>")
	eng.KeyPress("v")
	if eng.Mode() != engine.ModeVisual {
		t.Fatal("want visual")
	}
	eng.KeyPress("q")
	if eng.Mode() != engine.ModeNormal {
		t.Errorf("q want normal, got %v", eng.Mode())
	}
}

func TestUnmapRemovesBinding(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.Nmap("X", "dd")
	eng.SetCursor(0, 1)
	r0 := eng.Rows()
	if !eng.Unmap(engine.ModeNormal, "X") {
		t.Fatal("Unmap")
	}
	eng.KeyPress("X")
	if eng.Rows() != r0 {
		t.Errorf("after unmap X should not delete, rows=%d", eng.Rows())
	}
}

func TestJKMotionUnchangedWithoutImap(t *testing.T) {
	eng := newTestEngine(testCols, testRows)
	eng.SetCursor(0, 1)
	eng.KeyPress("j")
	if eng.CursorY() != 2 {
		t.Errorf("j want row 2, got %d", eng.CursorY())
	}
}
