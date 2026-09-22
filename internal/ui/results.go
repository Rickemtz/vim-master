package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/kyrcovarick/vimdojo/internal/scoring"
)

// resultsModel muestra el puntaje final de una sesión terminada. No hay
// interacción propia: F10 (manejado en App) vuelve al menú.
type resultsModel struct {
	summary      scoring.Summary
	isNewRecord  bool
	unlockedNext bool
	nextModule   int
	improvedRank bool
	official     bool // Examen final: el rango mostrado es "oficial"
}

func newResultsModel(s scoring.Summary, isNewRecord, improvedRank, unlockedNext bool, nextModule int, official bool) resultsModel {
	return resultsModel{
		summary:      s,
		isNewRecord:  isNewRecord,
		improvedRank: improvedRank,
		unlockedNext: unlockedNext,
		nextModule:   nextModule,
		official:     official,
	}
}

func (m resultsModel) View() string {
	s := m.summary
	var b strings.Builder

	title := strResultsTitle
	if m.official {
		title = strResultsTitleExam
	}
	b.WriteString(titleStyle.Render(title) + "\n\n")

	rankLabel := strLabelRank
	if m.official {
		rankLabel = strLabelRankOfficial
	}
	fmt.Fprintf(&b, "Puntaje:     %d\n", s.TotalPoints)
	fmt.Fprintf(&b, "%s %s\n", rankLabel, rankStyle(s.Rank).Render(string(s.Rank)))
	fmt.Fprintf(&b, "Porcentaje:  %.1f%%\n\n", s.Percentage)

	fmt.Fprintf(&b, "Tiempo total:      %s\n", time.Duration(s.TotalTimeSec*float64(time.Second)).Round(time.Second))
	fmt.Fprintf(&b, "Teclas:            %d (par %d)\n", s.TotalKeystrokes, s.TotalPar)
	fmt.Fprintf(&b, "Precisión:         %d/%d (%.0f%%)\n", s.Completed, s.Attempted, s.Accuracy)
	fmt.Fprintf(&b, "Combo máximo:      x%.1f\n\n", s.MaxCombo)

	if len(s.Commands) > 0 {
		b.WriteString(strCommandsUsedLabel + "\n")
		for _, c := range topCommands(s.Commands, 5) {
			fmt.Fprintf(&b, "  %-6s %d\n", c.name, c.count)
		}
		b.WriteString("\n")
	}

	if m.isNewRecord {
		b.WriteString(titleStyle.Render(strNewRecord) + "\n")
	} else if m.improvedRank {
		b.WriteString(titleStyle.Render(strImprovedRank) + "\n")
	}
	if m.unlockedNext {
		fmt.Fprintf(&b, "%s\n", titleStyle.Render(strModuleUnlocked(m.nextModule)))
	}
	if m.isNewRecord || m.improvedRank || m.unlockedNext {
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render(strHelpBackToMenu))
	return b.String()
}

func rankStyle(r scoring.Rank) lipgloss.Style {
	switch r {
	case scoring.RankS, scoring.RankA:
		return titleStyle
	default:
		return helpStyle
	}
}

type commandCount struct {
	name  string
	count int
}

func topCommands(commands map[string]int, n int) []commandCount {
	list := make([]commandCount, 0, len(commands))
	for name, count := range commands {
		list = append(list, commandCount{name, count})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].count != list[j].count {
			return list[i].count > list[j].count
		}
		return list[i].name < list[j].name
	})
	if len(list) > n {
		list = list[:n]
	}
	return list
}
