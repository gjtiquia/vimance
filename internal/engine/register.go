package engine

// Register holds yanked or deleted content for paste (unnamed register).
type Register struct {
	Cells    [][]string
	Linewise bool
}
