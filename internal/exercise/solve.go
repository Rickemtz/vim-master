package exercise

import (
	"fmt"

	"github.com/kyrcovarick/vimdojo/internal/engine"
)

// Result es el resultado de resolver un ejercicio con su Solution usando
// el engine real (usado por `make validate`).
type Result struct {
	OK        bool
	GotText   string
	GotCursor engine.Pos
	Reason    string
}

// Solve alimenta ex.Solution al engine partiendo de ex.Initial/ex.CursorStart
// y comprueba que el resultado alcanza ex.Target (y ex.TargetCursor, si el
// tipo del ejercicio lo exige).
func Solve(ex Exercise) Result {
	e := engine.New(ex.Initial, ex.CursorStart)
	for _, k := range engine.Keys(ex.Solution) {
		e.Feed(k)
	}

	res := Result{GotText: e.Text(), GotCursor: e.Cursor(), OK: true}

	if ex.Type == TypeEdit || ex.Type == TypeBoth {
		if res.GotText != ex.Target {
			res.OK = false
			res.Reason = fmt.Sprintf("texto final %q, se esperaba %q", res.GotText, ex.Target)
			return res
		}
	}

	if ex.Type == TypeCursor || ex.Type == TypeBoth {
		if ex.TargetCursor != nil && res.GotCursor != *ex.TargetCursor {
			res.OK = false
			res.Reason = fmt.Sprintf("cursor final %+v, se esperaba %+v", res.GotCursor, *ex.TargetCursor)
			return res
		}
	}

	return res
}
