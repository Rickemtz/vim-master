package engine

// Special identifica teclas no imprimibles.
type Special int

const (
	KeyNone Special = iota
	KeyEsc
	KeyEnter
	KeyBackspace
	KeyCtrlR
	KeyCtrlV
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
)

// Key representa una tecla enviada al engine: o bien un rune imprimible
// (Special == KeyNone) o una tecla especial.
type Key struct {
	Rune    rune
	Special Special
}

func RuneKey(r rune) Key { return Key{Rune: r} }
func EscKey() Key        { return Key{Special: KeyEsc} }
func EnterKey() Key      { return Key{Special: KeyEnter} }
func BackspaceKey() Key  { return Key{Special: KeyBackspace} }
func CtrlRKey() Key      { return Key{Special: KeyCtrlR} }
func CtrlVKey() Key      { return Key{Special: KeyCtrlV} }
func UpKey() Key         { return Key{Special: KeyUp} }
func DownKey() Key       { return Key{Special: KeyDown} }
func LeftKey() Key       { return Key{Special: KeyLeft} }
func RightKey() Key      { return Key{Special: KeyRight} }

// IsArrow indica si la tecla es una flecha (para penalizar en scoring
// cuando el ejercicio las prohíbe).
func (k Key) IsArrow() bool {
	switch k.Special {
	case KeyUp, KeyDown, KeyLeft, KeyRight:
		return true
	default:
		return false
	}
}

// EventType clasifica lo que produjo un Feed.
type EventType int

const (
	EventNone EventType = iota
	EventMoved
	EventChanged
	EventModeChanged
	EventUnsupported
	EventWrite
	EventQuit
)

// Event es el resultado de procesar una tecla.
type Event struct {
	Type EventType
}

// Stats acumula el uso de teclas para el sistema de puntuación.
type Stats struct {
	TotalKeys int
	ArrowKeys int
	Commands  map[string]int
}
