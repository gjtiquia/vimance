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
