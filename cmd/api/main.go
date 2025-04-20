package main

import (
	"fmt"
	"github.com/leowie93/watchforest/internal/server"
	"log/slog"
	"net/http"

	"github.com/leowie93/goenv"
)

func main() {
	if err := goenv.LoadIntoEnv(".env"); err != nil {
		panic(err)
	}

	server := server.NewServer()
	slog.Info("starting server")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// middleware that knows all types of auth middlewares
	// initialises all types (.env, clients etc)
	// e.Use(auth.userMiddleware())
}
