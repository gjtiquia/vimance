package engine

// SimpleCommandFunc is a normal-mode action that is not a motion (enter insert, visual, etc.).
type SimpleCommandFunc func(eng *Engine)

// SimpleCommandRegistry maps a single KeyPress string to a command.
type SimpleCommandRegistry struct {
	commands map[string]SimpleCommandFunc
}

func NewSimpleCommandRegistry() *SimpleCommandRegistry {
	r := &SimpleCommandRegistry{commands: map[string]SimpleCommandFunc{}}
	r.registerBuiltins()
	return r
}

func (r *SimpleCommandRegistry) Register(key string, fn SimpleCommandFunc) {
	r.commands[key] = fn
}

func (r *SimpleCommandRegistry) Get(key string) (SimpleCommandFunc, bool) {
	fn, ok := r.commands[key]
	return fn, ok
}

func (r *SimpleCommandRegistry) registerBuiltins() {
	r.Register("i", func(eng *Engine) {
		eng.setMode(ModeInsert, InsertPositionBefore)
	})
	r.Register("a", func(eng *Engine) {
		eng.setMode(ModeInsert, InsertPositionAfter)
	})
	r.Register("Enter", func(eng *Engine) {
		eng.setMode(ModeInsert, InsertPositionHighlight)
	})
	r.Register("v", func(eng *Engine) {
		eng.setMode(ModeVisual, InsertPositionNone)
	})
}
