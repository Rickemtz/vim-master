package engine

// Mode enumera los modos de Vim. E1 solo implementa comportamiento para
// ModeNormal, ModeInsert y ModeCommand; el resto se añaden en fases
// posteriores (E7 visual).
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeVisualLine
	ModeVisualBlock
	ModeCommand
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "Normal"
	case ModeInsert:
		return "Insert"
	case ModeVisual:
		return "Visual"
	case ModeVisualLine:
		return "Visual Line"
	case ModeVisualBlock:
		return "Visual Block"
	case ModeCommand:
		return "Command"
	default:
		return "?"
	}
}
