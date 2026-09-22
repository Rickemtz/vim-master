package engine

import (
	"strconv"
	"strings"
)

// Engine emula un subconjunto de Vim por defecto (nocompatible, sin
// vimrc):
//
//   - E1: buffer, cursor, modos Normal/Insert, hjkl, i a I A o O, Esc, x,
//     :w / :q simulados.
//   - E2: w b e W B E 0 ^ $ gg G {n}G y counts.
//   - E3: operadores d c y + motion, dd cc yy D C, p P, r s, u Ctrl-r.
//   - E4: f F t T ; , %, . (repetir el último cambio).
//   - E5: text objects iw aw iW aW i" a" i' a' i( a( i[ a[ i{ a{ ib ab
//     iB aB ip ap it at.
//   - E6: búsqueda / ? n N * # (subcadena literal, no regex de Vim).
//   - E7: Visual v V Ctrl-v con operadores.
//   - E8: > < ~ J, registros con nombre "a.
//   - E9: Ex :s :%s :g básico (patrón literal, no regex).
//   - E10: macros q{reg} @{reg} @@.
//
// :w / :q no tocan disco; la UI decide qué hacer con EventWrite/EventQuit.
type Engine struct {
	buf        *Buffer
	cursor     Pos
	desiredCol int
	mode       Mode
	cmdline    string
	cmdPrefix  rune // ':' '/' '?': qué tipo de línea de comando está activa
	stats      Stats

	// Estado del parser de comandos Normal: [count][operador][count][motion].
	count1   int  // dígitos acumulados antes de resolver un comando (0 = ninguno)
	operator rune // 'd' 'c' 'y' '>' '<', o 0 si no hay operador pendiente
	opCount  int  // count "bancado" al fijar el operador, antes del segundo count

	pending      pendingKind
	pendingFwd   bool // f/t (adelante) vs F/T (atrás)
	pendingTill  bool // t/T (para justo antes) vs f/F (sobre el carácter)
	awaitingG    bool // se vio una 'g', esperando la segunda tecla de gg
	textObjInner bool // i vs a, para el text object pendiente

	lastFind findState // para ; ,

	lastSearch          string
	lastSearchForward   bool
	lastSearchWholeWord bool // true tras * o # (coincidencia de palabra completa)

	visualAnchor Pos // punto fijo de la selección en modo Visual

	unnamed        Register
	namedRegisters map[rune]Register
	selectedReg    rune // registro elegido con "<letra>; 0 = usar el sin nombre

	undoStack []undoState
	redoStack []undoState

	macros       map[rune][]Key
	recordingReg rune // 0 = no se está grabando ninguna macro
	recordedKeys []Key
	lastMacroReg rune

	// cmdKeys acumula las teclas del comando Normal en curso; al
	// completarse, si cambió el buffer respecto a cmdKeysBeforeText (el
	// texto al INICIO del comando, no tecla a tecla), se guarda en
	// lastChange para '.'.
	cmdKeys           []Key
	cmdKeysBeforeText string
	lastChange        []Key
	replaying         bool
}

type pendingKind int

const (
	pendingNone pendingKind = iota
	pendingReplace
	pendingFindChar
	pendingTextObject
	pendingRegisterSelect
	pendingMacroRecord
	pendingMacroPlay
)

type findState struct {
	ch      rune
	forward bool
	till    bool
	valid   bool
}

func New(text string, cursor Pos) *Engine {
	buf := NewBuffer(text)
	cursor.Line = clampInt(cursor.Line, 0, buf.LineCount()-1)
	cursor.Col = ClampCol(buf.LineLen(cursor.Line), cursor.Col, ModeNormal)
	return &Engine{
		buf:        buf,
		cursor:     cursor,
		desiredCol: cursor.Col,
		mode:       ModeNormal,
		stats:      Stats{Commands: map[string]int{}},
	}
}

func (e *Engine) Text() string { return e.buf.Text() }
func (e *Engine) Cursor() Pos  { return e.cursor }
func (e *Engine) Mode() Mode   { return e.mode }

// VisualAnchor devuelve el punto fijo de la selección en modo Visual
// (solo tiene sentido cuando Mode() es Visual/VisualLine/VisualBlock).
func (e *Engine) VisualAnchor() Pos { return e.visualAnchor }

// PendingKeys devuelve una representación de lo que se lleva tecleado de
// un comando todavía incompleto (p.ej. "d2" mientras se escribe "d2w", o
// ":w" mientras se escribe una línea de comando), para mostrar en la
// barra de estado.
func (e *Engine) PendingKeys() string {
	if e.mode == ModeCommand {
		return string(e.cmdPrefix) + e.cmdline
	}

	var b strings.Builder
	if e.opCount > 0 {
		b.WriteString(strconv.Itoa(e.opCount))
	}
	if e.operator != 0 {
		b.WriteRune(e.operator)
	}
	if e.count1 > 0 {
		b.WriteString(strconv.Itoa(e.count1))
	}
	if e.awaitingG {
		b.WriteRune('g')
	}
	switch e.pending {
	case pendingFindChar:
		switch {
		case e.pendingFwd && !e.pendingTill:
			b.WriteRune('f')
		case !e.pendingFwd && !e.pendingTill:
			b.WriteRune('F')
		case e.pendingFwd && e.pendingTill:
			b.WriteRune('t')
		default:
			b.WriteRune('T')
		}
	case pendingTextObject:
		if e.textObjInner {
			b.WriteRune('i')
		} else {
			b.WriteRune('a')
		}
	case pendingReplace:
		b.WriteRune('r')
	case pendingRegisterSelect:
		b.WriteRune('"')
	case pendingMacroRecord:
		b.WriteRune('q')
	case pendingMacroPlay:
		b.WriteRune('@')
	}
	return b.String()
}

// Stats devuelve una copia; el mapa interno no se expone para que el
// llamador no pueda mutar el estado del engine.
func (e *Engine) Stats() Stats {
	cmds := make(map[string]int, len(e.stats.Commands))
	for k, v := range e.stats.Commands {
		cmds[k] = v
	}
	return Stats{TotalKeys: e.stats.TotalKeys, ArrowKeys: e.stats.ArrowKeys, Commands: cmds}
}

// Feed procesa una tecla: registra estadísticas, graba la secuencia de
// teclas del comando en curso (para '.'), graba macros si hay una en
// curso, y despacha según el modo.
func (e *Engine) Feed(key Key) Event {
	e.stats.TotalKeys++
	if key.IsArrow() {
		e.stats.ArrowKeys++
	}

	// '.' es un meta-comando: nunca debe convertirse a sí mismo en "el
	// último cambio" (eso rompería el repetir y podría recursar sin fin).
	recording := !e.replaying && e.mode != ModeCommand && key.Rune != '.'
	if recording {
		if len(e.cmdKeys) == 0 {
			e.cmdKeysBeforeText = e.buf.Text() // snapshot al INICIO del comando, no tecla a tecla
		}
		e.cmdKeys = append(e.cmdKeys, key)
	}

	// Grabación de macro: se registra la tecla ANTES de despacharla,
	// salvo la 'q' que la detiene (esa no debe quedar grabada dentro).
	if e.recordingReg != 0 && !e.replaying {
		stopKey := e.mode == ModeNormal && e.pending == pendingNone && key.Rune == 'q'
		if !stopKey {
			e.recordedKeys = append(e.recordedKeys, key)
		}
	}

	ev := e.dispatch(key)

	if !e.replaying {
		commandDone := e.mode == ModeNormal && e.operator == 0 &&
			e.pending == pendingNone && !e.awaitingG && e.count1 == 0
		if commandDone && len(e.cmdKeys) > 0 {
			if e.buf.Text() != e.cmdKeysBeforeText {
				e.lastChange = append([]Key(nil), e.cmdKeys...)
			}
			e.cmdKeys = nil
		}
	}

	return ev
}

// dispatch despacha una tecla ya contabilizada. Los estados "pendientes"
// de una sola tecla más (r<char>, f/F/t/T<char>, text objects, "<reg>,
// q<reg>, @<reg>) se resuelven aquí, antes de mirar el modo, porque
// aplican tanto en Normal como en Visual.
func (e *Engine) dispatch(key Key) Event {
	switch e.pending {
	case pendingReplace:
		return e.finishReplace(key)
	case pendingFindChar:
		return e.finishFindChar(key)
	case pendingTextObject:
		return e.finishTextObject(key)
	case pendingRegisterSelect:
		return e.finishRegisterSelect(key)
	case pendingMacroRecord:
		return e.finishMacroRecord(key)
	case pendingMacroPlay:
		return e.finishMacroPlay(key)
	}

	switch e.mode {
	case ModeNormal:
		return e.feedNormal(key)
	case ModeInsert:
		return e.feedInsert(key)
	case ModeCommand:
		return e.feedCommand(key)
	case ModeVisual, ModeVisualLine, ModeVisualBlock:
		return e.feedVisual(key)
	default:
		return Event{Type: EventUnsupported}
	}
}

func (e *Engine) recordCommand(name string) { e.stats.Commands[name]++ }

func (e *Engine) finishRegisterSelect(key Key) Event {
	e.pending = pendingNone
	if key.Special != KeyNone && !isRegisterName(key.Rune) {
		return Event{Type: EventUnsupported}
	}
	if !isRegisterName(key.Rune) {
		return Event{Type: EventUnsupported}
	}
	e.selectedReg = key.Rune
	e.recordCommand("\"" + string(key.Rune))
	return Event{Type: EventNone}
}

func isRegisterName(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func (e *Engine) feedInsert(key Key) Event {
	switch {
	case key.Special == KeyEsc:
		e.mode = ModeNormal
		e.cursor.Col = ClampCol(e.buf.LineLen(e.cursor.Line), e.cursor.Col-1, ModeNormal)
		e.desiredCol = e.cursor.Col
		return Event{Type: EventModeChanged}

	case key.Special == KeyEnter:
		e.buf.SplitLine(e.cursor)
		e.cursor = Pos{Line: e.cursor.Line + 1, Col: 0}
		e.desiredCol = 0
		return Event{Type: EventChanged}

	case key.Special == KeyBackspace:
		// Vim por defecto (sin 'backspace' en el vimrc) no permite borrar
		// hacia atrás del punto donde empezó la inserción ni unir líneas;
		// aquí simplificamos a: no borra nada en la columna 0.
		if e.cursor.Col > 0 {
			e.cursor.Col--
			e.buf.DeleteRune(e.cursor)
			return Event{Type: EventChanged}
		}
		return Event{Type: EventNone}

	case key.Special == KeyNone:
		e.buf.InsertRune(e.cursor, key.Rune)
		e.cursor.Col++
		return Event{Type: EventChanged}

	default:
		return Event{Type: EventUnsupported}
	}
}

func (e *Engine) feedCommand(key Key) Event {
	switch {
	case key.Special == KeyEsc:
		e.mode = ModeNormal
		e.cmdline = ""
		return Event{Type: EventModeChanged}

	case key.Special == KeyEnter:
		cmd := e.cmdline
		prefix := e.cmdPrefix
		e.cmdline = ""
		e.mode = ModeNormal

		if prefix == '/' || prefix == '?' {
			forward := prefix == '/'
			if cmd == "" {
				if e.lastSearch == "" {
					return Event{Type: EventNone}
				}
			} else {
				e.lastSearch = cmd
				e.lastSearchWholeWord = false
			}
			e.lastSearchForward = forward
			e.recordCommand(string(prefix))
			if !e.searchMove(e.lastSearch, forward, false, 1) {
				return Event{Type: EventUnsupported}
			}
			return Event{Type: EventMoved}
		}

		return e.execExCommand(cmd)

	case key.Special == KeyBackspace:
		if len(e.cmdline) > 0 {
			r := []rune(e.cmdline)
			e.cmdline = string(r[:len(r)-1])
		}
		return Event{Type: EventNone}

	case key.Special == KeyNone:
		e.cmdline += string(key.Rune)
		return Event{Type: EventNone}

	default:
		return Event{Type: EventUnsupported}
	}
}

func firstNonBlank(line string) int {
	for i, r := range []rune(line) {
		if r != ' ' && r != '\t' {
			return i
		}
	}
	return 0
}

func orOne(n int) int {
	if n <= 0 {
		return 1
	}
	return n
}
