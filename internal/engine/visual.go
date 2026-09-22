package engine

import "strings"

// feedVisual despacha una tecla en Visual/Visual Line/Visual Block. Los
// motions mueven el cursor extendiendo la selección (desde
// e.visualAnchor); d/c/y/x aplican el operador a la selección actual.
func (e *Engine) feedVisual(key Key) Event {
	if e.awaitingG && key.Rune != 'g' {
		e.awaitingG = false
	}

	switch {
	case key.Special == KeyEsc:
		e.mode = ModeNormal
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Special == KeyNone && key.Rune >= '1' && key.Rune <= '9':
		e.count1 = e.count1*10 + int(key.Rune-'0')
		return Event{Type: EventNone}
	case key.Special == KeyNone && key.Rune == '0' && e.count1 > 0:
		e.count1 = e.count1 * 10
		return Event{Type: EventNone}

	case key.Rune == 'v':
		if e.mode == ModeVisual {
			e.mode = ModeNormal
		} else {
			e.mode = ModeVisual
		}
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Rune == 'V':
		if e.mode == ModeVisualLine {
			e.mode = ModeNormal
		} else {
			e.mode = ModeVisualLine
		}
		e.count1 = 0
		return Event{Type: EventModeChanged}

	case key.Special == KeyCtrlV:
		if e.mode == ModeVisualBlock {
			e.mode = ModeNormal
		} else {
			e.mode = ModeVisualBlock
		}
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
		e.recordCommand(string(key.Rune))
		e.executeMotion(motionWordForward(e.buf, e.cursor, e.totalCount(), key.Rune == 'W'))
		return Event{Type: EventMoved}

	case key.Rune == 'b' || key.Rune == 'B':
		e.recordCommand(string(key.Rune))
		e.executeMotion(motionWordBackward(e.buf, e.cursor, e.totalCount(), key.Rune == 'B'))
		return Event{Type: EventMoved}

	case key.Rune == 'e' || key.Rune == 'E':
		e.recordCommand(string(key.Rune))
		e.executeMotion(motionWordEnd(e.buf, e.cursor, e.totalCount(), key.Rune == 'E'))
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
		e.executeMotion(motionFindChar(e.buf, e.cursor, e.lastFind.ch, e.totalCount(), fwd, e.lastFind.till))
		return Event{Type: EventMoved}

	case key.Rune == '%':
		e.recordCommand("%")
		e.executeMotion(motionMatchPair(e.buf, e.cursor))
		return Event{Type: EventMoved}

	case key.Rune == 'i' || key.Rune == 'a':
		e.textObjInner = key.Rune == 'i'
		e.pending = pendingTextObject
		return Event{Type: EventNone}

	case key.Rune == 'd' || key.Rune == 'x':
		e.applyVisualOperator('d')
		return Event{Type: EventChanged}

	case key.Rune == 'c':
		e.applyVisualOperator('c')
		return Event{Type: EventChanged}

	case key.Rune == 'y':
		e.applyVisualOperator('y')
		return Event{Type: EventChanged}

	case key.Rune == '>':
		e.applyVisualOperator('>')
		return Event{Type: EventChanged}

	case key.Rune == '<':
		e.applyVisualOperator('<')
		return Event{Type: EventChanged}

	default:
		e.count1 = 0
		return Event{Type: EventUnsupported}
	}
}

// applyVisualOperator aplica op a la selección actual [visualAnchor,
// cursor] y vuelve a Normal (salvo 'c', que deja en Insert).
func (e *Engine) applyVisualOperator(op rune) {
	anchor := e.visualAnchor
	cur := e.cursor
	mode := e.mode

	switch mode {
	case ModeVisualLine:
		e.applyOperatorLines(op, anchor.Line, cur.Line)
	case ModeVisualBlock:
		e.applyOperatorBlock(op, anchor, cur)
	default: // ModeVisual: charwise, inclusive
		start, end := anchor, cur
		if posLess(end, start) {
			start, end = end, start
		}
		if next, ok := e.buf.Next(end); ok {
			end = next
		} else {
			end = e.buf.LastPos()
		}
		e.applyOperatorRange(op, start, end)
	}

	if e.mode != ModeInsert {
		e.mode = ModeNormal
	}
	e.count1 = 0
}

// applyOperatorBlock aplica op a un bloque rectangular [minLine,maxLine]
// x [minCol,maxCol]. Simplificación documentada: Vim real permite
// seleccionar más allá del final de líneas cortas (virtualedit); aquí se
// recorta cada línea a su longitud.
func (e *Engine) applyOperatorBlock(op rune, anchor, cur Pos) {
	minLine, maxLine := anchor.Line, cur.Line
	if maxLine < minLine {
		minLine, maxLine = maxLine, minLine
	}
	minCol, maxCol := anchor.Col, cur.Col
	if maxCol < minCol {
		minCol, maxCol = maxCol, minCol
	}

	switch op {
	case 'y':
		parts := make([]string, 0, maxLine-minLine+1)
		for l := minLine; l <= maxLine; l++ {
			line := []rune(e.buf.Line(l))
			s := clampInt(minCol, 0, len(line))
			en := clampInt(maxCol+1, 0, len(line))
			if en < s {
				en = s
			}
			parts = append(parts, string(line[s:en]))
		}
		e.writeRegister(strings.Join(parts, "\n"), false)
		e.cursor = Pos{Line: minLine, Col: minCol}

	case 'd', 'c':
		e.pushUndo()
		parts := make([]string, 0, maxLine-minLine+1)
		for l := minLine; l <= maxLine; l++ {
			line := []rune(e.buf.Line(l))
			s := clampInt(minCol, 0, len(line))
			en := clampInt(maxCol+1, 0, len(line))
			if en < s {
				en = s
			}
			parts = append(parts, string(line[s:en]))
			e.buf.SetLine(l, string(line[:s])+string(line[en:]))
		}
		e.writeRegister(strings.Join(parts, "\n"), false)
		e.cursor = Pos{Line: minLine, Col: ClampCol(e.buf.LineLen(minLine), minCol, ModeNormal)}
		if op == 'c' {
			e.mode = ModeInsert
		}
	}
}
