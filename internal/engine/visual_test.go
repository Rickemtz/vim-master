package engine

import "testing"

func TestVisualCharwise(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantText   string
		wantCursor Pos
		wantMode   Mode
	}{
		{"v entra en Visual", "hola", Pos{0, 0}, "v", "hola", Pos{0, 0}, ModeVisual},
		{"v seguido de Esc vuelve a Normal", "hola", Pos{0, 0}, "v<Esc>", "hola", Pos{0, 0}, ModeNormal},
		{"vlld borra la seleccion", "hola mundo", Pos{0, 0}, "vlld", "a mundo", Pos{0, 0}, ModeNormal},
		{"vlly copia sin borrar", "hola mundo", Pos{0, 0}, "vlly", "hola mundo", Pos{0, 0}, ModeNormal},
		{"vllc cambia la seleccion", "hola mundo", Pos{0, 0}, "vllcXY<Esc>", "XYa mundo", Pos{0, 1}, ModeNormal},
		{"V selecciona la linea completa", "uno\ndos\ntres", Pos{1, 1}, "Vd", "uno\ntres", Pos{1, 0}, ModeNormal},
		{"V hacia abajo selecciona varias lineas", "uno\ndos\ntres", Pos{0, 0}, "Vjd", "tres", Pos{0, 0}, ModeNormal},
		{"viw selecciona la palabra interna", "el coche rojo", Pos{0, 4}, "viwd", "el  rojo", Pos{0, 3}, ModeNormal},
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
			if got := e.Mode(); got != tt.wantMode {
				t.Errorf("Mode() = %v, want %v", got, tt.wantMode)
			}
		})
	}
}

func TestVisualCountAlEntrar(t *testing.T) {
	e := New("uno\ndos\ntres\ncuatro", Pos{0, 0})
	for _, k := range Keys("3Vd") {
		e.Feed(k)
	}
	if got := e.Text(); got != "cuatro" {
		t.Fatalf("Text() = %q, want %q (3V debería preseleccionar 3 líneas)", got, "cuatro")
	}
}

func TestVisualBlock(t *testing.T) {
	// Ctrl-v selecciona un bloque rectangular; d borra esa columna en
	// cada línea del bloque.
	e := New("abc\ndef\nghi", Pos{0, 0})
	e.Feed(CtrlVKey())
	e.Feed(RuneKey('j'))
	e.Feed(RuneKey('j'))
	e.Feed(RuneKey('d'))
	if got := e.Text(); got != "bc\nef\nhi" {
		t.Fatalf("Text() = %q, want %q", got, "bc\nef\nhi")
	}
	if got := e.Mode(); got != ModeNormal {
		t.Fatalf("Mode() = %v, want ModeNormal tras aplicar el operador", got)
	}
}

func TestVisualToggleSameKeyVuelveANormal(t *testing.T) {
	e := New("hola", Pos{0, 0})
	e.Feed(RuneKey('v'))
	if e.Mode() != ModeVisual {
		t.Fatalf("Mode() = %v, want ModeVisual", e.Mode())
	}
	e.Feed(RuneKey('v'))
	if e.Mode() != ModeNormal {
		t.Fatalf("Mode() = %v, want ModeNormal (v de nuevo alterna)", e.Mode())
	}
}
