package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kyrcovarick/vimdojo/internal/storage"
)

// leaderboardModel muestra el top 10 de cada modo de juego. Navega entre
// modos con ←/→; F10 (manejado en App) vuelve al menú.
type leaderboardModel struct {
	scores storage.Scores
	cursor int
}

func newLeaderboardModel(scores storage.Scores) leaderboardModel {
	return leaderboardModel{scores: scores}
}

func (m leaderboardModel) Update(msg tea.Msg) (leaderboardModel, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor < len(leaderboardModeLabels)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m leaderboardModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(strLeaderboardTitle) + "\n\n")

	var tabs strings.Builder
	for i, mode := range leaderboardModeLabels {
		if i > 0 {
			tabs.WriteString("   ")
		}
		if i == m.cursor {
			tabs.WriteString(selectedItemStyle.Render(mode.Label))
		} else {
			tabs.WriteString(itemStyle.Render(mode.Label))
		}
	}
	b.WriteString(tabs.String() + "\n\n")

	entries := m.scores.Top(leaderboardModeLabels[m.cursor].Key)
	if len(entries) == 0 {
		b.WriteString(helpStyle.Render(strLeaderboardEmpty) + "\n")
	} else {
		for i, e := range entries {
			fmt.Fprintf(&b, "%2d. %-15s %6d pts  %s  %s\n",
				i+1, e.Name, e.Score, e.Rank, e.Date.Format("2006-01-02"))
		}
	}

	b.WriteString("\n" + helpStyle.Render(strLeaderboardHelp))
	return b.String()
}
