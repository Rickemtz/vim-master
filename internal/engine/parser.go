package engine

// feedNormal despacha una tecla en modo Normal. Los estados pendientes de
// una tecla más (r<char>, f/F/t/T<char>, text objects, "<reg>, q<reg>,
// @<reg>) ya se resolvieron en dispatch() antes de llegar aquí.
// [count1][operador][count2][motion|comando].
func (e *Engine) feedNormal(key Key) Event {
	if e.awaitingG && key.Rune != 'g' {
		e.awaitingG = false // 'g' seguido de algo que no es 'g': no soportado aún, cancelar
	}

	switch {
	case key.Special == KeyNone && key.Rune >= '1' && key.Rune <= '9':
		e.count1 = e.count1*10 + int(key.Rune-'0')
		return Event{Type: EventNone}
	case key.Special == KeyNone && key.Rune == '0' && e.count1 > 0:
		e.count1 = e.count1 * 10
		return Event{Type: EventNone}

	case key.Special == KeyCtrlR:
		e.recordCommand("ctrl+r")
		e.redo()
		e.count1 = 0
		return Event{Type: EventChanged}

	case key.Special == KeyCtrlV:
		e.recordCommand("ctrl+v")
		e.visualAnchor = e.cursor
		e.mode = ModeVisualBlock
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Special == KeyLeft || key.Rune == 'h':
		e.recordCommand("h")
		e.executeMotion(motionCharLeft(e.cursor, e.totalCount()))
		return Event{Type: EventMoved}

	case key.Special == KeyRight || key.Rune == 'l':
		e.recordCommand("l")
		e.executeMotion(motionCharRight(e.buf, e.cursor, e.totalCount()))
		return Event{Type: EventMoved}

	case key.Special == KeyDown || key.Rune == 'j':
		e.recordCommand("j")
		e.executeVerticalMotion(1, e.totalCount())
		return Event{Type: EventMoved}

	case key.Special == KeyUp || key.Rune == 'k':
		e.recordCommand("k")
		e.executeVerticalMotion(-1, e.totalCount())
		return Event{Type: EventMoved}

	case key.Rune == '0':
		e.recordCommand("0")
		e.executeMotion(motionLineStart(e.cursor))
		return Event{Type: EventMoved}

	case key.Rune == '^':
		e.recordCommand("^")
		e.executeMotion(motionFirstNonBlank(e.buf, e.cursor))
		return Event{Type: EventMoved}

	case key.Rune == '$':
		e.recordCommand("$")
		e.executeMotion(motionLineEnd(e.buf, e.cursor, e.totalCount()))
		return Event{Type: EventMoved}

	case key.Rune == 'w' || key.Rune == 'W':
		big := key.Rune == 'W'
		e.recordCommand(string(key.Rune))
		count := e.totalCount()
		if e.operator == 'c' && classAt(e.buf, e.cursor, big) != classSpace {
			// cw/cW se comportan como ce/cE: Vim no incluye el espacio
			// posterior a la palabra al cambiarla (:help cw).
			e.executeMotion(motionWordEnd(e.buf, e.cursor, count, big))
		} else {
			e.executeMotion(motionWordForward(e.buf, e.cursor, count, big))
		}
		return Event{Type: EventMoved}

	case key.Rune == 'b' || key.Rune == 'B':
		big := key.Rune == 'B'
		e.recordCommand(string(key.Rune))
		e.executeMotion(motionWordBackward(e.buf, e.cursor, e.totalCount(), big))
		return Event{Type: EventMoved}

	case key.Rune == 'e' || key.Rune == 'E':
		big := key.Rune == 'E'
		e.recordCommand(string(key.Rune))
		e.executeMotion(motionWordEnd(e.buf, e.cursor, e.totalCount(), big))
		return Event{Type: EventMoved}

	case key.Rune == 'G':
		e.recordCommand("G")
		e.executeMotion(motionGotoLine(e.buf, e.totalCountRaw(), true))
		return Event{Type: EventMoved}

	case key.Rune == 'g':
		if e.awaitingG {
			e.awaitingG = false
			e.recordCommand("gg")
			e.executeMotion(motionGotoLine(e.buf, e.totalCountRaw(), false))
			return Event{Type: EventMoved}
		}
		e.awaitingG = true
		return Event{Type: EventNone}

	case key.Rune == 'f', key.Rune == 'F', key.Rune == 't', key.Rune == 'T':
		e.pending = pendingFindChar
		e.pendingFwd = key.Rune == 'f' || key.Rune == 't'
		e.pendingTill = key.Rune == 't' || key.Rune == 'T'
		return Event{Type: EventNone}

	case key.Rune == ';' || key.Rune == ',':
		e.recordCommand(string(key.Rune))
		if !e.lastFind.valid {
			e.count1 = 0
			return Event{Type: EventNone}
		}
		fwd := e.lastFind.forward
		if key.Rune == ',' {
			fwd = !fwd
		}
		res := motionFindChar(e.buf, e.cursor, e.lastFind.ch, e.totalCount(), fwd, e.lastFind.till)
		e.executeMotion(res)
		return Event{Type: EventMoved}

	case key.Rune == '%':
		e.recordCommand("%")
		e.executeMotion(motionMatchPair(e.buf, e.cursor))
		return Event{Type: EventMoved}

	case key.Rune == '/' || key.Rune == '?':
		e.recordCommand(string(key.Rune))
		e.mode = ModeCommand
		e.cmdPrefix = key.Rune
		e.cmdline = ""
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Rune == 'n':
		e.recordCommand("n")
		return e.repeatSearch(e.lastSearchForward)

	case key.Rune == 'N':
		e.recordCommand("N")
		return e.repeatSearch(!e.lastSearchForward)

	case key.Rune == '*' || key.Rune == '#':
		e.recordCommand(string(key.Rune))
		e.count1 = 0
		if !e.searchWordUnderCursor(key.Rune == '*') {
			return Event{Type: EventUnsupported}
		}
		return Event{Type: EventMoved}

	// Text objects: i/a solo se leen como prefijo de text object cuando
	// hay un operador pendiente; sin operador, i/a son "entrar a Insert".
	case (key.Rune == 'i' || key.Rune == 'a') && e.operator != 0:
		e.textObjInner = key.Rune == 'i'
		e.pending = pendingTextObject
		return Event{Type: EventNone}

	case key.Rune == 'd', key.Rune == 'c', key.Rune == 'y', key.Rune == '>', key.Rune == '<':
		return e.feedOperatorKey(key.Rune)

	case key.Rune == 'D':
		e.recordCommand("D")
		e.operator = 'd'
		e.executeMotion(motionLineEnd(e.buf, e.cursor, 1))
		return Event{Type: EventChanged}

	case key.Rune == 'C':
		e.recordCommand("C")
		e.operator = 'c'
		e.executeMotion(motionLineEnd(e.buf, e.cursor, 1))
		return Event{Type: EventChanged}

	case key.Rune == '~':
		e.recordCommand("~")
		return e.doToggleCase(orOne(e.count1))

	case key.Rune == 'J':
		e.recordCommand("J")
		count := orOne(e.count1)
		if count < 2 {
			count = 2
		}
		e.count1 = 0
		return e.doJoin(count)

	case key.Rune == 'r':
		e.recordCommand("r")
		e.pending = pendingReplace
		return Event{Type: EventNone}

	case key.Rune == 's':
		e.recordCommand("s")
		count := e.totalCount()
		e.count1 = 0
		e.pushUndo()
		for i := 0; i < count && e.buf.LineLen(e.cursor.Line) > e.cursor.Col; i++ {
			e.buf.DeleteRune(e.cursor)
		}
		e.mode = ModeInsert
		return Event{Type: EventModeChanged}

	case key.Rune == '"':
		e.pending = pendingRegisterSelect
		return Event{Type: EventNone}

	case key.Rune == 'p':
		e.recordCommand("p")
		e.doPaste(true, e.totalCount())
		e.count1 = 0
		return Event{Type: EventChanged}

	case key.Rune == 'P':
		e.recordCommand("P")
		e.doPaste(false, e.totalCount())
		e.count1 = 0
		return Event{Type: EventChanged}

	case key.Rune == 'u':
		e.recordCommand("u")
		e.undo()
		e.count1 = 0
		return Event{Type: EventChanged}

	case key.Rune == '.':
		e.count1 = 0
		if len(e.lastChange) == 0 {
			return Event{Type: EventNone}
		}
		e.replaying = true
		keys := e.lastChange
		var last Event
		for _, k := range keys {
			last = e.dispatch(k)
		}
		e.replaying = false
		return last

	case key.Rune == 'q':
		if e.recordingReg != 0 {
			reg := e.recordingReg
			if e.macros == nil {
				e.macros = map[rune][]Key{}
			}
			e.macros[reg] = e.recordedKeys
			e.recordingReg = 0
			e.recordedKeys = nil
			e.recordCommand("q")
			return Event{Type: EventNone}
		}
		e.pending = pendingMacroRecord
		return Event{Type: EventNone}

	case key.Rune == '@':
		e.pending = pendingMacroPlay
		return Event{Type: EventNone}

	case key.Rune == 'v':
		e.recordCommand("v")
		count := e.totalCount()
		e.visualAnchor = e.cursor
		e.mode = ModeVisual
		if count > 1 {
			res := motionCharRight(e.buf, e.cursor, count-1)
			e.cursor = Pos{Line: res.Pos.Line, Col: ClampCol(e.buf.LineLen(res.Pos.Line), res.Pos.Col, ModeNormal)}
		}
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Rune == 'V':
		e.recordCommand("V")
		count := e.totalCount()
		e.visualAnchor = e.cursor
		e.mode = ModeVisualLine
		if count > 1 {
			e.cursor.Line = clampInt(e.cursor.Line+count-1, 0, e.buf.LineCount()-1)
		}
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Rune == 'i':
		e.recordCommand("i")
		e.pushUndo()
		e.mode = ModeInsert
		return Event{Type: EventModeChanged}

	case key.Rune == 'a':
		e.recordCommand("a")
		e.pushUndo()
		if e.buf.LineLen(e.cursor.Line) > 0 {
			e.cursor.Col++
		}
		e.mode = ModeInsert
		return Event{Type: EventModeChanged}

	case key.Rune == 'I':
		e.recordCommand("I")
		e.pushUndo()
		e.cursor.Col = firstNonBlank(e.buf.Line(e.cursor.Line))
		e.mode = ModeInsert
		return Event{Type: EventModeChanged}

	case key.Rune == 'A':
		e.recordCommand("A")
		e.pushUndo()
		e.cursor.Col = e.buf.LineLen(e.cursor.Line)
		e.mode = ModeInsert
		return Event{Type: EventModeChanged}

	case key.Rune == 'o':
		e.recordCommand("o")
		e.pushUndo()
		idx := e.buf.OpenLineBelow(e.cursor.Line)
		e.cursor = Pos{Line: idx, Col: 0}
		e.desiredCol = 0
		e.mode = ModeInsert
		return Event{Type: EventChanged}

	case key.Rune == 'O':
		e.recordCommand("O")
		e.pushUndo()
		idx := e.buf.OpenLineAbove(e.cursor.Line)
		e.cursor = Pos{Line: idx, Col: 0}
		e.desiredCol = 0
		e.mode = ModeInsert
		return Event{Type: EventChanged}

	case key.Rune == 'x':
		e.recordCommand("x")
		return e.doDeleteChars(e.totalCount())

	case key.Rune == ':':
		e.recordCommand(":")
		e.mode = ModeCommand
		e.cmdPrefix = ':'
		e.cmdline = ""
		e.count1 = 0
		return Event{Type: EventModeChanged}

	default:
		e.count1 = 0
		e.operator = 0
		e.opCount = 0
		return Event{Type: EventUnsupported}
	}
}

// totalCount combina el count antes y después de un operador
// ([count1]operador[count2]motion se multiplican). Sin count escrito,
// cada parte vale 1.
func (e *Engine) totalCount() int {
	if e.operator != 0 {
		return orOne(e.opCount) * orOne(e.count1)
	}
	return orOne(e.count1)
}

// totalCountRaw es igual que totalCount pero devuelve 0 si no se escribió
// ningún dígito, para distinguir "sin count" (gg/G van al extremo) de
// "count 1" ({1}G va a la línea 1).
func (e *Engine) totalCountRaw() int {
	if e.opCount == 0 && e.count1 == 0 {
		return 0
	}
	return orOne(e.opCount) * orOne(e.count1)
}

func (e *Engine) feedOperatorKey(op rune) Event {
	if e.operator == op {
		// Doble letra (dd/cc/yy/>>/<<): linewise sobre `count` líneas.
		e.recordCommand(string(op) + string(op))
		count := e.totalCount()
		target := clampInt(e.cursor.Line+count-1, 0, e.buf.LineCount()-1)
		e.applyOperatorLines(op, e.cursor.Line, target)
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return Event{Type: EventChanged}
	}
	if e.operator != 0 {
		// Un operador ya pendiente seguido de otra letra de operador
		// distinta tampoco es válido en Vim real: se ignora.
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return Event{Type: EventUnsupported}
	}
	e.recordCommand(string(op))
	e.operator = op
	e.opCount = e.count1
	e.count1 = 0
	return Event{Type: EventNone}
}

// executeMotion aplica el motion al operador pendiente (si lo hay) o
// mueve el cursor con él, y limpia el estado de count/operador.
func (e *Engine) executeMotion(res MotionResult) {
	if e.operator != 0 {
		op := e.operator
		from := e.cursor
		e.applyOperatorMotion(op, from, res)
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return
	}
	e.count1 = 0
	if !res.Found {
		return
	}
	e.cursor = Pos{Line: res.Pos.Line, Col: ClampCol(e.buf.LineLen(res.Pos.Line), res.Pos.Col, ModeNormal)}
	e.desiredCol = e.cursor.Col
}

// executeVerticalMotion es como executeMotion pero para j/k: preserva
// desiredCol (columna deseada) tal como en E1, y trata el motion como
// linewise cuando hay un operador pendiente.
func (e *Engine) executeVerticalMotion(delta, count int) {
	target := clampInt(e.cursor.Line+delta*count, 0, e.buf.LineCount()-1)
	if e.operator != 0 {
		op := e.operator
		from := e.cursor
		e.applyOperatorLines(op, from.Line, target)
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return
	}
	e.cursor = Pos{Line: target, Col: ClampCol(e.buf.LineLen(target), e.desiredCol, ModeNormal)}
	e.count1 = 0
}

func (e *Engine) finishReplace(key Key) Event {
	e.pending = pendingNone
	count := orOne(e.count1)
	e.count1 = 0
	if key.Special != KeyNone {
		return Event{Type: EventUnsupported}
	}
	line := []rune(e.buf.Line(e.cursor.Line))
	if e.cursor.Col+count > len(line) {
		return Event{Type: EventUnsupported}
	}
	e.pushUndo()
	for i := 0; i < count; i++ {
		line[e.cursor.Col+i] = key.Rune
	}
	e.buf.SetLine(e.cursor.Line, string(line))
	e.cursor.Col += count - 1
	return Event{Type: EventChanged}
}

func (e *Engine) finishFindChar(key Key) Event {
	e.pending = pendingNone
	if key.Special != KeyNone {
		e.count1 = 0
		e.operator = 0
		e.opCount = 0
		return Event{Type: EventUnsupported}
	}
	count := e.totalCount()
	res := motionFindChar(e.buf, e.cursor, key.Rune, count, e.pendingFwd, e.pendingTill)
	if res.Found {
		e.lastFind = findState{ch: key.Rune, forward: e.pendingFwd, till: e.pendingTill, valid: true}
	}
	e.executeMotion(res)
	return Event{Type: EventMoved}
}

// finishTextObject resuelve la segunda tecla de un text object (la que
// nombra el objeto: w, ", (, p...). En Visual EXTIENDE la selección; con
// un operador pendiente en Normal, lo APLICA de inmediato.
func (e *Engine) finishTextObject(key Key) Event {
	e.pending = pendingNone
	if key.Special != KeyNone {
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return Event{Type: EventUnsupported}
	}
	e.recordCommand(map[bool]string{true: "i", false: "a"}[e.textObjInner] + string(key.Rune))

	start, end, linewise, ok := textObjectRange(e.buf, e.cursor, key.Rune, e.textObjInner)
	if !ok {
		e.operator = 0
		e.opCount = 0
		e.count1 = 0
		return Event{Type: EventNone}
	}

	if e.mode == ModeVisual || e.mode == ModeVisualLine || e.mode == ModeVisualBlock {
		e.visualAnchor = start
		if linewise {
			e.cursor = Pos{Line: end.Line, Col: 0}
			e.mode = ModeVisualLine
		} else if last, ok := e.buf.Prev(end); ok && !posLess(last, start) {
			e.cursor = last
		} else {
			e.cursor = start
		}
		e.count1 = 0
		return Event{Type: EventMoved}
	}

	op := e.operator
	e.operator = 0
	e.opCount = 0
	e.count1 = 0
	if linewise {
		e.applyOperatorLines(op, start.Line, end.Line)
	} else {
		e.applyOperatorRange(op, start, end)
	}
	return Event{Type: EventChanged}
}

// repeatSearch repite la última búsqueda (n/N).
func (e *Engine) repeatSearch(forward bool) Event {
	count := e.totalCount()
	e.count1 = 0
	if e.lastSearch == "" {
		return Event{Type: EventNone}
	}
	if !e.searchMove(e.lastSearch, forward, e.lastSearchWholeWord, count) {
		return Event{Type: EventUnsupported}
	}
	return Event{Type: EventMoved}
}

func (e *Engine) searchWordUnderCursor(forward bool) bool {
	word, ok := wordUnderCursor(e.buf, e.cursor)
	if !ok {
		return false
	}
	e.lastSearch = word
	e.lastSearchForward = forward
	e.lastSearchWholeWord = true
	return e.searchMove(word, forward, true, 1)
}

func (e *Engine) finishMacroRecord(key Key) Event {
	e.pending = pendingNone
	if key.Special != KeyNone || !isRegisterName(key.Rune) {
		return Event{Type: EventUnsupported}
	}
	e.recordingReg = key.Rune
	e.recordedKeys = nil
	e.recordCommand("q" + string(key.Rune))
	return Event{Type: EventNone}
}

func (e *Engine) finishMacroPlay(key Key) Event {
	e.pending = pendingNone
	count := orOne(e.count1)
	e.count1 = 0

	var reg rune
	switch {
	case key.Special == KeyNone && key.Rune == '@':
		reg = e.lastMacroReg
	case key.Special == KeyNone && isRegisterName(key.Rune):
		reg = key.Rune
	default:
		return Event{Type: EventUnsupported}
	}
	if reg == 0 {
		return Event{Type: EventNone}
	}
	keys, ok := e.macros[reg]
	if !ok || len(keys) == 0 {
		return Event{Type: EventNone}
	}
	e.lastMacroReg = reg
	e.recordCommand("@" + string(reg))

	wasReplaying := e.replaying
	e.replaying = true
	var last Event
	for i := 0; i < count; i++ {
		for _, k := range keys {
			last = e.dispatch(k)
		}
	}
	e.replaying = wasReplaying
	return last
}

// doDeleteChars implementa x (con count): borra hasta count caracteres
// desde el cursor.
func (e *Engine) doDeleteChars(count int) Event {
	e.count1 = 0
	e.pushUndo()
	var deleted []rune
	for i := 0; i < count; i++ {
		r, ok := e.buf.DeleteRune(e.cursor)
		if !ok {
			break
		}
		deleted = append(deleted, r)
	}
	if len(deleted) == 0 {
		e.undoStack = e.undoStack[:len(e.undoStack)-1]
		return Event{Type: EventNone}
	}
	e.writeRegister(string(deleted), false)
	e.cursor.Col = ClampCol(e.buf.LineLen(e.cursor.Line), e.cursor.Col, ModeNormal)
	e.desiredCol = e.cursor.Col
	return Event{Type: EventChanged}
}
