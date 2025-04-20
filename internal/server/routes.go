package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/leowie93/watchforest/internal/helper"
	"github.com/leowie93/watchforest/internal/oauth"
)

type Data struct {
	Data string
}

var clientId string = os.Getenv("GITHUB_CLIENT_ID")
var clientSecret string = os.Getenv("GITHUB_CLIENT_SECRET")
var sessionName string = "session_token_v1"

// TODO split this. Defining routes / defining middlewares (both can be multiple files?)
func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", s.ShowAuthHandler)
	mux.HandleFunc("/oauth/github/callback", s.OauthGithubCallbackHandler)

	mux.HandleFunc("/autharea", s.wrapProtected(s.ShowLoggedInHandler))

	//TODO use a stackbuilder or something
	return s.errorMiddleware(
		s.loggingMiddleware(
			s.corsMiddleware(
				s.sessionMiddleware(mux),
			),
		),
	)
}

func (s *Server) errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO listen for the response of servehttp and handle err if given
		next.ServeHTTP(w, r)
	})
}

// TODO do we want to bild this logger out? for each environment etc?
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		time := time.Now().Sub(start)
		slog.Info(fmt.Sprintf("%s to %s in %vms", r.Method, r.URL.Path, time.Milliseconds()))
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//TODO set via env
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie(sessionName); err != nil {
			s.sessionCookie = cookie
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) wrapProtected(handler http.HandlerFunc) http.HandlerFunc {
	return s.authMiddleware(handler)
}

//cookie code

// fmt.Println("we create a cookie")
// sessionId, err := uuid.NewV7()
//
// if err != nil {
// 	//TODO we return an error and let the erro middleware do its thing
// 	fmt.Println("some uuid error")
// 	fmt.Println(err)
// }
//
// http.SetCookie(w, &http.Cookie{
// 	Name:    sessionName,
// 	Value:   sessionId.String(),
// 	Domain:  s.domain,
// 	Expires: time.Now().Add(s.sessionLength),
// })

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// clientId = os.Getenv("GITHUB_CLIENT_ID")
		// clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

		next.ServeHTTP(w, r)

		// if not valid
		// http.Redirect(w, r, "/", 402)
	})
}

func (s *Server) ShowLoggedInHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"page": "logged in"}

	jsonResp, _ := json.Marshal(data)
	w.Write(jsonResp)
}

func (s *Server) ShowAuthHandler(w http.ResponseWriter, r *http.Request) {
	//TODO move this somewhere
	clientId := os.Getenv("GITHUB_CLIENT_ID")
	data := Data{Data: clientId}

	w.WriteHeader(http.StatusOK)
	s.templates.Render(w, "auth", data)
}

func (s *Server) OauthGithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	if code == "" {
		helper.HandleError(errors.New("code query param does not exist or is empty"))
		//TODO set error message (message back or session back handling whatever. mem-db would be fun)
		http.Redirect(w, r, "/login", 302)
		return
	}

	atResponse, err := oauth.RequestAccessToken(code)
	if err != nil {
		http.Redirect(w, r, "/login", 302)
	}
	req, _ := http.NewRequest(http.MethodGet, "https://api.github.com/user/public_emails?per_page=1", nil)
	req.Header.Add("Authorization", "Bearer "+atResponse.AccessToken)
	req.Header.Add("Accept", "application/vnd.github+json")
	res, _ := http.DefaultClient.Do(req)

	body, _ := io.ReadAll(res.Body)
	res.Body.Close()

	w.Write(body)

	// use oauth.GetUserMail(access_token)

	// create or get user if not exists
}
