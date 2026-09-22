package engine

// undoState es una foto completa del buffer y el cursor. Vim real hace
// undo por diffs; aquí usamos snapshots completos por simplicidad, lo que
// es correcto y suficientemente rápido para los tamaños de buffer de
// VimDojo.
type undoState struct {
	text   string
	cursor Pos
}

// pushUndo guarda el estado ANTES de un cambio. Se llama una sola vez por
// "unidad de cambio" (p.ej. al entrar a Insert, no en cada tecla tecleada
// dentro de Insert), para que u deshaga la unidad completa de una vez.
func (e *Engine) pushUndo() {
	e.undoStack = append(e.undoStack, undoState{text: e.buf.Text(), cursor: e.cursor})
	e.redoStack = nil
}

func (e *Engine) undo() {
	if len(e.undoStack) == 0 {
		return
	}
	e.redoStack = append(e.redoStack, undoState{text: e.buf.Text(), cursor: e.cursor})
	last := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.buf = NewBuffer(last.text)
	e.cursor = last.cursor
}

func (e *Engine) redo() {
	if len(e.redoStack) == 0 {
		return
	}
	e.undoStack = append(e.undoStack, undoState{text: e.buf.Text(), cursor: e.cursor})
	last := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
	e.buf = NewBuffer(last.text)
	e.cursor = last.cursor
}
