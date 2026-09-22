package engine

import "testing"

func TestEngineE8(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantText   string
		wantCursor Pos
	}{
		// "ayw guarda "uno " en el registro a; w mueve a "dos"; x borra la
		// "d" (va al registro sin nombre, no toca "a"); "ap pega desde el
		// registro a, demostrando que es independiente del sin nombre.
		{"registro con nombre: yank y paste independiente del sin nombre", "uno dos", Pos{0, 0}, `"aywwx"ap`, "uno ouno s", Pos{0, 8}},
		{">> indenta la linea", "hola", Pos{0, 0}, ">>", "    hola", Pos{0, 4}},
		{"<< quita la indentacion", "    hola", Pos{0, 0}, "<<", "hola", Pos{0, 0}},
		{"~ alterna mayusculas y avanza", "hola", Pos{0, 0}, "~~", "HOla", Pos{0, 2}},
		{"J une dos lineas con un espacio", "hola\nmundo", Pos{0, 0}, "J", "hola mundo", Pos{0, 4}},
		{"3J une tres lineas", "uno\ndos\ntres", Pos{0, 0}, "3J", "uno dos tres", Pos{0, 7}},
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

func TestEngineE9Substitute(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		cursor   Pos
		keys     string
		wantText string
	}{
		{":s reemplaza la primera coincidencia de la linea", "foo foo foo", Pos{0, 0}, ":s/foo/bar/<CR>", "bar foo foo"},
		{":s con g reemplaza todas en la linea", "foo foo foo", Pos{0, 0}, ":s/foo/bar/g<CR>", "bar bar bar"},
		{":%s reemplaza la primera de cada linea", "foo uno\nfoo dos", Pos{0, 0}, ":%s/foo/bar/<CR>", "bar uno\nbar dos"},
		{":%s con g reemplaza todas en todas las lineas", "foo foo\nfoo foo", Pos{0, 0}, ":%s/foo/bar/g<CR>", "bar bar\nbar bar"},
		{":g borra las lineas que coinciden", "uno\nbadline\ndos\nbadline", Pos{0, 0}, ":g/badline/d<CR>", "uno\ndos"},
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
		})
	}
}

func TestEngineE10Macros(t *testing.T) {
	// Graba "x" con qa...q (esa x se ejecuta de verdad al grabarla) y la
	// repite dos veces con @a: 3 borrados en total.
	e := New("holamundo", Pos{0, 0})
	for _, k := range Keys("qaxq@a@a") {
		e.Feed(k)
	}
	if got := e.Text(); got != "amundo" {
		t.Fatalf("Text() = %q, want %q", got, "amundo")
	}
}

func TestEngineE10MacroConCount(t *testing.T) {
	e := New("holamundo", Pos{0, 0})
	for _, k := range Keys("qaxq3@a") {
		e.Feed(k)
	}
	if got := e.Text(); got != "mundo" {
		t.Fatalf("Text() = %q, want %q", got, "mundo")
	}
}

func TestEngineE10ArrobaArrobaRepiteLaUltimaMacro(t *testing.T) {
	e := New("holamundo", Pos{0, 0})
	for _, k := range Keys("qaxq@a@@") {
		e.Feed(k)
	}
	if got := e.Text(); got != "amundo" {
		t.Fatalf("Text() = %q, want %q", got, "amundo")
	}
}

func TestEngineE10MacroGrabaSoloLoOcurridoDentro(t *testing.T) {
	e := New("aaaa", Pos{0, 0})
	// qa x q -> graba solo esa x (que además se ejecuta al grabarla);
	// la x suelta de después queda fuera de la macro; @a repite solo la
	// grabada. 3 borrados en total sobre 4 "a": 1 grabada + 1 manual +
	// 1 repetida.
	for _, k := range Keys("qaxqx@a") {
		e.Feed(k)
	}
	if got := e.Text(); got != "a" {
		t.Fatalf("Text() = %q, want %q", got, "a")
	}
}
