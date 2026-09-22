package scoring

import "testing"

func TestAccumulatorFlujoBasico(t *testing.T) {
	a := NewAccumulator()

	// Ejercicio 1: perfecto (máximo teórico, combo sube a 1.1).
	a.Add(ExerciseInput{
		Difficulty: 1, TimeLimitSec: 30, TimeUsedSec: 0,
		ParKeystrokes: 4, KeystrokesUsed: 4,
	}, map[string]int{"l": 4})

	// Ejercicio 2: se agota el tiempo, rompe el combo.
	a.Add(ExerciseInput{
		Difficulty: 1, TimeLimitSec: 20, TimedOut: true,
	}, nil)

	sum := a.Summary()

	if sum.Attempted != 2 {
		t.Fatalf("Attempted = %d, want 2", sum.Attempted)
	}
	if sum.Completed != 1 {
		t.Fatalf("Completed = %d, want 1", sum.Completed)
	}
	if sum.MaxTheoretical != 400 { // 2 ejercicios de dificultad 1 -> 200 c/u
		t.Fatalf("MaxTheoretical = %d, want 400", sum.MaxTheoretical)
	}
	if sum.TotalPoints != 200 { // ej1 al máximo (200), ej2 vale 0
		t.Fatalf("TotalPoints = %d, want 200", sum.TotalPoints)
	}
	if sum.Percentage != 50 {
		t.Fatalf("Percentage = %v, want 50", sum.Percentage)
	}
	if sum.Rank != RankC {
		t.Fatalf("Rank = %v, want C", sum.Rank)
	}
	if sum.MaxCombo != 1.1 {
		t.Fatalf("MaxCombo = %v, want 1.1 (se alcanzó en el ejercicio 1)", sum.MaxCombo)
	}
	if sum.Commands["l"] != 4 {
		t.Fatalf("Commands[l] = %d, want 4", sum.Commands["l"])
	}
}

func TestAccumulatorSummaryVacio(t *testing.T) {
	a := NewAccumulator()
	sum := a.Summary()
	if sum.Percentage != 0 || sum.Accuracy != 0 || sum.Rank != RankD {
		t.Fatalf("Summary() de un acumulador vacío = %+v, want ceros y rango D", sum)
	}
}

func TestAccumulatorCommandsCopyIsIndependent(t *testing.T) {
	a := NewAccumulator()
	a.Add(ExerciseInput{Difficulty: 1, TimeLimitSec: 10, ParKeystrokes: 1, KeystrokesUsed: 1}, map[string]int{"x": 1})
	sum := a.Summary()
	sum.Commands["x"] = 999
	if got := a.Summary().Commands["x"]; got != 1 {
		t.Fatalf("mutar el resumen afectó al acumulador: Commands[x] = %d", got)
	}
}
