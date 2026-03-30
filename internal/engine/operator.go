package engine

// OperatorContext is passed when a linewise doubled operator runs (dd, yy, cc).
type OperatorContext struct {
	Count      int
	CountGiven bool
}

// OperatorRegistry lists keys that start an operator-pending sequence (d, y, c).
type OperatorRegistry struct {
	ops map[string]struct{}
}

func NewOperatorRegistry() *OperatorRegistry {
	r := &OperatorRegistry{ops: map[string]struct{}{
		"d": {},
		"y": {},
		"c": {},
	}}
	return r
}

// IsOperator reports whether key begins an operator sequence in normal mode.
func (r *OperatorRegistry) IsOperator(key string) bool {
	if len(key) != 1 {
		return false
	}
	_, ok := r.ops[key]
	return ok
}
