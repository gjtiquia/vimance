package engine

import "testing"

func emptyGridEngine(cols, rows int) Engine {
	cells := make([][]string, rows)
	for y := range cells {
		cells[y] = make([]string, cols)
	}
	return New(&StaticDataSource{Cells: cells})
}

func TestNormalKeyHandlerGG(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	eng.moveCursorTo(3, 4)

	kh := newNormalKeyHandler()
	if r := kh.Feed(&eng, "g"); r != ParseIncomplete {
		t.Fatalf("first g: ParseIncomplete, got %v", r)
	}
	if r := kh.Feed(&eng, "g"); r != ParseExecuted {
		t.Fatalf("second g: ParseExecuted, got %v", r)
	}
	if eng.CursorX() != 3 || eng.CursorY() != 0 {
		t.Fatalf("gg: expected (3,0), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestNormalKeyHandlerInvalidThenRetry(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	kh := newNormalKeyHandler()
	_ = kh.Feed(&eng, "g")
	// "x" is a normal-mode command; use an unknown key after incomplete "g".
	if r := kh.Feed(&eng, "z"); r != ParseInvalid {
		t.Fatalf("gz: expected ParseInvalid after retry, got %v", r)
	}
}

func TestKeyPressLastKeyCaptured(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	eng.KeyPress("z")
	if eng.LastKeyCaptured() {
		t.Fatal("unknown key should not capture")
	}
	eng.KeyPress("l")
	if !eng.LastKeyCaptured() {
		t.Fatal("l should capture")
	}
}

func TestKeyPressGGIncompleteCaptures(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	eng.KeyPress("g")
	if !eng.LastKeyCaptured() {
		t.Fatal("incomplete gg prefix should capture")
	}
	eng.KeyPress("g")
	if !eng.LastKeyCaptured() {
		t.Fatal("completing gg should capture")
	}
}

func TestCountThreeJMovesDown(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	eng.KeyPress("3")
	eng.KeyPress("j")
	if eng.CursorX() != 0 || eng.CursorY() != 3 {
		t.Fatalf("3j: expected (0,3), got (%d,%d)", eng.CursorX(), eng.CursorY())
	}
}

func TestCountFiveGToLineFive(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	eng.KeyPress("5")
	eng.KeyPress("G")
	if eng.CursorY() != 4 {
		t.Fatalf("5G: expected last row y=4, got %d", eng.CursorY())
	}
}

func TestCountFiveGGToRowFive(t *testing.T) {
	eng := emptyGridEngine(6, 5)
	for range 3 {
		eng.KeyPress("j")
	}
	if eng.CursorY() != 3 {
		t.Fatalf("setup y=3, got %d", eng.CursorY())
	}
	eng.KeyPress("5")
	eng.KeyPress("g")
	eng.KeyPress("g")
	if eng.CursorY() != 4 {
		t.Fatalf("5gg: expected y=4 (line 5), got %d", eng.CursorY())
	}
}
