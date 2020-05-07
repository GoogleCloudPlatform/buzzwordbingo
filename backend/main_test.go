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
			Phrase{Row: "1", Column: "B", Clicked: true},
			Phrase{Row: "1", Column: "I", Clicked: true},
			Phrase{Row: "1", Column: "N", Clicked: true},
			Phrase{Row: "1", Column: "G", Clicked: true},
			Phrase{Row: "1", Column: "O", Clicked: true}}}, true},
		{Board{Phrases: []Phrase{
			Phrase{Row: "1", Column: "B", Clicked: true},
			Phrase{Row: "2", Column: "I", Clicked: true},
			Phrase{Row: "3", Column: "N", Clicked: true},
			Phrase{Row: "4", Column: "G", Clicked: true},
			Phrase{Row: "5", Column: "O", Clicked: true}}}, true},
		{Board{Phrases: []Phrase{
			Phrase{Row: "1", Column: "B", Clicked: true},
			Phrase{Row: "2", Column: "I", Clicked: true},
			Phrase{Row: "3", Column: "N", Clicked: true},
			Phrase{Row: "2", Column: "G", Clicked: true},
			Phrase{Row: "1", Column: "O", Clicked: true}}}, false},
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

func TestGameBingo(t *testing.T) {

	cases := []struct {
		in   Game
		want int
	}{
		{Game{}, 0},
		{Game{
			Board{},
			[]Board{Board{Phrases: []Phrase{
				Phrase{Row: "1", Column: "B", Clicked: true},
				Phrase{Row: "1", Column: "I", Clicked: true},
				Phrase{Row: "1", Column: "N", Clicked: true},
				Phrase{Row: "1", Column: "G", Clicked: true},
				Phrase{Row: "1", Column: "O", Clicked: true}}},
				Board{Phrases: []Phrase{
					Phrase{Row: "1", Column: "B", Clicked: true},
					Phrase{Row: "2", Column: "I", Clicked: true},
					Phrase{Row: "3", Column: "N", Clicked: true},
					Phrase{Row: "4", Column: "G", Clicked: true},
					Phrase{Row: "5", Column: "O", Clicked: true}}},
				Board{Phrases: []Phrase{
					Phrase{Row: "1", Column: "B", Clicked: true},
					Phrase{Row: "2", Column: "I", Clicked: true},
					Phrase{Row: "3", Column: "N", Clicked: true},
					Phrase{Row: "2", Column: "G", Clicked: true},
					Phrase{Row: "1", Column: "O", Clicked: true}}}}}, 2},
	}

	for _, c := range cases {
		got := c.in.Bingo()
		if len(got) != c.want {
			t.Errorf("Game.Bingo() got %d, want %d", len(got), c.want)
		}
	}

}
