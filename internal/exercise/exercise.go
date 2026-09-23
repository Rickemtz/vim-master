// Package exercise carga, valida y resuelve los ejercicios YAML de
// Vim Master. No depende de la UI.
package exercise

import (
	"fmt"

	"github.com/Rickemtz/vim-master/internal/engine"
)

type Type string

const (
	TypeEdit   Type = "edit"
	TypeCursor Type = "cursor"
	TypeBoth   Type = "both"
)

// Exercise es un ejercicio ya parseado y validado, listo para jugarse o
// para resolverse con su solución (`make validate`).
type Exercise struct {
	ID            string
	Module        int
	Title         string
	Difficulty    int
	Type          Type
	Instructions  string
	Initial       string
	CursorStart   engine.Pos
	Target        string
	TargetCursor  *engine.Pos
	TimeLimitSec  int
	ParKeystrokes int
	Solution      string
	Allowed       []string
	ForbidArrows  bool
	Hints         []string
}

// Validate comprueba que el ejercicio tiene los campos obligatorios y
// consistentes. Se llama automáticamente al cargarlo con LoadAll.
func (e Exercise) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("falta id")
	}
	if e.Module <= 0 {
		return fmt.Errorf("module debe ser >= 1")
	}
	if e.Difficulty < 1 || e.Difficulty > 5 {
		return fmt.Errorf("difficulty debe estar entre 1 y 5, es %d", e.Difficulty)
	}
	switch e.Type {
	case TypeEdit, TypeCursor, TypeBoth:
	default:
		return fmt.Errorf("type inválido: %q (debe ser edit, cursor o both)", e.Type)
	}
	if e.Solution == "" {
		return fmt.Errorf("falta solution")
	}
	if e.TimeLimitSec <= 0 {
		return fmt.Errorf("time_limit_sec debe ser > 0")
	}
	if e.ParKeystrokes <= 0 {
		return fmt.Errorf("par_keystrokes debe ser > 0")
	}
	if (e.Type == TypeCursor || e.Type == TypeBoth) && e.TargetCursor == nil {
		return fmt.Errorf("target_cursor es obligatorio cuando type=%s", e.Type)
	}
	return nil
}

// SolutionKeyCount devuelve cuántas teclas representa Solution al
// interpretarla en notación Vim.
func (e Exercise) SolutionKeyCount() int {
	return len(engine.Keys(e.Solution))
}
