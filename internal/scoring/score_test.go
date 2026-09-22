package scoring

import "testing"

func TestScoreExerciseMaximoTeorico(t *testing.T) {
	// tiempo 0 (instantáneo) y teclas == par: bonus_tiempo y
	// bonus_eficiencia al máximo -> subtotal == base*2.
	in := ExerciseInput{
		Difficulty:     2,
		TimeLimitSec:   30,
		TimeUsedSec:    0,
		ParKeystrokes:  4,
		KeystrokesUsed: 4,
		Combo:          1.0,
	}
	res := ScoreExercise(in)
	if res.Base != 200 {
		t.Fatalf("Base = %d, want 200", res.Base)
	}
	if res.Subtotal != 400 {
		t.Fatalf("Subtotal = %v, want 400 (base*2)", res.Subtotal)
	}
	if res.Points != 400 {
		t.Fatalf("Points = %d, want 400", res.Points)
	}
}

func TestScoreExerciseTeclasMenoresQuePar(t *testing.T) {
	// Usar menos teclas que par no debe dar más del bonus máximo
	// (min(1, par/usadas) se satura en 1).
	in := ExerciseInput{
		Difficulty:     1,
		TimeLimitSec:   30,
		TimeUsedSec:    0,
		ParKeystrokes:  4,
		KeystrokesUsed: 2,
		Combo:          1.0,
	}
	res := ScoreExercise(in)
	if res.BonusEficiencia != 50 { // base(100) * 0.5 * 1
		t.Fatalf("BonusEficiencia = %v, want 50", res.BonusEficiencia)
	}
}

func TestScoreExerciseTiempoAgotado(t *testing.T) {
	in := ExerciseInput{Difficulty: 3, TimedOut: true, Combo: 1.7}
	res := ScoreExercise(in)
	if res.Points != 0 {
		t.Fatalf("Points = %d, want 0 cuando se agota el tiempo", res.Points)
	}
	if !res.ComboBroken {
		t.Fatal("ComboBroken = false, want true cuando se agota el tiempo")
	}
	if res.NewCombo != 1.0 {
		t.Fatalf("NewCombo = %v, want 1.0 tras agotar el tiempo", res.NewCombo)
	}
}

func TestScoreExerciseSkipped(t *testing.T) {
	in := ExerciseInput{Difficulty: 1, Skipped: true, Combo: 1.5}
	res := ScoreExercise(in)
	if res.Points != 0 || res.NewCombo != 1.0 || !res.ComboBroken {
		t.Fatalf("resultado inesperado al saltar ejercicio: %+v", res)
	}
}

func TestScoreExercisePenalizacionesNoBajanDeCero(t *testing.T) {
	in := ExerciseInput{
		Difficulty:     1, // base 100
		TimeLimitSec:   30,
		TimeUsedSec:    30,
		ParKeystrokes:  2,
		KeystrokesUsed: 20,
		HintsUsed:      5, // 5 * 100 * 0.15 = 75
		ForbidArrows:   true,
		ArrowKeys:      10, // 10*5 = 50
		ForbiddenCmds:  5,  // 5*10 = 50
		Combo:          1.0,
	}
	res := ScoreExercise(in)
	if res.Subtotal != 0 {
		t.Fatalf("Subtotal = %v, want 0 (penalizaciones superan la base+bonus)", res.Subtotal)
	}
	if res.Points != 0 {
		t.Fatalf("Points = %d, want 0", res.Points)
	}
}

func TestScoreExercisePistaRompeComboPeroPuntuaAlgo(t *testing.T) {
	in := ExerciseInput{
		Difficulty:     1,
		TimeLimitSec:   30,
		TimeUsedSec:    10,
		ParKeystrokes:  4,
		KeystrokesUsed: 4,
		HintsUsed:      1,
		Combo:          1.5,
	}
	res := ScoreExercise(in)
	if res.NewCombo != 1.0 || !res.ComboBroken {
		t.Fatalf("usar una pista debería resetear el combo: %+v", res)
	}
	if res.Points <= 0 {
		t.Fatalf("Points = %d, debería seguir puntuando algo aunque use pista", res.Points)
	}
}

func TestScoreExerciseComboSubeYTopeEnDos(t *testing.T) {
	in := ExerciseInput{
		Difficulty:     1,
		TimeLimitSec:   30,
		TimeUsedSec:    10,
		ParKeystrokes:  4,
		KeystrokesUsed: 4, // <= par*1.2
		Combo:          1.95,
	}
	res := ScoreExercise(in)
	if got := round1(res.NewCombo); got != 2.0 {
		t.Fatalf("NewCombo = %v, want 2.0 (tope), sumando 0.1 a 1.95", got)
	}

	in.Combo = 2.0
	res = ScoreExercise(in)
	if res.NewCombo != 2.0 {
		t.Fatalf("NewCombo = %v, want 2.0, el combo no debe superar el tope", res.NewCombo)
	}
}

func TestScoreExerciseComboNoSubeSiSuperaMargen(t *testing.T) {
	// teclas_usadas > par*1.2: no sube el combo, pero tampoco se rompe.
	in := ExerciseInput{
		Difficulty:     1,
		TimeLimitSec:   30,
		TimeUsedSec:    10,
		ParKeystrokes:  4,
		KeystrokesUsed: 10, // > 4*1.2
		Combo:          1.3,
	}
	res := ScoreExercise(in)
	if res.NewCombo != 1.3 {
		t.Fatalf("NewCombo = %v, want 1.3 sin cambios", res.NewCombo)
	}
	if res.ComboBroken {
		t.Fatal("ComboBroken = true, no debería romperse solo por exceder el margen de teclas")
	}
}

func TestScoreExerciseComboPorDefectoUno(t *testing.T) {
	// Combo == 0 (valor zero de Go) se trata como 1.0, no como
	// "multiplicar todo por cero". Tiempo agotado al límite y 0 teclas
	// usadas anulan ambos bonus, así que subtotal == base.
	in := ExerciseInput{
		Difficulty:     1,
		TimeLimitSec:   30,
		TimeUsedSec:    30,
		ParKeystrokes:  4,
		KeystrokesUsed: 0,
	}
	res := ScoreExercise(in)
	if res.Points != res.Base {
		t.Fatalf("Points = %d, want %d con combo por defecto 1.0", res.Points, res.Base)
	}
}

func round1(f float64) float64 {
	return float64(int(f*10+0.5)) / 10
}
