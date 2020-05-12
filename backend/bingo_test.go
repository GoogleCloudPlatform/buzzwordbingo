package main

import (
	"testing"
)

func TestBoardBingo(t *testing.T) {

	cases := []struct {
		label string
		in    Board
		want  bool
	}{
		{"Empty", Board{}, false},
		{"Top Row", Board{Phrases: []Phrase{
			Phrase{Row: "0", Column: "B", Selected: true},
			Phrase{Row: "0", Column: "I", Selected: true},
			Phrase{Row: "0", Column: "N", Selected: true},
			Phrase{Row: "0", Column: "G", Selected: true},
			Phrase{Row: "0", Column: "O", Selected: true}}}, true},
		{"Diagonal", Board{Phrases: []Phrase{
			Phrase{Row: "0", Column: "B", Selected: true},
			Phrase{Row: "1", Column: "I", Selected: true},
			Phrase{Row: "2", Column: "N", Selected: true},
			Phrase{Row: "3", Column: "G", Selected: true},
			Phrase{Row: "4", Column: "O", Selected: true}}}, true},
		{"V pattern", Board{Phrases: []Phrase{
			Phrase{Row: "0", Column: "B", Selected: true},
			Phrase{Row: "1", Column: "I", Selected: true},
			Phrase{Row: "2", Column: "N", Selected: true},
			Phrase{Row: "1", Column: "G", Selected: true},
			Phrase{Row: "0", Column: "O", Selected: true}}}, false},
	}

	for _, c := range cases {
		got := c.in.Bingo()
		if got != c.want {
			t.Errorf("Board.TestBingo(%s) got %t, want %t", c.label, got, c.want)
		}
	}

}

func TestBoardLoad(t *testing.T) {
	phrases := getTestPhrases()

	cases := []struct {
		in    func() int64
		first string
		last  string
	}{
		{func() int64 { return int64(1) }, "18", "16"},
		{func() int64 { return int64(2) }, "16", "6"},
		{func() int64 { return int64(3) }, "23", "7"},
	}

	for _, c := range cases {
		b := Board{}
		randseedfunc = c.in
		b.Load(phrases)
		gotfirst := b.Phrases[0].ID
		if gotfirst != c.first {
			t.Errorf("Board.Load() first got %s, want %s", gotfirst, c.first)
		}

		gotlast := b.Phrases[len(b.Phrases)-1].ID
		if gotlast != c.last {
			t.Errorf("Board.Load() last got %s, want %s", gotlast, c.last)
		}
	}

}

func TestRowCalc(t *testing.T) {
	cases := []struct {
		in     int
		column string
		row    string
	}{
		{0, "B", "0"},
		{1, "I", "0"},
		{2, "N", "0"},
		{3, "G", "0"},
		{4, "O", "0"},
		{5, "B", "1"},
		{6, "I", "1"},
		{7, "N", "1"},
		{8, "G", "1"},
		{9, "O", "1"},
		{10, "B", "2"},
		{11, "I", "2"},
		{12, "N", "2"},
		{13, "G", "2"},
		{14, "O", "2"},
		{15, "B", "3"},
		{16, "I", "3"},
		{17, "N", "3"},
		{18, "G", "3"},
		{19, "O", "3"},
		{20, "B", "4"},
		{21, "I", "4"},
		{22, "N", "4"},
		{23, "G", "4"},
		{24, "O", "4"},
	}
	for _, c := range cases {
		gotcolumn, gotrow := calcColumnsRows(c.in)
		if gotcolumn != c.column {
			t.Errorf("Board.CalcColumnsRows(%d) column got %s, want %s", c.in, gotcolumn, c.column)
		}

		if gotrow != c.row {
			t.Errorf("Board.CalcColumnsRows(%d) row got %s, want %s", c.in, gotrow, c.row)
		}
	}

}

func TestDubiousBingo(t *testing.T) {

}

func getTestBoard() Board {
	board := Board{}
	board.Load(getTestPhrases())

	return board
}

func getTestGame() Game {
	game := Game{}
	game.Master.Load(getTestPhrases())

	return game
}

func getTestPhrases() []Phrase {
	phrases := []Phrase{
		Phrase{"1", "", false, "", "", 0},
		Phrase{"2", "", false, "", "", 1},
		Phrase{"3", "", false, "", "", 2},
		Phrase{"4", "", false, "", "", 3},
		Phrase{"5", "", false, "", "", 4},
		Phrase{"6", "", false, "", "", 5},
		Phrase{"7", "", false, "", "", 6},
		Phrase{"8", "", false, "", "", 0},
		Phrase{"9", "", false, "", "", 1},
		Phrase{"10", "", false, "", "", 2},
		Phrase{"11", "", false, "", "", 3},
		Phrase{"12", "", false, "", "", 4},
		Phrase{"13", "", false, "", "", 5},
		Phrase{"14", "", false, "", "", 6},
		Phrase{"15", "", false, "", "", 0},
		Phrase{"16", "", false, "", "", 1},
		Phrase{"17", "", false, "", "", 2},
		Phrase{"18", "", false, "", "", 3},
		Phrase{"19", "", false, "", "", 4},
		Phrase{"20", "", false, "", "", 5},
		Phrase{"21", "", false, "", "", 6},
		Phrase{"22", "", false, "", "", 3},
		Phrase{"23", "", false, "", "", 4},
		Phrase{"24", "", false, "", "", 5},
		Phrase{"25", "", false, "", "", 6},
	}

	return phrases
}
