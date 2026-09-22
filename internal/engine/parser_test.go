package engine

import "testing"

func TestEngineE2E3E4(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantText   string
		wantCursor Pos
		wantMode   Mode
	}{
		// E2: motions de palabra y línea, counts.
		{"w avanza a la siguiente palabra", "el coche rojo", Pos{0, 0}, "w", "el coche rojo", Pos{0, 3}, ModeNormal},
		{"2w avanza dos palabras", "el coche rojo", Pos{0, 0}, "2w", "el coche rojo", Pos{0, 9}, ModeNormal},
		{"b retrocede una palabra", "el coche rojo", Pos{0, 9}, "b", "el coche rojo", Pos{0, 3}, ModeNormal},
		{"e va al final de la palabra", "el coche rojo", Pos{0, 0}, "e", "el coche rojo", Pos{0, 1}, ModeNormal},
		{"dollar va al final de la linea", "hola mundo", Pos{0, 0}, "$", "hola mundo", Pos{0, 9}, ModeNormal},
		{"0 va al inicio de la linea", "hola mundo", Pos{0, 9}, "0", "hola mundo", Pos{0, 0}, ModeNormal},
		{"caret va al primer no blanco", "  hola", Pos{0, 5}, "^", "  hola", Pos{0, 2}, ModeNormal},
		{"gg va a la primera linea", "uno\ndos\ntres", Pos{2, 0}, "gg", "uno\ndos\ntres", Pos{0, 0}, ModeNormal},
		{"G va a la ultima linea", "uno\ndos\ntres", Pos{0, 0}, "G", "uno\ndos\ntres", Pos{2, 0}, ModeNormal},
		{"2G va a la linea 2", "uno\ndos\ntres", Pos{0, 0}, "2G", "uno\ndos\ntres", Pos{1, 0}, ModeNormal},
		{"3l con count mueve 3", "hola mundo", Pos{0, 0}, "3l", "hola mundo", Pos{0, 3}, ModeNormal},

		// E3: operadores + motion, dd/cc/yy, D/C, p/P, r, s, u/Ctrl-r.
		{"dw borra hasta la siguiente palabra", "el coche rojo", Pos{0, 3}, "dw", "el rojo", Pos{0, 3}, ModeNormal},
		{"cw se comporta como ce (no incluye el espacio)", "el coche rojo", Pos{0, 3}, "cwXXX<Esc>", "el XXX rojo", Pos{0, 5}, ModeNormal},
		{"dd borra la linea completa", "uno\ndos\ntres", Pos{1, 0}, "dd", "uno\ntres", Pos{1, 0}, ModeNormal},
		{"2dd borra dos lineas", "uno\ndos\ntres\ncuatro", Pos{0, 0}, "2dd", "tres\ncuatro", Pos{0, 0}, ModeNormal},
		{"yy despues p duplica la linea", "uno\ndos", Pos{0, 0}, "yyp", "uno\nuno\ndos", Pos{1, 0}, ModeNormal},
		{"yw despues p pega la palabra", "hola mundo", Pos{0, 0}, "ywwP", "hola hola mundo", Pos{0, 9}, ModeNormal},
		{"D borra hasta el final de la linea", "hola mundo", Pos{0, 4}, "D", "hola", Pos{0, 3}, ModeNormal},
		{"C cambia hasta el final de la linea", "hola mundo", Pos{0, 4}, "CXY<Esc>", "holaXY", Pos{0, 5}, ModeNormal},
		{"x despues p pega el caracter borrado", "hola", Pos{0, 0}, "xllp", "olah", Pos{0, 3}, ModeNormal},
		{"r reemplaza el caracter bajo el cursor", "hola", Pos{0, 0}, "rX", "Xola", Pos{0, 0}, ModeNormal},
		{"3rx reemplaza 3 caracteres por x", "hola mundo", Pos{0, 0}, "3rx", "xxxa mundo", Pos{0, 2}, ModeNormal},
		{"s sustituye el caracter y entra en insert", "hola", Pos{0, 0}, "sXY<Esc>", "XYola", Pos{0, 1}, ModeNormal},
		{"u deshace el ultimo cambio", "hola", Pos{0, 0}, "xu", "hola", Pos{0, 0}, ModeNormal},
		{"ctrl+r rehace lo deshecho", "hola", Pos{0, 0}, "xu<C-r>", "ola", Pos{0, 0}, ModeNormal},

		// E4: f F t T ; , % y . (repetir).
		{"fx mueve al siguiente caracter x", "el coche rojo", Pos{0, 0}, "fr", "el coche rojo", Pos{0, 9}, ModeNormal},
		{"tx para justo antes del caracter", "el coche rojo", Pos{0, 0}, "tr", "el coche rojo", Pos{0, 8}, ModeNormal},
		{"punto y coma repite el ultimo f", "a-b-c-d", Pos{0, 0}, "f-;", "a-b-c-d", Pos{0, 3}, ModeNormal},
		{"coma repite en direccion opuesta", "a-b-c-d", Pos{0, 0}, "f-;,", "a-b-c-d", Pos{0, 1}, ModeNormal},
		{"porcentaje salta a la pareja", "foo(bar)", Pos{0, 0}, "%", "foo(bar)", Pos{0, 7}, ModeNormal},
		{"dfx borra hasta incluir el caracter", "el coche, rojo", Pos{0, 0}, "df,", " rojo", Pos{0, 0}, ModeNormal},
		{"punto repite el ultimo cambio (x)", "hola", Pos{0, 0}, "x.", "la", Pos{0, 0}, ModeNormal},
		{"punto repite un cambio de insert", "mundo", Pos{0, 0}, "iHola <Esc>w.", "Hola Hola mundo", Pos{0, 9}, ModeNormal},
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

func TestEnginePendingKeysCountYOperador(t *testing.T) {
	e := New("hola mundo", Pos{0, 0})
	for _, k := range Keys("2d") {
		e.Feed(k)
	}
	if got := e.PendingKeys(); got != "2d" {
		t.Fatalf("PendingKeys() = %q, want %q", got, "2d")
	}
}

func TestEnginePendingKeysTextObject(t *testing.T) {
	e := New("hola mundo", Pos{0, 0})
	for _, k := range Keys("di") {
		e.Feed(k)
	}
	if got := e.PendingKeys(); got != "di" {
		t.Fatalf("PendingKeys() = %q, want %q", got, "di")
	}
}

func TestEngineDotNoSeAutoRepite(t *testing.T) {
	// x . . . debe borrar 4 caracteres en total (el original x + 3 puntos
	// que repiten x, no "repetir el punto anterior").
	e := New("holamundo", Pos{0, 0})
	for _, k := range Keys("x...") {
		e.Feed(k)
	}
	if got := e.Text(); got != "mundo" {
		t.Fatalf("Text() = %q, want %q", got, "mundo")
	}
}

func TestEngineUndoRedoStackVacioNoPanica(t *testing.T) {
	e := New("hola", Pos{0, 0})
	e.Feed(RuneKey('u')) // nada que deshacer
	if got := e.Text(); got != "hola" {
		t.Fatalf("Text() = %q, want %q sin cambios", got, "hola")
	}
	e.Feed(CtrlRKey()) // nada que rehacer
	if got := e.Text(); got != "hola" {
		t.Fatalf("Text() = %q, want %q sin cambios", got, "hola")
	}
}

func TestEngineFNoEncuentraNoMueveElCursor(t *testing.T) {
	e := New("hola", Pos{0, 1})
	e.Feed(RuneKey('f'))
	e.Feed(RuneKey('z'))
	if got := e.Cursor(); got != (Pos{0, 1}) {
		t.Fatalf("Cursor() = %+v, want {0 1} (f sin coincidencia no mueve)", got)
	}
}

func TestEngineOperadorConMotionSinCoincidenciaNoCambiaNada(t *testing.T) {
	e := New("hola mundo", Pos{0, 0})
	for _, k := range Keys("dfz") { // no hay 'z' en la línea
		e.Feed(k)
	}
	if got := e.Text(); got != "hola mundo" {
		t.Fatalf("Text() = %q, want sin cambios", got)
	}
}
