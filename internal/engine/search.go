package engine

// searchMove busca pattern desde el cursor y mueve el cursor a la
// coincidencia. Simplificación documentada: busca subcadena literal, no
// expresiones regulares de Vim. Da la vuelta al final/inicio del buffer
// (wrapscan, el comportamiento por defecto de Vim). Devuelve false si no
// hay ninguna coincidencia.
func (e *Engine) searchMove(pattern string, forward, wholeWord bool, count int) bool {
	if pattern == "" {
		return false
	}
	found := false
	for i := 0; i < count; i++ {
		pos, ok := findNext(e.buf, e.cursor, pattern, forward, wholeWord)
		if !ok {
			return found
		}
		e.cursor = pos
		e.desiredCol = pos.Col
		found = true
	}
	return found
}

// findNext busca pattern en el buffer a partir de from, en la dirección
// dada, dando la vuelta si hace falta. wholeWord exige que la
// coincidencia no esté pegada a otro carácter de palabra (usado por * y
// #).
func findNext(b *Buffer, from Pos, pattern string, forward, wholeWord bool) (Pos, bool) {
	n := b.LineCount()
	pat := []rune(pattern)
	if n == 0 || len(pat) == 0 {
		return Pos{}, false
	}

	boundaryOK := func(text []rune, start int) bool {
		if !wholeWord {
			return true
		}
		if start > 0 && classify(text[start-1], false) == classWord {
			return false
		}
		end := start + len(pat)
		if end < len(text) && classify(text[end], false) == classWord {
			return false
		}
		return true
	}

	// step recorre n+1 líneas: la línea de partida se revisita al final
	// (step==n) para permitir dar la vuelta completa (wrapscan).
	for step := 0; step <= n; step++ {
		var line int
		if forward {
			line = (from.Line + step) % n
		} else {
			line = ((from.Line-step)%n + n) % n
		}
		text := []rune(b.Line(line))

		var positions []int
		for i := 0; i+len(pat) <= len(text); i++ {
			if string(text[i:i+len(pat)]) == pattern && boundaryOK(text, i) {
				positions = append(positions, i)
			}
		}

		restrictToAfterCursor := line == from.Line && step == 0

		if forward {
			for _, p := range positions {
				if restrictToAfterCursor && p <= from.Col {
					continue
				}
				return Pos{Line: line, Col: p}, true
			}
		} else {
			for i := len(positions) - 1; i >= 0; i-- {
				p := positions[i]
				if restrictToAfterCursor && p >= from.Col {
					continue
				}
				return Pos{Line: line, Col: p}, true
			}
		}
	}
	return Pos{}, false
}

// wordUnderCursor devuelve la palabra bajo (o después de, en la misma
// línea) el cursor, para * y #.
func wordUnderCursor(b *Buffer, pos Pos) (string, bool) {
	line := []rune(b.Line(pos.Line))
	if len(line) == 0 {
		return "", false
	}
	col := clampInt(pos.Col, 0, len(line)-1)
	if classify(line[col], false) != classWord {
		for col < len(line) && classify(line[col], false) != classWord {
			col++
		}
		if col >= len(line) {
			return "", false
		}
	}
	start := col
	for start > 0 && classify(line[start-1], false) == classWord {
		start--
	}
	end := col
	for end+1 < len(line) && classify(line[end+1], false) == classWord {
		end++
	}
	return string(line[start : end+1]), true
}
