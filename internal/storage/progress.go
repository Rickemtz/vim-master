package storage

import "path/filepath"

// Progress es el avance del jugador entre sesiones: módulos desbloqueados,
// mejor rango alcanzado por módulo y qué ejercicios completó.
type Progress struct {
	UnlockedModules []int           `json:"unlocked_modules"`
	BestRank        map[int]string  `json:"best_rank"`
	Completed       map[string]bool `json:"completed_exercises"`
}

// DefaultProgress es el estado inicial de un jugador nuevo: solo el
// módulo 1 desbloqueado.
func DefaultProgress() Progress {
	return Progress{
		UnlockedModules: []int{1},
		BestRank:        map[int]string{},
		Completed:       map[string]bool{},
	}
}

func progressPath(dir string) string { return filepath.Join(dir, "progress.json") }

// LoadProgress siempre devuelve un Progress utilizable: si el archivo no
// existe, el valor por defecto; si existe pero está dañado, lo respalda
// como .bak, devuelve el valor por defecto y el error (para que el
// llamador pueda registrarlo, sin que sea obligatorio tratarlo).
func (s *Store) LoadProgress() (Progress, error) {
	var p Progress
	err := readJSON(progressPath(s.dir), &p)
	switch {
	case err == nil:
		p.normalize()
		return p, nil
	case isNotExist(err):
		return DefaultProgress(), nil
	default:
		return DefaultProgress(), err
	}
}

// SaveProgress escribe p de forma atómica.
func (s *Store) SaveProgress(p Progress) error {
	return writeJSONAtomic(progressPath(s.dir), p)
}

func (p *Progress) normalize() {
	if p.BestRank == nil {
		p.BestRank = map[int]string{}
	}
	if p.Completed == nil {
		p.Completed = map[string]bool{}
	}
	if len(p.UnlockedModules) == 0 {
		p.UnlockedModules = []int{1}
	}
}

// IsUnlocked indica si el jugador puede jugar ese módulo.
func (p Progress) IsUnlocked(module int) bool {
	for _, m := range p.UnlockedModules {
		if m == module {
			return true
		}
	}
	return false
}

// Unlock añade module a la lista de desbloqueados si no estaba ya.
func (p *Progress) Unlock(module int) {
	if p.IsUnlocked(module) {
		return
	}
	p.UnlockedModules = append(p.UnlockedModules, module)
}

// MarkCompleted registra un ejercicio como completado alguna vez.
func (p *Progress) MarkCompleted(exerciseID string) {
	if p.Completed == nil {
		p.Completed = map[string]bool{}
	}
	p.Completed[exerciseID] = true
}

var rankOrder = map[string]int{"S": 5, "A": 4, "B": 3, "C": 2, "D": 1}

// RecordModuleResult guarda rank como mejor rango de module si supera al
// anterior, y desbloquea el módulo siguiente cuando rank es C o mejor
// ("Un módulo se desbloquea al completar el anterior con rango C o
// superior"). Devuelve true si rank mejoró el mejor rango previo de ese
// módulo.
func (p *Progress) RecordModuleResult(module int, rank string) (improved bool) {
	if p.BestRank == nil {
		p.BestRank = map[int]string{}
	}
	prev := p.BestRank[module]
	if rankOrder[rank] > rankOrder[prev] {
		p.BestRank[module] = rank
		improved = true
	}
	if rankOrder[rank] >= rankOrder["C"] {
		p.Unlock(module + 1)
	}
	return improved
}
