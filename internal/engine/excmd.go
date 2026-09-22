package engine

import "strings"

// execExCommand interpreta la línea de comando ':' tecleada (sin el
// ':' inicial). Simplificación documentada: :s/:g usan una subcadena
// literal como patrón, no una expresión regular de Vim.
func (e *Engine) execExCommand(cmd string) Event {
	switch {
	case cmd == "w":
		e.recordCommand(":w")
		return Event{Type: EventWrite}
	case cmd == "q":
		e.recordCommand(":q")
		return Event{Type: EventQuit}
	case strings.HasPrefix(cmd, "s/"):
		e.recordCommand(":s")
		return e.execSubstitute(cmd[1:], false)
	case strings.HasPrefix(cmd, "%s/"):
		e.recordCommand(":%s")
		return e.execSubstitute(cmd[2:], true)
	case strings.HasPrefix(cmd, "g/"):
		e.recordCommand(":g")
		return e.execGlobal(cmd[1:])
	default:
		return Event{Type: EventUnsupported}
	}
}

// splitExParts separa "/a/b/c" en ["a","b","c"]. No soporta barras
// escapadas dentro de a/b/c (simplificación).
func splitExParts(s string) []string {
	if len(s) == 0 || s[0] != '/' {
		return nil
	}
	return strings.Split(s[1:], "/")
}

// execSubstitute implementa :s///[g] (línea actual) y :%s///[g] (todo el
// buffer, allLines=true).
func (e *Engine) execSubstitute(rangeCmd string, allLines bool) Event {
	parts := splitExParts(rangeCmd)
	if len(parts) < 2 || parts[0] == "" {
		return Event{Type: EventUnsupported}
	}
	pattern, replacement := parts[0], parts[1]
	global := len(parts) >= 3 && strings.Contains(parts[2], "g")

	lineFrom, lineTo := e.cursor.Line, e.cursor.Line
	if allLines {
		lineFrom, lineTo = 0, e.buf.LineCount()-1
	}

	e.pushUndo()
	changed := false
	for l := lineFrom; l <= lineTo; l++ {
		line := e.buf.Line(l)
		if !strings.Contains(line, pattern) {
			continue
		}
		var newLine string
		if global {
			newLine = strings.ReplaceAll(line, pattern, replacement)
		} else {
			newLine = strings.Replace(line, pattern, replacement, 1)
		}
		e.buf.SetLine(l, newLine)
		changed = true
	}
	if !changed {
		e.undoStack = e.undoStack[:len(e.undoStack)-1]
		return Event{Type: EventUnsupported}
	}
	e.cursor.Col = ClampCol(e.buf.LineLen(e.cursor.Line), e.cursor.Col, ModeNormal)
	return Event{Type: EventChanged}
}

// execGlobal implementa :g/patrón/d (soporte "básico": solo la acción
// de borrar las líneas que coinciden).
func (e *Engine) execGlobal(rest string) Event {
	parts := splitExParts(rest)
	if len(parts) < 2 || parts[0] == "" || strings.TrimSpace(parts[1]) != "d" {
		return Event{Type: EventUnsupported}
	}
	pattern := parts[0]

	var keep []string
	matched := false
	for l := 0; l < e.buf.LineCount(); l++ {
		line := e.buf.Line(l)
		if strings.Contains(line, pattern) {
			matched = true
			continue
		}
		keep = append(keep, line)
	}
	if !matched {
		return Event{Type: EventUnsupported}
	}

	e.pushUndo()
	if len(keep) == 0 {
		keep = []string{""}
	}
	e.buf = NewBuffer(strings.Join(keep, "\n"))
	e.cursor = Pos{Line: 0, Col: firstNonBlank(e.buf.Line(0))}
	return Event{Type: EventChanged}
}
