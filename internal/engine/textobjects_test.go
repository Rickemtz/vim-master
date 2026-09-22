package engine

import "testing"

func TestTextObjects(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantText   string
		wantCursor Pos
	}{
		{"diw borra la palabra interna", "el coche rojo", Pos{0, 4}, "diw", "el  rojo", Pos{0, 3}},
		{"daw borra la palabra y su espacio", "el coche rojo", Pos{0, 4}, "daw", "el rojo", Pos{0, 3}},
		{"ciw cambia la palabra", "el coche rojo", Pos{0, 4}, "ciwauto<Esc>", "el auto rojo", Pos{0, 6}},
		{"di\" borra dentro de comillas", `di "hola mundo" fin`, Pos{0, 6}, `di"`, `di "" fin`, Pos{0, 4}},
		{"da\" borra las comillas tambien", `di "hola mundo" fin`, Pos{0, 6}, `da"`, `di  fin`, Pos{0, 3}},
		{"di( borra dentro del parentesis", "foo(bar, baz)", Pos{0, 6}, "di(", "foo()", Pos{0, 4}},
		{"da( borra el parentesis tambien", "foo(bar, baz)", Pos{0, 6}, "da(", "foo", Pos{0, 2}},
		{"di{ funciona anidado", "func(bar(baz))", Pos{0, 10}, "di(", "func(bar())", Pos{0, 9}},
		{"dip borra el parrafo", "uno\ndos\n\ncuatro", Pos{0, 0}, "dip", "\ncuatro", Pos{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New(tt.initial, tt.cursor)
			for _, k := range Keys(tt.keys) {
				e.Feed(k)
			}
			if got := e.Text(); got != tt.wantText {
				t.Errorf("Text() = %q, want %q", got, tt.wantText)
			}
			if got := e.Cursor(); got != tt.wantCursor {
				t.Errorf("Cursor() = %+v, want %+v", got, tt.wantCursor)
			}
		})
	}
}

func TestTextObjectTag(t *testing.T) {
	e := New("<p>hola <b>mundo</b></p>", Pos{0, 12})
	for _, k := range Keys("dit") {
		e.Feed(k)
	}
	if got := e.Text(); got != "<p>hola <b></b></p>" {
		t.Fatalf("Text() = %q, want %q", got, "<p>hola <b></b></p>")
	}
}

func TestTextObjectSinCoincidenciaNoCambiaNada(t *testing.T) {
	e := New("sin comillas aqui", Pos{0, 0})
	for _, k := range Keys(`di"`) {
		e.Feed(k)
	}
	if got := e.Text(); got != "sin comillas aqui" {
		t.Fatalf("Text() = %q, want sin cambios", got)
	}
}
