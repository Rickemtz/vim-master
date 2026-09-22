package engine

import "testing"

func TestSearchBasico(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantCursor Pos
	}{
		{"/ busca hacia adelante", "uno dos tres dos", Pos{0, 0}, "/dos<CR>", Pos{0, 4}},
		{"n repite la busqueda", "uno dos tres dos", Pos{0, 0}, "/dos<CR>n", Pos{0, 13}},
		{"n da la vuelta (wrapscan)", "uno dos tres dos", Pos{0, 13}, "/dos<CR>n", Pos{0, 13}},
		{"N repite en direccion opuesta", "uno dos tres dos", Pos{0, 13}, "/dos<CR>N", Pos{0, 13}},
		{"? busca hacia atras", "uno dos tres dos", Pos{0, 15}, "?dos<CR>", Pos{0, 13}},
		{"* busca la palabra completa bajo el cursor", "foo bar foo baz", Pos{0, 0}, "*", Pos{0, 8}},
		{"# busca hacia atras la palabra completa", "foo bar foo baz", Pos{0, 8}, "#", Pos{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := New(tt.initial, tt.cursor)
			for _, k := range Keys(tt.keys) {
				e.Feed(k)
			}
			if got := e.Cursor(); got != tt.wantCursor {
				t.Errorf("Cursor() = %+v, want %+v", got, tt.wantCursor)
			}
		})
	}
}

func TestSearchSinCoincidenciaNoMueve(t *testing.T) {
	e := New("uno dos tres", Pos{0, 0})
	for _, k := range Keys("/zzz<CR>") {
		e.Feed(k)
	}
	if got := e.Cursor(); got != (Pos{0, 0}) {
		t.Fatalf("Cursor() = %+v, want {0 0} (sin coincidencia no debe mover)", got)
	}
}

func TestSearchEstrellaNoCoincideConSubcadena(t *testing.T) {
	// "foo" bajo el cursor no debería saltar a "foobar" (coincidencia
	// parcial), solo a "foo" completo.
	e := New("foo foobar foo", Pos{0, 0})
	e.Feed(RuneKey('*'))
	if got := e.Cursor(); got != (Pos{0, 11}) {
		t.Fatalf("Cursor() = %+v, want {0 11} (el segundo \"foo\" completo)", got)
	}
}
