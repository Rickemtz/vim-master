package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kyrcovarick/vimdojo/internal/engine"
)

// translateKey traduce una tecla de Bubble Tea a la representación que
// entiende el engine. El segundo valor es false para teclas que el engine
// aún no maneja (p.ej. F1/F2/F10, que la UI intercepta antes de llegar aquí).
func translateKey(msg tea.KeyMsg) (engine.Key, bool) {
	switch msg.Type {
	case tea.KeyEsc:
		return engine.EscKey(), true
	case tea.KeyEnter:
		return engine.EnterKey(), true
	case tea.KeyBackspace:
		return engine.BackspaceKey(), true
	case tea.KeyCtrlR:
		return engine.CtrlRKey(), true
	case tea.KeyCtrlV:
		return engine.CtrlVKey(), true
	case tea.KeyUp:
		return engine.UpKey(), true
	case tea.KeyDown:
		return engine.DownKey(), true
	case tea.KeyLeft:
		return engine.LeftKey(), true
	case tea.KeyRight:
		return engine.RightKey(), true
	case tea.KeySpace:
		return engine.RuneKey(' '), true
	case tea.KeyRunes:
		if len(msg.Runes) != 1 {
			return engine.Key{}, false
		}
		return engine.RuneKey(msg.Runes[0]), true
	default:
		return engine.Key{}, false
	}
}
