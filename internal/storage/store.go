// Package storage lee y escribe el progreso y los récords de Vim Master en
// ~/.config/vim-master/ (JSON). Escribe en un archivo temporal y hace rename
// para no corromper datos a medio escribir; si un archivo existente está
// dañado, lo respalda como .bak y sigue con valores por defecto.
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Store apunta al directorio de configuración donde viven progress.json y
// scores.json.
type Store struct {
	dir string
}

// New crea (si hace falta) y devuelve el Store en el directorio de
// configuración estándar del sistema operativo.
func New() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("obtener directorio de configuración: %w", err)
	}
	return NewAt(filepath.Join(base, "vim-master"))
}

// NewAt crea un Store en un directorio explícito (usado en tests, o para
// una ubicación alternativa).
func NewAt(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("crear directorio de configuración %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// writeJSONAtomic serializa v y lo escribe en path mediante un archivo
// temporal + rename, para que un fallo a mitad de escritura nunca deje un
// JSON corrupto en su lugar final.
func writeJSONAtomic(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar %s: %w", path, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("escribir %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("renombrar %s a %s: %w", tmp, path, err)
	}
	return nil
}

// readJSON lee y decodifica path en v. Si el archivo no existe, devuelve
// os.ErrNotExist tal cual para que el llamador use sus valores por
// defecto. Si existe pero el JSON está dañado, lo respalda como .bak y
// devuelve el error de decodificación.
func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err // incluye os.ErrNotExist, sin envolver, para poder usar errors.Is
	}

	if err := json.Unmarshal(data, v); err != nil {
		backup := path + ".bak"
		if werr := os.WriteFile(backup, data, 0o644); werr != nil {
			return fmt.Errorf("json dañado en %s (no se pudo respaldar: %v): %w", path, werr, err)
		}
		return fmt.Errorf("json dañado en %s (respaldado en %s): %w", path, backup, err)
	}
	return nil
}

func isNotExist(err error) bool { return errors.Is(err, os.ErrNotExist) }
