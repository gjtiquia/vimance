package engine

import "testing"

func TestTrieInsertExact(t *testing.T) {
	var tr Trie
	tr.Insert([]string{"a"}, 1)
	mr, v := tr.Match([]string{"a"})
	if mr != MatchExact || v != 1 {
		t.Fatalf("MatchExact, got mr=%v v=%v", mr, v)
	}
}

func TestTriePrefixVsExact(t *testing.T) {
	var tr Trie
	tr.Insert([]string{"g", "g"}, "gg")
	mr, _ := tr.Match([]string{"g"})
	if mr != MatchPrefix {
		t.Fatalf("expected MatchPrefix for g, got %v", mr)
	}
	mr2, v := tr.Match([]string{"g", "g"})
	if mr2 != MatchExact || v != "gg" {
		t.Fatalf("expected MatchExact gg, got mr=%v v=%v", mr2, v)
	}
}

func TestTrieMatchNone(t *testing.T) {
	var tr Trie
	tr.Insert([]string{"h"}, 1)
	mr, _ := tr.Match([]string{"x"})
	if mr != MatchNone {
		t.Fatalf("expected MatchNone, got %v", mr)
	}
}
