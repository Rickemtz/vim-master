package scoring

// Rank es el rango final de una sesión.
type Rank string

const (
	RankS Rank = "S"
	RankA Rank = "A"
	RankB Rank = "B"
	RankC Rank = "C"
	RankD Rank = "D"
)

// RankFor devuelve el rango correspondiente a un porcentaje
// (puntaje_total / máximo_teórico_sin_combo × 100). Con combos el
// porcentaje puede superar 100; sigue siendo S.
func RankFor(percentage float64) Rank {
	switch {
	case percentage >= 95:
		return RankS
	case percentage >= 85:
		return RankA
	case percentage >= 70:
		return RankB
	case percentage >= 50:
		return RankC
	default:
		return RankD
	}
}
