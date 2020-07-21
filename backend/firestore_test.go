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
	"context"
	"log"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
)

func newFirestoreTestClient(ctx context.Context) *firestore.Client {
	client, err := firestore.NewClient(ctx, "test")
	if err != nil {
		log.Fatalf("firebase.NewClient err: %v", err)
	}

	return client
}

const FirestoreEmulatorHost = "FIRESTORE_EMULATOR_HOST"

func agentTestSetup() {
	agent := Agent{}
	agent.ProjectID = projectID
	agent.ctx = context.Background()
	agent.client = newFirestoreTestClient(agent.ctx)
	a = agent
}

func TestEditAdmin(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"

	if err := a.AddAdmin(player); err != nil {
		t.Errorf("Agent.AddAdmin() err want %v got %s ", nil, err)
	}

	admins, err := a.GetAdmins()
	if err != nil {
		t.Errorf("Agent.GetAdmins() err want %v got %s ", nil, err)
	}

	isAdmin, err := a.IsAdmin(player.Email)
	if err != nil {
		t.Errorf("Agent.IsAdmin() err want %v got %s ", nil, err)
	}

	if !isAdmin {
		t.Errorf("Agent.AddAdmin() said it added the admin, but clearly didn't")
	}

	if !admins.IsMember(player) {
		t.Errorf("Agent.AddAdmin() said it added the admin, but clearly didn't")
	}

	if err := a.DeleteAdmin(player); err != nil {
		t.Errorf("Agent.DeleteAdmin() err want %v got %s ", nil, err)
	}

	adminsPostDelete, err := a.GetAdmins()
	if err != nil {
		t.Errorf("Agent.GetAdmins() err want %v got %s ", nil, err)
	}

	isAdminPostDelete, err := a.IsAdmin(player.Email)
	if err != nil {
		t.Errorf("Agent.IsAdmin() err want %v got %s ", nil, err)
	}

	if adminsPostDelete.IsMember(player) {
		t.Errorf("Agent.AddAdmin() said it deleted the admin, but clearly didn't")
	}

	if isAdminPostDelete {
		t.Errorf("Agent.AddAdmin() said it deleted the admin, but clearly didn't")
	}
}

func TestEditMasterPhrases(t *testing.T) {
	phrases := getTestPhrases()

	if err := a.LoadPhrases(phrases); err != nil {
		t.Errorf("Agent.LoadPhrases() err want %v got %s ", nil, err)
	}

	phrasesFromData, err := a.GetPhrases()
	if err != nil {
		t.Errorf("Agent.GetPhrases() err want %v got %s ", nil, err)
	}

	if !masterPhrasesEquals(phrases, phrasesFromData) {
		t.Errorf("Agent.GetPhrases() should return an unchanged set of phrases")
	}

	updatePhrase := phrases[0]

	if err := a.UpdateMasterPhrase(updatePhrase); err != nil {
		t.Errorf("Agent.UpdateMasterPhrase() err want %v got %s ", nil, err)
	}

	phrasesFromDataPostUpdate, err := a.GetPhrases()
	if err != nil {
		t.Errorf("Agent.GetPhrases() err want %v got %s ", nil, err)
	}

	updatePhraseFromDataPostUpdate := phrasesFromDataPostUpdate[0]

	if updatePhraseFromDataPostUpdate.Text != updatePhrase.Text {
		t.Errorf("Agent.UpdateMasterPhrase() didn't update want %s got %s ", updatePhrase.Text, updatePhraseFromDataPostUpdate.Text)
	}
}

func TestUpdateGame(t *testing.T) {

	game, _, _, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	gameFromData, err := a.GetGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGame() err want %v got %s ", nil, err)
	}

	if !gameEquals(game, gameFromData) {
		t.Errorf("Agent.GetGame() should return an unchanged game")
	}

	game.Active = false
	if err := a.SaveGame(game); err != nil {
		t.Errorf("Agent.SaveGame() err want %v got %s ", nil, err)
	}

	gameFromDataPostEdit, err := a.GetGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGame() err want %v got %s ", nil, err)
	}

	if !gameEquals(game, gameFromDataPostEdit) {
		t.Errorf("Agent.GetGame() should return an unchanged game")
	}

	if err := a.DeleteGame(gameFromDataPostEdit); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestUpdatePhrasesForGame(t *testing.T) {
	game, board, _, phrase, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	// Altered because the phrase gets stored will nil values.
	phrase.Text = "Altered text"
	phrase.DisplayOrder = 0
	phrase.Row = ""
	phrase.Column = ""

	board.UpdatePhrase(phrase)
	game.UpdatePhrase(phrase)

	if _, err := a.SaveBoard(board); err != nil {
		t.Errorf("Agent.SaveBoard() err want %v got %s ", nil, err)
	}

	if err := a.UpdatePhrase(game, phrase); err != nil {
		t.Errorf("Agent.UpdatePhrase() err want %v got %s ", nil, err)
	}

	gameFromDataPostEdit, err := a.GetGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGame() err want %v got %s ", nil, err)
	}

	if !gameEquals(game, gameFromDataPostEdit) {
		t.Errorf("Agent.GetGame() should return an unchanged game")
	}

	if game.Boards[board.ID].Phrases[phrase.ID].Text != phrase.Text {
		t.Errorf("Agent.UpdatePhrase() board should be the same as board in game")
	}

	if err := a.DeleteGame(game); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestSelectPhrase(t *testing.T) {
	game, board, player, phrase, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}
	phrase.Selected = true

	phrase = board.Select(phrase)
	record := game.Select(phrase, player)

	if err := a.SelectPhrase(board, phrase, record); err != nil {
		t.Errorf("Agent.SelectPhrase() err want %v got %s ", nil, err)
	}

	gameFromDataPostEdit, err := a.GetGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGame() err want %v got %s ", nil, err)
	}

	if !gameEquals(game, gameFromDataPostEdit) {
		t.Errorf("Agent.GetGame() should return an unchanged game")
	}

	boardFromDataPostEdit, err := a.GetBoard(board.ID, game.ID)
	if err != nil {
		t.Errorf("Agent.GetBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromDataPostEdit) {
		t.Errorf("Agent.GetBoard() should return an unchanged board.")
	}

	if err := a.DeleteGame(gameFromDataPostEdit); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestGamesEditing(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"
	player2 := Player{}
	player2.Email = "test2@example.com"
	phrases := getTestPhrases()

	if err := a.LoadPhrases(phrases); err != nil {
		t.Errorf("Agent.LoadPhrases() err want %v got %s ", nil, err)
	}

	game1, err := a.NewGame("test game1", player)
	if err != nil {
		t.Errorf("Agent.NewGame() err want %v got %s ", nil, err)
	}

	game2, err := a.NewGame("test game2", player)
	if err != nil {
		t.Errorf("Agent.NewGame() err want %v got %s ", nil, err)
	}

	game3, err := a.NewGame("test game3", player2)
	if err != nil {
		t.Errorf("Agent.NewGame() err want %v got %s ", nil, err)
	}

	games, err := a.GetGames(10, time.Now())
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(games) != 3 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 3, len(games))
	}

	games2, err := a.GetGamesForKey(player2.Email)
	if err != nil {
		t.Errorf("Agent.GetGames() err want %v got %s ", nil, err)
	}

	if len(games2) != 1 {
		t.Errorf("Agent.GetGames() count want %d got %d ", 1, len(games))
	}

	if err := a.DeleteGame(game1); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
	if err := a.DeleteGame(game2); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
	if err := a.DeleteGame(game3); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}
}

func TestBoardDelete(t *testing.T) {

	game, board, _, phrase, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	player2 := Player{}
	player2.Email = "test2@example.com"
	phrase.Selected = true

	board2 := game.NewBoard(player2)

	if _, err := a.SaveBoard(board2); err != nil {
		t.Errorf("Agent.SaveBoard() err want %v got %s ", nil, err)
	}

	if err := a.SaveGame(game); err != nil {
		t.Errorf("Agent.SaveGame() err want %v got %s ", nil, err)
	}

	if err := a.DeleteBoard(board, game); err != nil {
		t.Errorf("Agent.DeleteBoard() err want %v got %s ", nil, err)
	}

	gameFromDataPostEdit, err := a.GetGame(game.ID)
	if err != nil {
		t.Errorf("Agent.GetGame() err want %v got %s ", nil, err)
	}

	if len(gameFromDataPostEdit.Boards) != 1 {
		t.Errorf("Agent.DeleteBoard() count want %d got %d ", 1, len(gameFromDataPostEdit.Boards))
	}

	if err := a.DeleteGame(gameFromDataPostEdit); err != nil {
		t.Errorf("Agent.DeleteGame() err want %v got %s ", nil, err)
	}

}

func TestGetGameForPlayer(t *testing.T) {
	game, board, player, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	boardForEmail, err := a.GetBoardForPlayer(game.ID, player)
	if err != nil {
		t.Errorf("Agent.GetBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardForEmail) {
		t.Errorf("Agent.GetBoardForPlayer() should return an unchanged board.")
	}

}

func TestMessageAcknowledge(t *testing.T) {
	game, _, player, _, err := initFirestoreBaseState()
	if err != nil {
		t.Errorf("initFirestoreBaseState() err want %v got %s ", nil, err)
	}

	messages := []Message{}
	m1 := Message{}
	m1.ID = "1"
	m1.SetText("Test message")
	m1.SetAudience(player.Email)
	messages = append(messages, m1)

	if err := a.AddMessagesToGame(game, messages); err != nil {
		t.Errorf("Agent.AddMessagesToGame() err want %v got %s ", nil, err)
	}

	if err := a.AcknowledgeMessage(game, m1); err != nil {
		t.Errorf("Agent.AcknowledgeMessage() err want %v got %s ", nil, err)
	}

}

func initFirestoreBaseState() (Game, Board, Player, Phrase, error) {
	player := Player{}
	player.Email = "test@example.com"
	phrases := getTestPhrases()
	phrase := phrases[0]

	if err := a.LoadPhrases(phrases); err != nil {
		return Game{}, Board{}, player, phrase, err
	}

	game, err := a.NewGame("test game", player)
	if err != nil {
		return game, Board{}, player, phrase, err
	}

	board := game.NewBoard(player)

	if _, err := a.SaveBoard(board); err != nil {
		return game, board, player, phrase, err
	}

	if err := a.SaveGame(game); err != nil {
		return game, board, player, phrase, err
	}
	return game, board, player, phrase, nil
}
