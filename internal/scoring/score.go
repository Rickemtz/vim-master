// Package scoring calcula los puntos, combos y rangos de Vim Master. Son
// funciones puras: no dependen de la UI ni del engine.
package scoring

import "math"

// ExerciseInput son los datos de un ejercicio ya jugado, listos para
// puntuar.
type ExerciseInput struct {
	Difficulty     int     // 1-5
	TimeLimitSec   int     // time_limit_sec del ejercicio
	TimeUsedSec    float64 // tiempo que tardó el jugador
	ParKeystrokes  int     // par_keystrokes del ejercicio
	KeystrokesUsed int     // teclas realmente usadas
	HintsUsed      int     // veces que pidió pista (F1)
	ForbidArrows   bool    // forbid_arrows del ejercicio
	ArrowKeys      int     // teclas de flecha usadas
	ForbiddenCmds  int     // comandos usados fuera de "allowed"
	TimedOut       bool    // se agotó el tiempo
	Skipped        bool    // el jugador saltó el ejercicio
	Combo          float64 // multiplicador de combo ANTES de este ejercicio
}

// ExerciseResult es el desglose de puntuación de un ejercicio.
type ExerciseResult struct {
	Base            int
	BonusTiempo     float64
	BonusEficiencia float64
	Penalizaciones  float64
	Subtotal        float64
	Points          int     // subtotal * combo, redondeado
	NewCombo        float64 // combo resultante tras este ejercicio
	ComboBroken     bool
}

// ScoreExercise aplica las fórmulas de puntuación por ejercicio descritas
// en CLAUDE.md.
func ScoreExercise(in ExerciseInput) ExerciseResult {
	base := 100 * in.Difficulty

	if in.TimedOut || in.Skipped {
		return ExerciseResult{Base: base, Points: 0, NewCombo: 1.0, ComboBroken: true}
	}

	combo := in.Combo
	if combo <= 0 {
		combo = 1.0
	}

	timeFrac := 0.0
	if in.TimeLimitSec > 0 {
		timeFrac = math.Max(0, 1-in.TimeUsedSec/float64(in.TimeLimitSec))
	}
	bonusTiempo := float64(base) * 0.5 * timeFrac

	effFrac := 0.0
	if in.KeystrokesUsed > 0 {
		effFrac = math.Min(1, float64(in.ParKeystrokes)/float64(in.KeystrokesUsed))
	}
	bonusEficiencia := float64(base) * 0.5 * effFrac

	penal := float64(in.HintsUsed) * float64(base) * 0.15
	if in.ForbidArrows {
		penal += float64(in.ArrowKeys) * 5
	}
	penal += float64(in.ForbiddenCmds) * 10

	subtotal := math.Max(0, float64(base)+bonusTiempo+bonusEficiencia-penal)
	points := subtotal * combo

	newCombo := combo
	comboBroken := false
	switch {
	case in.HintsUsed > 0:
		newCombo = 1.0
		comboBroken = true
	case float64(in.KeystrokesUsed) <= float64(in.ParKeystrokes)*1.2:
		newCombo = math.Min(2.0, combo+0.1)
	}

	return ExerciseResult{
		Base:            base,
		BonusTiempo:     bonusTiempo,
		BonusEficiencia: bonusEficiencia,
		Penalizaciones:  penal,
		Subtotal:        subtotal,
		Points:          int(math.Round(points)),
		NewCombo:        newCombo,
		ComboBroken:     comboBroken,
	}
}
