package exercise

import (
	"testing"

	"github.com/Rickemtz/vim-master/internal/engine"
)

func TestSolveEditOK(t *testing.T) {
	ex, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	res := Solve(ex)
	if !res.OK {
		t.Fatalf("Solve() no pasó: %s", res.Reason)
	}
	if res.GotText != "hola" {
		t.Errorf("GotText = %q, want %q", res.GotText, "hola")
	}
}

func TestSolveEditFalla(t *testing.T) {
	ex, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	ex.Solution = "l" // no borra nada, el target no se alcanza
	res := Solve(ex)
	if res.OK {
		t.Fatal("Solve() debería fallar con una solución que no llega al target")
	}
	if res.Reason == "" {
		t.Error("Reason vacío en un resultado fallido")
	}
}

func TestSolveCursorOK(t *testing.T) {
	ex := Exercise{
		ID:            "cursor-001",
		Module:        1,
		Difficulty:    1,
		Type:          TypeCursor,
		Initial:       "hola mundo",
		CursorStart:   engine.Pos{Line: 0, Col: 0},
		TargetCursor:  &engine.Pos{Line: 0, Col: 2},
		TimeLimitSec:  10,
		ParKeystrokes: 2,
		Solution:      "ll",
	}
	res := Solve(ex)
	if !res.OK {
		t.Fatalf("Solve() no pasó: %s", res.Reason)
	}
}

func TestSolutionKeyCount(t *testing.T) {
	ex, err := Parse([]byte(validYAML))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := ex.SolutionKeyCount(); got != 1 {
		t.Errorf("SolutionKeyCount() = %d, want 1", got)
	}
}
