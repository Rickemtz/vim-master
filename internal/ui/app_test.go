package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Tests mínimos de Update() enviando mensajes, sin comparar el render
// píxel a píxel (solo que el texto esperado aparezca donde corresponde).

func TestAppWindowSizeMsgMuestraAviso(t *testing.T) {
	a := NewApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	app := m.(App)
	if !strings.Contains(app.View(), "Terminal demasiado pequeño") {
		t.Fatalf("View() con terminal chico no muestra el aviso: %q", app.View())
	}
}

func TestAppWindowSizeMsgSuficienteNoMuestraAviso(t *testing.T) {
	a := NewApp()
	m, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	app := m.(App)
	if strings.Contains(app.View(), "Terminal demasiado pequeño") {
		t.Fatalf("View() con terminal suficiente no debería mostrar el aviso")
	}
}

func TestAppSinWindowSizeMsgAunNoMuestraAviso(t *testing.T) {
	// Antes del primer WindowSizeMsg (width==0) no debe parpadear el
	// aviso al arrancar.
	a := NewApp()
	if strings.Contains(a.View(), "Terminal demasiado pequeño") {
		t.Fatal("View() antes de recibir WindowSizeMsg no debería mostrar el aviso")
	}
}

func TestAppMenuNavegacion(t *testing.T) {
	a := NewApp()
	if a.menu.cursor != 0 {
		t.Fatalf("cursor inicial = %d, want 0", a.menu.cursor)
	}
	m, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown})
	app := m.(App)
	if app.menu.cursor != 1 {
		t.Fatalf("cursor tras KeyDown = %d, want 1", app.menu.cursor)
	}
}

func TestAppQuitDesdeElMenu(t *testing.T) {
	a := NewApp()
	_, cmd := a.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("q en el menú debería devolver un comando (tea.Quit)")
	}
}

func TestAppEnterEnLeccionesArrancaLaLeccion(t *testing.T) {
	a := NewApp()
	m, cmd := a.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app := m.(App)
	if app.screen != screenLesson {
		t.Fatalf("screen = %v, want screenLesson", app.screen)
	}
	if cmd == nil {
		t.Fatal("entrar a Lecciones debería devolver el Cmd de Init() del timer")
	}
}

func TestAppContrarrelojVaAPantallaDeDuracion(t *testing.T) {
	a := NewApp()
	m, _ := a.Update(tea.KeyMsg{Type: tea.KeyDown}) // cursor -> Contrarreloj
	app := m.(App)
	m2, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app2 := m2.(App)
	if app2.screen != screenDuration {
		t.Fatalf("screen = %v, want screenDuration", app2.screen)
	}
}

func TestAppF10DesdeDuracionVuelveAlMenu(t *testing.T) {
	a := NewApp()
	a.screen = screenDuration
	m, _ := a.Update(tea.KeyMsg{Type: tea.KeyF10})
	app := m.(App)
	if app.screen != screenMenu {
		t.Fatalf("screen = %v, want screenMenu", app.screen)
	}
}
