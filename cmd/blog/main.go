package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/arumandesu/blog/assets"
	"github.com/arumandesu/blog/internal/transport"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fileServer := http.FileServerFS(assets.StaticFiles)

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	h := &transport.HTTP{}
	transport.Handle(mux, h)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
