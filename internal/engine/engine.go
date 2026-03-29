package engine

type Engine struct {
	listeners []EventListener
	mode Mode
}

type EventListener interface {
	OnModeChanged()
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
		listeners: []EventListener{},
		mode: ModeNormal,
	}
}

func (eng *Engine) AddListener(listener EventListener) {
	eng.listeners = append(eng.listeners, listener)
}

func (eng *Engine) Mode() Mode {
	return eng.mode
}

func (eng *Engine) setMode(mode Mode) {
	eng.mode = mode

	for _, listener := range eng.listeners {
		listener.OnModeChanged()
	}
}

func (eng *Engine) KeyPress(key string) {
	switch eng.mode {

	case ModeNormal:
		switch key {

		case "i":
			eng.setMode(ModeInsert)
			// TODO : handle cursor position

		case "a":
			eng.setMode(ModeInsert)
			// TODO : handle cursor position

		case "v":
			eng.setMode(ModeVisual)
		}

	case ModeInsert:
		if key == KeyEsc {
			eng.setMode(ModeNormal)
		}

	case ModeVisual:
		if key == KeyEsc {
			eng.setMode(ModeNormal)
		}
	}
}
