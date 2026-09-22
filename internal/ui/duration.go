package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// durationModel es la pantalla previa a Contrarreloj: elegir cuántos
// segundos dura la partida.
type durationModel struct {
	options []int // segundos
	cursor  int
}

func newDurationModel() durationModel {
	return durationModel{options: []int{60, 120, 300}}
}

func (m durationModel) Seconds() int { return m.options[m.cursor] }

func (m durationModel) Update(msg tea.Msg) (durationModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m durationModel) View() string {
	s := titleStyle.Render(strDurationTitle) + "\n\n" + strDurationPrompt + "\n\n"
	for i, secs := range m.options {
		label := strDurationLabel(secs)
		if i == m.cursor {
			s += selectedItemStyle.Render("> "+label) + "\n"
		} else {
			s += itemStyle.Render(label) + "\n"
		}
	}
	s += "\n" + helpStyle.Render(strDurationHelp)
	return s
}
