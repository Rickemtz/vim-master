package ui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Rickemtz/vim-master/internal/engine"
	"github.com/Rickemtz/vim-master/internal/exercise"
	"github.com/Rickemtz/vim-master/internal/scoring"
	"github.com/Rickemtz/vim-master/internal/session"
)

// lessonKind distingue los cuatro modos de juego que comparten esta
// pantalla: cambia cómo se cuenta el tiempo y cuándo termina la sesión.
type lessonKind int

const (
	kindLecciones  lessonKind = iota // timer por ejercicio; termina al agotar el módulo
	kindTimeAttack                   // un solo timer global; termina cuando se acaba
	kindGolf                         // sin timer; termina al agotar el pool de ejercicios
	kindExam                         // como Lecciones, pero mixto, sin pistas y con rango oficial
)

const examExerciseCount = 20

// lessonModel es la pantalla de juego: un ejercicio a la vez, con timer,
// editor emulado, objetivo y puntuación.
type lessonModel struct {
	kind   lessonKind
	sess   *session.Session
	acc    *scoring.Accumulator
	eng    *engine.Engine
	ex     exercise.Exercise
	module int // módulo de los ejercicios (Lecciones; 0 en Contrarreloj/Golf)

	timer         timer.Model   // Lecciones: por ejercicio. Contrarreloj: global (no se reinicia por ejercicio). Golf: sin uso.
	timerTotal    time.Duration // duración total del timer activo, para el color de aviso
	exerciseStart time.Time     // para medir el tiempo usado en Contrarreloj/Golf

	message      string
	hintIdx      int // siguiente pista a mostrar (para no repetir)
	hintsUsed    int // veces que se pidió pista en este ejercicio (penaliza)
	completedIDs []string
	finished     bool

	// exitToMenu se pone en true cuando el jugador pide salir con F10.
	exitToMenu bool
}

// newLessonModel crea una sesión de Lecciones: secuencial, dentro de un
// único módulo, con timer por ejercicio.
func newLessonModel(exs []exercise.Exercise) lessonModel {
	m := lessonModel{kind: kindLecciones, sess: session.New(exs), acc: scoring.NewAccumulator()}
	if len(exs) > 0 {
		m.module = exs[0].Module
	}
	m.loadCurrent()
	return m
}

// newTimeAttackModel crea una sesión de Contrarreloj: ejercicios
// aleatorios (repitiendo si hace falta) de exs hasta que se acaben
// seconds segundos globales.
func newTimeAttackModel(exs []exercise.Exercise, seconds int) lessonModel {
	m := lessonModel{kind: kindTimeAttack, sess: session.NewShuffled(exs, true), acc: scoring.NewAccumulator()}
	m.timer = timer.NewWithInterval(time.Duration(seconds)*time.Second, time.Second)
	m.timerTotal = time.Duration(seconds) * time.Second
	m.loadCurrent()
	return m
}

// newGolfModel crea una sesión de Golf: todos los ejercicios de exs una
// vez, en orden aleatorio, sin límite de tiempo.
func newGolfModel(exs []exercise.Exercise) lessonModel {
	m := lessonModel{kind: kindGolf, sess: session.NewShuffled(exs, false), acc: scoring.NewAccumulator()}
	m.loadCurrent()
	return m
}

// newExamModel crea el Examen final: 20 ejercicios mixtos de todos los
// módulos (sin repetir), timer por ejercicio como en Lecciones, pero sin
// pistas (F1 no hace nada) y con el rango "oficial".
func newExamModel(exs []exercise.Exercise) lessonModel {
	m := lessonModel{kind: kindExam, sess: session.New(pickRandom(exs, examExerciseCount)), acc: scoring.NewAccumulator()}
	m.loadCurrent()
	return m
}

func pickRandom(exs []exercise.Exercise, n int) []exercise.Exercise {
	shuffled := make([]exercise.Exercise, len(exs))
	copy(shuffled, exs)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	if len(shuffled) > n {
		shuffled = shuffled[:n]
	}
	return shuffled
}

func (m *lessonModel) loadCurrent() {
	ex, ok := m.sess.Current()
	if !ok {
		m.finished = true
		return
	}
	m.ex = ex
	m.eng = engine.New(ex.Initial, ex.CursorStart)
	m.exerciseStart = time.Now()
	if m.kind == kindLecciones || m.kind == kindExam {
		m.timer = timer.NewWithInterval(time.Duration(ex.TimeLimitSec)*time.Second, time.Second)
		m.timerTotal = time.Duration(ex.TimeLimitSec) * time.Second
	}
	m.message = ""
	m.hintIdx = 0
	m.hintsUsed = 0
}

func (m lessonModel) Init() tea.Cmd {
	if m.finished || m.kind == kindGolf {
		return nil
	}
	return m.timer.Init()
}

// Kind devuelve el modo de esta sesión (para que App decida cómo
// persistir el resultado).
func (m lessonModel) Kind() lessonKind { return m.kind }

// Finished indica si ya se jugaron todos los ejercicios de la sesión
// (Lecciones/Golf) o si se acabó el tiempo global (Contrarreloj).
func (m lessonModel) Finished() bool { return m.finished }

// Summary devuelve el resumen final de puntuación (válido una vez
// Finished() es true, pero se puede consultar en cualquier momento para
// ver el progreso acumulado).
func (m lessonModel) Summary() scoring.Summary { return m.acc.Summary() }

// Module devuelve el módulo de los ejercicios jugados en esta sesión (0
// en Contrarreloj/Golf, que mezclan varios módulos).
func (m lessonModel) Module() int { return m.module }

// CompletedIDs devuelve los ids de los ejercicios que se completaron
// (no los que agotaron el tiempo), para registrar el progreso.
func (m lessonModel) CompletedIDs() []string { return m.completedIDs }

func (m lessonModel) Update(msg tea.Msg) (lessonModel, tea.Cmd) {
	if m.finished {
		return m, nil
	}

	switch msg := msg.(type) {
	case timer.TickMsg:
		if m.kind == kindGolf {
			return m, nil
		}
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		switch m.kind {
		case kindGolf:
			return m, nil
		case kindTimeAttack:
			// Se acabó el reloj global: la sesión termina ya; el ejercicio
			// en curso, si no se llegó a completar, no se puntúa.
			m.finished = true
			return m, nil
		default:
			m.finishExercise(true)
			return m, m.Init()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "f10":
			m.exitToMenu = true
			return m, nil
		case "f2":
			m.loadCurrent()
			return m, m.Init()
		case "f1":
			if m.kind == kindExam {
				// El examen final es sin pistas.
				return m, nil
			}
			if len(m.ex.Hints) > 0 {
				m.message = m.ex.Hints[m.hintIdx]
				if m.hintIdx < len(m.ex.Hints)-1 {
					m.hintIdx++
				}
			}
			m.hintsUsed++
			return m, nil
		}

		if key, ok := translateKey(msg); ok {
			m.eng.Feed(key)
			if m.exerciseComplete() {
				m.finishExercise(false)
				return m, m.Init()
			}
		}
	}

	return m, nil
}

func (m lessonModel) exerciseComplete() bool {
	if m.ex.Type == exercise.TypeEdit || m.ex.Type == exercise.TypeBoth {
		if m.eng.Text() != m.ex.Target {
			return false
		}
	}
	if m.ex.Type == exercise.TypeCursor || m.ex.Type == exercise.TypeBoth {
		if m.ex.TargetCursor == nil || m.eng.Cursor() != *m.ex.TargetCursor {
			return false
		}
	}
	return true
}

// finishExercise puntúa el ejercicio actual (completado o por tiempo
// agotado; solo Lecciones llega aquí con timedOut=true, ya que
// Contrarreloj corta la sesión entera en su lugar), suma el resultado al
// acumulador de la sesión y avanza al siguiente.
func (m *lessonModel) finishExercise(timedOut bool) {
	stats := m.eng.Stats()
	in := scoring.ExerciseInput{
		Difficulty:     m.ex.Difficulty,
		TimeLimitSec:   m.ex.TimeLimitSec,
		TimeUsedSec:    m.timeUsedSec(),
		ParKeystrokes:  m.ex.ParKeystrokes,
		KeystrokesUsed: stats.TotalKeys,
		HintsUsed:      m.hintsUsed,
		ForbidArrows:   m.ex.ForbidArrows,
		ArrowKeys:      stats.ArrowKeys,
		ForbiddenCmds:  m.forbiddenCommandCount(stats),
		TimedOut:       timedOut,
	}
	res := m.acc.Add(in, stats.Commands)

	if timedOut {
		m.message = strTimeUp
	} else {
		m.message = strExerciseDone(res.Points)
		m.completedIDs = append(m.completedIDs, m.ex.ID)
	}

	m.sess.Advance()
	m.loadCurrent()
}

// timeUsedSec mide cuánto tardó el jugador en el ejercicio actual. En
// Golf no cuenta (no hay presión de tiempo, así que el bono de tiempo
// siempre es máximo: solo importa la eficiencia de teclas). En
// Contrarreloj se mide con el reloj de pared, porque el timer visible es
// el global, no uno por ejercicio.
func (m lessonModel) timeUsedSec() float64 {
	switch m.kind {
	case kindGolf:
		return 0
	case kindTimeAttack:
		used := time.Since(m.exerciseStart).Seconds()
		if used < 0 {
			return 0
		}
		return used
	default:
		total := float64(m.ex.TimeLimitSec)
		used := total - m.timer.Timeout.Seconds()
		switch {
		case used < 0:
			return 0
		case used > total:
			return total
		default:
			return used
		}
	}
}

// forbiddenCommandCount cuenta los comandos usados que no están en
// ex.Allowed. Una lista vacía significa "todos permitidos".
func (m lessonModel) forbiddenCommandCount(stats engine.Stats) int {
	if len(m.ex.Allowed) == 0 {
		return 0
	}
	allowed := make(map[string]bool, len(m.ex.Allowed))
	for _, c := range m.ex.Allowed {
		allowed[c] = true
	}
	count := 0
	for cmd, n := range stats.Commands {
		if !allowed[cmd] {
			count += n
		}
	}
	return count
}

func (m lessonModel) View() string {
	if m.finished {
		return titleStyle.Render(strSessionComplete) + "\n\n" +
			helpStyle.Render(strHelpBackToMenu)
	}

	var b strings.Builder

	fmt.Fprintf(&b, "%s", titleStyle.Render(m.progressLabel()))
	if m.kind != kindGolf {
		fmt.Fprintf(&b, "   %s", m.timerView())
	}
	fmt.Fprintf(&b, strFmtHUDScore, m.acc.Summary().TotalPoints, m.acc.Combo())
	fmt.Fprintf(&b, "%s\n\n", helpStyle.Render(m.ex.Title))

	b.WriteString(m.ex.Instructions)
	b.WriteString("\n\n")
	b.WriteString(renderBuffer(m.eng))
	b.WriteString("\n")

	fmt.Fprintf(&b, "%s\n%s\n\n", strObjectiveLabel, renderTarget(m.ex.Target, m.eng.Text()))

	fmt.Fprintf(&b, "-- %s --", strings.ToUpper(m.eng.Mode().String()))
	if pending := m.eng.PendingKeys(); pending != "" {
		fmt.Fprintf(&b, "  %s", pending)
	}
	b.WriteString("\n")

	if m.message != "" {
		b.WriteString(helpStyle.Render(m.message) + "\n")
	}
	if m.kind == kindExam {
		b.WriteString(helpStyle.Render(strExamHelp))
	} else {
		b.WriteString(helpStyle.Render(strLessonHelp))
	}

	return b.String()
}

// progressLabel muestra "Ejercicio X/N" en Lecciones y Golf (listas
// finitas), y solo un contador en Contrarreloj, donde el pool se repite
// y un "/N" sería engañoso.
func (m lessonModel) progressLabel() string {
	cur, total := m.sess.Progress()
	if m.kind == kindTimeAttack {
		return strExerciseLabel(cur)
	}
	return strExerciseLabelTotal(cur, total)
}

func (m lessonModel) timerView() string {
	remaining := m.timer.Timeout
	total := m.timerTotal
	s := fmt.Sprintf("%02.0f s", remaining.Seconds())

	if total <= 0 {
		return s
	}
	frac := float64(remaining) / float64(total)
	switch {
	case frac < 0.10:
		return timerDangerStyle.Render(s)
	case frac < 0.30:
		return timerWarnStyle.Render(s)
	default:
		return s
	}
}

// renderBuffer dibuja el buffer con números de línea, el cursor resaltado
// (bloque en Normal, barra aproximada con subrayado en Insert) y, en
// modo Visual, la selección resaltada con un fondo distinto.
func renderBuffer(e *engine.Engine) string {
	lines := strings.Split(e.Text(), "\n")
	cur := e.Cursor()
	mode := e.Mode()
	insert := mode == engine.ModeInsert
	visual := mode == engine.ModeVisual || mode == engine.ModeVisualLine || mode == engine.ModeVisualBlock
	var anchor engine.Pos
	if visual {
		anchor = e.VisualAnchor()
	}

	var b strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&b, "%3d  ", i+1)
		b.WriteString(renderLine([]rune(line), i, cur, insert, visual, mode, anchor))
		b.WriteString("\n")
	}
	return b.String()
}

func renderLine(runes []rune, lineIdx int, cur engine.Pos, insert, visual bool, mode engine.Mode, anchor engine.Pos) string {
	if len(runes) == 0 {
		if lineIdx == cur.Line {
			return cellStyle(insert).Render(" ")
		}
		return ""
	}

	var b strings.Builder
	for col, r := range runes {
		pos := engine.Pos{Line: lineIdx, Col: col}
		switch {
		case lineIdx == cur.Line && col == cur.Col:
			b.WriteString(cellStyle(insert).Render(string(r)))
		case visual && inVisualSelection(mode, anchor, cur, pos):
			b.WriteString(selectionStyle.Render(string(r)))
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func cellStyle(insert bool) lipgloss.Style {
	if insert {
		return insertCursorStyle
	}
	return cursorStyle
}

// inVisualSelection indica si pos está dentro de la selección Visual
// actual, según el submodo (charwise inclusive, linewise, o bloque
// rectangular).
func inVisualSelection(mode engine.Mode, anchor, cursor, pos engine.Pos) bool {
	switch mode {
	case engine.ModeVisualLine:
		lo, hi := anchor.Line, cursor.Line
		if hi < lo {
			lo, hi = hi, lo
		}
		return pos.Line >= lo && pos.Line <= hi

	case engine.ModeVisualBlock:
		loLine, hiLine := anchor.Line, cursor.Line
		if hiLine < loLine {
			loLine, hiLine = hiLine, loLine
		}
		loCol, hiCol := anchor.Col, cursor.Col
		if hiCol < loCol {
			loCol, hiCol = hiCol, loCol
		}
		return pos.Line >= loLine && pos.Line <= hiLine && pos.Col >= loCol && pos.Col <= hiCol

	default: // ModeVisual: charwise, inclusive
		start, end := anchor, cursor
		if posGreater(start, end) {
			start, end = end, start
		}
		if pos.Line < start.Line || pos.Line > end.Line {
			return false
		}
		if pos.Line == start.Line && pos.Col < start.Col {
			return false
		}
		if pos.Line == end.Line && pos.Col > end.Col {
			return false
		}
		return true
	}
}

func posGreater(a, b engine.Pos) bool {
	if a.Line != b.Line {
		return a.Line > b.Line
	}
	return a.Col > b.Col
}

// renderTarget muestra el texto esperado línea a línea, coloreando en
// verde el prefijo que ya coincide con el buffer actual y en rojo lo que
// todavía falta.
func renderTarget(target, current string) string {
	targetLines := strings.Split(target, "\n")
	currentLines := strings.Split(current, "\n")

	var b strings.Builder
	for i, tLine := range targetLines {
		cLine := ""
		if i < len(currentLines) {
			cLine = currentLines[i]
		}
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  ")
		b.WriteString(renderDiffLine(tLine, cLine))
	}
	return b.String()
}

func renderDiffLine(target, current string) string {
	t := []rune(target)
	c := []rune(current)
	n := 0
	for n < len(t) && n < len(c) && t[n] == c[n] {
		n++
	}
	var b strings.Builder
	if n > 0 {
		b.WriteString(diffMatchStyle.Render(string(t[:n])))
	}
	if n < len(t) {
		b.WriteString(diffPendingStyle.Render(string(t[n:])))
	}
	return b.String()
}
