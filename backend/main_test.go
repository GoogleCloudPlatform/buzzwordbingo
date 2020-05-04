package main

import "testing"

func TestBingo(t *testing.T) {

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
	}

	for _, c := range cases {
		got := c.in.CheckBingo()
		if got != c.want {
			t.Errorf("Board.TestBingo() got %t, want %t", got, c.want)
		}
	}

}
