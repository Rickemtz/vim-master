package storage

import (
	"path/filepath"
	"sort"
	"time"
)

// Modos de juego reconocidos por el ranking. Solo "lecciones" tiene
// contenido jugable por ahora; el resto se usarán en fases posteriores.
const (
	ModeLecciones    = "lecciones"
	ModeContrarreloj = "contrarreloj"
	ModeGolf         = "golf"
	ModeExamen       = "examen"
)

const topN = 10

// ScoreEntry es una entrada del top 10 de un modo de juego.
type ScoreEntry struct {
	Name  string    `json:"name"`
	Score int       `json:"score"`
	Rank  string    `json:"rank"`
	Date  time.Time `json:"date"`
}

// Scores agrupa el top 10 de cada modo de juego.
type Scores struct {
	Modes map[string][]ScoreEntry `json:"modes"`
}

func DefaultScores() Scores {
	return Scores{Modes: map[string][]ScoreEntry{}}
}

func scoresPath(dir string) string { return filepath.Join(dir, "scores.json") }

// LoadScores siempre devuelve un Scores utilizable, igual que LoadProgress.
func (s *Store) LoadScores() (Scores, error) {
	var sc Scores
	err := readJSON(scoresPath(s.dir), &sc)
	switch {
	case err == nil:
		if sc.Modes == nil {
			sc.Modes = map[string][]ScoreEntry{}
		}
		return sc, nil
	case isNotExist(err):
		return DefaultScores(), nil
	default:
		return DefaultScores(), err
	}
}

// SaveScores escribe sc de forma atómica.
func (s *Store) SaveScores(sc Scores) error {
	return writeJSONAtomic(scoresPath(s.dir), sc)
}

// AddEntry inserta entry en el top de mode, reordena de mayor a menor
// puntaje y recorta a los primeros topN. Devuelve true si entry quedó en
// el primer lugar (nuevo récord de ese modo).
func (sc *Scores) AddEntry(mode string, entry ScoreEntry) (isRecord bool) {
	if sc.Modes == nil {
		sc.Modes = map[string][]ScoreEntry{}
	}
	list := sc.Modes[mode]
	isRecord = len(list) == 0 || entry.Score > list[0].Score

	list = append(list, entry)
	sort.SliceStable(list, func(i, j int) bool { return list[i].Score > list[j].Score })
	if len(list) > topN {
		list = list[:topN]
	}
	sc.Modes[mode] = list
	return isRecord
}

// Top devuelve el top 10 actual de mode (posiblemente vacío).
func (sc Scores) Top(mode string) []ScoreEntry {
	return sc.Modes[mode]
}
