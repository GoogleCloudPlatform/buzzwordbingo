package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/idtoken"
)

var (
	randseedfunc  = randomseed
	a             Agent
	cache         Cache
	cacheEnabled  = true
	port          = ":8080"
	noisy         = true
	projectID     = ""
	projectNumber = ""
	// ErrNotAdmin is an error that indicates that the user is not an admin
	ErrNotAdmin = fmt.Errorf("not an admin or game admin")
	ctx         = context.Background()
)

func main() {
	var err error

	redisHost := os.Getenv("REDISHOST")
	redisPort := os.Getenv("REDISPORT")

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

	cache, err = NewCache(redisHost, redisPort, cacheEnabled)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/healthz", handleHealth)
	r.HandleFunc("/api/board", handleBoardGet)
	r.HandleFunc("/api/board/delete", handleBoardDelete)
	r.HandleFunc("/api/record", handleRecordSelect)
	r.HandleFunc("/api/game", handleGameGet)
	r.HandleFunc("/api/game/new", handleGameNew)
	r.HandleFunc("/api/game/list", handleGameList)
	r.HandleFunc("/api/player/game/list", handlePlayerGameList)
	r.HandleFunc("/api/game/admin/add", handleGameAdminAdd)
	r.HandleFunc("/api/game/admin/remove", handleGameAdminDelete)
	r.HandleFunc("/api/game/deactivate", handleGameDeactivate)
	r.HandleFunc("/api/game/phrase/update", handleGamePhraseUpdate)
	r.HandleFunc("/api/phrase/update", handleMasterPhraseUpdate)
	r.HandleFunc("/api/game/isadmin", handleIsGameAdmin)
	r.HandleFunc("/api/player/identify", handleIAPUsernameGet)
	r.HandleFunc("/api/player/isadmin", handleIsAdmin)
	r.HandleFunc("/api/admin/add", handleAdminAdd)
	r.HandleFunc("/api/admin/remove", handleAdminDelete)
	r.HandleFunc("/api/admin/list", handleAdminList)
	r.HandleFunc("/api/message/receive", handleMessageAcknowledge)
	r.HandleFunc("/api/cache/clear", handleClearCache)

	routes := []string{"login", "invite", "game", "manage", "gamenew", "gamepicker", "admin"}

	for _, v := range routes {
		r.PathPrefix("/" + v).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "static/index.html")
		})
	}

	fs := wrapHandler(http.FileServer(http.Dir("./static")))
	r.PathPrefix("/").HandlerFunc(fs)

	http.Handle("/", r)
	weblog(fmt.Sprintf("Starting server on port %s\n", port))
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}

}

func handleClearCache(w http.ResponseWriter, r *http.Request) {
	weblog("/api/cache/clear called")

	if err := cache.Clear(); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")

}

func handleIsAdmin(w http.ResponseWriter, r *http.Request) {
	weblog("/api/player/isadmin called")
	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		if err == ErrNotAdmin {
			writeResponse(w, http.StatusOK, fmt.Sprintf("%t", false))
			return
		}
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("%t", true))

}

func handleIsGameAdmin(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/isadmin called")

	gid, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	statusCode, err := isGameAdmin(r, gid)
	if err != nil {
		if err == ErrNotAdmin {
			writeResponse(w, http.StatusOK, fmt.Sprintf("%t", false))
			return
		}
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
		return
	}

	writeResponse(w, http.StatusOK, fmt.Sprintf("%t", true))

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

	statusCode, err := isAdmin(r, gid)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
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

	status, err := isGlobalAdmin(r)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, status, msg)
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

func handleMessageAcknowledge(w http.ResponseWriter, r *http.Request) {
	weblog("/api/message/receive called")

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

	mid := r.Form.Get("m")
	gid := r.Form.Get("g")

	g := Game{}
	m := Message{}
	g.ID = gid
	m.ID = mid

	if err := a.AcknowledgeMessage(g, m); err != nil {
		writeError(w, err.Error())
		return
	}

	writeSuccess(w, "ok")

}

func handlePlayerGameList(w http.ResponseWriter, r *http.Request) {
	weblog("/api/player/game/list called")
	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	games, err := getGamesForKey(email)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	writeJSON(w, games)
	return
}

func handleGameList(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/list called")

	status, err := isGlobalAdmin(r)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, status, msg)
		return
	}

	games, err := getGamesForKey("admin-list")
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

	board, err := getBoard(b, g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if statusCode, err := isAdmin(r, g); err != nil && err != ErrNotAdmin {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
	}

	if !(board.Player.Email == email) && err == ErrNotAdmin {
		msg := fmt.Sprintf("{\"error\":\"Not an admin, game admin or player\"}")
		writeResponse(w, http.StatusForbidden, msg)
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

	pname, err := getFirstQuery("pname", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	p := Player{Name: pname, Email: email}

	game, err := getNewGame(name, p)
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

	email, err := getPlayerEmail(r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	game, err := getGame(g)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	if _, err := isAdmin(r, g); err != nil {
		if err != ErrNotAdmin {
			writeError(w, err.Error())
			return

		}
		game.Obscure(email)
	}

	writeJSON(w, game)
	return
}

func handleGameDeactivate(w http.ResponseWriter, r *http.Request) {
	weblog("/api/game/deactivate called")

	g, err := getFirstQuery("g", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	statusCode, err := isAdmin(r, g)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
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

	st, err := getFirstQuery("selected", r)
	if err != nil {
		writeError(w, err.Error())
		return
	}

	selected := st == "true"
	msg := fmt.Sprintf("Selecte result: %t %s", selected, st)
	weblog(msg)

	if err := recordSelect(bid, gid, pid, selected); err != nil {
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

	statusCode, err := isAdmin(r, g)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
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

	statusCode, err := isAdmin(r, g)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, statusCode, msg)
		return
	}

	game, err := getGame(g)
	if err != nil {
		writeError(w, err.Error())
		return
	}
	p := Player{}
	p.Email = email

	game.Admins.Remove(p)

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

	status, err := isGlobalAdmin(r)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, status, msg)
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

	status, err := isGlobalAdmin(r)
	if err != nil {
		msg := fmt.Sprintf("{\"error\":\"%s\"}", err)
		writeResponse(w, status, msg)
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

func isAdmin(r *http.Request, gid string) (int, error) {

	isGameAdm, err := isGameAdmin(r, gid)
	if err != nil {
		if err != ErrNotAdmin {
			return isGameAdm, err
		}

	}

	isAdm, err := isGlobalAdmin(r)
	if err != nil {
		if err != ErrNotAdmin {
			return isAdm, err
		}
	}

	if isGameAdm == http.StatusOK || isAdm == http.StatusOK {
		return http.StatusOK, nil
	}

	return http.StatusForbidden, ErrNotAdmin
}

func isGlobalAdmin(r *http.Request) (int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	result, err := a.IsAdmin(email)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if result {
		return http.StatusOK, nil
	}
	return http.StatusForbidden, ErrNotAdmin
}

func isGameAdmin(r *http.Request, gid string) (int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	game, err := getGame(gid)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	p := Player{}
	p.Email = email

	result := game.IsAdmin(p)
	if result {
		return http.StatusOK, nil
	}
	return http.StatusForbidden, ErrNotAdmin
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

	payload, err := validateJWT(jwt, projectNumber, projectID)
	if err != nil {
		return "", fmt.Errorf("could not validate IAP JWT: %s", err)
	}

	var ok bool
	email, ok = payload.Claims["email"].(string)

	if !ok {
		return "", fmt.Errorf("could not get email from IAP JWT")
	}

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

func weblog(msg string) {
	if noisy {
		log.Printf("Webserver : %s", msg)
	}
}
