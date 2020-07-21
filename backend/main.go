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
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	ctx           = context.Background()
	// ErrNotAdmin is an error that indicates that the user is not an admin
	ErrNotAdmin = fmt.Errorf("not an admin or game admin")
	// ErrNotAdminOrPlayer is an error that indicates that the user is not an
	// admin nor an owner of the board they are editing.
	ErrNotAdminOrPlayer = fmt.Errorf("not an admin or game admin, or owner of board")
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
	r.Handle("/api/board", JSONHandler(boardGetHandle, "none"))
	r.Handle("/api/board/delete", PrefetechHandler(boardDeleteHandle, http.MethodDelete, "none"))
	r.Handle("/api/record", SimpleHandler(recordSelectHandle, "none"))
	r.Handle("/api/game", JSONHandler(gameGetHandle, "none"))
	r.Handle("/api/game/new", JSONHandler(gameNewHandle, "none"))
	r.Handle("/api/game/list", JSONHandler(gameListHandle, "global"))
	r.Handle("/api/player/game/list", JSONHandler(playerGameListHandle, "none"))
	r.Handle("/api/game/admin/add", PrefetechHandler(gameAdminAddHandle, http.MethodPost, "game"))
	r.Handle("/api/game/admin/remove", PrefetechHandler(gameAdminDeleteHandle, http.MethodDelete, "game"))
	r.Handle("/api/game/deactivate", SimpleHandler(gameDeactivateHandle, "game"))
	r.Handle("/api/game/purge", SimpleHandler(purgeHandle, "none"))
	r.Handle("/api/game/phrase/update", SimpleHandler(gamePhraseUpdateHandle, "game"))
	r.Handle("/api/phrase/update", SimpleHandler(masterPhraseUpdateHandle, "global"))
	r.Handle("/api/game/isadmin", AdminHandler(isGameAdminHandle))
	r.Handle("/api/player/identify", JSONHandler(iapUsernameGetHandle, "none"))
	r.Handle("/api/player/isadmin", AdminHandler(isAdminHandle))
	r.Handle("/api/admin/add", PrefetechHandler(adminAddHandle, http.MethodPost, "global"))
	r.Handle("/api/admin/remove", PrefetechHandler(adminDeleteHandle, http.MethodDelete, "global"))
	r.Handle("/api/admin/list", JSONHandler(adminListHandle, "global"))
	r.Handle("/api/message/receive", PrefetechHandler(messageAcknowledgeHandle, http.MethodPost, "none"))
	r.Handle("/api/cache/clear", SimpleHandler(clearCacheHandle, "global"))

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

// ErrorEmitter is a http.Handler that emits an error for much better reporting
type ErrorEmitter func(http.ResponseWriter, *http.Request) error

// JSONEmitter is a http.Handler that emits a jsoning file
type JSONEmitter func(http.ResponseWriter, *http.Request) (JSONProducer, error)

// AdminEmitter is a http.Handler checks a boolean condition
type AdminEmitter func(http.ResponseWriter, *http.Request) (int, error)

// AdminHandler is a http.Handler checks the conditions of a isadmin request
func AdminHandler(h AdminEmitter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))

		_, err := h(w, r)

		if err != nil {
			if err == ErrNotAdmin {
				writeResponse(w, http.StatusOK, fmt.Sprintf("%t", false))
				return
			}
			writeErrorMsg(w, err)
			return
		}
		writeResponse(w, http.StatusOK, fmt.Sprintf("%t", true))
		return

	})
}

// SimpleHandler is a http.Handler thta does a simple request
func SimpleHandler(h ErrorEmitter, adminlevel string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))

		if err := IsAdminChecker(w, r, adminlevel); err != nil {
			return
		}

		err := h(w, r)

		if err != nil {
			writeError(w, err.Error())
			return
		}
		writeSuccess(w, "ok")
	})
}

// IsAdminChecker does the proper test for whether or not something is an admin
func IsAdminChecker(w http.ResponseWriter, r *http.Request, adminlevel string) error {
	switch adminlevel {
	case "game":
		queries, err := getQueries(r, "g")
		if err != nil {
			writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("{\"error\":\"%s\"}", err))
			return err
		}

		statusCode, err := isAdmin(r, queries["g"])
		if err != nil {
			weblog(fmt.Sprintf("IsAdminCheck failed in the handler"))
			writeResponse(w, statusCode, fmt.Sprintf("{\"error\":\"%s\"}", err))
			return err
		}
		weblog(fmt.Sprintf("IsAdminCheck passed in the handler"))
	case "global":
		statusCode, err := isGlobalAdmin(r)
		if err != nil {
			weblog(fmt.Sprintf("IsGlobalCheck failed in the handler"))
			writeResponse(w, statusCode, fmt.Sprintf("{\"error\":\"%s\"}", err))
			return err
		}
		weblog(fmt.Sprintf("IsGlobalCheck passed in the handler"))
	default:
		return nil
	}
	return nil
}

// JSONHandler is a http.Handler that handles returning json
func JSONHandler(h JSONEmitter, adminlevel string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))

		if err := IsAdminChecker(w, r, adminlevel); err != nil {
			return
		}

		jsonProducer, err := h(w, r)

		if err != nil {
			writeError(w, err.Error())
			return
		}
		writeJSON(w, jsonProducer)
		return
	})
}

// PrefetechHandler is a http.Handler that handles preflight requests
func PrefetechHandler(h ErrorEmitter, method string, adminlevel string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		weblog(fmt.Sprintf("%s called", r.URL.Path))

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", method)
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			return
		}

		if r.Method != method {
			msg := fmt.Sprintf("{\"error\":\"Must use http method %s you had %s\"}", method, r.Method)
			writeResponse(w, http.StatusMethodNotAllowed, msg)
			return
		}

		if err := IsAdminChecker(w, r, adminlevel); err != nil {
			return
		}

		if err := h(w, r); err != nil {
			if err == ErrNotAdminOrPlayer {
				writeResponse(w, http.StatusForbidden, fmt.Sprintf("{\"error\":\"%s\"}", err))
			}

			writeError(w, err.Error())
			return
		}
		writeSuccess(w, "ok")
	})
}

func clearCacheHandle(w http.ResponseWriter, r *http.Request) error {
	return cache.Clear()
}

func isAdminHandle(w http.ResponseWriter, r *http.Request) (int, error) {
	return isGlobalAdmin(r)
}

func isGameAdminHandle(w http.ResponseWriter, r *http.Request) (int, error) {

	queries, err := getQueries(r, "g")

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return isGameAdmin(r, queries["g"])
}

func purgeHandle(w http.ResponseWriter, r *http.Request) error {
	return a.PurgeOldGames()
}

func gamePhraseUpdateHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "g", "p", "text")
	if err != nil {
		return err
	}

	phrase := Phrase{}
	phrase.ID = queries["p"]
	phrase.Text = queries["text"]

	return updateGamePhrases(queries["g"], phrase)
}

func masterPhraseUpdateHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "p", "text")
	if err != nil {
		return err
	}

	phrase := Phrase{}
	phrase.ID = queries["p"]
	phrase.Text = queries["text"]

	return updateMasterPhrase(phrase)
}

func iapUsernameGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {

	email, err := getPlayerEmail(r)
	if err != nil {
		return Player{}, err
	}

	return Player{"", email}, nil
}

func messageAcknowledgeHandle(w http.ResponseWriter, r *http.Request) error {

	queries, err := getQueries(r, "m", "g")
	if err != nil {
		return err
	}

	g := Game{}
	m := Message{}
	g.ID = queries["g"]
	m.ID = queries["m"]

	return a.AcknowledgeMessage(g, m)
}

// TODO: Change to handle things passed from request.
func playerGameListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {

	email, err := getPlayerEmail(r)
	if err != nil {
		return Games{}, err
	}
	return getGamesForKey(email, 10, time.Now())
}

func gameListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {
	queries, err := getQueries(r, "l", "t")
	if err != nil {
		return Games{}, err
	}

	limit, err := strconv.Atoi(queries["l"])
	if err != nil {
		return Games{}, err
	}

	tokenint, err := strconv.Atoi(queries["t"])
	if err != nil {
		return Games{}, err
	}

	token := time.Unix(int64(tokenint), 0)

	if err != nil {
		return Games{}, err
	}

	key := fmt.Sprintf("admin-list-%d-%d", limit, tokenint)

	return getGamesForKey(key, limit, token)
}

func boardGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Board{}, err
	}

	queries, err := getQueries(r, "g", "name")
	if err != nil {
		return Board{}, err
	}

	p := Player{Name: queries["name"], Email: email}

	g, err := getGame(queries["g"])
	if err != nil {
		return Board{}, err
	}

	return getBoardForPlayer(p, g)
}

func boardDeleteHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "b", "g")
	if err != nil {
		return err
	}

	board, err := getBoard(queries["b"], queries["g"])
	if err != nil {
		return err
	}

	email, err := getPlayerEmail(r)
	if err != nil {
		return err
	}

	if _, err := isAdmin(r, queries["g"]); err != nil && err != ErrNotAdmin {
		return err
	}

	if !(board.Player.Email == email) && err == ErrNotAdmin {
		return ErrNotAdminOrPlayer
	}

	return deleteBoard(queries["b"], queries["g"])
}

func gameNewHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Game{}, err
	}

	queries, err := getQueries(r, "name", "pname")
	if err != nil {
		return Game{}, err
	}

	p := Player{Name: queries["pname"], Email: email}

	return getNewGame(queries["name"], p)
}

func gameGetHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {
	email, err := getPlayerEmail(r)
	if err != nil {
		return Game{}, err
	}

	queries, err := getQueries(r, "g")
	if err != nil {
		return Game{}, err
	}

	game, err := getGame(queries["g"])
	if err != nil {
		return Game{}, err
	}

	if _, err := isAdmin(r, queries["g"]); err != nil {
		if err != ErrNotAdmin {
			return Game{}, err
		}
		game.Obscure(email)
	}
	return game, nil
}

func gameDeactivateHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "g")
	if err != nil {
		return err
	}

	return deactivateGame(queries["g"])
}

func recordSelectHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "p", "b", "g", "selected")
	if err != nil {
		return err
	}

	selected := queries["selected"] == "true"
	return recordSelect(queries["b"], queries["g"], queries["p"], selected)
}

func gameAdminAddHandle(w http.ResponseWriter, r *http.Request) error {

	queries, err := getQueries(r, "g", "email")
	if err != nil {
		return err
	}

	game, err := getGame(queries["g"])
	if err != nil {
		return err
	}
	p := Player{}
	p.Email = queries["email"]
	game.Admins.Add(p)

	if err := cache.SaveGame(game); err != nil {
		return err
	}

	return a.SaveGame(game)
}

func gameAdminDeleteHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "g", "email")
	if err != nil {
		return err
	}

	game, err := getGame(queries["g"])
	if err != nil {
		return err
	}
	p := Player{"", queries["email"]}
	game.Admins.Remove(p)

	if err := cache.SaveGame(game); err != nil {
		return err
	}

	return a.SaveGame(game)
}

func adminAddHandle(w http.ResponseWriter, r *http.Request) error {

	queries, err := getQueries(r, "email")
	if err != nil {
		return err
	}

	p := Player{"", queries["email"]}

	return a.AddAdmin(p)
}

func adminDeleteHandle(w http.ResponseWriter, r *http.Request) error {
	queries, err := getQueries(r, "email")
	if err != nil {
		return err
	}

	p := Player{"", queries["email"]}

	return a.DeleteAdmin(p)
}

func adminListHandle(w http.ResponseWriter, r *http.Request) (JSONProducer, error) {
	return a.GetAdmins()
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

func writeErrorMsg(w http.ResponseWriter, err error) {
	s := fmt.Sprintf("{\"error\":\"%s\"}", err)
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

type notFoundRedirectRespWr struct {
	http.ResponseWriter // We embed http.ResponseWriter
	status              int
}

func (w *notFoundRedirectRespWr) WriteHeader(status int) {
	w.status = status // Store the status for our own use
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *notFoundRedirectRespWr) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &notFoundRedirectRespWr{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == 404 {
			http.Redirect(w, r, "/index.html", http.StatusFound)
		}
	}
}

func getQueries(r *http.Request, queries ...string) (map[string]string, error) {
	results := make(map[string]string)

	switch r.Method {
	case http.MethodPost:
		if err := r.ParseMultipartForm(160000); err != nil {
			return results, err
		}
		for _, v := range queries {
			result := r.Form.Get(v)
			if len(result) < 1 {
				fmt.Printf("POST '%+v'\n", v)
				fmt.Printf("Form '%+v'\n", r.Form)
				return results, fmt.Errorf("query parameter '%s' is missing", v)
			}
			results[v] = result
		}

	default:
		for _, v := range queries {
			result, ok := r.URL.Query()[v]
			if !ok || len(result[0]) < 1 || result[0] == "undefined" {
				err := fmt.Errorf("query parameter '%s' is missing", v)
				return results, err
			}
			results[v] = result[0]
		}
	}

	return results, nil
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

	p := Player{"", email}

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
