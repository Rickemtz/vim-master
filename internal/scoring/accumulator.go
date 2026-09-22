package scoring

// Accumulator agrega los resultados de los ejercicios sucesivos de una
// partida y lleva el combo de uno al siguiente.
type Accumulator struct {
	combo          float64
	maxCombo       float64
	totalPoints    int
	maxTheoretical int
	completed      int
	attempted      int
	totalTimeSec   float64
	totalKeys      int
	totalPar       int
	commands       map[string]int
}

func NewAccumulator() *Accumulator {
	return &Accumulator{combo: 1.0, maxCombo: 1.0, commands: map[string]int{}}
}

// Add puntúa un ejercicio con el combo actual del acumulador, actualiza el
// combo y suma las estadísticas agregadas. commands son los comandos
// usados en ese ejercicio (para el resumen de "más usados"); no afectan
// el cálculo de puntos.
func (a *Accumulator) Add(in ExerciseInput, commands map[string]int) ExerciseResult {
	in.Combo = a.combo
	res := ScoreExercise(in)

	a.combo = res.NewCombo
	if a.combo > a.maxCombo {
		a.maxCombo = a.combo
	}
	a.totalPoints += res.Points
	a.maxTheoretical += res.Base * 2
	a.attempted++
	if !in.TimedOut && !in.Skipped {
		a.completed++
	}
	a.totalTimeSec += in.TimeUsedSec
	a.totalKeys += in.KeystrokesUsed
	a.totalPar += in.ParKeystrokes
	for name, count := range commands {
		a.commands[name] += count
	}

	return res
}

// Combo devuelve el multiplicador de combo actual (para mostrarlo en el
// HUD durante la partida).
func (a *Accumulator) Combo() float64 { return a.combo }

// Summary es el resumen final de la partida, para la pantalla de
// resultados.
type Summary struct {
	TotalPoints     int
	MaxTheoretical  int
	Percentage      float64
	Rank            Rank
	TotalTimeSec    float64
	TotalKeystrokes int
	TotalPar        int
	Completed       int
	Attempted       int
	Accuracy        float64 // completados / intentados × 100
	MaxCombo        float64
	Commands        map[string]int
}

func (a *Accumulator) Summary() Summary {
	pct := 0.0
	if a.maxTheoretical > 0 {
		pct = float64(a.totalPoints) / float64(a.maxTheoretical) * 100
	}
	acc := 0.0
	if a.attempted > 0 {
		acc = float64(a.completed) / float64(a.attempted) * 100
	}

	commands := make(map[string]int, len(a.commands))
	for k, v := range a.commands {
		commands[k] = v
	}

	return Summary{
		TotalPoints:     a.totalPoints,
		MaxTheoretical:  a.maxTheoretical,
		Percentage:      pct,
		Rank:            RankFor(pct),
		TotalTimeSec:    a.totalTimeSec,
		TotalKeystrokes: a.totalKeys,
		TotalPar:        a.totalPar,
		Completed:       a.completed,
		Attempted:       a.attempted,
		Accuracy:        acc,
		MaxCombo:        a.maxCombo,
		Commands:        commands,
	}
}
