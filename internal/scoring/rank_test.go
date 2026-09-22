package scoring

import "testing"

func TestRankFor(t *testing.T) {
	tests := []struct {
		pct  float64
		want Rank
	}{
		{0, RankD},
		{49.9, RankD},
		{50, RankC},
		{69.9, RankC},
		{70, RankB},
		{84.9, RankB},
		{85, RankA},
		{94.9, RankA},
		{95, RankS},
		{130, RankS}, // con combo se puede superar 100% y sigue siendo S
	}
	for _, tt := range tests {
		if got := RankFor(tt.pct); got != tt.want {
			t.Errorf("RankFor(%v) = %v, want %v", tt.pct, got, tt.want)
		}
	}
}
