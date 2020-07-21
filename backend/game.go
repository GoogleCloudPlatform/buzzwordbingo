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
	"fmt"
	"strings"
	"time"
)

func getBoardForPlayer(player Player, game Game) (Board, error) {
	var err error
	b := Board{}
	messages := []Message{}
	weblog("Trying Cache")
	b, err = cache.GetBoardForPlayer(game.ID, player.Email)
	if err != nil {
		if err == ErrCacheMiss {
			weblog("Cache Empty trying DB")
			b, err = a.GetBoardForPlayer(game.ID, player)
			if err != nil {
				return b, fmt.Errorf("error getting board for player: %v", err)
			}
		}
		if err := cache.DeleteGamesForKey([]string{player.Email}); err != nil {
			return b, fmt.Errorf("error clearing game cache for player: %v", err)
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board for player: %v", err)
		}
	}
	m := Message{}
	m.SetText("<strong>%s</strong> rejoined the game.", b.Player.Name)
	m.SetAudience("admin", b.Player.Email)

	if b.ID == "" {
		b = game.NewBoard(player)
		b, err = a.SaveBoard(b)
		if err != nil {
			return b, fmt.Errorf("error saving board for player: %v", err)
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board for player: %v", err)
		}
		game.Boards[b.ID] = b

		if err := cache.SaveGame(game); err != nil {
			return b, fmt.Errorf("error caching game for player: %v", err)
		}
		m.SetText("<strong>%s</strong> got a board and joined the game.", b.Player.Name)
		m.SetAudience("all")

	}

	messages = append(messages, m)

	bingo := b.Bingo()
	if bingo {
		msg := generateBingoMessages(b, game, false)
		messages = append(messages, msg...)
	}

	if err := a.AddMessagesToGame(game, messages); err != nil {
		return b, fmt.Errorf("could not send message to notify player of bingo: %s", err)
	}

	return b, nil
}

func getBoard(bid string, gid string) (Board, error) {

	b, err := cache.GetBoard(bid)
	if err != nil {
		if err == ErrCacheMiss {
			b, err = a.GetBoard(bid, gid)
			if err != nil {
				return b, fmt.Errorf("error getting board: %v", err)
			}
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board : %v", err)
		}
	}

	return b, nil
}

func deleteBoard(bid, gid string) error {
	b, err := getBoard(bid, gid)
	if err != nil {
		return fmt.Errorf("could not retrieve board from firestore: %s", err)
	}

	game, err := getGame(b.Game)
	if err != nil {
		return fmt.Errorf("failed to get active game to delete board: %v", err)
	}

	game.DeleteBoard(b)

	if err := a.DeleteBoard(b, game); err != nil {
		return fmt.Errorf("could not delete board from firestore: %s", err)
	}
	if err := cache.DeleteBoard(b); err != nil {
		return fmt.Errorf("could not delete board from cache: %s", err)
	}

	if err := cache.SaveGame(game); err != nil {
		return fmt.Errorf("could not reset game in cache: %s", err)
	}

	messages := []Message{}
	m := Message{}
	m.SetText("Your game is being reset")
	m.SetAudience(b.Player.Email)
	m.Operation = "reset"
	messages = append(messages, m)

	if err := a.AddMessagesToGame(game, messages); err != nil {
		return fmt.Errorf("could not send message to delete board: %s", err)
	}

	return nil
}

func generateBingoMessages(board Board, game Game, first bool) []Message {

	messages := []Message{}

	bingoMsg := "<strong>You</strong> already had <em><strong>BINGO</strong></em> on your board."
	dubiousMsg := fmt.Sprintf("<strong>%s</strong> might have just redeclared a dubious <em><strong>BINGO</strong></em> on their board.", board.Player.Name)

	if first {
		bingoMsg = fmt.Sprintf("<strong>%s</strong> just got <em><strong>BINGO</strong></em> on their board.", board.Player.Name)
		dubiousMsg = fmt.Sprintf("<strong>%s</strong> might have just declared a dubious <em><strong>BINGO</strong></em> on their board.", board.Player.Name)
	}

	m1 := Message{}
	m1.SetText(bingoMsg)
	m1.SetAudience(board.Player.Email)
	if first {
		m1.SetAudience("all", board.Player.Email)
	}

	m1.Bingo = true
	messages = append(messages, m1)

	reports := game.CheckBoard(board)
	if reports.IsDubious() {
		board.log("REPORTED BINGO IS DUBIOUS")
		m2 := Message{}
		m2.SetText(dubiousMsg)
		m2.SetAudience("admin", board.Player.Email)
		m2.Bingo = true
		messages = append(messages, m2)

		for _, v := range reports {
			mr := Message{}

			if v.Percent > .5 {
				mr.SetText("<strong>%s</strong> was selected by %d of the other %d players", v.Phrase.Text, v.Count-1, v.Total-1)
			} else {
				mr.SetText("<strong>%s</strong> was selected by only <strong>%d of the other %d players</strong>", v.Phrase.Text, v.Count-1, v.Total-1)
				if v.Count == 1 {
					mr.SetText("<strong>%s</strong> was selected by <strong>none</strong> of the other %d players", v.Phrase.Text, v.Total-1)
				}
			}

			mr.SetAudience("admin", board.Player.Email)
			mr.Bingo = true
			messages = append(messages, mr)
		}
	}
	return messages
}

func getGamesForKey(key string, limit int, token time.Time) (Games, error) {
	g := Games{}
	var err error

	g, err = cache.GetGamesForKey(key)
	if err != nil {
		if err == ErrCacheMiss {

			if strings.Contains(key, "admin-list") {
				g, err = a.GetGames(limit, token)
				if err != nil {
					return g, fmt.Errorf("error getting games: %v", err)
				}
			} else {
				g, err = a.GetGamesForKey(key)
				if err != nil {
					return g, fmt.Errorf("error getting games: %v", err)
				}
			}

		}
		if err := cache.SaveGamesForKey(key, g); err != nil {
			return g, fmt.Errorf("error caching games : %v", err)
		}
	}

	return g, nil
}

func getNewGame(name string, player Player) (Game, error) {

	game, err := a.NewGame(name, player)
	if err != nil {
		return game, fmt.Errorf("failed to get new game: %v", err)
	}
	if err := cache.DeleteGamesForKey([]string{player.Email, "admin-list"}); err != nil {
		return game, fmt.Errorf("failed to clear cache: %v", err)
	}
	if err := cache.SaveGame(game); err != nil {
		return game, fmt.Errorf("error caching game : %v", err)
	}

	return game, nil
}

func getGame(gid string) (Game, error) {
	game, err := cache.GetGame(gid)
	if err != nil {
		if err == ErrCacheMiss {
			game, err = a.GetGame(gid)
			if err != nil {
				return Game{}, fmt.Errorf("error getting game: %v", err)
			}
		}
		if err := cache.SaveGame(game); err != nil {
			return Game{}, fmt.Errorf("error caching game : %v", err)
		}
	}

	if len(game.Boards) == 0 {
		weblog("WARNING a game was retrieved without its boards - fixing")

		game, err = a.loadGameWithBoards(game)
		if err != nil {
			return Game{}, fmt.Errorf("error loading game : %v", err)
		}
		if err := cache.SaveGame(game); err != nil {
			return Game{}, fmt.Errorf("error caching game : %v", err)
		}

	}

	return game, nil
}

func deactivateGame(gid string) error {
	game, err := cache.GetGame(gid)
	if err != nil {
		if err == ErrCacheMiss {
			game, err = a.GetGame(gid)
			if err != nil {
				return fmt.Errorf("error getting game: %v", err)
			}
		}
	}
	game.Active = false

	if err := cache.SaveGame(game); err != nil {
		return fmt.Errorf("error caching game : %v", err)
	}

	if err := a.SaveGame(game); err != nil {
		return fmt.Errorf("error saving game to firestore : %v", err)
	}

	keys := []string{"admin-list"}
	for _, v := range game.Boards {
		keys = append(keys, v.Player.Email)
	}

	msg := fmt.Sprintf("Deleting games for caches: %+v", keys)
	cache.log(msg)
	if err := cache.DeleteGamesForKey(keys); err != nil {
		return fmt.Errorf("error caching game : %v", err)
	}

	return nil
}

func recordSelect(bid, gid, pid string, selected bool) error {
	p := Phrase{}
	p.ID = pid
	p.Selected = selected
	messages := []Message{}

	b, err := getBoard(bid, gid)
	if err != nil {
		return fmt.Errorf("could not get board id(%s): %s", bid, err)
	}

	g, err := getGame(b.Game)
	if err != nil {
		return fmt.Errorf("could not get game id(%s): %s", b.Game, err)
	}

	p = b.Select(p)
	r := g.Select(p, b.Player)
	bingo := b.Bingo()

	if err := a.SelectPhrase(b, p, r); err != nil {
		return fmt.Errorf("record click to firestore: %s", err)
	}

	if err := cache.SaveGame(g); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

	if err := cache.SaveBoard(b); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

	indicator := "unselected"
	if p.Selected {
		indicator = "selected"
	}

	m := Message{}
	m.SetText("<strong>%s</strong> %s <em>%s</em> on their board.", b.Player.Name, indicator, p.Text)
	m.SetAudience("admin", b.Player.Email)
	messages = append(messages, m)

	if bingo {
		msg := generateBingoMessages(b, g, true)
		messages = append(messages, msg...)
	}

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return fmt.Errorf("could not send message announce bingo on select: %s", err)
	}

	return nil
}

func updateMasterPhrase(phrase Phrase) error {

	if err := a.UpdateMasterPhrase(phrase); err != nil {
		return fmt.Errorf("error updating master phrase : %v", err)
	}

	return nil
}

func updateGamePhrases(gid string, phrase Phrase) error {
	messages := []Message{}
	m := Message{}
	m.SetText("A square has been changed and reset for all players. ")
	m.SetAudience("all")
	messages = append(messages, m)

	g, err := getGame(gid)
	if err != nil {
		return fmt.Errorf("could not get game id(%s): %s", g.ID, err)
	}

	bingos := make(map[string]Board)

	for _, v := range g.Boards {
		if v.Bingo() {
			bingos[v.ID] = v
		}
	}

	g.UpdatePhrase(phrase)

	for _, v := range bingos {
		if !v.Bingo() {
			m := Message{}
			m.SetText("An action from the <strong>game managers</strong> has rescinded your <em><strong>BINGO</strong></em>")
			m.SetAudience(v.Player.Email)
			m.Bingo = true
			messages = append(messages, m)

			m2 := Message{}
			m2.SetText("<strong>%s</strong> just lost their <em><strong>BINGO</strong></em>", v.Player.Name)
			m2.SetAudience("admin")
			m2.Bingo = true
			messages = append(messages, m2)
		}
	}

	if err := a.UpdatePhrase(g, phrase); err != nil {
		return fmt.Errorf("error saving update phrase in firebase: %v", err)
	}

	if err := cache.UpdatePhrase(g, phrase); err != nil {
		return fmt.Errorf("error saving update phrase in cache: %v", err)
	}

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return fmt.Errorf("could not send message announce bingo on select: %s", err)
	}

	return nil
}
