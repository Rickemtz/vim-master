package engine

import "testing"

func TestEngineE1(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursor     Pos
		keys       string
		wantText   string
		wantCursor Pos
		wantMode   Mode
	}{
		{"h mueve a la izquierda", "hola", Pos{0, 2}, "h", "hola", Pos{0, 1}, ModeNormal},
		{"h no pasa de columna 0", "hola", Pos{0, 0}, "h", "hola", Pos{0, 0}, ModeNormal},
		{"l mueve a la derecha", "hola", Pos{0, 0}, "l", "hola", Pos{0, 1}, ModeNormal},
		{"l no pasa del ultimo caracter", "hola", Pos{0, 3}, "l", "hola", Pos{0, 3}, ModeNormal},
		{"j conserva columna deseada", "hola\nmu", Pos{0, 3}, "j", "hola\nmu", Pos{1, 1}, ModeNormal},
		{"j recupera columna deseada en linea mas larga", "hola\nmu\nmundo", Pos{0, 3}, "jj", "hola\nmu\nmundo", Pos{2, 3}, ModeNormal},
		{"k mueve arriba", "hola\nmundo", Pos{1, 0}, "k", "hola\nmundo", Pos{0, 0}, ModeNormal},
		{"j no pasa de la ultima linea", "hola\nmundo", Pos{1, 0}, "j", "hola\nmundo", Pos{1, 0}, ModeNormal},
		{"flechas se comportan como hjkl", "hola", Pos{0, 0}, "<Right><Right>", "hola", Pos{0, 2}, ModeNormal},

		{"i entra en insert sin mover el cursor", "hola", Pos{0, 0}, "i", "hola", Pos{0, 0}, ModeInsert},
		{"a entra en insert despues del caracter", "hola", Pos{0, 0}, "a", "hola", Pos{0, 1}, ModeInsert},
		{"a en linea vacia no avanza", "", Pos{0, 0}, "a", "", Pos{0, 0}, ModeInsert},
		{"I va al primer no blanco", "  hola", Pos{0, 4}, "I", "  hola", Pos{0, 2}, ModeInsert},
		{"A va al final de la linea", "hola", Pos{0, 0}, "A", "hola", Pos{0, 4}, ModeInsert},
		{"o abre linea debajo y entra en insert", "uno\ndos", Pos{0, 0}, "o", "uno\n\ndos", Pos{1, 0}, ModeInsert},
		{"O abre linea arriba y entra en insert", "uno\ndos", Pos{1, 0}, "O", "uno\n\ndos", Pos{1, 0}, ModeInsert},

		{"x borra el caracter bajo el cursor", "hola", Pos{0, 0}, "x", "ola", Pos{0, 0}, ModeNormal},
		{"x en el ultimo caracter retrocede el cursor", "hola", Pos{0, 3}, "x", "hol", Pos{0, 2}, ModeNormal},
		{"x en linea vacia no hace nada", "", Pos{0, 0}, "x", "", Pos{0, 0}, ModeNormal},

		{"escribir texto en insert y volver con Esc", "", Pos{0, 0}, "iHola<Esc>", "Hola", Pos{0, 3}, ModeNormal},
		{"Esc retrocede el cursor una columna", "hola", Pos{0, 0}, "aXY<Esc>", "hXYola", Pos{0, 2}, ModeNormal},
		{"Enter en insert divide la linea", "holamundo", Pos{0, 4}, "i<CR>", "hola\nmundo", Pos{1, 0}, ModeInsert},
		{"Backspace borra el caracter anterior", "hola", Pos{0, 2}, "i<BS>", "hla", Pos{0, 1}, ModeInsert},
		{"Backspace en columna 0 no hace nada", "hola", Pos{0, 0}, "i<BS>", "hola", Pos{0, 0}, ModeInsert},
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

func TestEngineCommandModeWrite(t *testing.T) {
	e := New("hola", Pos{0, 0})
	e.Feed(RuneKey(':'))
	e.Feed(RuneKey('w'))
	if got := e.PendingKeys(); got != ":w" {
		t.Fatalf("PendingKeys() = %q, want %q", got, ":w")
	}
	if got := e.Mode(); got != ModeCommand {
		t.Fatalf("Mode() = %v, want %v", got, ModeCommand)
	}
	ev := e.Feed(EnterKey())
	if ev.Type != EventWrite {
		t.Fatalf("Event = %v, want EventWrite", ev.Type)
	}
	if got := e.Mode(); got != ModeNormal {
		t.Fatalf("Mode() tras :w = %v, want ModeNormal", got)
	}
}

func TestEngineCommandModeQuit(t *testing.T) {
	e := New("hola", Pos{0, 0})
	for _, k := range Keys(":q") {
		e.Feed(k)
	}
	ev := e.Feed(EnterKey())
	if ev.Type != EventQuit {
		t.Fatalf("Event = %v, want EventQuit", ev.Type)
	}
}

func TestEngineCommandModeEscCancela(t *testing.T) {
	e := New("hola", Pos{0, 0})
	e.Feed(RuneKey(':'))
	e.Feed(RuneKey('w'))
	e.Feed(EscKey())
	if got := e.Mode(); got != ModeNormal {
		t.Fatalf("Mode() = %v, want ModeNormal", got)
	}
	if got := e.PendingKeys(); got != "" {
		t.Fatalf("PendingKeys() = %q, want vacío", got)
	}
}

func TestEngineStats(t *testing.T) {
	e := New("hola mundo", Pos{0, 0})
	e.Feed(RuneKey('l'))
	e.Feed(RightKey())
	e.Feed(RuneKey('x'))

	stats := e.Stats()
	if stats.TotalKeys != 3 {
		t.Fatalf("TotalKeys = %d, want 3", stats.TotalKeys)
	}
	if stats.ArrowKeys != 1 {
		t.Fatalf("ArrowKeys = %d, want 1", stats.ArrowKeys)
	}
	// Right cuenta como flecha en ArrowKeys pero registra el mismo comando
	// lógico "l" que la tecla de letra, así que aquí suma 2.
	if stats.Commands["l"] != 2 {
		t.Fatalf("Commands[l] = %d, want 2", stats.Commands["l"])
	}
	if stats.Commands["x"] != 1 {
		t.Fatalf("Commands[x] = %d, want 1", stats.Commands["x"])
	}
}

func TestEngineStatsCopyIsIndependent(t *testing.T) {
	e := New("hola", Pos{0, 0})
	e.Feed(RuneKey('l'))
	stats := e.Stats()
	stats.Commands["l"] = 999
	if got := e.Stats().Commands["l"]; got != 1 {
		t.Fatalf("mutar la copia de Stats afectó al engine: Commands[l] = %d", got)
	}
}

func TestEngineUnsupportedEvent(t *testing.T) {
	e := New("hola", Pos{0, 0})
	ev := e.Feed(RuneKey('Z')) // 'Z' no es ningún comando de Vim implementado
	if ev.Type != EventUnsupported {
		t.Fatalf("Event = %v, want EventUnsupported", ev.Type)
	}
}
