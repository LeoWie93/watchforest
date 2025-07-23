package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/leowie93/watchforest/internal/database/models"
	"github.com/leowie93/watchforest/internal/templates"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Server struct {
	env           string
	domain        string
	sessionLength time.Duration
	port          int
	templates     *templates.Templates
	sessionCookie *http.Cookie
	db            *gorm.DB
}

func NewServer() *http.Server {
	env := os.Getenv("APP_ENV")
	domain := os.Getenv("APP_DOMAIN")
	port, _ := strconv.Atoi(os.Getenv("APP_PORT"))
	dbPath := os.Getenv("DB_FS_PATH")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	//TODO keep this here or move?
	//run migrations and general db stuff
	if err := db.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		panic(err)
	}

	newServer := &Server{
		env:           env,
		domain:        domain,
		sessionLength: time.Hour * 24 * 7,
		port:          port,
		templates:     templates.NewTemplates(),
		db:            db,
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
