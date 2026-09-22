package engine

import "testing"

func TestKeys(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []Key
	}{
		{"vacio", "", nil},
		{"runas simples", "hjkl", []Key{RuneKey('h'), RuneKey('j'), RuneKey('k'), RuneKey('l')}},
		{"especial esc", "i<Esc>", []Key{RuneKey('i'), EscKey()}},
		{"especial cr", ":w<CR>", []Key{RuneKey(':'), RuneKey('w'), EnterKey()}},
		{"especial insensible a mayusculas", "<ESC>", []Key{EscKey()}},
		{"token desconocido se trata como literal", "<x>", []Key{RuneKey('<'), RuneKey('x'), RuneKey('>')}},
		{"backspace y ctrl", "<BS><C-r><C-v>", []Key{BackspaceKey(), CtrlRKey(), CtrlVKey()}},
		{"flechas", "<Up><Down><Left><Right>", []Key{UpKey(), DownKey(), LeftKey(), RightKey()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Keys(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("Keys(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("Keys(%q)[%d] = %v, want %v", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}
