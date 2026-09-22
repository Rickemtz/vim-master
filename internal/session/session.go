// Package session orquesta una partida: la lista de ejercicios a jugar y
// en cuál va el jugador. No depende de la UI ni del engine directamente;
// solo conoce el tipo exercise.Exercise.
package session

import (
	"math/rand"

	"github.com/kyrcovarick/vimdojo/internal/exercise"
)

// Session recorre una lista de ejercicios en orden. Con loop=true (usado
// en Contrarreloj) nunca termina: Current/Advance dan la vuelta al llegar
// al final, porque ahí el límite lo pone el reloj global, no la lista.
type Session struct {
	exercises []exercise.Exercise
	index     int
	loop      bool
}

// New crea una sesión secuencial (Lecciones): recorre exs en el orden
// dado y termina al agotarlos.
func New(exs []exercise.Exercise) *Session {
	return &Session{exercises: exs}
}

// NewShuffled crea una sesión con exs en orden aleatorio (Golf y
// Contrarreloj). Con loop=false termina al agotar la lista (Golf); con
// loop=true da la vuelta indefinidamente (Contrarreloj).
func NewShuffled(exs []exercise.Exercise, loop bool) *Session {
	shuffled := make([]exercise.Exercise, len(exs))
	copy(shuffled, exs)
	rand.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
	return &Session{exercises: shuffled, loop: loop}
}

// Current devuelve el ejercicio actual. ok es false si ya no quedan
// ejercicios (nunca ocurre si la sesión es loop).
func (s *Session) Current() (ex exercise.Exercise, ok bool) {
	if len(s.exercises) == 0 {
		return exercise.Exercise{}, false
	}
	if s.loop {
		return s.exercises[s.index%len(s.exercises)], true
	}
	if s.Done() {
		return exercise.Exercise{}, false
	}
	return s.exercises[s.index], true
}

// Done indica si se llegó al final de la lista. Una sesión loop nunca
// está "Done".
func (s *Session) Done() bool {
	if s.loop {
		return false
	}
	return s.index >= len(s.exercises)
}

// Advance pasa al siguiente ejercicio.
func (s *Session) Advance() {
	s.index++
}

// Progress devuelve la posición actual en base 1 y el total (el tamaño
// del pool, no de "la vuelta" en sesiones loop), para mostrar
// "ejercicio X/N" en el HUD.
func (s *Session) Progress() (current, total int) {
	return s.index + 1, len(s.exercises)
}
