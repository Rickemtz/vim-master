package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// menuModel es la pantalla inicial: elige el modo de juego o el ranking.
type menuModel struct {
	options []string
	cursor  int
}

func newMenuModel() menuModel {
	return menuModel{options: menuOptions}
}

func (m menuModel) Init() tea.Cmd {
	return nil
}

func (m menuModel) Update(msg tea.Msg) (menuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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

func (m menuModel) View() string {
	s := titleStyle.Render(appTitle) + "\n\n"
	for i, opt := range m.options {
		if i == m.cursor {
			s += selectedItemStyle.Render("> "+opt) + "\n"
		} else {
			s += itemStyle.Render(opt) + "\n"
		}
	}
	s += "\n" + helpStyle.Render(strMenuHelp)
	return s
}
