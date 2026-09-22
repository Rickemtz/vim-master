package engine

import (
	"strings"
	"unicode"
)

func posLess(a, b Pos) bool {
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Col < b.Col
}

// writeRegister guarda text en el registro seleccionado con "<letra> (si
// lo hay) o en el registro sin nombre, y limpia la selección.
func (e *Engine) writeRegister(text string, linewise bool) {
	reg := Register{Text: text, Linewise: linewise}
	if e.selectedReg != 0 {
		if e.namedRegisters == nil {
			e.namedRegisters = map[rune]Register{}
		}
		e.namedRegisters[e.selectedReg] = reg
		e.selectedReg = 0
		return
	}
	e.unnamed = reg
}

// readRegister lee el registro seleccionado con "<letra> (si lo hay) o el
// sin nombre, y limpia la selección.
func (e *Engine) readRegister() Register {
	if e.selectedReg != 0 {
		r := e.namedRegisters[e.selectedReg]
		e.selectedReg = 0
		return r
	}
	return e.unnamed
}

// applyOperatorMotion aplica op ('d', 'c', 'y', '>' o '<') entre from y
// el destino de un motion, según su Kind. > y < siempre son linewise en
// Vim real, sin importar el motion.
func (e *Engine) applyOperatorMotion(op rune, from Pos, to MotionResult) {
	if !to.Found {
		return
	}
	if to.Kind == MotionLinewise || op == '>' || op == '<' {
		e.applyOperatorLines(op, from.Line, to.Pos.Line)
		return
	}

	start, end := from, to.Pos
	if posLess(end, start) {
		start, end = end, start
	}
	if to.Kind == MotionInclusive {
		if next, ok := e.buf.Next(end); ok {
			end = next
		} else {
			end = e.buf.LastPos()
		}
	}
	e.applyOperatorRange(op, start, end)
}

// applyOperatorRange aplica op sobre el rango charwise [start, end).
func (e *Engine) applyOperatorRange(op rune, start, end Pos) {
	switch op {
	case 'y':
		e.writeRegister(e.buf.TextRange(start, end), false)
		e.cursor = start

	case 'd', 'c':
		e.pushUndo()
		text := e.buf.DeleteRange(start, end)
		e.writeRegister(text, false)
		e.cursor = Pos{Line: start.Line, Col: ClampCol(e.buf.LineLen(start.Line), start.Col, ModeNormal)}
		if op == 'c' {
			e.cursor.Col = start.Col // en Insert se permite la columna justo tras el borrado
			e.mode = ModeInsert
		}
	}
}

// applyOperatorLines aplica op linewise sobre las líneas [lineA, lineB]
// (en cualquier orden).
func (e *Engine) applyOperatorLines(op rune, lineA, lineB int) {
	origCol := e.cursor.Col
	if lineB < lineA {
		lineA, lineB = lineB, lineA
	}

	switch op {
	case 'y':
		lines := make([]string, 0, lineB-lineA+1)
		for l := lineA; l <= lineB; l++ {
			lines = append(lines, e.buf.Line(l))
		}
		e.writeRegister(strings.Join(lines, "\n"), true)
		e.cursor = Pos{Line: lineA, Col: ClampCol(e.buf.LineLen(lineA), origCol, ModeNormal)}

	case 'd':
		e.pushUndo()
		text := e.buf.DeleteLines(lineA, lineB)
		e.writeRegister(text, true)
		line := clampInt(lineA, 0, e.buf.LineCount()-1)
		e.cursor = Pos{Line: line, Col: firstNonBlank(e.buf.Line(line))}

	case 'c':
		e.pushUndo()
		text := e.buf.DeleteLines(lineA, lineB)
		e.writeRegister(text, true)
		idx := e.buf.OpenLineAbove(clampInt(lineA, 0, e.buf.LineCount()))
		e.cursor = Pos{Line: idx, Col: 0}
		e.mode = ModeInsert

	case '>':
		e.pushUndo()
		for l := lineA; l <= lineB; l++ {
			if e.buf.LineLen(l) > 0 {
				e.buf.SetLine(l, "    "+e.buf.Line(l))
			}
		}
		e.cursor = Pos{Line: lineA, Col: firstNonBlank(e.buf.Line(lineA))}

	case '<':
		e.pushUndo()
		for l := lineA; l <= lineB; l++ {
			line := []rune(e.buf.Line(l))
			trim := 0
			for trim < 4 && trim < len(line) && line[trim] == ' ' {
				trim++
			}
			e.buf.SetLine(l, string(line[trim:]))
		}
		e.cursor = Pos{Line: lineA, Col: firstNonBlank(e.buf.Line(lineA))}
	}
}

// pastePutNoUndo pega reg. after=true es p, false es P. No empuja undo
// (lo hace el llamador una sola vez, incluso si se pega varias veces por
// un count) ni resuelve el registro (lo hace el llamador una sola vez,
// no una por repetición, como en Vim real).
func (e *Engine) pastePutNoUndo(reg Register, after bool) {
	if reg.Text == "" {
		return
	}

	if reg.Linewise {
		lines := strings.Split(reg.Text, "\n")
		idx := e.cursor.Line
		if after {
			e.buf.InsertLinesBelow(idx, lines)
			idx++
		} else {
			e.buf.InsertLinesBelow(idx-1, lines)
		}
		e.cursor = Pos{Line: idx, Col: firstNonBlank(e.buf.Line(idx))}
		return
	}

	pos := e.cursor
	if after && e.buf.LineLen(pos.Line) > 0 {
		pos.Col++
	}
	e.buf.InsertText(pos, reg.Text)

	// Simplificación: el cursor queda al final de lo pegado solo cuando el
	// registro es de una sola línea (caso común: x, r, d/y de un motion
	// charwise en la misma línea). Para un registro charwise multilínea
	// (p.ej. de un bloque Visual) se deja al inicio.
	if !strings.Contains(reg.Text, "\n") {
		newCol := pos.Col + len([]rune(reg.Text)) - 1
		if newCol < pos.Col {
			newCol = pos.Col
		}
		e.cursor = Pos{Line: pos.Line, Col: ClampCol(e.buf.LineLen(pos.Line), newCol, ModeNormal)}
	} else {
		e.cursor = pos
	}
}

// doPaste pega count veces, como una sola unidad de undo.
func (e *Engine) doPaste(after bool, count int) {
	reg := e.readRegister()
	if reg.Text == "" {
		return
	}
	e.pushUndo()
	for i := 0; i < count; i++ {
		e.pastePutNoUndo(reg, after)
	}
}

func toggleCase(r rune) rune {
	switch {
	case unicode.IsUpper(r):
		return unicode.ToLower(r)
	case unicode.IsLower(r):
		return unicode.ToUpper(r)
	default:
		return r
	}
}

// doToggleCase implementa ~ (con count): alterna mayúsculas/minúsculas de
// los siguientes count caracteres y avanza el cursor.
func (e *Engine) doToggleCase(count int) Event {
	e.count1 = 0
	line := []rune(e.buf.Line(e.cursor.Line))
	end := clampInt(e.cursor.Col+count, 0, len(line))
	if end <= e.cursor.Col {
		return Event{Type: EventNone}
	}
	e.pushUndo()
	for i := e.cursor.Col; i < end; i++ {
		line[i] = toggleCase(line[i])
	}
	e.buf.SetLine(e.cursor.Line, string(line))
	e.cursor.Col = ClampCol(e.buf.LineLen(e.cursor.Line), end, ModeNormal)
	e.desiredCol = e.cursor.Col
	return Event{Type: EventChanged}
}

// doJoin implementa J (con count): une count líneas (la actual + las
// siguientes), reemplazando el salto de línea por un espacio (ninguno si
// alguno de los dos lados queda vacío) y recortando la indentación de la
// línea que se une.
func (e *Engine) doJoin(count int) Event {
	if e.cursor.Line >= e.buf.LineCount()-1 {
		return Event{Type: EventNone}
	}
	e.pushUndo()
	joins := count - 1
	joinCol := e.buf.LineLen(e.cursor.Line)
	for i := 0; i < joins && e.cursor.Line < e.buf.LineCount()-1; i++ {
		cur := e.buf.Line(e.cursor.Line)
		next := strings.TrimLeft(e.buf.Line(e.cursor.Line+1), " \t")
		sep := " "
		if cur == "" || next == "" {
			sep = ""
		}
		joinCol = len([]rune(cur))
		e.buf.SetLine(e.cursor.Line, cur+sep+next)
		e.buf.DeleteLines(e.cursor.Line+1, e.cursor.Line+1)
	}
	e.cursor.Col = ClampCol(e.buf.LineLen(e.cursor.Line), joinCol, ModeNormal)
	e.desiredCol = e.cursor.Col
	return Event{Type: EventChanged}
}
