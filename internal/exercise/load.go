package exercise

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Rickemtz/vim-master/internal/engine"
)

// rawExercise refleja el YAML tal cual, antes de convertir los pares
// [línea, columna] a engine.Pos y validar.
type rawExercise struct {
	ID            string   `yaml:"id"`
	Module        int      `yaml:"module"`
	Title         string   `yaml:"title"`
	Difficulty    int      `yaml:"difficulty"`
	Type          string   `yaml:"type"`
	Instructions  string   `yaml:"instructions"`
	Initial       string   `yaml:"initial"`
	CursorStart   [2]int   `yaml:"cursor_start"`
	Target        string   `yaml:"target"`
	TargetCursor  *[2]int  `yaml:"target_cursor"`
	TimeLimitSec  int      `yaml:"time_limit_sec"`
	ParKeystrokes int      `yaml:"par_keystrokes"`
	Solution      string   `yaml:"solution"`
	Allowed       []string `yaml:"allowed"`
	ForbidArrows  bool     `yaml:"forbid_arrows"`
	Hints         []string `yaml:"hints"`
}

// Parse interpreta el contenido de un archivo YAML de ejercicio y lo
// valida.
func Parse(data []byte) (Exercise, error) {
	var raw rawExercise
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Exercise{}, fmt.Errorf("yaml inválido: %w", err)
	}

	ex := Exercise{
		ID:            raw.ID,
		Module:        raw.Module,
		Title:         raw.Title,
		Difficulty:    raw.Difficulty,
		Type:          Type(raw.Type),
		Instructions:  strings.TrimSuffix(raw.Instructions, "\n"),
		Initial:       strings.TrimSuffix(raw.Initial, "\n"),
		CursorStart:   engine.Pos{Line: raw.CursorStart[0], Col: raw.CursorStart[1]},
		Target:        strings.TrimSuffix(raw.Target, "\n"),
		TimeLimitSec:  raw.TimeLimitSec,
		ParKeystrokes: raw.ParKeystrokes,
		Solution:      raw.Solution,
		Allowed:       raw.Allowed,
		ForbidArrows:  raw.ForbidArrows,
		Hints:         raw.Hints,
	}
	if raw.TargetCursor != nil {
		ex.TargetCursor = &engine.Pos{Line: raw.TargetCursor[0], Col: raw.TargetCursor[1]}
	}

	if err := ex.Validate(); err != nil {
		return Exercise{}, err
	}
	return ex, nil
}

// LoadAll recorre fsys en busca de archivos .yaml, los parsea y valida, y
// devuelve los ejercicios ordenados por módulo y luego por id.
func LoadAll(fsys fs.FS) ([]Exercise, error) {
	var out []Exercise
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("leer %s: %w", path, err)
		}
		ex, err := Parse(data)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		out = append(out, ex)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Module != out[j].Module {
			return out[i].Module < out[j].Module
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}
