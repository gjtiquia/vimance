package engine

import "testing"

func TestValidateRectangularGrid(t *testing.T) {
	cols, rows, ok := validateRectangularGrid([][]string{
		{"a", "b"},
		{"c", "d"},
	})
	if !ok || cols != 2 || rows != 2 {
		t.Fatalf("expected 2x2 ok, got cols=%d rows=%d ok=%v", cols, rows, ok)
	}
	_, _, ok = validateRectangularGrid([][]string{
		{"a", "b"},
		{"c"},
	})
	if ok {
		t.Fatal("jagged grid should fail")
	}
	_, _, ok = validateRectangularGrid(nil)
	if ok {
		t.Fatal("nil should fail")
	}
}

func TestStubDataSourceLoad(t *testing.T) {
	var s StubDataSource
	cells := s.Load()
	cols, rows, ok := validateRectangularGrid(cells)
	if !ok || cols != 6 || rows != 5 {
		t.Fatalf("stub grid: expected 6x5, got %dx%d ok=%v", cols, rows, ok)
	}
}

func TestStaticDataSourceSaveRoundTrip(t *testing.T) {
	s := &StaticDataSource{
		Cells: [][]string{
			{"x", "y"},
			{"1", "2"},
		},
	}
	eng := New(s)
	eng.SetCellValue(0, 1, "edited")
	if err := eng.SaveBuffer(); err != nil {
		t.Fatal(err)
	}
	if s.Cells[1][0] != "edited" {
		t.Fatalf("expected save to update StaticDataSource, got %q", s.Cells[1][0])
	}
}
