package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kyrcovarick/vimdojo/internal/ui"
)

// version/commit/date se rellenan en tiempo de build vía -ldflags (ver
// .goreleaser.yaml); "dev" es el valor para `go run`/`make build` locales.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("vimdojo %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	p := tea.NewProgram(ui.NewApp())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
