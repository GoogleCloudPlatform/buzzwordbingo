package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/idtoken"
)

var (
	randseedfunc  = randomseed
	a             = Agent{}
	cache         = Cache{}
	port          = ":8080"
	boards        = make(map[string]Board)
	games         = make(map[string]Game)
	ctx           = context.Background()
	noisy         = true
	projectID     = ""
	projectNumber = ""
	client        *firestore.Client
	errNotAdmin   = fmt.Errorf("not an admin")
)

func main() {
	var err error
	projectID, err = getProjectID()
	if err != nil {
		log.Fatal(err)
	}

	projectNumber, err = getProjectNumber(projectID)
	if err != nil {
		log.Fatal(err)
	}

	a.ProjectID = projectID

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

	weblog(fmt.Sprintf("Starting server on port %s\n", port))
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleGetIsAdmin(w http.ResponseWriter, r *http.Request) {
	weblog("/api/player/isadmin called")
	isAdm, err := isAdmin(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("%t", isAdm))

}

func handleGetIAPUsername(w http.ResponseWriter, r *http.Request) {
	weblog("/api/player/identify called")
	p := Player{}

	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	p.Email = email

	writeJSON(w, p)

}

func handleGetBoard(w http.ResponseWriter, r *http.Request) {
	weblog("/api/board called")
	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	name, err := getFirstQuery("name", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	p := Player{}
	p.Email = email
	p.Name = name

	board, err := getBoardForPlayer(p)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, board)
	return
}

func handleDeleteBoard(w http.ResponseWriter, r *http.Request) {
	weblog("/api/board/delete called")

	isAdm, err := isAdmin(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	b, err := getFirstQuery("b", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if err := deleteBoard(b); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleNewGame(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/new called")

	isAdm, err := isAdmin(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	name, err := getFirstQuery("name", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	game, err := getNewGame(name)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, game)
	return
}

func handleResetActiveGame(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/reset called")

	isAdm, err := isAdmin(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	game, err := resetGame()
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, fmt.Sprintf("Game %s Reset", game.ID))

}

func handleActiveGame(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/active called")

	game, err := getActiveGame()
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, game)

}

func handleGetGame(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game called")

	id, err := getFirstQuery("id", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	game, err := getGame(id)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, game)

}

func handleRecordSelect(w http.ResponseWriter, r *http.Request) {
	weblog("/api/record called")

	pid, err := getFirstQuery("p", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	bid, err := getFirstQuery("b", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if err := recordSelect(bid, pid); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")

}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, "ok")
	return
}

type JSONProducer interface {
	JSON() (string, error)
}

func writeJSON(w http.ResponseWriter, j JSONProducer) {
	json, err := j.JSON()
	if err != nil {
		writeError(w, err.Error())
		return
	}
	writeResponse(w, http.StatusOK, json)
	return
}

func writeSuccess(w http.ResponseWriter, msg string) {
	s := fmt.Sprintf("{\"msg\":\"%s\"}", msg)
	writeResponse(w, http.StatusOK, s)
	return
}

func writeError(w http.ResponseWriter, msg string) {
	s := fmt.Sprintf("{\"error\":\"%s\"}", msg)
	writeResponse(w, http.StatusInternalServerError, s)
	return
}

func writeResponse(w http.ResponseWriter, code int, msg string) {

	if code != http.StatusOK {
		weblog(fmt.Sprintf(msg))
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

func getFirstQuery(query string, r *http.Request) (string, error) {
	result, ok := r.URL.Query()[query]

	if !ok || len(result[0]) < 1 || result[0] == "undefined" {
		return "", fmt.Errorf("query parameter %s is missing", query)
	}
	return result[0], nil
}

func getProjectID() (string, error) {
	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		return "", fmt.Errorf("could not determine this project id: %v", err)
	}
	return credentials.ProjectID, nil
}

func getProjectNumber(projectID string) (string, error) {

	c, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
	if err != nil {
		return "", fmt.Errorf("could not get cloudresourcemanager client: %v", err)
	}

	svc, err := cloudresourcemanager.New(c)
	if err != nil {
		return "", fmt.Errorf("could not get cloudresourcemanager service: %v", err)
	}

	resp, err := svc.Projects.Get(projectID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("could not get project data: %v", err)
	}

	return strconv.Itoa(int(resp.ProjectNumber)), nil
}

func validateJWT(iapJWT, projectNumber, projectID string) (*idtoken.Payload, error) {
	aud := fmt.Sprintf("/projects/%s/apps/%s", projectNumber, projectID)
	return idtoken.Validate(ctx, iapJWT, aud)
}

func isAdmin(r *http.Request) (bool, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return false, err
	}
	result, err := a.IsAdmin(email)
	if err != nil {
		return false, err
	}
	return result, nil
}

func getPlayerEmail(r *http.Request) (string, error) {
	email, err := getValidatedEmail(r)
	if err != nil {
		return "", err
	}

	// If it's not behind IAP, it's developemnt
	if email == "" {
		username := os.Getenv("USER")
		email = fmt.Sprintf("%s@example.com", username)
	}

	return email, nil
}

func getValidatedEmail(r *http.Request) (string, error) {
	arr := r.Header.Get("X-Goog-Authenticated-User-Email")

	if arr == "" {
		return "", nil
	}

	jwt := r.Header.Get("X-Goog-IAP-JWT-Assertion")

	payload, err := validateJWT(jwt, projectNumber, projectID)
	if err != nil {
		return "", fmt.Errorf("could not validate IAP JWT: %s", err)
	}

	return payload.Email, nil

}

func getBoardForPlayer(p Player) (Board, error) {
	b := Board{}
	messages := []Message{}

	game, err := getActiveGame()
	if err != nil {
		return b, fmt.Errorf("failed to get active game: %v", err)
	}

	b, err = cache.GetBoard(game.ID + "_" + p.Email)
	if err != nil {
		if err == ErrCacheMiss {
			b, err = a.GetBoardForPlayer(game.ID, p)
			if err != nil {
				return b, fmt.Errorf("error getting board for player: %v", err)
			}
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board for player: %v", err)
		}
	}

	m := Message{}
	m.SetText("<strong>%s</strong> rejoined the game.", b.Player.Name)
	m.SetAudience("admin", b.Player.Email)

	if b.ID == "" {
		b = game.NewBoard(p)
		b, err = a.SaveBoard(b)
		if err != nil {
			return b, fmt.Errorf("error saving board for player: %v", err)
		}
		if err := cache.SaveBoard(b); err != nil {
			return b, fmt.Errorf("error caching board for player: %v", err)
		}
		m.SetText("<strong>%s</strong> got a board and joined the game.", b.Player.Name)
		m.SetAudience("all", b.Player.Email)

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

func generateBingoMessages(b Board, g Game, first bool) []Message {

	bingoMsg := "<strong>You</strong> already had <em><strong>BINGO</strong></em> on your board."
	dubiousMsg := fmt.Sprintf("<strong>%s</strong> might have just redeclared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)

	if first {
		bingoMsg = fmt.Sprintf("<strong>%s</strong> just got <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
		dubiousMsg = fmt.Sprintf("<strong>%s</strong> might have just declared a dubious <em><strong>BINGO</strong></em> on their board.", b.Player.Name)
	}

	messages := []Message{}

	m1 := Message{}
	m1.SetText(bingoMsg)
	m1.SetAudience(b.Player.Email)
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

func getBoard(bid string) (Board, error) {
	b := Board{}

	game, err := getActiveGame()
	if err != nil {
		return b, fmt.Errorf("could not get active game for board: %s", err)
	}

	b, err = cache.GetBoard(bid)
	if err != nil {
		if err == ErrCacheMiss {
			b, err = a.GetBoard(bid, game.ID)
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

func deleteBoard(bid string) error {
	b, err := getBoard(bid)
	if err != nil {
		return fmt.Errorf("could not retrieve board from firestore: %s", err)
	}

	game, err := getActiveGame()
	if err != nil {
		return fmt.Errorf("failed to get active game to delete board: %v", err)
	}

	if err := cache.DeleteBoard(b); err != nil {
		return fmt.Errorf("could not delete board from cache: %s", err)
	}

	if err := a.DeleteBoard(bid, game.ID); err != nil {
		return fmt.Errorf("could not delete board from firestore: %s", err)
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
	return getGame("active")
}

func getGame(id string) (Game, error) {
	game, err := cache.GetGame(id)
	if err != nil {
		if err == ErrCacheMiss {
			game, err = a.GetActiveGame()
			if err != nil {
				return Game{}, fmt.Errorf("error getting game: %v", err)
			}
		}
		if err := cache.SaveGame(game); err != nil {
			return Game{}, fmt.Errorf("error caching game : %v", err)
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

	if err := cache.SaveGame(g); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

	if err := cache.SaveBoard(b); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

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
		msg := generateBingoMessages(b, g, false)
		messages = append(messages, msg...)
	}

	if err := a.AddMessagesToGame(g, messages); err != nil {
		return fmt.Errorf("could not send message announce bingo on select: %s", err)
	}

	return nil
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
	cache.Clear()

	return a.ResetActiveGame()
}

func weblog(msg string) {
	if noisy {
		log.Printf("Webserver: %s", msg)
	}
}
