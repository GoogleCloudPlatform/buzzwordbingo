package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func TestMain(m *testing.M) {
	// command to start firestore emulator
	cmd := exec.Command("gcloud", "beta", "emulators", "firestore", "start", fmt.Sprintf("--host-port=localhost:%d", 8181), "--quiet")

	// this makes it killable
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// we need to capture it's output to know when it's started
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stderr.Close()

	// start her up!
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// ensure the process is killed when we're finished, even if an error occurs
	// (thanks to Brian Moran for suggestion)
	var result int
	defer func() {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		os.Exit(result)
	}()

	// we're going to wait until it's running to start
	var wg sync.WaitGroup
	wg.Add(1)

	// by starting a separate go routine
	go func() {
		// reading it's output
		buf := make([]byte, 256, 256)
		for {
			n, err := stderr.Read(buf[:])
			if err != nil {
				// until it ends
				if err == io.EOF {
					break
				}
				log.Fatalf("reading stderr %v", err)
			}

			if n > 0 {
				d := string(buf[:n])

				// only required if we want to see the emulator output
				log.Printf("%s", d)

				// checking for the message that it's started
				if strings.Contains(d, "Dev App Server is now running") {
					wg.Done()
				}

				// and capturing the FIRESTORE_EMULATOR_HOST value to set
				pos := strings.Index(d, FirestoreEmulatorHost+"=")
				if pos > 0 {
					host := d[pos+len(FirestoreEmulatorHost)+1 : len(d)-1]
					os.Setenv(FirestoreEmulatorHost, host)
				}
			}
		}
	}()

	// wait until the running message has been received
	wg.Wait()

	agentTestSetup()
	cacheTestSetup()
	noisy = false

	// now it's running, we can run our unit tests
	result = m.Run()
}

func TestGetUsername(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"accounts.google.com:example@gmail.com", "example@gmail.com"},
		{"", ""},
	}
	for _, c := range cases {
		got := getEmailFromString(c.in)
		if got != c.want {
			t.Errorf("getEmailFromString(%s)  got %s, want %s", c.in, got, c.want)
		}

	}

}

func TestSimpleHandlers(t *testing.T) {

	var table = []struct {
		in      string
		out     string
		handler http.Handler
	}{
		{"/healthz", `{"msg":"ok"}`, http.HandlerFunc(handleHealth)},
		{"/api/player/identify", fmt.Sprintf(`{"name":"","email":"%s@google.com"}`, os.Getenv("USER")), JSONHandler(iapUsernameGetHandle, "none")},
		{"/api/player/isadmin", "false", AdminHandler(isAdminHandle)},
		{"/api/admin/list", "[]", JSONHandler(adminListHandle, "none")},
		{"/api/game/list", "[]", JSONHandler(gameListHandle, "none")},
		{"/api/player/game/list", "[]", JSONHandler(playerGameListHandle, "none")},
		{"/api/cache/clear", `{"msg":"ok"}`, SimpleHandler(clearCacheHandle, "none")},
	}

	for _, v := range table {

		// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
		// pass 'nil' as the third parameter.
		req, err := http.NewRequest("GET", v.in, nil)
		if err != nil {
			t.Fatal(err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		v.handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// Check the response body is what we expect.
		expected := v.out
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}

	}

}

func TestAuthErrorsGetHandlers(t *testing.T) {

	errmsg := fmt.Sprintf(`{"error":"%s"}`, ErrNotAdmin)

	var table = []struct {
		in      string
		out     string
		handler http.Handler
	}{
		{"/api/admin/list", errmsg, JSONHandler(adminListHandle, "global")},
		{"/api/game/list", errmsg, JSONHandler(gameListHandle, "global")},
		{"/api/cache/clear", errmsg, SimpleHandler(clearCacheHandle, "global")},
	}

	for _, v := range table {

		// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
		// pass 'nil' as the third parameter.
		req, err := http.NewRequest("GET", v.in, nil)
		if err != nil {
			t.Fatal(err)
		}

		// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
		rr := httptest.NewRecorder()

		// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
		// directly and pass in our Request and ResponseRecorder.
		v.handler.ServeHTTP(rr, req)

		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusForbidden {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusForbidden)
		}

		// Check the response body is what we expect.
		expected := v.out
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v",
				rr.Body.String(), expected)
		}

	}

}
