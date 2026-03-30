package engine

import "testing"

func TestKeyBufferAppendReset(t *testing.T) {
	var kb KeyBuffer
	kb.Append("g")
	kb.Append("g")
	if kb.Len() != 2 {
		t.Fatalf("len 2, got %d", kb.Len())
	}
	if kb.Keys()[0] != "g" || kb.Keys()[1] != "g" {
		t.Fatalf("keys: %v", kb.Keys())
	}
	kb.Reset()
	if kb.Len() != 0 {
		t.Fatalf("after reset len 0, got %d", kb.Len())
	}
}
