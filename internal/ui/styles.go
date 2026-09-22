package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42"))

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Bold(true).
				Foreground(lipgloss.Color("212"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	// cursorStyle simula un cursor "bloque" (modo Normal): vídeo inverso.
	cursorStyle = lipgloss.NewStyle().
			Reverse(true)

	// insertCursorStyle simula un cursor "barra" (modo Insert): un
	// terminal de texto no puede dibujar una barra fina entre caracteres,
	// así que se aproxima con subrayado para distinguirlo del bloque.
	insertCursorStyle = lipgloss.NewStyle().
				Underline(true).
				Bold(true)

	// selectionStyle resalta el texto seleccionado en modo Visual (sin
	// ser la celda exacta del cursor, que usa cursorStyle).
	selectionStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237"))

	timerWarnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220"))

	timerDangerStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	// diffMatchStyle / diffPendingStyle colorean el objetivo: lo que el
	// buffer actual ya iguala (verde) y lo que falta (rojo).
	diffMatchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	diffPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203"))
)
