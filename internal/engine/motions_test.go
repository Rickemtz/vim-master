package engine

import "testing"

func TestMotionWordForward(t *testing.T) {
	b := NewBuffer("el coche, rojo")
	tests := []struct {
		name  string
		start Pos
		count int
		big   bool
		want  Pos
	}{
		{"al inicio de la siguiente palabra", Pos{0, 0}, 1, false, Pos{0, 3}},
		{"la puntuacion es su propia palabra (w)", Pos{0, 3}, 1, false, Pos{0, 8}}, // "coche" -> ","
		{"con W la puntuacion pegada no separa", Pos{0, 3}, 1, true, Pos{0, 10}},   // "coche," -> "rojo"
		{"count repite el motion", Pos{0, 0}, 2, false, Pos{0, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := motionWordForward(b, tt.start, tt.count, tt.big)
			if res.Pos != tt.want {
				t.Errorf("Pos = %+v, want %+v", res.Pos, tt.want)
			}
		})
	}
}

func TestMotionWordForwardLineaVaciaEsPalabra(t *testing.T) {
	b := NewBuffer("hola\n\nmundo")
	res := motionWordForward(b, Pos{0, 0}, 1, false)
	if res.Pos != (Pos{1, 0}) {
		t.Fatalf("Pos = %+v, want {1 0} (la línea vacía cuenta como palabra)", res.Pos)
	}
}

func TestMotionWordBackward(t *testing.T) {
	b := NewBuffer("el coche rojo")
	res := motionWordBackward(b, Pos{0, 9}, 1, false) // sobre 'r' de rojo
	if res.Pos != (Pos{0, 3}) {
		t.Fatalf("Pos = %+v, want {0 3}", res.Pos)
	}
}

func TestMotionWordEnd(t *testing.T) {
	b := NewBuffer("el coche rojo")
	res := motionWordEnd(b, Pos{0, 0}, 1, false)
	if res.Pos != (Pos{0, 1}) { // fin de "el"
		t.Fatalf("Pos = %+v, want {0 1}", res.Pos)
	}
	if res.Kind != MotionInclusive {
		t.Errorf("Kind = %v, want MotionInclusive", res.Kind)
	}
}

func TestMotionLineStartFirstNonBlankEnd(t *testing.T) {
	b := NewBuffer("  hola mundo")
	if res := motionLineStart(Pos{0, 8}); res.Pos != (Pos{0, 0}) {
		t.Errorf("motionLineStart Pos = %+v, want {0 0}", res.Pos)
	}
	if res := motionFirstNonBlank(b, Pos{0, 8}); res.Pos != (Pos{0, 2}) {
		t.Errorf("motionFirstNonBlank Pos = %+v, want {0 2}", res.Pos)
	}
	if res := motionLineEnd(b, Pos{0, 0}, 1); res.Pos != (Pos{0, 11}) {
		t.Errorf("motionLineEnd Pos = %+v, want {0 11}", res.Pos)
	}
}

func TestMotionGotoLine(t *testing.T) {
	b := NewBuffer("uno\ndos\ntres")

	if res := motionGotoLine(b, 0, false); res.Pos.Line != 0 { // gg
		t.Errorf("gg Line = %d, want 0", res.Pos.Line)
	}
	if res := motionGotoLine(b, 0, true); res.Pos.Line != 2 { // G
		t.Errorf("G Line = %d, want 2", res.Pos.Line)
	}
	if res := motionGotoLine(b, 2, false); res.Pos.Line != 1 { // 2G / 2gg
		t.Errorf("2G Line = %d, want 1", res.Pos.Line)
	}
	if res := motionGotoLine(b, 0, false); res.Kind != MotionLinewise {
		t.Errorf("Kind = %v, want MotionLinewise", res.Kind)
	}
}

func TestMotionFindChar(t *testing.T) {
	b := NewBuffer("el coche rojo, veloz")

	res := motionFindChar(b, Pos{0, 0}, 'r', 1, true, false) // f
	if !res.Found || res.Pos != (Pos{0, 9}) {
		t.Fatalf("f r = %+v, %v, want {0 9}, true", res.Pos, res.Found)
	}

	res = motionFindChar(b, Pos{0, 0}, 'r', 1, true, true) // t
	if !res.Found || res.Pos != (Pos{0, 8}) {
		t.Fatalf("t r = %+v, %v, want {0 8}, true", res.Pos, res.Found)
	}

	res = motionFindChar(b, Pos{0, 13}, 'c', 1, false, false) // F: "coche" tiene 'c' en 3 y 5, la más cercana hacia atrás es 5
	if !res.Found || res.Pos != (Pos{0, 5}) {
		t.Fatalf("F c = %+v, %v, want {0 5}, true", res.Pos, res.Found)
	}

	res = motionFindChar(b, Pos{0, 0}, 'z', 1, true, false)
	if !res.Found {
		t.Fatal("f z debería encontrar la 'z' de veloz")
	}

	res = motionFindChar(b, Pos{0, 0}, 'x', 1, true, false)
	if res.Found {
		t.Fatal("f x no debería encontrar nada en esta línea")
	}
}

func TestMotionMatchPair(t *testing.T) {
	b := NewBuffer("foo(bar(baz))")

	res := motionMatchPair(b, Pos{0, 0}) // busca desde el cursor: encuentra "(" en col 3
	if !res.Found || res.Pos != (Pos{0, 12}) {
		t.Fatalf("%% desde col0 = %+v, %v, want {0 12}, true", res.Pos, res.Found)
	}

	res = motionMatchPair(b, Pos{0, 7}) // sobre el "(" de bar(
	if !res.Found || res.Pos != (Pos{0, 11}) {
		t.Fatalf("%% desde col7 = %+v, %v, want {0 11}, true", res.Pos, res.Found)
	}

	res = motionMatchPair(b, Pos{0, 12}) // sobre el ")" final, buscando hacia atrás
	if !res.Found || res.Pos != (Pos{0, 3}) {
		t.Fatalf("%% desde col12 = %+v, %v, want {0 3}, true", res.Pos, res.Found)
	}
}

func TestMotionMatchPairSinCoincidencia(t *testing.T) {
	b := NewBuffer("sin parentesis")
	if res := motionMatchPair(b, Pos{0, 0}); res.Found {
		t.Fatal("no debería encontrar nada sin brackets en la línea")
	}
}
