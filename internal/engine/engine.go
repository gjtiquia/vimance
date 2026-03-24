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

const KeyEsc string = "esc"

func New() Engine {
	return Engine{
		mode: ModeNormal,
	}
}

func (eng *Engine) Mode() Mode {
	return eng.mode
}

func (eng *Engine) KeyPress(key string) {
	switch eng.mode {

	case ModeNormal:
		switch key {
		case "i":
			eng.mode = ModeInsert
			// TODO : handle cursor position
		case "a":
			eng.mode = ModeInsert
			// TODO : handle cursor position
		case "v":
			eng.mode = ModeVisual
		}

	case ModeInsert:
		if key == KeyEsc {
			eng.mode = ModeNormal
		}

	case ModeVisual:
		if key == KeyEsc {
			eng.mode = ModeNormal
		}
	}
}
