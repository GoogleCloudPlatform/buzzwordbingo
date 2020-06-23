// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		{func() int64 { return int64(1) }, "1", "16"},
		{func() int64 { return int64(2) }, "24", "6"},
		{func() int64 { return int64(3) }, "24", "7"},
	}

	for _, c := range cases {
		b := InitBoard()
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

func TestBoardPhraseUpdate(t *testing.T) {
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

func TestPlayerAddAndRemove(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	players := Players{}

	players.Add(pl)

	if len(players) != 1 {
		t.Errorf("Players.Add() expected there to be 1 player")
	}

	players.Add(pl2)
	if len(players) != 2 {
		t.Errorf("Players.Add() expected there to be 2 player")
	}

	players.Add(pl)

	if len(players) != 2 {
		t.Errorf("Players.Add() expected there to be 1 player")
	}

	if !players.IsMember(pl) {
		t.Errorf("Players.Add() expected added player to be present")
	}

	players.Remove(pl)

	if len(players) != 1 {
		t.Errorf("Players.Remove() expected there to be 1 player")
	}

	players.Remove(pl)

	if len(players) != 1 {
		t.Errorf("Players.Remove() expected there to be 1 player")
	}

	if players.IsMember(pl) {
		t.Errorf("Players.Remove() expected removed player to not be present")
	}
}

func TestGamePhraseUpdate(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	game := NewGame("test name", pl, getTestPhrases())
	game.Admins.Add(pl2)
	_ = game.NewBoard(pl2)

	phrase := Phrase{"1", "Test Phrase", false, "", "", 0}

	game.UpdatePhrase(phrase)

	for _, v := range game.Master.Records {
		if v.Phrase.ID == phrase.ID {
			if v.Phrase.Text != phrase.Text {
				t.Errorf("Board.UpdatePhrase() got %s, want %s", v.Phrase.Text, phrase.Text)
			}
			return
		}
	}
}

func TestGameCheckBoardAndDubious(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	pl3 := Player{}
	pl3.Email = "test3@example.com"
	game := NewGame("test name", pl, getTestPhrases())
	board := game.NewBoard(pl)
	board2 := game.NewBoard(pl2)
	_ = game.NewBoard(pl3)

	temp := getTestPhrases()
	for _, v := range temp {
		board.Phrases[v.ID] = v
		board2.Phrases[v.ID] = v
	}
	game.Boards[board.ID] = board
	game.Boards[board2.ID] = board2

	phrase1 := board.Phrases["1"]
	phrase2 := board.Phrases["2"]
	phrase3 := board.Phrases["3"]
	phrase4 := board.Phrases["4"]
	phrase5 := board.Phrases["5"]

	phrase1.Selected = true
	phrase2.Selected = true
	phrase3.Selected = true
	phrase4.Selected = true
	phrase5.Selected = true

	game.Select(phrase1, pl)
	board.Select(phrase1)

	game.Select(phrase2, pl)
	board.Select(phrase2)

	game.Select(phrase3, pl)
	board.Select(phrase3)

	game.Select(phrase4, pl)
	board.Select(phrase4)

	game.Select(phrase5, pl)
	board.Select(phrase5)

	if !board.Bingo() {
		t.Errorf("Board.Select sequence should have made bingo")
	}

	inGameBoard := game.Boards[board.ID]
	if !inGameBoard.Bingo() {
		t.Errorf("Game.Select sequence should have made bingo")
	}

	results := game.CheckBoard(board)

	for _, v := range results {
		if v.Percent > 34 {
			t.Errorf("Game.CheckBoard() Percents off,  want %f got %f ", .33, v.Percent)
		}
	}

	if !results.IsDubious() {
		t.Errorf("Reports.IsDubious() should have been true. ")
	}

	game.Select(phrase1, pl2)
	board2.Select(phrase1)

	game.Select(phrase2, pl2)
	board2.Select(phrase2)

	game.Select(phrase3, pl2)
	board2.Select(phrase3)

	game.Select(phrase4, pl2)
	board2.Select(phrase4)

	game.Select(phrase5, pl2)
	board2.Select(phrase5)

	if !board2.Bingo() {
		t.Errorf("Board.Select sequence should have made bingo")
	}

	inGameBoard2 := game.Boards[board.ID]
	if !inGameBoard2.Bingo() {
		t.Errorf("Game.Select sequence should have made bingo")
	}

	results2 := game.CheckBoard(board2)

	for _, v := range results2 {
		if v.Percent > 67 {
			t.Errorf("Game.CheckBoard() Percents off,  want %f got %f ", .33, v.Percent)
		}
	}

	if results2.IsDubious() {
		t.Errorf("Reports.IsDubious() should have been false. ")
	}

}

func TestNewGame(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	game := NewGame("test name", pl, getTestPhrases())

	if !game.IsAdmin(pl) {
		t.Errorf("NewGame() expected player passed into be an admin, but they were not")
	}

	if game.IsAdmin(pl2) {
		t.Errorf("NewGame() expected player not passed into to not be an admin, but they were")
	}

	if !game.Players.IsMember(pl) {
		t.Errorf("NewGame() expected player passed into be a player, but they were not")
	}

	if game.Players.IsMember(pl2) {
		t.Errorf("NewGame() expected player not passed into to not be a player, but they were")
	}

}

func TestNewMessage(t *testing.T) {
	m := Message{}
	m.SetText("%s if %s works", "test", "this")
	m.SetAudience("all", "test@test.com")

	testText := "test if this works"
	if m.Text != "test if this works" {
		t.Errorf("Message.SetText() got %s, want %s", m.Text, testText)
	}

	foundExpected := 0
	for _, v := range m.Audience {
		if v == "all" || v == "test@test.com" {
			foundExpected++
		}
	}
	if foundExpected != 2 {
		t.Errorf("Message.SetAudience() got %s, want %s", m.Audience, []string{"all", "test@test.com"})
	}

}

func TestGameNewBoard(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	game := NewGame("test name", pl, getTestPhrases())

	board := game.NewBoard(pl2)

	if !game.IsAdmin(pl) {
		t.Errorf("NewGame() expected player passed into be an admin, but they were not")
	}

	if game.IsAdmin(pl2) {
		t.Errorf("NewGame() expected player not passed into to not be an admin, but they were")
	}

	if !game.Players.IsMember(pl) {
		t.Errorf("NewGame() expected player passed into be a player, but they were not")
	}

	if !game.Players.IsMember(pl2) {
		t.Errorf("NewGame() expected player getting board to be a player, but they were not")
	}

	if board.Game != game.ID {
		t.Errorf("NewGame() expected board to have game.id set as board.game, it was not. ")
	}

	_, ok := game.Boards[board.ID]

	if !ok {
		t.Errorf("NewGame() expected board to be in the list of boards for the game, it was not. ")
	}
}

func TestGameDeletingBoard(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	game := NewGame("test name", pl, getTestPhrases())

	board := game.NewBoard(pl2)

	game.DeleteBoard(board)

	_, ok := game.Boards[board.ID]

	if ok {
		t.Errorf("Game.Delete() expected board to not be in the list of boards for the game, it was. ")
	}
}

func TestGameObscure(t *testing.T) {
	pl := Player{}
	pl.Email = "test@example.com"
	pl2 := Player{}
	pl2.Email = "test2@example.com"
	game := NewGame("test name", pl, getTestPhrases())
	game.Admins.Add(pl2)
	board := game.NewBoard(pl2)

	game.Obscure("test@example.com")

	targetFoundInPlayers := false
	for _, v := range game.Players {
		if v.Email == "test2@example.com" {
			t.Errorf("Game.Obscure() expected email address to be xxxxxx@xxxxxx.xxx got %s", v.Email)
		}
		if v.Email == "test@example.com" {
			targetFoundInPlayers = true
		}
	}

	if !targetFoundInPlayers {
		t.Errorf("Game.Obscure() expected email address to find email address %s", pl)
	}

	targetFoundInAdmins := false
	for _, v := range game.Admins {
		if v.Email == "test2@example.com" {
			t.Errorf("Game.Obscure() expected email address to be xxxxxx@xxxxxx.xxx got %s", v.Email)
		}
		if v.Email == "test@example.com" {
			targetFoundInAdmins = true
		}
	}

	if !targetFoundInAdmins {
		t.Errorf("Game.Obscure() expected email address to find email address %s", pl)
	}

	savedBoard, _ := game.Boards[board.ID]

	if savedBoard.Player.Email == "test2@example.com" {
		t.Errorf("Game.Obscure() expected email address to be xxxxxx@xxxxxx.xxx got %s", board.Player.Email)
	}
}

func TestGameSelectAndUnselect(t *testing.T) {
	phrases := getTestPhrases()
	phrase := phrases[0]
	phrase.Selected = true
	pl := Player{}
	pl.Email = "test@example.com"
	g := NewGame("test game", pl, phrases)
	_ = g.NewBoard(pl)

	g.Select(phrase, pl)

	i, record := g.FindRecord(phrase)

	if i != 0 {
		t.Errorf("Game.Select() GameFindRecord() want %d got %d ", 0, i)
	}

	if !record.Phrase.Selected {
		t.Errorf("Game.Select() GameFindRecord() Phrase.Selected want %t got %t ", true, record.Phrase.Selected)
	}

	phrase.Selected = false
	g.Select(phrase, pl)

	_, record2 := g.FindRecord(phrase)

	if record2.Phrase.Selected {
		t.Errorf("Game.Select() GameFindRecord() Phrase.Selected want %t got %t ", false, record2.Phrase.Selected)
	}

}

func TestMasterDoesNotExist(t *testing.T) {
	phrases := getTestPhrases()
	phrase := phrases[0]
	phrase.ID = "1021212"
	phrase.Selected = true
	pl := Player{}
	pl.Email = "test@example.com"
	g := NewGame("test game", pl, phrases)

	g.Select(phrase, pl)

	i, _ := g.FindRecord(phrase)

	if i != -1 {
		t.Errorf("Game.Select() GameFindRecord() want %d got %d ", -1, i)
	}

}

func getTestBoard() Board {
	board := InitBoard()
	board.Load(getTestPhrases())
	board.ID = "1"

	return board
}

func getTestGame() Game {
	game := NewGame("A Test Game", Player{"Test", "t@t"}, getTestPhrases())

	return game
}

func getTestPhrases() []Phrase {
	phrases := []Phrase{
		{"1", "Filler 1", false, "0", "B", 0},
		{"2", "Filler 2", false, "0", "I", 1},
		{"3", "Filler 3", false, "0", "N", 2},
		{"4", "Filler 4", false, "0", "G", 3},
		{"5", "Filler 5", false, "0", "O", 4},
		{"6", "Filler 6", false, "1", "B", 5},
		{"7", "Filler 7", false, "1", "I", 6},
		{"8", "Filler 8", false, "1", "N", 7},
		{"9", "Filler 9", false, "1", "G", 8},
		{"10", "Filler 10", false, "1", "O", 9},
		{"11", "Filler 11", false, "2", "B", 10},
		{"12", "Filler 12", false, "2", "I", 11},
		{"13", "FREE", false, "2", "N", 12},
		{"14", "Filler 14", false, "2", "G", 13},
		{"15", "Filler 15", false, "2", "O", 14},
		{"16", "Filler 16", false, "3", "B", 15},
		{"17", "Filler 17", false, "3", "I", 16},
		{"18", "Filler 18", false, "3", "N", 17},
		{"19", "Filler 19", false, "3", "G", 18},
		{"20", "Filler 20", false, "3", "O", 19},
		{"21", "Filler 21", false, "4", "B", 20},
		{"22", "Filler 22", false, "4", "I", 21},
		{"23", "Filler 23", false, "4", "N", 22},
		{"24", "Filler 24", false, "4", "G", 23},
		{"25", "Filler 25", false, "4", "O", 24},
	}

	return phrases
}
