// Package ui contiene el modelo Bubble Tea (Elm-like) de Vim Master: pantallas
// y su enrutado. No contiene lógica del motor de Vim ni de puntuación.
package ui

import (
	"os/user"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Rickemtz/vim-master/exercises"
	"github.com/Rickemtz/vim-master/internal/exercise"
	"github.com/Rickemtz/vim-master/internal/storage"
)

type screen int

const (
	screenMenu screen = iota
	screenDuration
	screenLesson
	screenResults
	screenLeaderboard
)

// App es el modelo raíz: enruta entre el menú y las pantallas de juego, y
// persiste el progreso y el ranking al terminar una sesión.
const (
	minWidth  = 80
	minHeight = 24
)

type App struct {
	screen      screen
	menu        menuModel
	duration    durationModel
	lesson      lessonModel
	results     resultsModel
	leaderboard leaderboardModel
	err         string

	width, height int // último tamaño de terminal conocido (0 = aún sin WindowSizeMsg)

	store    *storage.Store // nil si no se pudo abrir ~/.config/vim-master
	progress storage.Progress
}

func NewApp() App {
	a := App{menu: newMenuModel(), progress: storage.DefaultProgress()}
	if store, err := storage.New(); err == nil {
		a.store = store
		if p, lerr := store.LoadProgress(); lerr == nil {
			a.progress = p
		}
	}
	return a
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if sizeMsg, ok := msg.(tea.WindowSizeMsg); ok {
		a.width, a.height = sizeMsg.Width, sizeMsg.Height
		return a, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "q", "f10":
			if a.screen == screenMenu {
				return a, tea.Quit
			}
		}
	}

	switch a.screen {
	case screenLesson:
		return a.updateLesson(msg)
	case screenResults:
		return a.updateResults(msg)
	case screenDuration:
		return a.updateDuration(msg)
	case screenLeaderboard:
		return a.updateLeaderboard(msg)
	default:
		return a.updateMenu(msg)
	}
}

func (a App) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "enter" {
		switch a.menu.cursor {
		case 0: // Lecciones
			module := a.currentModule()
			exs, err := loadModuleExercises(module)
			for err == nil && len(exs) == 0 && module > 1 {
				// El módulo desbloqueado más alto todavía no tiene
				// ejercicios (fase futura): cae al último módulo jugable.
				module--
				exs, err = loadModuleExercises(module)
			}
			if err != nil {
				a.err = err.Error()
				return a, nil
			}
			if len(exs) == 0 {
				a.err = strNoExercisesAvailable
				return a, nil
			}
			a.err = ""
			a.lesson = newLessonModel(exs)
			a.screen = screenLesson
			return a, a.lesson.Init()

		case 1: // Contrarreloj
			a.err = ""
			a.duration = newDurationModel()
			a.screen = screenDuration
			return a, nil

		case 2: // Golf
			exs, err := loadUnlockedExercises(a.progress)
			if err != nil {
				a.err = err.Error()
				return a, nil
			}
			if len(exs) == 0 {
				a.err = strNoExercisesAvailable
				return a, nil
			}
			a.err = ""
			a.lesson = newGolfModel(exs)
			a.screen = screenLesson
			return a, a.lesson.Init()

		case 3: // Examen final
			exs, err := loadAllExercises()
			if err != nil {
				a.err = err.Error()
				return a, nil
			}
			if len(exs) == 0 {
				a.err = strNoExercisesAvailable
				return a, nil
			}
			a.err = ""
			a.lesson = newExamModel(exs)
			a.screen = screenLesson
			return a, a.lesson.Init()

		case 4: // Ranking
			scores := storage.DefaultScores()
			if a.store != nil {
				if s, err := a.store.LoadScores(); err == nil {
					scores = s
				}
			}
			a.err = ""
			a.leaderboard = newLeaderboardModel(scores)
			a.screen = screenLeaderboard
			return a, nil

		default:
			a.err = strModeNotAvailable
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.menu, cmd = a.menu.Update(msg)
	return a, cmd
}

func (a App) updateLeaderboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "f10" {
		a.screen = screenMenu
		return a, nil
	}
	var cmd tea.Cmd
	a.leaderboard, cmd = a.leaderboard.Update(msg)
	return a, cmd
}

func (a App) updateDuration(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			exs, err := loadUnlockedExercises(a.progress)
			if err != nil {
				a.err = err.Error()
				a.screen = screenMenu
				return a, nil
			}
			if len(exs) == 0 {
				a.err = strNoExercisesAvailable
				a.screen = screenMenu
				return a, nil
			}
			a.err = ""
			a.lesson = newTimeAttackModel(exs, a.duration.Seconds())
			a.screen = screenLesson
			return a, a.lesson.Init()
		case "f10":
			a.screen = screenMenu
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.duration, cmd = a.duration.Update(msg)
	return a, cmd
}

func (a App) updateLesson(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.lesson, cmd = a.lesson.Update(msg)
	if a.lesson.Finished() {
		a.results = a.finishSession()
		a.screen = screenResults
		return a, nil
	}
	if a.lesson.exitToMenu {
		a.screen = screenMenu
		return a, nil
	}
	return a, cmd
}

// finishSession persiste el progreso y el ranking de la sesión que acaba
// de terminar, y arma la pantalla de resultados con lo que cambió. Solo
// Lecciones desbloquea módulos (Contrarreloj y Golf mezclan varios
// módulos a la vez, así que no hay un "mejor rango de módulo" que
// registrar ahí).
func (a *App) finishSession() resultsModel {
	summary := a.lesson.Summary()

	for _, id := range a.lesson.CompletedIDs() {
		a.progress.MarkCompleted(id)
	}

	var improvedRank, unlockedNext bool
	var nextModule int
	mode := storage.ModeGolf
	official := false

	switch a.lesson.Kind() {
	case kindLecciones:
		mode = storage.ModeLecciones
		module := a.lesson.Module()
		wasUnlocked := a.progress.IsUnlocked(module + 1)
		improvedRank = a.progress.RecordModuleResult(module, string(summary.Rank))
		unlockedNext = !wasUnlocked && a.progress.IsUnlocked(module+1)
		nextModule = module + 1
	case kindTimeAttack:
		mode = storage.ModeContrarreloj
	case kindGolf:
		mode = storage.ModeGolf
	case kindExam:
		mode = storage.ModeExamen
		official = true
	}

	isNewRecord := false
	if a.store != nil {
		_ = a.store.SaveProgress(a.progress)

		scores, _ := a.store.LoadScores()
		entry := storage.ScoreEntry{
			Name:  currentUserName(),
			Score: summary.TotalPoints,
			Rank:  string(summary.Rank),
			Date:  time.Now(),
		}
		isNewRecord = scores.AddEntry(mode, entry)
		_ = a.store.SaveScores(scores)
	}

	return newResultsModel(summary, isNewRecord, improvedRank, unlockedNext, nextModule, official)
}

func (a App) updateResults(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "f10" {
		a.screen = screenMenu
	}
	return a, nil
}

func (a App) View() string {
	// Antes de tener un WindowSizeMsg (width==0) asumimos que el
	// terminal es suficientemente grande, para no parpadear el aviso al
	// arrancar.
	if a.width > 0 && (a.width < minWidth || a.height < minHeight) {
		return smallTerminalWarning(a.width, a.height)
	}

	switch a.screen {
	case screenLesson:
		return a.lesson.View()
	case screenResults:
		return a.results.View()
	case screenDuration:
		return a.duration.View()
	case screenLeaderboard:
		return a.leaderboard.View()
	default:
		v := a.menu.View()
		if a.err != "" {
			v += "\n\n" + helpStyle.Render(a.err)
		}
		return v
	}
}

// currentModule devuelve el módulo desbloqueado más alto: es "por dónde
// va" el jugador en el modo Lecciones.
func (a App) currentModule() int {
	best := 1
	for _, m := range a.progress.UnlockedModules {
		if m > best {
			best = m
		}
	}
	return best
}

func loadModuleExercises(module int) ([]exercise.Exercise, error) {
	all, err := exercise.LoadAll(exercises.FS)
	if err != nil {
		return nil, err
	}
	var out []exercise.Exercise
	for _, ex := range all {
		if ex.Module == module {
			out = append(out, ex)
		}
	}
	return out, nil
}

// loadUnlockedExercises reúne los ejercicios de todos los módulos
// desbloqueados (Contrarreloj y Golf mezclan varios módulos, a
// diferencia de Lecciones).
func loadUnlockedExercises(progress storage.Progress) ([]exercise.Exercise, error) {
	all, err := exercise.LoadAll(exercises.FS)
	if err != nil {
		return nil, err
	}
	var out []exercise.Exercise
	for _, ex := range all {
		if progress.IsUnlocked(ex.Module) {
			out = append(out, ex)
		}
	}
	return out, nil
}

// loadAllExercises reúne los ejercicios de todos los módulos, sin
// filtrar por desbloqueo (usado por Examen final: "20 ejercicios mixtos
// de todos los módulos").
func loadAllExercises() ([]exercise.Exercise, error) {
	return exercise.LoadAll(exercises.FS)
}

func smallTerminalWarning(width, height int) string {
	return timerDangerStyle.Render(strTerminalTooSmall(width, height)) + "\n" +
		helpStyle.Render(strTerminalMinSize(minWidth, minHeight))
}

func currentUserName() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return strDefaultPlayerName
}
