package main

import (
	"log/slog"
	"mime"
	"net/http"
	"os"

	"github.com/arumandesu/blog/internal/transport"
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

	mux := http.NewServeMux()
	h := &transport.HTTP{}
	transport.Handle(mux, h)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
