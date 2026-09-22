package engine

import (
	"regexp"
	"strings"
)

// textObjectRange calcula el rango [start, end) de un text object
// (iw aw iW aW i" a" i' a' i( a( i[ a[ i{ a{ ib ab iB aB ip ap it at).
// linewise indica si debe tratarse como rango de líneas completas (solo
// ip/ap). ok es false si no se pudo determinar (p.ej. sin comillas en la
// línea, o sin brackets/tag que lo contengan).
func textObjectRange(b *Buffer, pos Pos, obj rune, inner bool) (start, end Pos, linewise, ok bool) {
	switch obj {
	case 'w':
		return wordObjectRange(b, pos, false, inner)
	case 'W':
		return wordObjectRange(b, pos, true, inner)
	case '"':
		return quoteObjectRange(b, pos, '"', inner)
	case '\'':
		return quoteObjectRange(b, pos, '\'', inner)
	case '(', ')', 'b':
		s, e, ok := bracketObjectRange(b, pos, '(', ')', inner)
		return s, e, false, ok
	case '[', ']':
		s, e, ok := bracketObjectRange(b, pos, '[', ']', inner)
		return s, e, false, ok
	case '{', '}', 'B':
		s, e, ok := bracketObjectRange(b, pos, '{', '}', inner)
		return s, e, false, ok
	case 'p':
		return paragraphObjectRange(b, pos, inner)
	case 't':
		s, e, ok := tagObjectRange(b, pos, inner)
		return s, e, false, ok
	default:
		return Pos{}, Pos{}, false, false
	}
}

// wordObjectRange implementa iw/aw (o iW/aW con big=true).
func wordObjectRange(b *Buffer, pos Pos, big, inner bool) (Pos, Pos, bool, bool) {
	line := []rune(b.Line(pos.Line))
	if len(line) == 0 {
		return pos, pos, false, false
	}
	col := clampInt(pos.Col, 0, len(line)-1)
	cls := classify(line[col], big)

	start := col
	for start > 0 && classify(line[start-1], big) == cls {
		start--
	}
	end := col
	for end+1 < len(line) && classify(line[end+1], big) == cls {
		end++
	}
	startPos := Pos{Line: pos.Line, Col: start}
	endPos := Pos{Line: pos.Line, Col: end + 1}

	if inner {
		return startPos, endPos, false, true
	}
	if end+1 < len(line) && classify(line[end+1], big) == classSpace {
		e2 := end + 1
		for e2+1 < len(line) && classify(line[e2+1], big) == classSpace {
			e2++
		}
		return startPos, Pos{Line: pos.Line, Col: e2 + 1}, false, true
	}
	if start > 0 && classify(line[start-1], big) == classSpace {
		s2 := start
		for s2 > 0 && classify(line[s2-1], big) == classSpace {
			s2--
		}
		return Pos{Line: pos.Line, Col: s2}, endPos, false, true
	}
	return startPos, endPos, false, true
}

// quoteObjectRange implementa i"/a" e i'/a'. Como en Vim real, busca los
// pares de comillas solo en la línea actual.
func quoteObjectRange(b *Buffer, pos Pos, q rune, inner bool) (Pos, Pos, bool, bool) {
	line := []rune(b.Line(pos.Line))
	var quotes []int
	for i, r := range line {
		if r == q {
			quotes = append(quotes, i)
		}
	}
	for i := 0; i+1 < len(quotes); i += 2 {
		open, close := quotes[i], quotes[i+1]
		if pos.Col <= close {
			if inner {
				return Pos{Line: pos.Line, Col: open + 1}, Pos{Line: pos.Line, Col: close}, false, true
			}
			return Pos{Line: pos.Line, Col: open}, Pos{Line: pos.Line, Col: close + 1}, false, true
		}
	}
	return Pos{}, Pos{}, false, false
}

// findEnclosingOpen busca, escaneando hacia atrás desde pos, el bracket
// de apertura que encierra a pos (o su propio par si pos ya está sobre
// uno de los dos brackets).
func findEnclosingOpen(b *Buffer, pos Pos, open, close rune) (Pos, bool) {
	if runeAtPos(b, pos) == open {
		return pos, true
	}
	depth := 1
	cur := pos
	for {
		prev, ok := b.Prev(cur)
		if !ok {
			return Pos{}, false
		}
		cur = prev
		switch runeAtPos(b, cur) {
		case close:
			depth++
		case open:
			depth--
			if depth == 0 {
				return cur, true
			}
		}
	}
}

func findMatchingClose(b *Buffer, openPos Pos, open, close rune) (Pos, bool) {
	depth := 1
	cur := openPos
	for {
		next, ok := b.Next(cur)
		if !ok {
			return Pos{}, false
		}
		cur = next
		switch runeAtPos(b, cur) {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return cur, true
			}
		}
	}
}

// bracketObjectRange implementa i(/a( i[/a[ i{/a{ (y sus alias ib/ab,
// iB/aB), buscando el par que encierra a pos, cruzando líneas.
func bracketObjectRange(b *Buffer, pos Pos, open, close rune, inner bool) (Pos, Pos, bool) {
	openPos, ok := findEnclosingOpen(b, pos, open, close)
	if !ok {
		return Pos{}, Pos{}, false
	}
	closePos, ok := findMatchingClose(b, openPos, open, close)
	if !ok {
		return Pos{}, Pos{}, false
	}
	if inner {
		innerStart, ok := b.Next(openPos)
		if !ok || !posLess(innerStart, closePos) {
			return closePos, closePos, true // "()" vacío: nada que seleccionar
		}
		return innerStart, closePos, true
	}
	end, ok := b.Next(closePos)
	if !ok {
		end = b.LastPos()
	}
	return openPos, end, true
}

// paragraphObjectRange implementa ip/ap: un bloque de líneas no vacías
// (o un bloque de líneas vacías, si el cursor está en una). ap además
// incluye las líneas en blanco siguientes (o las anteriores si no hay).
func paragraphObjectRange(b *Buffer, pos Pos, inner bool) (Pos, Pos, bool, bool) {
	isBlank := func(l int) bool { return b.LineLen(l) == 0 }
	start, end := pos.Line, pos.Line

	if isBlank(pos.Line) {
		for start > 0 && isBlank(start-1) {
			start--
		}
		for end < b.LineCount()-1 && isBlank(end+1) {
			end++
		}
		return Pos{Line: start}, Pos{Line: end}, true, true
	}

	for start > 0 && !isBlank(start-1) {
		start--
	}
	for end < b.LineCount()-1 && !isBlank(end+1) {
		end++
	}
	if !inner {
		if end < b.LineCount()-1 && isBlank(end+1) {
			for end < b.LineCount()-1 && isBlank(end+1) {
				end++
			}
		} else {
			for start > 0 && isBlank(start-1) {
				start--
			}
		}
	}
	return Pos{Line: start}, Pos{Line: end}, true, true
}

// tagRe reconoce etiquetas de apertura y cierre estilo HTML/XML.
// Simplificación: no distingue comentarios ni atributos con '<'/'>'
// dentro de cadenas.
var tagRe = regexp.MustCompile(`</?([A-Za-z][A-Za-z0-9]*)[^<>]*?(/?)>`)

// tagObjectRange implementa it/at: busca la pareja <tag>...</tag> que
// encierra a pos (la más interna), en todo el buffer.
func tagObjectRange(b *Buffer, pos Pos, inner bool) (Pos, Pos, bool) {
	text := b.Text()
	cursorOff := posToOffset(b, pos)
	matches := tagRe.FindAllStringSubmatchIndex(text, -1)

	type frame struct {
		openStart, openEnd int
		name               string
	}
	var stack []frame
	var bestOpen frame
	var bestCloseStart, bestCloseEnd int
	found := false

	for _, m := range matches {
		full := text[m[0]:m[1]]
		name := text[m[2]:m[3]]
		selfClose := m[4] != m[5]
		if selfClose {
			continue
		}
		closing := strings.HasPrefix(full, "</")
		if !closing {
			stack = append(stack, frame{m[0], m[1], name})
			continue
		}
		for i := len(stack) - 1; i >= 0; i-- {
			if stack[i].name == name {
				open := stack[i]
				stack = stack[:i]
				if cursorOff >= open.openStart && cursorOff < m[1] {
					if !found || open.openStart > bestOpen.openStart {
						bestOpen = open
						bestCloseStart, bestCloseEnd = m[0], m[1]
						found = true
					}
				}
				break
			}
		}
	}

	if !found {
		return Pos{}, Pos{}, false
	}
	if inner {
		return offsetToPos(b, bestOpen.openEnd), offsetToPos(b, bestCloseStart), true
	}
	return offsetToPos(b, bestOpen.openStart), offsetToPos(b, bestCloseEnd), true
}

func posToOffset(b *Buffer, p Pos) int {
	off := 0
	for l := 0; l < p.Line; l++ {
		off += b.LineLen(l) + 1
	}
	return off + p.Col
}

func offsetToPos(b *Buffer, off int) Pos {
	for l := 0; l < b.LineCount(); l++ {
		lineLen := b.LineLen(l)
		if off <= lineLen {
			return Pos{Line: l, Col: off}
		}
		off -= lineLen + 1
	}
	return b.LastPos()
}
