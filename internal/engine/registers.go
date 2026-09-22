package engine

// Register es el contenido de un registro de Vim. Linewise distingue si
// se guardó con un operador linewise (dd/yy/cc) o uno charwise (x, d con
// un motion charwise), lo que cambia cómo se pega con p/P.
//
// Por ahora solo existe el registro sin nombre (E3). Los registros con
// nombre ("a-"z) se añaden en E8.
type Register struct {
	Text     string
	Linewise bool
}
