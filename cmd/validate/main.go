// Comando validate: resuelve cada ejercicio en exercises/ con su
// "solution" usando el engine real y comprueba que alcanza el target.
package main

import (
	"fmt"
	"os"

	"github.com/kyrcovarick/vimdojo/exercises"
	"github.com/kyrcovarick/vimdojo/internal/exercise"
)

func main() {
	exs, err := exercise.LoadAll(exercises.FS)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error cargando ejercicios:", err)
		os.Exit(1)
	}
	if len(exs) == 0 {
		fmt.Fprintln(os.Stderr, "no se encontró ningún ejercicio en exercises/")
		os.Exit(1)
	}

	fail := 0
	for _, ex := range exs {
		res := exercise.Solve(ex)
		if !res.OK {
			fail++
			fmt.Printf("FAIL %s (%s): %s\n", ex.ID, ex.Title, res.Reason)
			continue
		}

		if n := ex.SolutionKeyCount(); ex.ParKeystrokes > n {
			fmt.Printf("WARN %s (%s): par_keystrokes=%d es mayor que las %d teclas de la solución\n",
				ex.ID, ex.Title, ex.ParKeystrokes, n)
		}
		fmt.Printf("ok   %s (%s)\n", ex.ID, ex.Title)
	}

	fmt.Printf("\n%d ejercicios, %d fallidos\n", len(exs), fail)
	if fail > 0 {
		os.Exit(1)
	}
}
