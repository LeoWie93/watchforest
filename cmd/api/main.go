package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"watchforest/internal/dotenv"
	"watchforest/internal/server"
)

func main() {
	if err := dotenv.Load(".env"); err != nil {
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
