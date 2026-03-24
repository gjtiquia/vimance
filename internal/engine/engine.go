package engine

type Engine struct {
	Mode Mode
}

type Mode uint 

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
)

func New() Engine {
	return Engine{
		Mode: ModeNormal,
	}
}
