package engine

import "testing"

func TestClampCol(t *testing.T) {
	tests := []struct {
		name    string
		lineLen int
		col     int
		mode    Mode
		want    int
	}{
		{"normal dentro de rango", 5, 2, ModeNormal, 2},
		{"normal no pasa del ultimo caracter", 5, 5, ModeNormal, 4},
		{"normal linea vacia", 0, 3, ModeNormal, 0},
		{"normal columna negativa", 5, -1, ModeNormal, 0},
		{"insert permite una columna extra", 5, 5, ModeInsert, 5},
		{"insert linea vacia", 0, 0, ModeInsert, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampCol(tt.lineLen, tt.col, tt.mode); got != tt.want {
				t.Fatalf("ClampCol(%d, %d, %v) = %d, want %d", tt.lineLen, tt.col, tt.mode, got, tt.want)
			}
		})
	}
}
