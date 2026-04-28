package home

import (
	"testing"

	"github.com/Pegorino82/lfcru_forum/internal/football"
)

func makeStandings(n, lfcPos int) []football.StandingsEntry {
	entries := make([]football.StandingsEntry, n)
	for i := range entries {
		pos := i + 1
		name := "Team"
		if pos == lfcPos {
			name = "Liverpool FC"
		}
		entries[i] = football.StandingsEntry{Position: pos, TeamName: name}
	}
	return entries
}

func TestCompactStandingsRange(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		lfcPos    int
		wantStart int
		wantEnd   int
	}{
		{name: "LFC 1st of 20", total: 20, lfcPos: 1, wantStart: 0, wantEnd: 5},
		{name: "LFC 2nd of 20", total: 20, lfcPos: 2, wantStart: 0, wantEnd: 5},
		{name: "LFC 3rd of 20", total: 20, lfcPos: 3, wantStart: 0, wantEnd: 5},
		{name: "LFC 10th of 20", total: 20, lfcPos: 10, wantStart: 7, wantEnd: 12},
		{name: "LFC 19th of 20", total: 20, lfcPos: 19, wantStart: 15, wantEnd: 20},
		{name: "LFC 20th of 20", total: 20, lfcPos: 20, wantStart: 15, wantEnd: 20},
		{name: "LFC not in standings", total: 5, lfcPos: 99, wantStart: 0, wantEnd: 0},
		{name: "fewer than 5 teams total", total: 3, lfcPos: 2, wantStart: 0, wantEnd: 3},
		{name: "empty standings", total: 0, lfcPos: 99, wantStart: 0, wantEnd: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			standings := makeStandings(tc.total, tc.lfcPos)
			start, end := compactStandingsRange(standings)
			if start != tc.wantStart || end != tc.wantEnd {
				t.Errorf("got [%d,%d), want [%d,%d)", start, end, tc.wantStart, tc.wantEnd)
			}
		})
	}
}
