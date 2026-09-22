// Todos los textos visibles al usuario viven en este archivo, para poder
// traducirlos más adelante sin tocar la lógica de cada pantalla.
package ui

import (
	"fmt"

	"github.com/kyrcovarick/vimdojo/internal/storage"
)

// General / errores, compartidos entre pantallas.
const (
	strNoExercisesAvailable = "Todavía no hay ejercicios disponibles"
	strModeNotAvailable     = "Este modo todavía no está disponible"
	strHelpBackToMenu       = "F10 volver al menú"
	strDefaultPlayerName    = "jugador" // cuando no se puede leer el usuario del sistema operativo
)

// Terminal demasiado pequeño (app.go).
func strTerminalTooSmall(width, height int) string {
	return fmt.Sprintf("Terminal demasiado pequeño: %dx%d", width, height)
}

func strTerminalMinSize(minW, minH int) string {
	return fmt.Sprintf("VimDojo necesita al menos %dx%d. Agranda la ventana.", minW, minH)
}

// Menú principal (menu.go).
const (
	appTitle    = "VimDojo"
	strMenuHelp = "↑/↓ navegar · F10/q salir"
)

var menuOptions = []string{"Lecciones", "Contrarreloj", "Golf", "Examen final", "Ranking"}

// Pantalla de duración de Contrarreloj (duration.go).
const (
	strDurationTitle  = "Contrarreloj"
	strDurationPrompt = "Elige la duración:"
	strDurationHelp   = "↑/↓ elegir · Enter empezar · F10 volver"
)

func strDurationLabel(seconds int) string { return fmt.Sprintf("%d segundos", seconds) }

// Pantalla de juego (lesson.go).
const (
	strSessionComplete = "¡Sesión completa!"
	strTimeUp          = "¡Se acabó el tiempo!"
	strObjectiveLabel  = "Objetivo:"
	strLessonHelp      = "F1 pista · F2 reiniciar · F10 salir al menú"
	strExamHelp        = "Examen final: sin pistas · F2 reiniciar · F10 salir al menú"
)

func strExerciseDone(points int) string       { return fmt.Sprintf("¡Bien hecho! +%d pts", points) }
func strExerciseLabel(cur int) string         { return fmt.Sprintf("Ejercicio %d", cur) }
func strExerciseLabelTotal(cur, tot int) string { return fmt.Sprintf("Ejercicio %d/%d", cur, tot) }

// Pantalla de resultados (results.go).
const (
	strResultsTitle      = "Resultado final"
	strResultsTitleExam  = "Examen final"
	strLabelRank         = "Rango:      "
	strLabelRankOfficial = "Rango oficial:"
	strNewRecord         = "¡Nuevo récord!"
	strImprovedRank      = "¡Mejoraste tu mejor rango en este módulo!"
	strCommandsUsedLabel = "Comandos más usados:"
)

func strModuleUnlocked(module int) string {
	return fmt.Sprintf("¡Módulo %d desbloqueado!", module)
}

// Ranking (leaderboard.go).
const (
	strLeaderboardTitle = "Ranking"
	strLeaderboardEmpty = "Todavía no hay puntajes en este modo."
	strLeaderboardHelp  = "←/→ cambiar de modo · F10 volver al menú"
)

// leaderboardModeLabels da el nombre visible de cada modo de storage, en
// el orden en que se muestran las pestañas del ranking.
var leaderboardModeLabels = []struct {
	Key   string
	Label string
}{
	{storage.ModeLecciones, "Lecciones"},
	{storage.ModeContrarreloj, "Contrarreloj"},
	{storage.ModeGolf, "Golf"},
	{storage.ModeExamen, "Examen final"},
}
