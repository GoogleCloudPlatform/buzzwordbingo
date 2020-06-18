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
	cache         *Cache
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
	r.Handle("/api/board", JSONHandler(boardGetHandle))
	r.Handle("/api/board/delete", PrefetechHandler(boardDeleteHandle, http.MethodDelete))
	r.Handle("/api/record", SimpleHandler(recordSelectHandle))
	r.Handle("/api/game", JSONHandler(gameGetHandle))
	r.Handle("/api/game/new", JSONHandler(gameNewHandle))
	r.Handle("/api/game/list", JSONHandler(gameListHandle))
	r.Handle("/api/player/game/list", JSONHandler(playerGameListHandle))
	r.Handle("/api/game/admin/add", PrefetechHandler(gameAdminAddHandle, http.MethodPost))
	r.Handle("/api/game/admin/remove", PrefetechHandler(gameAdminDeleteHandle, http.MethodDelete))
	r.Handle("/api/game/deactivate", SimpleHandler(gameDeactivateHandle))
	r.Handle("/api/game/phrase/update", SimpleHandler(gamePhraseUpdateHandle))
	r.Handle("/api/phrase/update", SimpleHandler(masterPhraseUpdateHandle))
	r.Handle("/api/game/isadmin", SimpleHandler(isGameAdminHandle))
	r.Handle("/api/player/identify", JSONHandler(iapUsernameGetHandle))
	r.Handle("/api/player/isadmin", SimpleHandler(isAdminHandle))
	r.Handle("/api/admin/add", PrefetechHandler(adminAddHandle, http.MethodPost))
	r.Handle("/api/admin/remove", PrefetechHandler(adminDeleteHandle, http.MethodDelete))
	r.Handle("/api/admin/list", JSONHandler(adminListHandle))
	r.Handle("/api/message/receive", PrefetechHandler(messageAcknowledgeHandle, http.MethodPost))
	r.Handle("/api/cache/clear", SimpleHandler(clearCacheHandle))

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

type Adapter func(http.Handler) http.Handler
type ErrorEmitter func(http.ResponseWriter, *http.Request) (string, int, error)
type JSONEmitter func(http.ResponseWriter, *http.Request) (JSONProducer, int, error)

func SimpleHandler(h ErrorEmitter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))
		response, statuscode, err := h(w, r)

		if err != nil {
			if statuscode != http.StatusInternalServerError {
				writeResponse(w, statuscode, response)
				return
			}
			writeError(w, err.Error())
			return
		}
		writeSuccess(w, response)
	})
}

func JSONHandler(h JSONEmitter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))
		jsonProducer, statuscode, err := h(w, r)

		if err != nil {
			if statuscode != http.StatusInternalServerError {
				writeResponse(w, statuscode, fmt.Sprintf("{\"error\":\"%s\"}", err))
				return
			}
			writeError(w, err.Error())
			return
		}
		writeJSON(w, jsonProducer)
	})
}

func PrefetechHandler(h ErrorEmitter, method string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", method)
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			return
		}

		response, statuscode, err := h(w, r)

		if r.Method != method {
			msg := fmt.Sprintf("{\"error\":\"Must use http method %s you had %s\"}", method, r.Method)
			writeResponse(w, http.StatusMethodNotAllowed, msg)
			return
		}

		if err != nil {
			if statuscode != http.StatusInternalServerError {
				writeResponse(w, statuscode, response)
				return
			}
			writeError(w, err.Error())
			return
		}
		writeSuccess(w, response)
	})
}

func clearCacheHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	return "ok", http.StatusOK, cache.Clear()
}

func isAdminHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		if err == ErrNotAdmin {
			return fmt.Sprintf("%t", false), http.StatusOK, err
		}
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}
	return fmt.Sprintf("%t", true), http.StatusOK, err
}

func isGameAdminHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	gid, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isGameAdmin(r, gid)
	if err != nil {
		if err == ErrNotAdmin {
			return fmt.Sprintf("%t", false), http.StatusOK, err
		}
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}
	return fmt.Sprintf("%t", true), http.StatusOK, err
}

func gamePhraseUpdateHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	gid, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	pid, err := getFirstQuery("p", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	text, err := getFirstQuery("text", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isAdmin(r, gid)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	phrase := Phrase{}
	phrase.ID = pid
	phrase.Text = text

	return "ok", http.StatusInternalServerError, updateGamePhrases(gid, phrase)
}

func masterPhraseUpdateHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	pid, err := getFirstQuery("p", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	text, err := getFirstQuery("text", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	phrase := Phrase{}
	phrase.ID = pid
	phrase.Text = text

	return "ok", http.StatusInternalServerError, updateMasterPhrase(phrase)
}

func iapUsernameGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	p := Player{}

	email, err := getPlayerEmail(r)
	if err != nil {
		return p, http.StatusInternalServerError, err
	}

	p.Email = email

	return p, http.StatusOK, nil
}

func messageAcknowledgeHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	if err := r.ParseMultipartForm(160000); err != nil {
		return "", http.StatusInternalServerError, err
	}

	mid := r.Form.Get("m")
	gid := r.Form.Get("g")

	if mid == "" {
		return "", http.StatusInternalServerError, fmt.Errorf("m is required")
	}

	if gid == "" {
		return "", http.StatusInternalServerError, fmt.Errorf("g is required")
	}

	g := Game{}
	m := Message{}
	g.ID = gid
	m.ID = mid

	return "ok", http.StatusOK, a.AcknowledgeMessage(g, m)
}

func playerGameListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Games{}, http.StatusInternalServerError, err
	}
	games, err := getGamesForKey(email)
	return games, http.StatusOK, err
}

func gameListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		return Games{}, statusCode, err
	}

	games, err := getGamesForKey("admin-list")
	return games, http.StatusOK, err
}

func boardGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Board{}, http.StatusInternalServerError, err
	}

	name, err := getFirstQuery("name", r)
	if err != nil {
		return Board{}, http.StatusInternalServerError, err
	}

	gid, err := getFirstQuery("g", r)
	if err != nil {
		return Board{}, http.StatusInternalServerError, err
	}

	p := Player{Name: name, Email: email}

	g, err := getGame(gid)
	if err != nil {
		return Board{}, http.StatusInternalServerError, err
	}

	board, err := getBoardForPlayer(p, g)
	if err != nil {
		return Board{}, http.StatusInternalServerError, err
	}

	return board, http.StatusOK, err
}

func boardDeleteHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	b, err := getFirstQuery("b", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	g, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	board, err := getBoard(b, g)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	email, err := getPlayerEmail(r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	if statusCode, err := isAdmin(r, g); err != nil && err != ErrNotAdmin {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	if !(board.Player.Email == email) && err == ErrNotAdmin {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), http.StatusForbidden, err
	}

	return "ok", http.StatusOK, deleteBoard(b, g)
}

func gameNewHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	name, err := getFirstQuery("name", r)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	pname, err := getFirstQuery("pname", r)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	p := Player{Name: pname, Email: email}

	game, err := getNewGame(name, p)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	return game, http.StatusOK, err
}

func gameGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	g, err := getFirstQuery("g", r)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	game, err := getGame(g)
	if err != nil {
		return Game{}, http.StatusInternalServerError, err
	}

	if _, err := isAdmin(r, g); err != nil {
		if err != ErrNotAdmin {
			return Game{}, http.StatusInternalServerError, err
		}
		game.Obscure(email)
	}
	return game, http.StatusOK, err
}

func gameDeactivateHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	g, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isAdmin(r, g)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	return "ok", http.StatusInternalServerError, deactivateGame(g)
}

func recordSelectHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	pid, err := getFirstQuery("p", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	bid, err := getFirstQuery("b", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	gid, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	st, err := getFirstQuery("selected", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	selected := st == "true"
	return "ok", http.StatusInternalServerError, recordSelect(bid, gid, pid, selected)
}

func gameAdminAddHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	if err := r.ParseMultipartForm(160000); err != nil {
		return "", http.StatusInternalServerError, err
	}
	g := r.Form.Get("g")
	email := r.Form.Get("email")

	if email == "" {
		return "", http.StatusInternalServerError, fmt.Errorf("email is required")
	}

	if g == "" {
		return "", http.StatusInternalServerError, fmt.Errorf("g is required")
	}

	statusCode, err := isAdmin(r, g)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	game, err := getGame(g)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	p := Player{}
	p.Email = email
	game.Admins.Add(p)

	if err := cache.SaveGame(game); err != nil {
		return "", http.StatusInternalServerError, err
	}

	if err := a.SaveGame(game); err != nil {
		return "", http.StatusInternalServerError, err
	}

	return "ok", http.StatusOK, nil
}

func gameAdminDeleteHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	g, err := getFirstQuery("g", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	email, err := getFirstQuery("email", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isAdmin(r, g)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	game, err := getGame(g)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}
	p := Player{"", email}
	game.Admins.Remove(p)

	if err := cache.SaveGame(game); err != nil {
		return "", http.StatusInternalServerError, err
	}

	if err := a.SaveGame(game); err != nil {
		return "", http.StatusInternalServerError, err
	}

	return "ok", http.StatusOK, nil
}

func adminAddHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	if err := r.ParseMultipartForm(160000); err != nil {
		return "", http.StatusInternalServerError, err
	}

	email := r.Form.Get("email")

	if email == "" {
		return "", http.StatusInternalServerError, fmt.Errorf("email is required")
	}

	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	p := Player{"", email}

	return "ok", http.StatusOK, a.AddAdmin(p)
}

func adminDeleteHandle(w http.ResponseWriter, r *http.Request) (string, int, error) {
	email, err := getFirstQuery("email", r)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		return fmt.Sprintf("{\"error\":\"%s\"}", err), statusCode, err
	}

	p := Player{"", email}

	return "ok", http.StatusOK, a.DeleteAdmin(p)
}

func adminListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, int, error) {
	statusCode, err := isGlobalAdmin(r)
	if err != nil {
		return Games{}, statusCode, err
	}

	players, err := a.GetAdmins()
	return players, http.StatusOK, err
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

	if msg == "true" || msg == "false" {
		s = msg
	}

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

func getQueries(r *http.Request, queries ...string) (map[string]string, error) {
	results := make(map[string]string)

	for _, v := range results {
		result, ok := r.URL.Query()[v]
		if !ok || len(result[0]) < 1 || result[0] == "undefined" {
			return results, fmt.Errorf("query parameter %s is missing", v)
		}
		results[v] = result[0]
	}

	return results, nil
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
