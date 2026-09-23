package session

import (
	"testing"

	"github.com/Rickemtz/vim-master/internal/exercise"
)

func TestSessionFlow(t *testing.T) {
	exs := []exercise.Exercise{{ID: "a"}, {ID: "b"}}
	s := New(exs)

	ex, ok := s.Current()
	if !ok || ex.ID != "a" {
		t.Fatalf("Current() = %+v, %v, want a, true", ex, ok)
	}
	if cur, total := s.Progress(); cur != 1 || total != 2 {
		t.Fatalf("Progress() = %d/%d, want 1/2", cur, total)
	}
	if s.Done() {
		t.Fatal("Done() = true, want false")
	}

	s.Advance()
	ex, ok = s.Current()
	if !ok || ex.ID != "b" {
		t.Fatalf("Current() tras Advance = %+v, %v, want b, true", ex, ok)
	}

	s.Advance()
	if !s.Done() {
		t.Fatal("Done() = false tras agotar los ejercicios, want true")
	}
	if _, ok := s.Current(); ok {
		t.Fatal("Current() debería devolver ok=false cuando la sesión terminó")
	}
}

func TestSessionShuffledSinLoopTermina(t *testing.T) {
	exs := []exercise.Exercise{{ID: "a"}, {ID: "b"}, {ID: "c"}}
	s := NewShuffled(exs, false)

	seen := map[string]bool{}
	for i := 0; i < 3; i++ {
		ex, ok := s.Current()
		if !ok {
			t.Fatalf("Current() en la posición %d = ok=false, want true", i)
		}
		seen[ex.ID] = true
		s.Advance()
	}
	if !s.Done() {
		t.Fatal("Done() = false tras agotar los 3 ejercicios, want true")
	}
	if len(seen) != 3 {
		t.Fatalf("se vieron %d ejercicios distintos, want 3 (sin repetir)", len(seen))
	}
}

func TestSessionShuffledConLoopNuncaTermina(t *testing.T) {
	exs := []exercise.Exercise{{ID: "a"}, {ID: "b"}}
	s := NewShuffled(exs, true)

	for i := 0; i < 10; i++ {
		if s.Done() {
			t.Fatalf("Done() = true en la iteración %d, una sesión loop nunca debería terminar", i)
		}
		if _, ok := s.Current(); !ok {
			t.Fatalf("Current() en la iteración %d = ok=false, want true", i)
		}
		s.Advance()
	}
}

func TestSessionShuffledVacia(t *testing.T) {
	s := NewShuffled(nil, true)
	if _, ok := s.Current(); ok {
		t.Fatal("Current() en una sesión vacía debería devolver ok=false, incluso con loop")
	}
}
