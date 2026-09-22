package storage

import (
	"testing"
	"time"
)

func TestLoadScoresPorDefecto(t *testing.T) {
	s, _ := NewAt(t.TempDir())
	sc, err := s.LoadScores()
	if err != nil {
		t.Fatalf("LoadScores() error = %v", err)
	}
	if got := sc.Top(ModeLecciones); len(got) != 0 {
		t.Errorf("Top() = %v, want vacío", got)
	}
}

func TestAddEntryOrdenaYRecorta(t *testing.T) {
	sc := DefaultScores()
	scores := []int{50, 90, 10, 200, 70, 30, 60, 40, 20, 80, 100} // 11 entradas
	for _, sVal := range scores {
		sc.AddEntry(ModeLecciones, ScoreEntry{Name: "jugador", Score: sVal, Rank: "B", Date: time.Now()})
	}

	top := sc.Top(ModeLecciones)
	if len(top) != topN {
		t.Fatalf("len(Top()) = %d, want %d", len(top), topN)
	}
	if top[0].Score != 200 {
		t.Errorf("top[0].Score = %d, want 200", top[0].Score)
	}
	for i := 1; i < len(top); i++ {
		if top[i-1].Score < top[i].Score {
			t.Fatalf("Top() no está ordenado de mayor a menor: %v", top)
		}
	}
	// La entrada más baja (10) debería haber quedado fuera al recortar a 10.
	for _, e := range top {
		if e.Score == 10 {
			t.Error("la puntuación más baja debería haberse recortado del top 10")
		}
	}
}

func TestAddEntryNuevoRecord(t *testing.T) {
	sc := DefaultScores()
	if isRecord := sc.AddEntry(ModeLecciones, ScoreEntry{Score: 100}); !isRecord {
		t.Error("la primera entrada siempre es un nuevo récord")
	}
	if isRecord := sc.AddEntry(ModeLecciones, ScoreEntry{Score: 50}); isRecord {
		t.Error("una puntuación menor no debería ser un nuevo récord")
	}
	if isRecord := sc.AddEntry(ModeLecciones, ScoreEntry{Score: 150}); !isRecord {
		t.Error("superar el máximo anterior debería ser un nuevo récord")
	}
}

func TestSaveLoadScoresRoundtrip(t *testing.T) {
	s, _ := NewAt(t.TempDir())
	sc := DefaultScores()
	date := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	sc.AddEntry(ModeLecciones, ScoreEntry{Name: "kyr", Score: 1234, Rank: "S", Date: date})

	if err := s.SaveScores(sc); err != nil {
		t.Fatalf("SaveScores() error = %v", err)
	}
	got, err := s.LoadScores()
	if err != nil {
		t.Fatalf("LoadScores() error = %v", err)
	}
	top := got.Top(ModeLecciones)
	if len(top) != 1 || top[0].Name != "kyr" || top[0].Score != 1234 || top[0].Rank != "S" {
		t.Fatalf("Top() = %+v, no coincide con lo guardado", top)
	}
	if !top[0].Date.Equal(date) {
		t.Errorf("Date = %v, want %v", top[0].Date, date)
	}
}
