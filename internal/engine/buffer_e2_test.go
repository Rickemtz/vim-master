package engine

import "testing"

func TestBufferDeleteLines(t *testing.T) {
	b := NewBuffer("uno\ndos\ntres\ncuatro")
	deleted := b.DeleteLines(1, 2)
	if deleted != "dos\ntres" {
		t.Fatalf("deleted = %q, want %q", deleted, "dos\ntres")
	}
	if got := b.Text(); got != "uno\ncuatro" {
		t.Fatalf("Text() = %q, want %q", got, "uno\ncuatro")
	}
}

func TestBufferDeleteLinesTodoElBufferDejaUnaLineaVacia(t *testing.T) {
	b := NewBuffer("uno\ndos")
	b.DeleteLines(0, 1)
	if got := b.Text(); got != "" {
		t.Fatalf("Text() = %q, want vacío", got)
	}
	if got := b.LineCount(); got != 1 {
		t.Fatalf("LineCount() = %d, want 1 (Vim siempre deja al menos una línea)", got)
	}
}

func TestBufferDeleteRangeMismaLinea(t *testing.T) {
	b := NewBuffer("el coche rojo")
	deleted := b.DeleteRange(Pos{0, 3}, Pos{0, 9})
	if deleted != "coche " {
		t.Fatalf("deleted = %q, want %q", deleted, "coche ")
	}
	if got := b.Text(); got != "el rojo" {
		t.Fatalf("Text() = %q, want %q", got, "el rojo")
	}
}

func TestBufferDeleteRangeEntreLineas(t *testing.T) {
	b := NewBuffer("hola\nmundo\ncruel")
	deleted := b.DeleteRange(Pos{0, 2}, Pos{2, 2})
	if deleted != "la\nmundo\ncr" {
		t.Fatalf("deleted = %q, want %q", deleted, "la\nmundo\ncr")
	}
	if got := b.Text(); got != "houel" {
		t.Fatalf("Text() = %q, want %q", got, "houel")
	}
}

func TestBufferInsertText(t *testing.T) {
	b := NewBuffer("hola")
	b.InsertText(Pos{0, 4}, " mundo")
	if got := b.Text(); got != "hola mundo" {
		t.Fatalf("Text() = %q, want %q", got, "hola mundo")
	}
}

func TestBufferInsertTextMultilinea(t *testing.T) {
	b := NewBuffer("holamundo")
	b.InsertText(Pos{0, 4}, "\n")
	if got := b.Text(); got != "hola\nmundo" {
		t.Fatalf("Text() = %q, want %q", got, "hola\nmundo")
	}
}

func TestBufferNextPrev(t *testing.T) {
	b := NewBuffer("ab\ncd")

	p, ok := b.Next(Pos{0, 2}) // fin de "ab" -> cruza a la línea siguiente
	if !ok || p != (Pos{1, 0}) {
		t.Fatalf("Next() = %+v, %v, want {1 0}, true", p, ok)
	}

	p, ok = b.Prev(Pos{1, 0})
	if !ok || p != (Pos{0, 2}) {
		t.Fatalf("Prev() = %+v, %v, want {0 2}, true", p, ok)
	}

	if _, ok := b.Prev(Pos{0, 0}); ok {
		t.Fatal("Prev() en la primera posición debería devolver ok=false")
	}
	if _, ok := b.Next(Pos{1, 2}); ok {
		t.Fatal("Next() en la última posición debería devolver ok=false")
	}
}
