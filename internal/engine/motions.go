package engine

import "unicode"

// MotionKind clasifica cómo un operador debe usar el resultado de un
// motion: linewise opera en líneas completas, inclusive incluye el
// carácter final, exclusive lo deja fuera.
type MotionKind int

const (
	MotionExclusive MotionKind = iota
	MotionInclusive
	MotionLinewise
)

// MotionResult es el resultado de aplicar un motion. Found es false
// cuando el motion no pudo ejecutarse (p.ej. f/t sin coincidencia), y en
// ese caso el cursor no debe moverse.
type MotionResult struct {
	Pos   Pos
	Kind  MotionKind
	Found bool
}

type charClass int

const (
	classSpace charClass = iota
	classWord
	classPunct
)

// classify clasifica una runa para los motions de palabra. En WORD
// (big=true, W/B/E) todo lo que no es espacio es una sola clase.
func classify(r rune, big bool) charClass {
	if r == ' ' || r == '\t' {
		return classSpace
	}
	if big {
		return classWord
	}
	if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
		return classWord
	}
	return classPunct
}

// classAt trata el final de línea (una columna más allá del último
// carácter) y las líneas vacías como espacio.
func classAt(b *Buffer, p Pos, big bool) charClass {
	line := []rune(b.Line(p.Line))
	if p.Col < 0 || p.Col >= len(line) {
		return classSpace
	}
	return classify(line[p.Col], big)
}

// wordForwardOnce implementa un solo paso de w/W. Una línea vacía cuenta
// como una palabra (:help word).
func wordForwardOnce(b *Buffer, pos Pos, big bool) Pos {
	if cls := classAt(b, pos, big); cls != classSpace {
		for {
			next, ok := b.Next(pos)
			if !ok || next.Line != pos.Line || classAt(b, next, big) != cls {
				break
			}
			pos = next
		}
	}
	for {
		next, ok := b.Next(pos)
		if !ok {
			return pos
		}
		pos = next
		if b.LineLen(pos.Line) == 0 || classAt(b, pos, big) != classSpace {
			return pos
		}
	}
}

// wordBackwardOnce implementa un solo paso de b/B.
func wordBackwardOnce(b *Buffer, pos Pos, big bool) Pos {
	prev, ok := b.Prev(pos)
	if !ok {
		return pos
	}
	pos = prev
	for b.LineLen(pos.Line) != 0 && classAt(b, pos, big) == classSpace {
		prev, ok := b.Prev(pos)
		if !ok {
			return pos
		}
		pos = prev
	}
	if b.LineLen(pos.Line) == 0 {
		return pos
	}
	cls := classAt(b, pos, big)
	for {
		prev, ok := b.Prev(pos)
		if !ok || prev.Line != pos.Line || classAt(b, prev, big) != cls {
			return pos
		}
		pos = prev
	}
}

// wordEndOnce implementa un solo paso de e/E.
func wordEndOnce(b *Buffer, pos Pos, big bool) Pos {
	next, ok := b.Next(pos)
	if !ok {
		return pos
	}
	pos = next
	for b.LineLen(pos.Line) != 0 && classAt(b, pos, big) == classSpace {
		next, ok := b.Next(pos)
		if !ok {
			return pos
		}
		pos = next
	}
	if b.LineLen(pos.Line) == 0 {
		return pos
	}
	cls := classAt(b, pos, big)
	for {
		next, ok := b.Next(pos)
		if !ok || next.Line != pos.Line || classAt(b, next, big) != cls {
			return pos
		}
		pos = next
	}
}

func motionWordForward(b *Buffer, pos Pos, count int, big bool) MotionResult {
	for i := 0; i < count; i++ {
		pos = wordForwardOnce(b, pos, big)
	}
	return MotionResult{Pos: pos, Kind: MotionExclusive, Found: true}
}

func motionWordBackward(b *Buffer, pos Pos, count int, big bool) MotionResult {
	for i := 0; i < count; i++ {
		pos = wordBackwardOnce(b, pos, big)
	}
	return MotionResult{Pos: pos, Kind: MotionExclusive, Found: true}
}

func motionWordEnd(b *Buffer, pos Pos, count int, big bool) MotionResult {
	for i := 0; i < count; i++ {
		pos = wordEndOnce(b, pos, big)
	}
	return MotionResult{Pos: pos, Kind: MotionInclusive, Found: true}
}

func motionCharLeft(pos Pos, count int) MotionResult {
	pos.Col -= count
	if pos.Col < 0 {
		pos.Col = 0
	}
	return MotionResult{Pos: pos, Kind: MotionExclusive, Found: true}
}

func motionCharRight(b *Buffer, pos Pos, count int) MotionResult {
	max := b.LineLen(pos.Line) - 1
	if max < 0 {
		max = 0
	}
	pos.Col += count
	if pos.Col > max {
		pos.Col = max
	}
	return MotionResult{Pos: pos, Kind: MotionExclusive, Found: true}
}

func motionLineStart(pos Pos) MotionResult {
	return MotionResult{Pos: Pos{Line: pos.Line, Col: 0}, Kind: MotionExclusive, Found: true}
}

func motionFirstNonBlank(b *Buffer, pos Pos) MotionResult {
	return MotionResult{Pos: Pos{Line: pos.Line, Col: firstNonBlank(b.Line(pos.Line))}, Kind: MotionExclusive, Found: true}
}

// motionLineEnd implementa $. Con count>1 baja count-1 líneas antes de ir
// al final (como en Vim real).
func motionLineEnd(b *Buffer, pos Pos, count int) MotionResult {
	line := clampInt(pos.Line+count-1, 0, b.LineCount()-1)
	col := b.LineLen(line) - 1
	if col < 0 {
		col = 0
	}
	return MotionResult{Pos: Pos{Line: line, Col: col}, Kind: MotionInclusive, Found: true}
}

// motionGotoLine implementa gg/G/{n}G. count<=0 significa "sin count":
// gg va a la primera línea, G a la última.
func motionGotoLine(b *Buffer, count int, defaultLast bool) MotionResult {
	var line int
	switch {
	case count > 0:
		line = clampInt(count-1, 0, b.LineCount()-1)
	case defaultLast:
		line = b.LineCount() - 1
	default:
		line = 0
	}
	return MotionResult{Pos: Pos{Line: line, Col: firstNonBlank(b.Line(line))}, Kind: MotionLinewise, Found: true}
}

// motionFindChar implementa f/F/t/T. forward=false es F/T (buscar hacia
// atrás); till=true es t/T (parar justo antes del carácter).
func motionFindChar(b *Buffer, pos Pos, ch rune, count int, forward, till bool) MotionResult {
	line := []rune(b.Line(pos.Line))
	found := pos.Col

	for i := 0; i < count; i++ {
		if forward {
			idx := -1
			for j := found + 1; j < len(line); j++ {
				if line[j] == ch {
					idx = j
					break
				}
			}
			if idx == -1 {
				return MotionResult{Found: false}
			}
			found = idx
		} else {
			idx := -1
			for j := found - 1; j >= 0; j-- {
				if line[j] == ch {
					idx = j
					break
				}
			}
			if idx == -1 {
				return MotionResult{Found: false}
			}
			found = idx
		}
	}

	resultCol := found
	kind := MotionInclusive
	if till {
		if forward {
			resultCol = found - 1
		} else {
			resultCol = found + 1
		}
	}
	if !forward {
		kind = MotionExclusive
	}
	return MotionResult{Pos: Pos{Line: pos.Line, Col: resultCol}, Kind: kind, Found: true}
}

var matchOpen = map[rune]rune{'(': ')', '[': ']', '{': '}'}
var matchClose = map[rune]rune{')': '(', ']': '[', '}': '{'}

func runeAtPos(b *Buffer, p Pos) rune {
	line := []rune(b.Line(p.Line))
	if p.Col < 0 || p.Col >= len(line) {
		return 0
	}
	return line[p.Col]
}

// motionMatchPair implementa %: salta al carácter que hace pareja con el
// primer (){}[] encontrado en la línea actual desde el cursor.
func motionMatchPair(b *Buffer, pos Pos) MotionResult {
	line := []rune(b.Line(pos.Line))
	start := -1
	var ch rune
	for i := pos.Col; i < len(line); i++ {
		if _, ok := matchOpen[line[i]]; ok {
			start, ch = i, line[i]
			break
		}
		if _, ok := matchClose[line[i]]; ok {
			start, ch = i, line[i]
			break
		}
	}
	if start == -1 {
		return MotionResult{Found: false}
	}
	cur := Pos{Line: pos.Line, Col: start}

	if close, isOpen := matchOpen[ch]; isOpen {
		depth := 1
		for {
			next, ok := b.Next(cur)
			if !ok {
				return MotionResult{Found: false}
			}
			cur = next
			switch runeAtPos(b, cur) {
			case ch:
				depth++
			case close:
				depth--
				if depth == 0 {
					return MotionResult{Pos: cur, Kind: MotionInclusive, Found: true}
				}
			}
		}
	}

	open := matchClose[ch]
	depth := 1
	for {
		prev, ok := b.Prev(cur)
		if !ok {
			return MotionResult{Found: false}
		}
		cur = prev
		switch runeAtPos(b, cur) {
		case ch:
			depth++
		case open:
			depth--
			if depth == 0 {
				return MotionResult{Pos: cur, Kind: MotionInclusive, Found: true}
			}
		}
	}
}
