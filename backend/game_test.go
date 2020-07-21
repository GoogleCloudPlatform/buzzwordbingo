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
	"time"
)

func TestGetBoardForPlayer(t *testing.T) {
	game, board, player, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	boardFromFirestore, err := getBoardForPlayer(player, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromFirestore) {
		t.Errorf("getBoardForPlayer() boardFromFirestore should return an unchanged board.")
	}

	boardFromCache, err := getBoardForPlayer(player, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromCache) {
		t.Errorf("getBoardForPlayer() boardFromCache should return an unchanged board.")
	}

	if err := deleteBoard(board.ID, game.ID); err != nil {
		t.Errorf("deleteBoard() err want %v got %s ", nil, err)
	}

	boardFromFirestoreNew, err := getBoardForPlayer(player, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	if boardEquals(board, boardFromFirestoreNew) {
		t.Errorf("getBoardForPlayer() boardFromFirestoreNew should return an changed board.")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGetBoard(t *testing.T) {
	game, board, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	boardFromFirestore, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromFirestore) {
		t.Errorf("getBoard() boardFromFirestore should return an unchanged board.")
	}

	boardFromCache, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromCache) {
		t.Errorf("getBoard() boardFromCache should return an unchanged board.")
	}

	if err := deleteBoard(board.ID, game.ID); err != nil {
		t.Errorf("deleteBoard() err want %v got %s ", nil, err)
	}

	_, err = getBoard(board.ID, game.ID)
	if err == nil {
		t.Errorf("getBoard() err want err got %s ", err)
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestBingo(t *testing.T) {
	game, board, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	bingoPhrases := getBingoPhrases(board)

	for _, v := range bingoPhrases {
		if err != recordSelect(board.ID, game.ID, v.ID, true) {
			t.Errorf("recordSelect() err want %v got %s ", nil, err)
		}
	}

	boardAfterBingo, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if !boardAfterBingo.BingoDeclared {
		t.Errorf("Should have created a bingo")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGenerateBingoMessages(t *testing.T) {
	game, board, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	bingoPhrases := getBingoPhrases(board)

	for _, v := range bingoPhrases {
		if err != recordSelect(board.ID, game.ID, v.ID, true) {
			t.Errorf("recordSelect() err want %v got %s ", nil, err)
		}
	}

	messages := generateBingoMessages(board, game, true)

	if len(messages) != 1 {
		t.Errorf("generateBingoMessages should have created only one messasge")
	}

	if !messages[0].Bingo {
		t.Errorf("generateBingoMessages should have a bingo")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestDubiousBingoMessages(t *testing.T) {
	game, board, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	player2 := Player{}
	player2.Email = "test2@example.com"

	player3 := Player{}
	player3.Email = "test3@example.com"

	player4 := Player{}
	player4.Email = "test4@example.com"

	_, err = getBoardForPlayer(player2, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	game, err = getGame(game.ID)
	if err != nil {
		t.Errorf("getGame() err want %v got %s ", nil, err)
	}

	_, err = getBoardForPlayer(player3, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	game, err = getGame(game.ID)
	if err != nil {
		t.Errorf("getGame() err want %v got %s ", nil, err)
	}

	_, err = getBoardForPlayer(player4, game)
	if err != nil {
		t.Errorf("getBoardForPlayer() err want %v got %s ", nil, err)
	}

	game, err = getGame(game.ID)
	if err != nil {
		t.Errorf("getGame() err want %v got %s ", nil, err)
	}

	bingoPhrases := getBingoPhrases(board)

	for _, v := range bingoPhrases {
		if err != recordSelect(board.ID, game.ID, v.ID, true) {
			t.Errorf("recordSelect() err want %v got %s ", nil, err)
		}
	}

	gameUpdated, err := getGame(game.ID)
	if err != nil {
		t.Errorf("getGame() err want %v got %s ", nil, err)
	}

	boardUpdated, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	messages := generateBingoMessages(boardUpdated, gameUpdated, true)

	if len(messages) < 2 {
		t.Errorf("generateBingoMessages should have created a glut of messages")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGamesUniqueID(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"
	phrases := getTestPhrases()

	if err := a.LoadPhrases(phrases); err != nil {
		t.Errorf("Agent.LoadPhrases() err want %v got %s ", nil, err)
	}

	game, err := a.NewGame("test game", player)
	if err != nil {
		t.Errorf("Agent.LoadPhrases() err want %v got %s ", nil, err)
	}

	game2, err := a.NewGame("test game2", player)
	if err != nil {
		t.Errorf("Agent.LoadPhrases() err want %v got %s ", nil, err)
	}

	if game.ID == game2.ID {
		t.Errorf("Expected different games to have different ids. ")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}

	if err := a.DeleteGame(game2); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGetGames(t *testing.T) {
	game, _, player, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	game2, err := a.NewGame("test game2", player)
	if err != nil {
		t.Errorf("Agent.NewGame() err want %v got %s ", nil, err)
	}

	gamesFromFirestore, err := getGamesForKey("admin-list", 10, time.Now())
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(gamesFromFirestore) != 2 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 2, len(gamesFromFirestore))
	}

	gamesFromCache, err := getGamesForKey("admin-list", 10, time.Now())
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(gamesFromCache) != 2 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 2, len(gamesFromCache))
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}

	if err := a.DeleteGame(game2); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGetGamesForPlayer(t *testing.T) {
	game, _, player, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	game2, err := a.NewGame("test game", player)
	if err != nil {
		t.Errorf("Agent.NewGame() err want %v got %s ", nil, err)
	}

	gamesFromFirestore, err := getGamesForKey(player.Email, 10, time.Now())
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(gamesFromFirestore) != 2 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 2, len(gamesFromFirestore))
	}

	gamesFromCache, err := getGamesForKey(player.Email, 10, time.Now())
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(gamesFromCache) != 2 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 2, len(gamesFromCache))
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}

	if err := a.DeleteGame(game2); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestDeactivateGame(t *testing.T) {
	game, _, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	if err := deactivateGame(game.ID); err != nil {
		t.Errorf("deactivateGame() err want %v got %s ", nil, err)
	}

	gamesFromCache, err := getGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if gamesFromCache.Active {
		t.Errorf("deactivateGame return want %t got %t ", false, gamesFromCache.Active)
	}

	if err := cache.DeleteGame(gamesFromCache); err != nil {
		t.Errorf("cache.DeleteGame() err want %v got %s ", nil, err)
	}

	gamesFromFirestore, err := getGame(game.ID)
	if err != nil {
		t.Errorf("getGames() err want %v got %s ", nil, err)
	}

	if gamesFromFirestore.Active {
		t.Errorf("deactivateGame return want %t got %t ", false, gamesFromFirestore.Active)
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestUpdateGamePhrases(t *testing.T) {
	game, _, _, phrase, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}
	phrase.Text = "I changed it in the db"

	if err := updateGamePhrases(game.ID, phrase); err != nil {
		t.Errorf("updateGamePhrases() err want %v got %s ", nil, err)
	}

	gamesFromCache, err := getGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	_, record := gamesFromCache.FindRecord(phrase)
	if record.Phrase.Text != phrase.Text {
		t.Errorf("updateGamePhrases() text want %s got %s ", phrase.Text, record.Phrase.Text)
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestUpdateGamePhrasesWithBingo(t *testing.T) {

	game, board, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	bingoPhrases := getBingoPhrases(board)

	for _, v := range bingoPhrases {
		if err != recordSelect(board.ID, game.ID, v.ID, true) {
			t.Errorf("recordSelect() err want %v got %s ", nil, err)
		}
	}

	boardAfterBingo, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if !boardAfterBingo.BingoDeclared {
		t.Errorf("Should have created a bingo")
	}

	phrase := bingoPhrases[0]
	phrase.Text = "I changed it in the db"

	if err := updateGamePhrases(game.ID, phrase); err != nil {
		t.Errorf("updateGamePhrases() err want %v got %s ", nil, err)
	}

	boardAfterBingoReverted, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if boardAfterBingoReverted.BingoDeclared {
		t.Errorf("Should have reverted a bingo")
	}

	if err := recordSelect(board.ID, game.ID, phrase.ID, true); err != nil {
		t.Errorf("updateGamePhrases() err want %v got %s ", nil, err)
	}

	boardAfterBingoRevertedThenRedone, err := getBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("getBoard() err want %v got %s ", nil, err)
	}

	if !boardAfterBingoRevertedThenRedone.BingoDeclared {
		t.Errorf("Should have created a bingo")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func getBingoPhrases(board Board) []Phrase {
	bingoPhrases := []Phrase{}

	for _, v := range board.Phrases {
		if v.Column == "B" {
			bingoPhrases = append(bingoPhrases, v)
		}
	}
	return bingoPhrases
}
