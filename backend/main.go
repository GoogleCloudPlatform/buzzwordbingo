package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	randseedfunc = randomseed
	a            = Agent{ProjectID: "bingo-collab"}
	port         = ":8080"
	boards       = make(map[string]Board)
	games        = make(map[string]Game)
)

func main() {

	fs := wrapHandler(http.FileServer(http.Dir("./static")))
	http.HandleFunc("/", fs)
	http.HandleFunc("/healthz", handleHealth)
	http.HandleFunc("/api/board", handleGetBoard)
	http.HandleFunc("/api/board/delete", handleDeleteBoard)
	http.HandleFunc("/api/record", handleRecordSelect)
	http.HandleFunc("/api/game", handleGetGame)
	http.HandleFunc("/api/game/new", handleNewGame)
	http.HandleFunc("/api/game/active", handleActiveGame)
	http.HandleFunc("/api/game/reset", handleResetActiveGame)
	http.HandleFunc("/api/player/identify", handleGetIAPUsername)
	http.HandleFunc("/api/player/isadmin", handleGetIsAdmin)

	log.Printf("Starting server on port %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleGetIsAdmin(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/player/isadmin called\n")
	email, ok := r.URL.Query()["email"]

	if !ok || len(email[0]) < 1 || email[0] == "undefined" {
		msg := "{\"error\":\"email is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	result, err := a.IsAdmin(email[0])
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	msg := fmt.Sprintf("%t", result)
	writeResponse(w, http.StatusOK, msg)

}

func handleGetIAPUsername(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/player/identify called\n")
	p := Player{}

	arr := r.Header.Get("X-Goog-Authenticated-User-Email")
	email := getEmailFromString(arr)

	if email == "" {
		username := os.Getenv("USER")
		email = fmt.Sprintf("%s@google.com_", username)
	}

	p.Email = email

	json, err := p.JSON()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	writeResponse(w, http.StatusOK, json)

}

func getEmailFromString(arr string) string {
	email := ""
	if len(arr) > 0 {

		iapstrings := strings.Split(arr, ":")
		if len(iapstrings) < 2 {
			return email
		}

		email = iapstrings[1]
	}
	return email
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

func handleDeleteBoard(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("/api/board/delete called\n")
	b, ok := r.URL.Query()["b"]

	if !ok || len(b[0]) < 1 || b[0] == "undefined" {
		msg := "{\"error\":\"b is missing\"}"
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	if err := deleteBoard(b[0]); err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	msg := fmt.Sprintf("{\"msg\":\"ok\"}")
	writeResponse(w, http.StatusOK, msg)

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

	game, err := resetGame()
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, http.StatusInternalServerError, msg)
		return
	}

	msg := fmt.Sprintf("{\"msg\":\"Game %s Reset\"}", game.ID)
	writeResponse(w, http.StatusOK, msg)

}

func resetGame() (Game, error) {

	game, err := getActiveGame()
	if err != nil {
		return Game{}, err
	}

	if game.ID != "" {
		messages := []Message{}
		m := Message{}
		m.SetText("Your game is being reset")
		m.SetAudience("all")
		m.Operation = "reset"
		messages = append(messages, m)

		if err := a.AddMessagesToGame(game, messages); err != nil {
			return game, fmt.Errorf("could not send message to reset: %s", err)
		}
	}
	boards = make(map[string]Board)
	games = make(map[string]Game)

	game, err = a.ResetActiveGame()
	if err != nil {
		return Game{}, err
	}

	return game, nil
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
	messages := []Message{}

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

	b.Player = p
	m := Message{}
	m.SetText("<strong>%s</strong> rejoined the game.", b.Player.Name)
	m.SetAudience("admin", b.Player.Email)

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

	messages = append(messages, m)

	bingo := b.Bingo()
	if bingo {
		m2 := Message{}
		m2.SetText("<strong>You</strong> already had <em><strong>BINGO</strong></em> on your board.")
		m2.SetAudience(b.Player.Email)
		m2.Bingo = bingo
		messages = append(messages, m2)

		reports := game.CheckBoard(b)
		if reports.IsDubious() {
			b.log("REPORTED BINGO IS DUBIOUS")
			m3 := Message{}
			m3.SetText("<strong>%s</strong> might have just redeclared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
			m3.SetAudience("admin", b.Player.Email)
			m3.Bingo = bingo
			messages = append(messages, m3)

			for _, v := range reports {
				mr := Message{}
				mr.SetText("<strong>%s</strong> was selected by %d of %d other players", v.Phrase.Text, v.Count, v.Total-1)
				mr.SetAudience("admin", b.Player.Email)
				mr.Bingo = bingo
				messages = append(messages, mr)
			}

		}
	}

	if err := a.AddMessagesToGame(game, messages); err != nil {
		return b, fmt.Errorf("could not send message to notify player of bingo: %s", err)
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

func deleteBoard(bid string) error {
	b, err := getBoard(bid)
	if err != nil {
		return fmt.Errorf("could not retrieve board from firestore: %s", err)
	}

	game, err := getActiveGame()
	if err != nil {
		return fmt.Errorf("failed to get active game: %v", err)
	}

	b.log(fmt.Sprintf("Cleaning from cache %s", bid))
	b.log(fmt.Sprintf("Cleaning from cache %s", b.Player.Email))
	delete(boards, bid)
	delete(boards, game.ID+"_"+b.Player.Email)

	if err := a.DeleteBoard(bid); err != nil {
		return fmt.Errorf("could not get board from firestore: %s", err)
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

func getNewGame(name string) (Game, error) {

	_, err := resetGame()
	if err != nil {
		return Game{}, err
	}

	game, err := a.NewGame(name)
	if err != nil {
		return game, fmt.Errorf("failed to get new game: %v", err)
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
		if game.Active {
			games["active"] = game
		}
	}
	return game, nil
}

func recordSelect(boardID string, phraseID string) error {
	p := Phrase{}
	p.ID = phraseID
	messages := []Message{}

	b, err := getBoard(boardID)
	if err != nil {
		return fmt.Errorf("could not get board id(%s): %s", boardID, err)
	}

	g, err := getGame(b.Game)
	if err != nil {
		return fmt.Errorf("could not get game id(%s): %s", b.Game, err)
	}

	p = b.Select(p)
	r := g.Master.Select(p, b.Player)
	bingo := b.Bingo()

	games[g.ID] = g
	games["active"] = g
	boards[b.ID] = b
	boards[g.ID+"_"+b.Player.Email] = b

	if err := a.UpdatePhrase(b, p, r); err != nil {
		return fmt.Errorf("record click to firestore: %s", err)
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
		m2 := Message{}
		m2.SetText("<strong>%s</strong> just got <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
		m2.SetAudience("all", b.Player.Email)
		m2.Bingo = bingo
		messages = append(messages, m2)

		reports := g.CheckBoard(b)
		if reports.IsDubious() {
			b.log("REPORTED BINGO IS DUBIOUS")
			m3 := Message{}
			m3.SetText("<strong>%s</strong> might have just declared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
			m3.SetAudience("admin", b.Player.Email)
			m3.Bingo = bingo
			messages = append(messages, m3)

			for _, v := range reports {
				mr := Message{}
				mr.SetText("<strong>%s</strong> was selected by %d of %d other players", v.Phrase.Text, v.Count, v.Total-1)
				mr.SetAudience("admin", b.Player.Email)
				mr.Bingo = bingo
				messages = append(messages, mr)
			}
		}

	}

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return fmt.Errorf("could not send message announce bingo on select: %s", err)
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

type NotFoundRedirectRespWr struct {
	http.ResponseWriter // We embed http.ResponseWriter
	status              int
}

func (w *NotFoundRedirectRespWr) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *NotFoundRedirectRespWr) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &NotFoundRedirectRespWr{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == 404 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
	}
}
