package main

import "fmt"

func getBoardForPlayer(p Player, g Game) (Board, error) {
	var err error
	b := Board{}
	messages := []Message{}
	weblog("Trying Cache")
	b, err = cache.GetBoard(g.ID + "_" + p.Email)
	if err != nil {
		if err == ErrCacheMiss {
			weblog("Cache Empty trying DB")
			b, err = a.GetBoardForPlayer(g.ID, p)
			if err != nil {
				return b, fmt.Errorf("error getting board for player: %v", err)
			}
		}
		if err := cache.DeleteGamesForKey([]string{p.Email}); err != nil {
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
		b = g.NewBoard(p)
		b, err = a.SaveBoard(b)
		if err != nil {
			return b, fmt.Errorf("error saving board for player: %v", err)
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board for player: %v", err)
		}
		m.SetText("<strong>%s</strong> got a board and joined the game.", b.Player.Name)
		m.SetAudience("all")

	}

	messages = append(messages, m)

	bingo := b.Bingo()
	if bingo {
		msg := generateBingoMessages(b, g, false)
		messages = append(messages, msg...)
	}

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return b, fmt.Errorf("could not send message to notify player of bingo: %s", err)
	}

	return b, nil
}

func generateBingoMessages(b Board, g Game, first bool) []Message {

	messages := []Message{}

	bingoMsg := "<strong>You</strong> already had <em><strong>BINGO</strong></em> on your board."
	dubiousMsg := fmt.Sprintf("<strong>%s</strong> might have just redeclared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)

	if first {
		bingoMsg = fmt.Sprintf("<strong>%s</strong> just got <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
		dubiousMsg = fmt.Sprintf("<strong>%s</strong> might have just declared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
	}

	m1 := Message{}
	m1.SetText(bingoMsg)
	m1.SetAudience(b.Player.Email)
	if first {
		m1.SetAudience("all", b.Player.Email)
	}

	m1.Bingo = true
	messages = append(messages, m1)

	reports := g.CheckBoard(b)
	if reports.IsDubious() {
		b.log("REPORTED BINGO IS DUBIOUS")
		m2 := Message{}
		m2.SetText(dubiousMsg)
		m2.SetAudience("admin", b.Player.Email)
		m2.Bingo = true
		messages = append(messages, m2)

		for _, v := range reports {
			mr := Message{}
			mr.SetText("<strong>%s</strong> was selected by %d of %d other players", v.Phrase.Text, v.Count, v.Total-1)
			mr.SetAudience("admin", b.Player.Email)
			mr.Bingo = true
			messages = append(messages, mr)
		}
	}
	return messages
}

func getGamesForKey(key string) (Games, error) {
	g := Games{}
	var err error

	g, err = cache.GetGamesForKey(key)
	if err != nil {
		if err == ErrCacheMiss {

			if key == "admin-list" {
				g, err = a.GetGames()
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
		if err := cache.SaveGamesForKey("games"+key, g); err != nil {
			return g, fmt.Errorf("error caching games : %v", err)
		}
	}

	return g, nil
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

	if err := a.DeleteBoard(bid, game.ID); err != nil {
		return fmt.Errorf("could not delete board from firestore: %s", err)
	}
	if err := cache.DeleteBoard(b); err != nil {
		return fmt.Errorf("could not delete board from cache: %s", err)
	}

	game, err = a.GetGame(game.ID)
	if err != nil {
		return fmt.Errorf("failed to get updated game from database: %v", err)
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

func getNewGame(name string, p Player) (Game, error) {

	game, err := a.NewGame(name, p)
	if err != nil {
		return game, fmt.Errorf("failed to get new game: %v", err)
	}
	if err := cache.DeleteGamesForKey([]string{p.Email}); err != nil {
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

	if err := cache.DeleteGamesForKey(keys); err != nil {
		return fmt.Errorf("error caching game : %v", err)
	}

	return nil
}

func recordSelect(boardID, gameID, phraseID string, selected bool) error {
	p := Phrase{}
	p.ID = phraseID
	p.Selected = selected
	messages := []Message{}

	b, err := getBoard(boardID, gameID)
	if err != nil {
		return fmt.Errorf("could not get board id(%s): %s", boardID, err)
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

func updateGamePhrases(gameID string, phrase Phrase) error {
	g, err := getGame(gameID)
	if err != nil {
		return fmt.Errorf("could not get game id(%s): %s", g.ID, err)
	}

	if err := a.UpdatePhrase(g, phrase); err != nil {
		return fmt.Errorf("error saving update phrase in firebase: %v", err)
	}

	if err := cache.UpdatePhrase(g, phrase); err != nil {
		return fmt.Errorf("error saving update phrase in cache: %v", err)
	}

	messages := []Message{}
	m := Message{}
	m.SetText("A square has been changed and reset for all players. ")
	m.SetAudience("all")
	messages = append(messages, m)

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return fmt.Errorf("could not send message announce bingo on select: %s", err)
	}

	return nil
}
