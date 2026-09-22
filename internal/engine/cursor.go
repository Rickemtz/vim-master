package engine

// Pos es una posición en el buffer, base 0 (línea, columna en runas).
type Pos struct {
	Line int
	Col  int
}

// ClampCol ajusta col al rango válido de una línea de longitud lineLen
// según el modo: en Normal (y Visual) el cursor nunca se posiciona una
// columna después del último carácter, salvo que la línea esté vacía; en
// Insert y Command puede llegar una posición más allá para permitir
// añadir al final.
func ClampCol(lineLen int, col int, mode Mode) int {
	max := lineLen
	if mode != ModeInsert && mode != ModeCommand {
		max = lineLen - 1
		if max < 0 {
			max = 0
		}
	}
	return clampInt(col, 0, max)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
