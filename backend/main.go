package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/idtoken"
)

var (
	randseedfunc  = randomseed
	a             Agent
	cache         Cache
	port          = ":8080"
	noisy         = true
	projectID     = ""
	projectNumber = ""
	// ErrNotAdmin is an error that indicates that the user is not an admin
	ErrNotAdmin = fmt.Errorf("not an admin")
	ctx         = context.Background()
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

	a, err = NewAgent(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	cache, err = NewCache()
	if err != nil {
		log.Fatal(err)
	}

	fs := wrapHandler(http.FileServer(http.Dir("./static")))
	http.HandleFunc("/", fs)
	http.HandleFunc("/healthz", handleHealth)
	http.HandleFunc("/api/board", handleBoardGet)
	http.HandleFunc("/api/board/delete", handleBoardDelete)
	http.HandleFunc("/api/record", handleRecordSelect)
	http.HandleFunc("/api/game", handleGameGet)
	http.HandleFunc("/api/game/new", handleGameNew)
	http.HandleFunc("/api/game/list", handleGameList)
	http.HandleFunc("/api/game/admin/add", handleGameAdminAdd)
	http.HandleFunc("/api/game/admin/remove", handleGameAdminDelete)
	http.HandleFunc("/api/game/deactivate", handleGameDeactivate)
	http.HandleFunc("/api/game/phrase/update", handleGamePhraseUpdate)
	http.HandleFunc("/api/phrase/update", handleMasterPhraseUpdate)
	http.HandleFunc("/api/game/isadmin", handleIsGameAdmin)
	http.HandleFunc("/api/player/identify", handleIAPUsernameGet)
	http.HandleFunc("/api/player/isadmin", handleIsAdmin)
	http.HandleFunc("/api/admin/add", handleAdminAdd)
	http.HandleFunc("/api/admin/remove", handleAdminDelete)
	http.HandleFunc("/api/admin/list", handleAdminList)

	weblog(fmt.Sprintf("Starting server on port %s\n", port))
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleIsAdmin(w http.ResponseWriter, r *http.Request) {
	weblog("/api/player/isadmin called")
	isAdm, err := isAdmin(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("%t", isAdm))

}

func handleIsGameAdmin(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/isadmin called")

	gid, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	isAdm, err := isGameAdmin(r, gid)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("%t", isAdm))

}

func handleGamePhraseUpdate(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/phrase/update called")

	gid, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	pid, err := getFirstQuery("p", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	text, err := getFirstQuery("text", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	isAdm, err := isGameAdmin(r, gid)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	phrase := Phrase{}
	phrase.ID = pid
	phrase.Text = text

	if err := updateGamePhrases(gid, phrase); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return

}

func handleMasterPhraseUpdate(w http.ResponseWriter, r *http.Request) {
	weblog("/api/phrase/update called")

	pid, err := getFirstQuery("p", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	text, err := getFirstQuery("text", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

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

	phrase := Phrase{}
	phrase.ID = pid
	phrase.Text = text

	if err := updateMasterPhrase(phrase); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return

}

func handleIAPUsernameGet(w http.ResponseWriter, r *http.Request) {
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

func handleGameList(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/list called")
	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	games, err := a.GetGamesForPlayer(email)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, games)
	return
}

func handleBoardGet(w http.ResponseWriter, r *http.Request) {
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

	gid, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	p := Player{}
	p.Email = email
	p.Name = name

	g, err := getGame(gid)
	if err != nil {
		writeError(w, err.Error())
	}

	board, err := getBoardForPlayer(p, g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, board)
	return
}

func handlePreflight(w http.ResponseWriter, method string) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", method)
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func handleBoardDelete(w http.ResponseWriter, r *http.Request) {
	weblog("/api/board/delete called")

	if r.Method == http.MethodOptions {
		handlePreflight(w, "DELETE")
		return
	}
	if r.Method != http.MethodDelete {
		msg := fmt.Sprintf("{\"error\":\"Must use http method DELETE you had %s\"}", r.Method)
		writeResponse(w, http.StatusMethodNotAllowed, msg)
		return
	}

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

	g, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if err := deleteBoard(b, g); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleGameNew(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/new called")

	name, err := getFirstQuery("name", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	game, err := getNewGame(name, email)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, game)
	return
}

func handleGameGet(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game called")

	g, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	game, err := getGame(g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, game)

}

func handleGameDeactivate(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/deactivate called")

	g, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	isAdm, err := isGameAdmin(r, g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	if err := deactivateGame(g); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")

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

	gid, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if err := recordSelect(bid, gid, pid); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")

}

func handleGameAdminAdd(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/admin/add called")
	if r.Method == http.MethodOptions {
		handlePreflight(w, "POST")
		return
	}
	if r.Method != http.MethodPost {
		msg := fmt.Sprintf("{\"error\":\"Must use http method POST you had %s\"}", r.Method)
		writeResponse(w, http.StatusMethodNotAllowed, msg)
		return
	}

	if err := r.ParseMultipartForm(160000); err != nil {
		writeError(w, err.Error())
		return
	}

	g := r.Form.Get("g")
	email := r.Form.Get("email")

	if email == "" {
		writeError(w, "email is required")
		return
	}

	if g == "" {
		writeError(w, "g is required")
		return
	}

	isAdm, err := isGameAdmin(r, g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	game, err := getGame(g)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	p := Player{}
	p.Email = email

	game.Admins.Add(p)

	if err := cache.SaveGame(game); err != nil {
		writeError(w, err.Error())
		return
	}

	if err := a.SaveGame(game); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleGameAdminDelete(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		handlePreflight(w, "DELETE")
		return
	}
	if r.Method != http.MethodDelete {
		msg := fmt.Sprintf("{\"error\":\"Must use http method DELETE you had %s\"}", r.Method)
		writeResponse(w, http.StatusMethodNotAllowed, msg)
		return
	}

	g, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	email, err := getFirstQuery("email", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	isAdm, err := isGameAdmin(r, g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if !isAdm {
		msg := fmt.Sprintf("{\"error\":\"Not an admin\"}")
		writeResponse(w, http.StatusForbidden, msg)
		return
	}

	game, err := getGame(g)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	p := Player{}
	p.Email = email

	new := game.Admins.Remove(p)
	game.Admins = new

	if err := cache.SaveGame(game); err != nil {
		writeError(w, err.Error())
		return
	}

	if err := a.SaveGame(game); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleAdminAdd(w http.ResponseWriter, r *http.Request) {
	weblog("/api/admin/add called")
	if r.Method == http.MethodOptions {
		handlePreflight(w, "POST")
		return
	}
	if r.Method != http.MethodPost {
		msg := fmt.Sprintf("{\"error\":\"Must use http method POST you had %s\"}", r.Method)
		writeResponse(w, http.StatusMethodNotAllowed, msg)
		return
	}

	if err := r.ParseMultipartForm(160000); err != nil {
		writeError(w, err.Error())
		return
	}

	email := r.Form.Get("email")

	if email == "" {
		writeError(w, "email is required")
		return
	}

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
	p := Player{}
	p.Email = email

	if err := a.AddAdmin(p); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleAdminDelete(w http.ResponseWriter, r *http.Request) {
	weblog("/api/admin/delete called")
	if r.Method == http.MethodOptions {
		handlePreflight(w, "DELETE")
		return
	}
	if r.Method != http.MethodDelete {
		msg := fmt.Sprintf("{\"error\":\"Must use http method DELETE you had %s\"}", r.Method)
		writeResponse(w, http.StatusMethodNotAllowed, msg)
		return
	}

	email, err := getFirstQuery("email", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

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

	p := Player{}
	p.Email = email

	if err := a.DeleteAdmin(p); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")
	return
}

func handleAdminList(w http.ResponseWriter, r *http.Request) {
	weblog("/api/admin/list called")

	players, err := a.GetAdmins()
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, players)
	return
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, "ok")
	return
}

// JSONProducer is an interface that spits out a JSON string version of itself
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

func isGameAdmin(r *http.Request, gid string) (bool, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return false, err
	}

	game, err := getGame(gid)
	if err != nil {
		return false, err
	}

	p := Player{}
	p.Email = email

	result := game.IsAdmin(p)
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
		email = fmt.Sprintf("%s@google.com", username)

	}

	return email, nil
}

func getValidatedEmail(r *http.Request) (string, error) {
	arr := r.Header.Get("X-Goog-Authenticated-User-Email")

	email := getEmailFromString(arr)
	if arr == "" {
		return "", nil
	}

	jwt := r.Header.Get("X-Goog-IAP-JWT-Assertion")

	_, err := validateJWT(jwt, projectNumber, projectID)
	if err != nil {
		return "", fmt.Errorf("could not validate IAP JWT: %s", err)
	}

	// TODO: Get rid of this and use the jwt token when it is done correctly
	return email, nil

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
		m.SetAudience("all", b.Player.Email)

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

func getBoard(bid string, gid string) (Board, error) {
	b := Board{}

	game, err := getGame(gid)
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

	if err := cache.SaveGame(game); err != nil {
		return fmt.Errorf("could not reset game in cache: %s", err)
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

func getNewGame(name, email string) (Game, error) {

	p := Player{}
	p.Email = email

	game, err := a.NewGame(name, p)
	if err != nil {
		return game, fmt.Errorf("failed to get new game: %v", err)
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

	return nil
}

func recordSelect(boardID, gameID, phraseID string) error {
	p := Phrase{}
	p.ID = phraseID
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

	if err := cache.SaveGame(g); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

	if err := cache.SaveBoard(b); err != nil {
		return fmt.Errorf("could not cache game: %s", err)
	}

	if err := a.SelectPhrase(b, p, r); err != nil {
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
	// TODO: Compenstate for display order somewhere in there.
	g.UpdatePhrase(phrase)

	if err := cache.SaveGame(g); err != nil {
		return fmt.Errorf("error caching game : %v", err)
	}

	for _, v := range g.Boards {
		b, err := getBoard(v.ID, g.ID)
		if err != nil {
			if strings.Contains(err.Error(), "failed to get board") {
				continue
			}
			return fmt.Errorf("could not get board id(%s): %s", v.ID, err)
		}
		b.UpdatePhrase(phrase)

		if err := cache.SaveBoard(b); err != nil {
			return fmt.Errorf("error caching board : %v", err)
		}
	}

	if err := a.UpdatePhrase(g, phrase); err != nil {
		return fmt.Errorf("error saving update phrase : %v", err)
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

func weblog(msg string) {
	if noisy {
		log.Printf("Webserver: %s", msg)
	}
}
