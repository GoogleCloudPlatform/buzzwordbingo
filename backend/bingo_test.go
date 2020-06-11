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
		{"Top Row", Board{Phrases: map[string]Phrase{
			"1": {Row: "0", Column: "B", Selected: true},
			"2": {Row: "0", Column: "I", Selected: true},
			"3": {Row: "0", Column: "N", Selected: true},
			"4": {Row: "0", Column: "G", Selected: true},
			"5": {Row: "0", Column: "O", Selected: true}}}, true},
		{"Diagonal", Board{Phrases: map[string]Phrase{
			"1": {Row: "0", Column: "B", Selected: true},
			"2": {Row: "1", Column: "I", Selected: true},
			"3": {Row: "2", Column: "N", Selected: true},
			"4": {Row: "3", Column: "G", Selected: true},
			"5": {Row: "4", Column: "O", Selected: true}}}, true},
		{"V pattern", Board{Phrases: map[string]Phrase{
			"1": {Row: "0", Column: "B", Selected: true},
			"2": {Row: "1", Column: "I", Selected: true},
			"3": {Row: "2", Column: "N", Selected: true},
			"4": {Row: "1", Column: "G", Selected: true},
			"5": {Row: "0", Column: "O", Selected: true}}}, false},
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
		b := NewBoard()
		randseedfunc = c.in
		b.Load(phrases)

		phrases := b.Phrases.ByDisplayOrder()

		gotfirst := phrases[0].ID
		if gotfirst != c.first {
			t.Errorf("Board.Load() first got %s, want %s", gotfirst, c.first)
		}

		gotlast := phrases[len(phrases)-1].ID
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

func TestPhraseUpdate(t *testing.T) {
	board := getTestBoard()
	phrase := Phrase{"1", "Test Phrase", false, "", "", 0}

	board.UpdatePhrase(phrase)

	for _, v := range board.Phrases {
		if v.ID == phrase.ID {
			if v.Text != phrase.Text {
				t.Errorf("Board.UpdatePhrase() got %s, want %s", v.Text, phrase.Text)
			}
			return
		}
	}

}

func getTestBoard() Board {
	board := NewBoard()
	board.Load(getTestPhrases())

	return board
}

func getTestGame() Game {
	game := NewGame("1", "A Test Game", Player{"Test", "t@t"}, getTestPhrases())

	return game
}

func getTestPhrases() []Phrase {
	phrases := []Phrase{
		{"1", "Filler 1", false, "", "", 0},
		{"2", "Filler 2", false, "", "", 1},
		{"3", "Filler 3", false, "", "", 2},
		{"4", "Filler 4", false, "", "", 3},
		{"5", "Filler 5", false, "", "", 4},
		{"6", "Filler 6", false, "", "", 5},
		{"7", "Filler 7", false, "", "", 6},
		{"8", "Filler 8", false, "", "", 0},
		{"9", "Filler 9", false, "", "", 1},
		{"10", "Filler 10", false, "", "", 2},
		{"11", "Filler 11", false, "", "", 3},
		{"12", "Filler 12", false, "", "", 4},
		{"13", "Filler 13", false, "", "", 5},
		{"14", "Filler 14", false, "", "", 6},
		{"15", "Filler 15", false, "", "", 0},
		{"16", "Filler 16", false, "", "", 1},
		{"17", "Filler 17", false, "", "", 2},
		{"18", "Filler 18", false, "", "", 3},
		{"19", "Filler 19", false, "", "", 4},
		{"20", "Filler 20", false, "", "", 5},
		{"21", "Filler 21", false, "", "", 6},
		{"22", "Filler 22", false, "", "", 3},
		{"23", "Filler 23", false, "", "", 4},
		{"24", "Filler 24", false, "", "", 5},
		{"25", "Filler 25", false, "", "", 6},
	}

	return phrases
}
