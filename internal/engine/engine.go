package engine

type Engine struct {
	listeners []EventListener
	mode      Mode
}

type EventListener interface {
	OnModeChanged(mode Mode)
}

type Mode string

const (
	ModeNormal Mode = "n"
	ModeInsert Mode = "i"
	ModeVisual Mode = "v"
)

const KeyEsc string = "Escape"

func New() Engine {
	return Engine{
		listeners: []EventListener{},
		mode:      ModeNormal,
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
		listener.OnModeChanged(mode)
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
