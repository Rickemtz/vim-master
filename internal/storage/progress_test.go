package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProgressPorDefecto(t *testing.T) {
	s, err := NewAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewAt() error = %v", err)
	}
	p, err := s.LoadProgress()
	if err != nil {
		t.Fatalf("LoadProgress() error = %v", err)
	}
	if !p.IsUnlocked(1) {
		t.Error("el módulo 1 debería estar desbloqueado por defecto")
	}
	if p.IsUnlocked(2) {
		t.Error("el módulo 2 no debería estar desbloqueado por defecto")
	}
}

func TestSaveLoadProgressRoundtrip(t *testing.T) {
	s, _ := NewAt(t.TempDir())
	p := DefaultProgress()
	p.Unlock(2)
	p.BestRank[1] = "A"
	p.MarkCompleted("supervivencia-001")

	if err := s.SaveProgress(p); err != nil {
		t.Fatalf("SaveProgress() error = %v", err)
	}

	got, err := s.LoadProgress()
	if err != nil {
		t.Fatalf("LoadProgress() error = %v", err)
	}
	if !got.IsUnlocked(2) {
		t.Error("módulo 2 debería seguir desbloqueado tras el roundtrip")
	}
	if got.BestRank[1] != "A" {
		t.Errorf("BestRank[1] = %q, want A", got.BestRank[1])
	}
	if !got.Completed["supervivencia-001"] {
		t.Error("supervivencia-001 debería seguir marcado como completado")
	}
}

func TestLoadProgressArchivoDañado(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewAt(dir)

	if err := os.WriteFile(progressPath(dir), []byte("{esto no es json"), 0o644); err != nil {
		t.Fatalf("escribir archivo dañado: %v", err)
	}

	p, err := s.LoadProgress()
	if err == nil {
		t.Fatal("LoadProgress() debería devolver un error con JSON dañado")
	}
	if !p.IsUnlocked(1) {
		t.Error("debería devolver el progreso por defecto ante un JSON dañado")
	}

	if _, err := os.Stat(progressPath(dir) + ".bak"); err != nil {
		t.Errorf("debería haberse creado un respaldo .bak: %v", err)
	}
}

func TestSaveProgressNoDejaArchivoTemporal(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewAt(dir)
	if err := s.SaveProgress(DefaultProgress()); err != nil {
		t.Fatalf("SaveProgress() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "progress.json.tmp")); !os.IsNotExist(err) {
		t.Error("no debería quedar un progress.json.tmp tras una escritura exitosa")
	}
}

func TestRecordModuleResultDesbloqueaConCOsuperior(t *testing.T) {
	p := DefaultProgress()

	if improved := p.RecordModuleResult(1, "D"); !improved {
		t.Error("la primera vez debería contar como mejora")
	}
	if p.IsUnlocked(2) {
		t.Error("rango D no debería desbloquear el módulo siguiente")
	}

	if improved := p.RecordModuleResult(1, "C"); !improved {
		t.Error("C mejora sobre D, debería contar como mejora")
	}
	if !p.IsUnlocked(2) {
		t.Error("rango C debería desbloquear el módulo siguiente")
	}
}

func TestRecordModuleResultNoEmpeoraElMejorRango(t *testing.T) {
	p := DefaultProgress()
	p.RecordModuleResult(1, "A")

	if improved := p.RecordModuleResult(1, "C"); improved {
		t.Error("un rango peor no debería contar como mejora")
	}
	if p.BestRank[1] != "A" {
		t.Errorf("BestRank[1] = %q, want A (no debería bajar)", p.BestRank[1])
	}
}
