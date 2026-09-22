package engine

import "strings"

// Buffer almacena el texto como líneas independientes. Las posiciones de
// columna se manejan en runas (no bytes) para soportar UTF-8 correctamente.
type Buffer struct {
	lines []string
}

func NewBuffer(text string) *Buffer {
	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &Buffer{lines: lines}
}

func (b *Buffer) Text() string { return strings.Join(b.lines, "\n") }

func (b *Buffer) LineCount() int { return len(b.lines) }

func (b *Buffer) Line(n int) string {
	if n < 0 || n >= len(b.lines) {
		return ""
	}
	return b.lines[n]
}

func (b *Buffer) LineLen(n int) int { return len([]rune(b.Line(n))) }

// SetLine reemplaza el contenido completo de la línea n.
func (b *Buffer) SetLine(n int, s string) { b.lines[n] = s }

// TextRange devuelve (sin borrar) el texto entre from (inclusive) y to
// (exclusive), recorriendo saltos de línea.
func (b *Buffer) TextRange(from, to Pos) string {
	if to.Line < from.Line || (to.Line == from.Line && to.Col <= from.Col) {
		return ""
	}
	if from.Line == to.Line {
		line := []rune(b.lines[from.Line])
		start := clampInt(from.Col, 0, len(line))
		end := clampInt(to.Col, 0, len(line))
		return string(line[start:end])
	}

	firstLine := []rune(b.lines[from.Line])
	start := clampInt(from.Col, 0, len(firstLine))
	lastLine := []rune(b.lines[to.Line])
	end := clampInt(to.Col, 0, len(lastLine))

	var out strings.Builder
	out.WriteString(string(firstLine[start:]))
	out.WriteByte('\n')
	for l := from.Line + 1; l < to.Line; l++ {
		out.WriteString(b.lines[l])
		out.WriteByte('\n')
	}
	out.WriteString(string(lastLine[:end]))
	return out.String()
}

// InsertRune inserta r en (pos.Line, pos.Col), desplazando el resto de la
// línea a la derecha.
func (b *Buffer) InsertRune(pos Pos, r rune) {
	line := []rune(b.lines[pos.Line])
	col := clampInt(pos.Col, 0, len(line))
	out := make([]rune, 0, len(line)+1)
	out = append(out, line[:col]...)
	out = append(out, r)
	out = append(out, line[col:]...)
	b.lines[pos.Line] = string(out)
}

// DeleteRune borra la runa en pos y la devuelve. ok es false si pos está
// fuera de rango (p.ej. línea vacía).
func (b *Buffer) DeleteRune(pos Pos) (r rune, ok bool) {
	line := []rune(b.lines[pos.Line])
	if pos.Col < 0 || pos.Col >= len(line) {
		return 0, false
	}
	r = line[pos.Col]
	line = append(line[:pos.Col], line[pos.Col+1:]...)
	b.lines[pos.Line] = string(line)
	return r, true
}

// SplitLine corta la línea de pos en pos.Col: lo anterior queda en la línea
// actual, lo posterior pasa a una línea nueva justo debajo.
func (b *Buffer) SplitLine(pos Pos) {
	line := []rune(b.lines[pos.Line])
	col := clampInt(pos.Col, 0, len(line))
	before := string(line[:col])
	after := string(line[col:])

	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:pos.Line]...)
	newLines = append(newLines, before, after)
	newLines = append(newLines, b.lines[pos.Line+1:]...)
	b.lines = newLines
}

// OpenLineBelow inserta una línea vacía justo debajo de n y devuelve su
// índice.
func (b *Buffer) OpenLineBelow(n int) int {
	idx := n + 1
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:idx]...)
	newLines = append(newLines, "")
	newLines = append(newLines, b.lines[idx:]...)
	b.lines = newLines
	return idx
}

// OpenLineAbove inserta una línea vacía justo encima de n y devuelve su
// índice (== n).
func (b *Buffer) OpenLineAbove(n int) int {
	newLines := make([]string, 0, len(b.lines)+1)
	newLines = append(newLines, b.lines[:n]...)
	newLines = append(newLines, "")
	newLines = append(newLines, b.lines[n:]...)
	b.lines = newLines
	return n
}

// DeleteLines borra las líneas [from, to] (ambas inclusive) y devuelve su
// texto. Si el buffer queda sin líneas, deja una línea vacía (un buffer
// siempre tiene al menos una línea, como en Vim real).
func (b *Buffer) DeleteLines(from, to int) string {
	from = clampInt(from, 0, len(b.lines)-1)
	to = clampInt(to, 0, len(b.lines)-1)
	if from > to {
		from, to = to, from
	}
	deleted := strings.Join(b.lines[from:to+1], "\n")
	rest := make([]string, 0, len(b.lines)-(to-from+1))
	rest = append(rest, b.lines[:from]...)
	rest = append(rest, b.lines[to+1:]...)
	if len(rest) == 0 {
		rest = []string{""}
	}
	b.lines = rest
	return deleted
}

// DeleteRange borra el texto entre from (inclusive) y to (exclusive),
// recorriendo saltos de línea, y lo devuelve.
func (b *Buffer) DeleteRange(from, to Pos) string {
	if to.Line < from.Line || (to.Line == from.Line && to.Col <= from.Col) {
		return ""
	}
	if from.Line == to.Line {
		line := []rune(b.lines[from.Line])
		start := clampInt(from.Col, 0, len(line))
		end := clampInt(to.Col, 0, len(line))
		deleted := string(line[start:end])
		b.lines[from.Line] = string(line[:start]) + string(line[end:])
		return deleted
	}

	firstLine := []rune(b.lines[from.Line])
	start := clampInt(from.Col, 0, len(firstLine))
	lastLine := []rune(b.lines[to.Line])
	end := clampInt(to.Col, 0, len(lastLine))

	var deleted strings.Builder
	deleted.WriteString(string(firstLine[start:]))
	deleted.WriteByte('\n')
	for l := from.Line + 1; l < to.Line; l++ {
		deleted.WriteString(b.lines[l])
		deleted.WriteByte('\n')
	}
	deleted.WriteString(string(lastLine[:end]))

	merged := string(firstLine[:start]) + string(lastLine[end:])
	newLines := make([]string, 0, len(b.lines)-(to.Line-from.Line))
	newLines = append(newLines, b.lines[:from.Line]...)
	newLines = append(newLines, merged)
	newLines = append(newLines, b.lines[to.Line+1:]...)
	b.lines = newLines
	return deleted.String()
}

// InsertText inserta text (que puede contener saltos de línea) en pos.
func (b *Buffer) InsertText(pos Pos, text string) {
	parts := strings.Split(text, "\n")
	line := []rune(b.lines[pos.Line])
	col := clampInt(pos.Col, 0, len(line))
	before := string(line[:col])
	after := string(line[col:])

	if len(parts) == 1 {
		b.lines[pos.Line] = before + parts[0] + after
		return
	}

	newLines := make([]string, 0, len(b.lines)+len(parts)-1)
	newLines = append(newLines, b.lines[:pos.Line]...)
	newLines = append(newLines, before+parts[0])
	newLines = append(newLines, parts[1:len(parts)-1]...)
	newLines = append(newLines, parts[len(parts)-1]+after)
	newLines = append(newLines, b.lines[pos.Line+1:]...)
	b.lines = newLines
}

// InsertLinesBelow inserta cada elemento de lines como una línea nueva
// justo debajo de n.
func (b *Buffer) InsertLinesBelow(n int, lines []string) {
	idx := n + 1
	newLines := make([]string, 0, len(b.lines)+len(lines))
	newLines = append(newLines, b.lines[:idx]...)
	newLines = append(newLines, lines...)
	newLines = append(newLines, b.lines[idx:]...)
	b.lines = newLines
}

// Next devuelve la posición siguiente a p, cruzando líneas. ok es false
// si p ya es la última posición del buffer.
func (b *Buffer) Next(p Pos) (Pos, bool) {
	lineLen := b.LineLen(p.Line)
	if p.Col < lineLen {
		return Pos{Line: p.Line, Col: p.Col + 1}, true
	}
	if p.Line < b.LineCount()-1 {
		return Pos{Line: p.Line + 1, Col: 0}, true
	}
	return p, false
}

// Prev devuelve la posición anterior a p, cruzando líneas. ok es false
// si p ya es la primera posición del buffer.
func (b *Buffer) Prev(p Pos) (Pos, bool) {
	if p.Col > 0 {
		return Pos{Line: p.Line, Col: p.Col - 1}, true
	}
	if p.Line > 0 {
		return Pos{Line: p.Line - 1, Col: b.LineLen(p.Line - 1)}, true
	}
	return p, false
}

// LastPos devuelve la última posición válida (fin de la última línea).
func (b *Buffer) LastPos() Pos {
	last := b.LineCount() - 1
	return Pos{Line: last, Col: b.LineLen(last)}
}
