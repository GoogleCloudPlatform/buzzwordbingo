package main

import (
	"testing"
)

func TestBoardBingo(t *testing.T) {

	cases := []struct {
		in   Board
		want bool
	}{
		{Board{}, false},
		{Board{Phrases: []Phrase{
			Phrase{Row: "1", Column: "B", Selected: true},
			Phrase{Row: "1", Column: "I", Selected: true},
			Phrase{Row: "1", Column: "N", Selected: true},
			Phrase{Row: "1", Column: "G", Selected: true},
			Phrase{Row: "1", Column: "O", Selected: true}}}, true},
		{Board{Phrases: []Phrase{
			Phrase{Row: "1", Column: "B", Selected: true},
			Phrase{Row: "2", Column: "I", Selected: true},
			Phrase{Row: "3", Column: "N", Selected: true},
			Phrase{Row: "4", Column: "G", Selected: true},
			Phrase{Row: "5", Column: "O", Selected: true}}}, true},
		{Board{Phrases: []Phrase{
			Phrase{Row: "1", Column: "B", Selected: true},
			Phrase{Row: "2", Column: "I", Selected: true},
			Phrase{Row: "3", Column: "N", Selected: true},
			Phrase{Row: "2", Column: "G", Selected: true},
			Phrase{Row: "1", Column: "O", Selected: true}}}, false},
	}

	for _, c := range cases {
		got := c.in.Bingo()
		if got != c.want {
			t.Errorf("Board.TestBingo() got %t, want %t", got, c.want)
		}
	}

}

func TestBoardLoad(t *testing.T) {
	phrases := []Phrase{
		Phrase{"1", "", false, "", ""},
		Phrase{"2", "", false, "", ""},
		Phrase{"3", "", false, "", ""},
		Phrase{"4", "", false, "", ""},
		Phrase{"5", "", false, "", ""},
		Phrase{"6", "", false, "", ""},
		Phrase{"7", "", false, "", ""},
	}

	cases := []struct {
		in    func() int64
		first string
		last  string
	}{
		{func() int64 { return int64(1) }, "1", "5"},
		{func() int64 { return int64(2) }, "2", "3"},
		{func() int64 { return int64(3) }, "7", "5"},
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
			t.Errorf("Board.Load() first got %s, want %s", gotlast, c.last)
		}
	}

}

func TestRowCalc(t *testing.T) {
	cases := []struct {
		in     int
		column string
		row    string
	}{
		{1, "B", "0"},
		{2, "I", "0"},
		{3, "N", "0"},
		{6, "B", "1"},
		{25, "O", "4"},
	}
	b := Board{}
	for _, c := range cases {
		gotcolumn, gotrow := b.CalcColumnsRows(c.in)
		if gotcolumn != c.column {
			t.Errorf("Board.CalcColumnsRows(%d) column got %s, want %s", c.in, gotcolumn, c.column)
		}

		if gotrow != c.row {
			t.Errorf("Board.CalcColumnsRows(%d) row got %s, want %s", c.in, gotrow, c.row)
		}
	}

}
