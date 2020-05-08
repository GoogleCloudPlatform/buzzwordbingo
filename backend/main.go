package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var (
	randseedfunc = randomseed
	a            = Agent{ProjectID: "bingo-collab"}
	port         = os.Getenv("PORT")
	boards       = make(map[string]Board)
	games        = make(map[string]Game)
)

func main() {
	// terry := Player{"Terrence Ryan", "tpryan@google.com", false}
	// jenn := Player{"Jenn Thomas", "thomasjennifer@google.com", false}
	// gameid := "PvleSnZLcW1g_8NJMPUQb7wKZXqwD9OhgMDJpTVkezM"
	// boardid := "H4smyjYHeDzjELBlmMS7Mt_k0Bidw6UHjg6IdFF8iwo"
	// phrase := Phrase{"101", "Achieve peak dead-pan", false, "", ""}

	// if err := recordSelect(boardid, phrase.ID); err != nil {
	// 	log.Fatalf("could select record: %s", err)
	// }

	// test, err := getBoardForPlayer(jenn)
	// if err != nil {
	// 	log.Fatalf("could select record: %s", err)
	// }
	// fmt.Printf("%s\n", test)
	// fmt.Printf("Done \n")

	if port == "" {
		port = ":8080"
	}

	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/healthz", handleHealth)
	http.HandleFunc("/api/board", handleGetBoard)
	http.HandleFunc("/api/record", handleRecordSelect)
	http.HandleFunc("/api/game", handleGetGame)
	http.HandleFunc("/api/game/new", handleNewGame)
	http.HandleFunc("/api/game/active", handleActiveGame)
	http.HandleFunc("/api/game/reset", handleResetActiveGame)

	log.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleGetBoard(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/board called\n")
	email, ok := r.URL.Query()["email"]

	if !ok || len(email[0]) < 1 || email[0] == "undefined" {
		msg := "{\"error\":\"email is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 || name[0] == "undefined" {
		msg := "{\"error\":\"name is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	p := Player{}
	p.Email = email[0]
	p.Name = name[0]

	board, err := getBoardForPlayer(p)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	json, err := board.JSON()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	writeResponse(w, http.StatusOK, json)

}

func handleNewGame(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/game/new called\n")

	name, ok := r.URL.Query()["name"]

	if !ok || len(name[0]) < 1 {
		msg := "{\"error\":\"name is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	game, err := getNewGame(name[0])
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	json, err := game.JSON()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	writeResponse(w, http.StatusOK, json)

}

func handleResetActiveGame(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/game/reset called\n")

	game, err := a.ResetActiveGame()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}
	boards = make(map[string]Board)
	games = make(map[string]Game)

	msg := fmt.Sprintf("{\"msg\":\"Game %s Reset\"}", game.ID)
	writeResponse(w, http.StatusOK, msg)

}

func handleActiveGame(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/game/active called\n")

	game, err := getActiveGame()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	json, err := game.JSON()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	writeResponse(w, http.StatusOK, json)

}

func handleGetGame(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/game called\n")

	id, ok := r.URL.Query()["id"]

	if !ok || len(id[0]) < 1 {
		msg := "{\"error\":\"id is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	game, err := getGame(id[0])
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	json, err := game.JSON()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	writeResponse(w, http.StatusOK, json)

}

func handleRecordSelect(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/record called\n")
	p, ok := r.URL.Query()["p"]

	if !ok || len(p[0]) < 1 {
		msg := "{\"error\":\"phrase is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	pid := p[0]

	b, ok := r.URL.Query()["b"]

	if !ok || len(b[0]) < 1 {
		msg := "{\"error\":\"board is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	bid := b[0]

	err := recordSelect(bid, pid)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	msg := fmt.Sprintf("{\"msg\":\"ok\"}")
	writeResponse(w, http.StatusOK, msg)

}

func getBoardForPlayer(p Player) (Board, error) {
	b := Board{}
	var ok bool

	game, err := getActiveGame()
	if err != nil {
		return b, fmt.Errorf("failed to get active game: %v", err)
	}

	b, ok = boards[game.ID+"_"+p.Email]
	if !ok {

		b, err = a.GetBoardForPlayer(game.ID, p.Email)
		if err != nil {
			return b, fmt.Errorf("error getting board for player: %v", err)
		}
		boards[game.ID+"_"+p.Email] = b

	}
	m := Message{}
	m.SetText("<strong>%s</strong> rejoined the game.", b.Player.Name)
	m.SetAudience("all")

	if b.ID == "" {
		b = game.NewBoard(p)
		b, err = a.SaveBoard(b)
		if err != nil {
			return b, fmt.Errorf("error saving board for player: %v", err)
		}
		boards[game.ID+"_"+p.Email] = b
		boards[b.ID] = b
		m.SetText("<strong>%s</strong> got a board and joined the game.", b.Player.Name)
		m.SetAudience("all", b.Player.Email)

	}

	if err := a.AddMessageToGame(game, m); err != nil {
		return b, fmt.Errorf("could not send message: %s", err)
	}

	bingo := b.Bingo()
	if bingo {
		m := Message{}
		m.SetText("<strong>%s</strong> already had <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
		m.SetAudience("all", b.Player.Email)
		m.Bingo = bingo

		if err := a.AddMessageToGame(game, m); err != nil {
			return b, fmt.Errorf("could not send message: %s", err)
		}
	}

	return b, nil
}

func getBoard(bid string) (Board, error) {
	b := Board{}
	var ok bool
	var err error

	b, ok = boards[bid]
	if !ok {
		b, err = a.GetBoard(bid)
		if err != nil {
			return b, fmt.Errorf("could not get board from firestore: %s", err)
		}
		boards[bid] = b

	}

	return b, nil
}

func getNewGame(name string) (Game, error) {
	game, err := a.NewGame(name)
	if err != nil {
		return game, fmt.Errorf("failed to get active game: %v", err)
	}

	return game, nil
}

func getActiveGame() (Game, error) {
	var err error
	game, ok := games["active"]
	if !ok {
		game, err = a.GetActiveGame()
		if err != nil {
			return game, fmt.Errorf("failed to get active game: %v", err)
		}

		games["active"] = game
		games[game.ID] = game
	}

	return game, nil
}

func getGame(id string) (Game, error) {
	game, ok := games[id]
	if !ok {
		game, err := a.GetGame(id)
		if err != nil {
			return game, fmt.Errorf("could not get game from cache or id(%s) from firestore: %s", id, err)
		}
		games[id] = game
	}
	return game, nil
}

func recordSelect(boardID string, phraseID string) error {
	p := Phrase{}
	p.ID = phraseID
	fmt.Printf("BoardID %v\n", boardID)

	b, err := getBoard(boardID)
	if err != nil {
		return fmt.Errorf("could not get board id(%s): %s", boardID, err)
	}

	g, err := getGame(b.Game)
	if err != nil {
		return fmt.Errorf("could not get game id(%s): %s", b.Game, err)
	}

	p = b.Select(p)
	record := g.Master.Select(p, b.Player)
	games[g.ID] = g
	boards[b.ID] = b

	if err := a.UpdatePhraseOnBoard(b, p); err != nil {
		return fmt.Errorf("could not update board to firestore: %s", err)
	}

	if err := a.UpdateRecordOnGame(g, record); err != nil {
		return fmt.Errorf("could not update game to firestore: %s", err)
	}

	indicator := "unselected"
	if p.Selected {
		indicator = "selected"
	}

	m := Message{}
	m.SetText("<strong>%s</strong> %s <em>%s</em> on their board.", b.Player.Name, indicator, p.Text)
	m.SetAudience("all")

	if err := a.AddMessageToGame(g, m); err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	bingo := b.Bingo()
	if bingo {
		boards[b.ID] = b
		boards[g.ID+"_"+b.Player.Email] = b
		m := Message{}
		m.SetText("<strong>%s</strong> just got <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
		m.SetAudience("all")
		m.Bingo = bingo

		if err := a.UpdateBingoOnBoard(b, bingo); err != nil {
			return fmt.Errorf("could not record bingo on board: %s", err)
		}

		if err := a.AddMessageToGame(g, m); err != nil {
			return fmt.Errorf("could not send message: %s", err)
		}
	}

	return nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, http.StatusOK, "ok")
	return
}

func writeResponse(w http.ResponseWriter, code int, msg string) {

	if code != http.StatusOK {
		log.Printf(msg)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write([]byte(msg))

	return
}
