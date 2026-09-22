package exercise

import (
	"strings"
	"testing"
	"testing/fstest"
)

const validYAML = `
id: supervivencia-001
module: 1
title: "Primer paso"
difficulty: 1
type: edit
instructions: |
  Borra la letra "x" con el comando x.
initial: |
  xhola
cursor_start: [0, 0]
target: |
  hola
target_cursor: null
time_limit_sec: 20
par_keystrokes: 1
solution: "x"
allowed: [x]
forbid_arrows: true
hints:
  - "x borra el carácter bajo el cursor"
`

func TestParseValid(t *testing.T) {
	ex, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if ex.ID != "supervivencia-001" {
		t.Errorf("ID = %q, want supervivencia-001", ex.ID)
	}
	if ex.Initial != "xhola" {
		t.Errorf("Initial = %q, want %q (sin salto de línea final)", ex.Initial, "xhola")
	}
	if ex.Target != "hola" {
		t.Errorf("Target = %q, want %q", ex.Target, "hola")
	}
	if ex.CursorStart.Line != 0 || ex.CursorStart.Col != 0 {
		t.Errorf("CursorStart = %+v, want {0 0}", ex.CursorStart)
	}
	if ex.TargetCursor != nil {
		t.Errorf("TargetCursor = %+v, want nil", ex.TargetCursor)
	}
	if ex.Type != TypeEdit {
		t.Errorf("Type = %q, want edit", ex.Type)
	}
}

func TestParseMissingSolution(t *testing.T) {
	bad := strings.Replace(validYAML, `solution: "x"`, `solution: ""`, 1)
	if _, err := Parse([]byte(bad)); err == nil {
		t.Fatal("Parse() no devolvió error con solution vacío")
	}
}

func TestParseInvalidType(t *testing.T) {
	bad := strings.Replace(validYAML, "type: edit", "type: invalido", 1)
	if _, err := Parse([]byte(bad)); err == nil {
		t.Fatal("Parse() no devolvió error con type inválido")
	}
}

func TestParseCursorTypeRequiresTargetCursor(t *testing.T) {
	bad := strings.Replace(validYAML, "type: edit", "type: cursor", 1)
	if _, err := Parse([]byte(bad)); err == nil {
		t.Fatal("Parse() no devolvió error cuando type=cursor no trae target_cursor")
	}
}

func TestLoadAll(t *testing.T) {
	fsys := fstest.MapFS{
		"01-supervivencia/001.yaml": &fstest.MapFile{Data: []byte(validYAML)},
		"01-supervivencia/002.yaml": &fstest.MapFile{Data: []byte(strings.Replace(
			validYAML, "supervivencia-001", "supervivencia-002", 1))},
		"README.md": &fstest.MapFile{Data: []byte("no es un ejercicio")},
	}

	exs, err := LoadAll(fsys)
	if err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}
	if len(exs) != 2 {
		t.Fatalf("LoadAll() devolvió %d ejercicios, want 2", len(exs))
	}
	if exs[0].ID != "supervivencia-001" || exs[1].ID != "supervivencia-002" {
		t.Errorf("orden inesperado: %q, %q", exs[0].ID, exs[1].ID)
	}
}

func TestLoadAllPropagaErrorDeParse(t *testing.T) {
	fsys := fstest.MapFS{
		"01-supervivencia/bad.yaml": &fstest.MapFile{Data: []byte("id: solo-id\n")},
	}
	if _, err := LoadAll(fsys); err == nil {
		t.Fatal("LoadAll() no devolvió error con un ejercicio inválido")
	}
}
