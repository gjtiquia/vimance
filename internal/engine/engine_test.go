package engine_test

import (
	"testing"

	"github.com/gjtiquia/vimance/internal/engine"
)

const testCols = 6
const testRows = 5

type TestEngineEventListener struct {
	OnModeChangedCounter     int
	OnCursorMovedCounter     int
	LastMode                 engine.Mode
	LastInsertPosition       engine.InsertPosition
	LastCursorX, LastCursorY int
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

func TestInitialState(t *testing.T) {
	eng := engine.New(testCols, testRows)

	if eng.Mode() != engine.ModeNormal {
		t.Errorf("expected initial mode to be normal, got %v", eng.Mode())
	}
	if eng.CursorX() != 0 || eng.CursorY() != 0 {
		t.Errorf("expected initial cursor (0,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestModeSwitching(t *testing.T) {

	eng := engine.New(testCols, testRows)

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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)

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
	eng := engine.New(testCols, testRows)
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

func TestEnterEntersInsertAfterLikeA(t *testing.T) {
	eng := engine.New(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.KeyPress("Enter")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected mode to be insert, got %v", eng.Mode())
	}
	if listener.LastInsertPosition != engine.InsertPositionAfter {
		t.Errorf("expected insert position after for Enter, got %v", listener.LastInsertPosition)
	}
}

func TestSetCursorMovesAndRespectsBounds(t *testing.T) {
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
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
	eng := engine.New(testCols, testRows)
	listener := TestEngineEventListener{}
	eng.AddListener(&listener)

	eng.SetCursorAndEdit(4, 3)
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected insert mode, got %v", eng.Mode())
	}
	if eng.CursorX() != 4 || eng.CursorY() != 3 {
		t.Errorf("expected cursor (4,3), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
	if listener.LastInsertPosition != engine.InsertPositionAfter {
		t.Errorf("expected insert position after, got %v", listener.LastInsertPosition)
	}
}

func TestSetCursorAndEditFromInsertMode(t *testing.T) {
	eng := engine.New(testCols, testRows)
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
	if listener.LastInsertPosition != engine.InsertPositionAfter {
		t.Errorf("expected insert position after, got %v", listener.LastInsertPosition)
	}
}
