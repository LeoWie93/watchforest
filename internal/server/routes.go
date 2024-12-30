package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

type Data struct {
	Data string
}

var clientId string = ""
var clientSecret string = ""

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", s.ShowAuthHandler)
	mux.HandleFunc("/oauth/github/callback", s.OauthGithubCallbackHandler)

	mux.HandleFunc("/autharea", s.wrapProtected(s.ShowLoggedInHandler))

	//TODO use a stackbuilder or something
	return s.errorMiddleware(s.loggingMiddleware(s.corsMiddleware(mux)))
}

func (s *Server) wrapProtected(handler http.HandlerFunc) http.HandlerFunc {
	return s.authMiddleware(handler)
}

var sessionName string = "session_token_v1"

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// clientId = os.Getenv("GITHUB_CLIENT_ID")
		// clientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

		// session middleware?
		if _, err := r.Cookie(sessionName); err == nil {
			sessionId, err := uuid.NewV7()

			if err != nil {
				//TODO we return an error and let the erro middleware do its thing
			}

			r.AddCookie(&http.Cookie{
				Name:    sessionName,
				Value:   sessionId.String(),
				Domain:  s.domain,
				Expires: time.Now().Add(s.sessionLength),
			})
		}

		next.ServeHTTP(w, r)

		// if not valid
		// http.Redirect(w, r, "/", 402)
	})
}

func (s *Server) errorMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// listen for the response of servehttp and handle err if given
		next.ServeHTTP(w, r)
	})
}

// TODO do we want to bild this logger out? for each environment etc?
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		time := start.Sub(time.Now())
		slog.Info(fmt.Sprintf("%s to %s in %vms", r.Method, r.URL.Path, time.Milliseconds()))
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
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

type OauthResponse struct {
	Access_Token string
	Token_Type   string
	Scope        string
}

func (s *Server) OauthGithubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	code := r.FormValue("code")
	if code == "" {
		// we want to show an internal error (most likely the user fiddled with some stuff)
		slog.Error("code query param does not exist")
		http.Redirect(w, r, "/login", 302)
		return
	}

	// reqUrl := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientId, clientSecret, code)
	// req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
	// if err != nil {
	// 	c.Logger().Error(err)
	// 	// give error message?
	// 	return c.Redirect(http.StatusInternalServerError, "/")
	// }
	//
	// req.Header.Add("accept", "application/json")
	// client := &http.Client{}
	// res, err := client.Do(req)
	// if err != nil {
	// 	c.Logger().Error(err)
	// 	// give error message?
	// 	return c.Redirect(http.StatusInternalServerError, "/")
	// }
	//
	// body, err := io.ReadAll(res.Body)
	// var oauthResponse OauthResponse
	//
	// err = json.Unmarshal(body, &oauthResponse)
	// if err != nil {
	// 	c.Logger().Error(err)
	// 	// give error message?
	// 	return c.Redirect(http.StatusInternalServerError, "/")
	// }
	//
	// //TODO github specific
	// // expand on the scope checking / can't be bothered right now
	// if oauthResponse.Scope != "user:email" {
	// 	c.Logger().Error("scope is not valid")
	// 	// give error message?
	// 	return c.Redirect(http.StatusInternalServerError, "/")
	// }
	//
	// //!!!! can receive multipe emails. first should always be primary
	// req, err = http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	// req.Header.Add("Authorization", "token "+oauthResponse.Access_Token)
	// res, err = client.Do(req)
	// body, err = io.ReadAll(res.Body)
	//
	// return c.String(200, string(body))
}
