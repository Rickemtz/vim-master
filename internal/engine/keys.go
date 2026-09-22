package engine

import "strings"

// Keys convierte una cadena en notación estilo Vim (p.ej. "dw", "iHola<Esc>",
// ":w<CR>") en la secuencia de Key que produciría un usuario tecleándola.
// Se usa en tests, en el campo "solution" de los ejercicios y en
// `make validate`.
func Keys(s string) []Key {
	runes := []rune(s)
	keys := make([]Key, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '<' {
			if end := indexRune(runes[i+1:], '>'); end >= 0 {
				token := string(runes[i+1 : i+1+end])
				if k, ok := specialFromToken(token); ok {
					keys = append(keys, k)
					i += end + 1
					continue
				}
			}
		}
		keys = append(keys, RuneKey(r))
	}
	return keys
}

func indexRune(rs []rune, target rune) int {
	for i, r := range rs {
		if r == target {
			return i
		}
	}
	return -1
}

func specialFromToken(token string) (Key, bool) {
	switch strings.ToLower(token) {
	case "esc":
		return EscKey(), true
	case "cr", "enter":
		return EnterKey(), true
	case "bs", "backspace":
		return BackspaceKey(), true
	case "c-r":
		return CtrlRKey(), true
	case "c-v":
		return CtrlVKey(), true
	case "up":
		return UpKey(), true
	case "down":
		return DownKey(), true
	case "left":
		return LeftKey(), true
	case "right":
		return RightKey(), true
	default:
		return Key{}, false
	}
}
