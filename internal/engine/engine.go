package engine

type Engine struct {
	mode Mode
}

type Mode uint

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
)

// RuneEsc is the rune value for the Escape key.
const RuneEsc rune = '\x1b'

func New() Engine {
	return Engine{
		mode: ModeNormal,
	}
}

func (eng *Engine) Mode() Mode {
	return eng.mode
}

// TODO : consider special keys and likely need to change to something else other than runes
// TODO : perhaps just strings for simplicity
func (eng *Engine) KeyPress(key rune) {
	switch eng.mode {

	case ModeNormal:
		if key == 'i' {
			eng.mode = ModeInsert
		} else if key == 'v' {
			eng.mode = ModeVisual
		}

	case ModeInsert:
		if key == RuneEsc { // Escape key
			eng.mode = ModeNormal
		}

	case ModeVisual:
		if key == RuneEsc { // Escape key
			eng.mode = ModeNormal
		}
	}
}
