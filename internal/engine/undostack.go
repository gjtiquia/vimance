package engine

// UndoEntry is a snapshot of grid + cursor for linear undo/redo.
type UndoEntry struct {
	Cells   [][]string
	CursorX int
	CursorY int
}

// UndoStack holds separate undo and redo histories; a new mutation clears redo.
type UndoStack struct {
	undos []UndoEntry
	redos []UndoEntry
}

// PushUndo appends a snapshot to the undo stack.
func (s *UndoStack) PushUndo(e UndoEntry) {
	s.undos = append(s.undos, e)
}

// PopUndo removes and returns the newest undo entry, if any.
func (s *UndoStack) PopUndo() (UndoEntry, bool) {
	if len(s.undos) == 0 {
		return UndoEntry{}, false
	}
	i := len(s.undos) - 1
	e := s.undos[i]
	s.undos = s.undos[:i]
	return e, true
}

// PushRedo appends a snapshot to the redo stack.
func (s *UndoStack) PushRedo(e UndoEntry) {
	s.redos = append(s.redos, e)
}

// PopRedo removes and returns the newest redo entry, if any.
func (s *UndoStack) PopRedo() (UndoEntry, bool) {
	if len(s.redos) == 0 {
		return UndoEntry{}, false
	}
	i := len(s.redos) - 1
	e := s.redos[i]
	s.redos = s.redos[:i]
	return e, true
}

// ClearRedo drops the redo branch (call before a new mutation).
func (s *UndoStack) ClearRedo() {
	s.redos = s.redos[:0]
}

// UndoLen returns the number of undo steps available.
func (s *UndoStack) UndoLen() int {
	return len(s.undos)
}

// RedoLen returns the number of redo steps available.
func (s *UndoStack) RedoLen() int {
	return len(s.redos)
}
