package engine

// DataSource loads the initial grid and persists it (e.g. SQLite later). Stub uses in-memory sample data.
type DataSource interface {
	Load() [][]string
	Save(cells [][]string) error
}

// StubDataSource returns a fixed finance-style grid for development. Replace with a DB-backed implementation later.
type StubDataSource struct{}

func (StubDataSource) Load() [][]string {
	return [][]string{
		{"Name", "Category", "Amount", "Date", "Note", "Status"},
		{"Latte", "Food", "4.50", "2026-03-01", "Morning", "posted"},
		{"Rent", "Housing", "1200", "2026-03-01", "March", "pending"},
		{"Guitar", "Hobby", "299", "2026-03-10", "Used", "posted"},
		{"Cloud", "Services", "42", "2026-03-15", "VPS", "posted"},
	}
}

func (StubDataSource) Save(_ [][]string) error {
	return nil
}

// StaticDataSource is for tests: fixed grid, Save updates the slice in memory.
type StaticDataSource struct {
	Cells [][]string
}

func (s *StaticDataSource) Load() [][]string {
	return cloneCells(s.Cells)
}

func (s *StaticDataSource) Save(cells [][]string) error {
	s.Cells = cloneCells(cells)
	return nil
}

func cloneCells(src [][]string) [][]string {
	if len(src) == 0 {
		return nil
	}
	out := make([][]string, len(src))
	for y := range src {
		out[y] = append([]string(nil), src[y]...)
	}
	return out
}

func validateRectangularGrid(cells [][]string) (cols, rows int, ok bool) {
	rows = len(cells)
	if rows == 0 {
		return 0, 0, false
	}
	cols = len(cells[0])
	if cols == 0 {
		return 0, 0, false
	}
	for y := 1; y < rows; y++ {
		if len(cells[y]) != cols {
			return 0, 0, false
		}
	}
	return cols, rows, true
}
