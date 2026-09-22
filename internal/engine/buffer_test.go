package engine

import "testing"

func TestBufferText(t *testing.T) {
	b := NewBuffer("hola\nmundo")
	if got := b.Text(); got != "hola\nmundo" {
		t.Fatalf("Text() = %q, want %q", got, "hola\nmundo")
	}
	if got := b.LineCount(); got != 2 {
		t.Fatalf("LineCount() = %d, want 2", got)
	}
}

func TestBufferInsertRune(t *testing.T) {
	b := NewBuffer("hola")
	b.InsertRune(Pos{0, 0}, 'X')
	if got := b.Line(0); got != "Xhola" {
		t.Fatalf("Line(0) = %q, want %q", got, "Xhola")
	}
	b.InsertRune(Pos{0, 5}, 'Y')
	if got := b.Line(0); got != "XholaY" {
		t.Fatalf("Line(0) = %q, want %q", got, "XholaY")
	}
}

func TestBufferDeleteRune(t *testing.T) {
	b := NewBuffer("hola")
	r, ok := b.DeleteRune(Pos{0, 0})
	if !ok || r != 'h' {
		t.Fatalf("DeleteRune = %q, %v, want 'h', true", r, ok)
	}
	if got := b.Line(0); got != "ola" {
		t.Fatalf("Line(0) = %q, want %q", got, "ola")
	}

	if _, ok := b.DeleteRune(Pos{0, 99}); ok {
		t.Fatalf("DeleteRune fuera de rango debería devolver ok=false")
	}

	empty := NewBuffer("")
	if _, ok := empty.DeleteRune(Pos{0, 0}); ok {
		t.Fatalf("DeleteRune en línea vacía debería devolver ok=false")
	}
}

func TestBufferSplitLine(t *testing.T) {
	b := NewBuffer("hola mundo")
	b.SplitLine(Pos{0, 4})
	if got := b.Text(); got != "hola\n mundo" {
		t.Fatalf("Text() = %q, want %q", got, "hola\n mundo")
	}
}

func TestBufferOpenLineBelowAbove(t *testing.T) {
	b := NewBuffer("uno\ndos")
	idx := b.OpenLineBelow(0)
	if idx != 1 {
		t.Fatalf("OpenLineBelow devolvió %d, want 1", idx)
	}
	if got := b.Text(); got != "uno\n\ndos" {
		t.Fatalf("Text() = %q, want %q", got, "uno\n\ndos")
	}

	b2 := NewBuffer("uno\ndos")
	idx2 := b2.OpenLineAbove(1)
	if idx2 != 1 {
		t.Fatalf("OpenLineAbove devolvió %d, want 1", idx2)
	}
	if got := b2.Text(); got != "uno\n\ndos" {
		t.Fatalf("Text() = %q, want %q", got, "uno\n\ndos")
	}
}

func TestBufferUTF8(t *testing.T) {
	b := NewBuffer("café")
	if got := b.LineLen(0); got != 4 {
		t.Fatalf("LineLen(0) = %d, want 4 (runas, no bytes)", got)
	}
	b.InsertRune(Pos{0, 4}, '!')
	if got := b.Line(0); got != "café!" {
		t.Fatalf("Line(0) = %q, want %q", got, "café!")
	}
}
