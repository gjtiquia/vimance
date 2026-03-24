package engine_test

import (
	"testing"

	"github.com/gjtiquia/vimance/internal/engine"
)

func TestInitialState(t *testing.T) {
	eng := engine.New()

	if eng.Mode() != engine.ModeNormal {
		t.Errorf("expected initial mode to be normal, got %v", eng.Mode())
	}
}

func TestModeSwitching(t *testing.T) {
	eng := engine.New()

	// Switch to insert mode
	eng.KeyPress("i")
	if eng.Mode() != engine.ModeInsert {
		t.Fatalf("expected mode to be insert, got %v", eng.Mode())
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
}
