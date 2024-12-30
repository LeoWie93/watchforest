package server

import (
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	env           string
	domain        string
	sessionLength time.Duration
	port          int
	templates     *Templates

	// authTypeHandler Array
	//// github, gitlab
	//// email

	// db database.Service
}

// Move into views/templates.go
type Templates struct {
	templates *template.Template
}

func (t *Templates) Render(wr io.Writer, name string, data interface{}) error {
	return t.templates.ExecuteTemplate(wr, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("internal/views/*.html")),
	}
}

func NewServer() *http.Server {
	env := os.Getenv("APP_ENV")
	domain := os.Getenv("APP_DOMAIN")
	port, _ := strconv.Atoi(os.Getenv("APP_PORT"))

	newServer := &Server{
		env:           env,
		domain:        domain,
		sessionLength: time.Hour,
		port:          port,
		templates:     NewTemplates(),
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	slog.Info("created server", "addr", server.Addr, "env", newServer.env)

	return server
}
