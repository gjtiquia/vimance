package engine

import (
	"reflect"
	"testing"
)

func TestParseKeys(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"dd", []string{"d", "d"}},
		{"<Escape>", []string{"Escape"}},
		{"jk", []string{"j", "k"}},
		{"d<Escape>", []string{"d", "Escape"}},
		{"<Ctrl+r>", []string{"Ctrl+r"}},
		{"ab", []string{"a", "b"}},
	}
	for _, tc := range tests {
		got := ParseKeys(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ParseKeys(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

func TestKeymapTableSetLookupDelete(t *testing.T) {
	var kt KeymapTable
	lhs := []string{"X"}
	rhs := []string{"d", "d"}
	kt.Set(lhs, rhs, false)

	mr, e := kt.Lookup([]string{"X"})
	if mr != MatchExact || e == nil || !reflect.DeepEqual(e.RHS, rhs) || e.Recursive {
		t.Fatalf("Lookup X: mr=%v entry=%v", mr, e)
	}

	mr, _ = kt.Lookup([]string{"g"})
	if mr != MatchNone {
		t.Errorf("Lookup g want None, got %v", mr)
	}

	kt.Set([]string{"g", "x"}, []string{"y"}, false)
	mr, _ = kt.Lookup([]string{"g"})
	if mr != MatchPrefix {
		t.Errorf("Lookup g with gx registered: want Prefix, got %v", mr)
	}
	mr, e = kt.Lookup([]string{"g", "x"})
	if mr != MatchExact || e.RHS[0] != "y" {
		t.Fatalf("Lookup gx: %v %+v", mr, e)
	}

	if !kt.Delete([]string{"X"}) {
		t.Fatal("Delete X")
	}
	mr, _ = kt.Lookup([]string{"X"})
	if mr != MatchNone {
		t.Errorf("after Delete X want None, got %v", mr)
	}
	if kt.Delete([]string{"X"}) {
		t.Error("second delete should fail")
	}
}
