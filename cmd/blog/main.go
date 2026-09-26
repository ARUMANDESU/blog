package main

import (
	"log/slog"
	"mime"
	"net/http"
	"os"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/repository"
	"github.com/arumandesu/blog/internal/transport"
	"github.com/arumandesu/blog/pkg"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Go falls back to the host's mime.types for these, and a scratch container has none.
	for ext, typ := range map[string]string{
		".woff2":       "font/woff2",
		".woff":        "font/woff",
		".webmanifest": "application/manifest+json",
	} {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
	}

	postRepo := repository.NewInMemoryPostRepo()
	a := app.New(&pkg.NoOpTxManager{}, postRepo, postRepo)

	mux := http.NewServeMux()
	h := transport.NewHTTP(a, logger)
	transport.Handle(mux, h)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
