package engine_test

import (
	"testing"

	"github.com/gjtiquia/vimance/internal/engine"
)

func TestInitialState(t *testing.T) {
	eng := engine.New()

	if eng.Mode != engine.ModeNormal {
		t.Errorf("expected initial mode to be normal, got %v", eng.Mode)
	}
}
