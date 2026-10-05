package main

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arumandesu/blog/internal/app"
	"github.com/arumandesu/blog/internal/garage"
	"github.com/arumandesu/blog/internal/repository/sqlite"
	"github.com/arumandesu/blog/internal/transport"
	"github.com/arumandesu/blog/internal/workers"
	"github.com/arumandesu/blog/pkg"
	"github.com/go-chi/chi/v5"
)

const DefaultDBPath = "./data/db.sqlite"

func main() {
	ctx := context.Background()
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

	dbPath, ok := os.LookupEnv("SQLITE_DB_PATH")
	if !ok {
		dbPath = DefaultDBPath
	}
	err := os.MkdirAll(filepath.Dir(dbPath), 0o755)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	wdb, rdb, err := pkg.ConnectToSQLite(ctx, dbPath)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer wdb.Close()
	defer rdb.Close()

	err = pkg.Migrate(wdb)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	postRepo := sqlite.NewPostRepo(wdb, rdb)
	// if err := seed(context.Background(), postRepo); err != nil {
	// 	logger.Error(err.Error())
	// 	os.Exit(1)
	// }
	mediaRepo := sqlite.NewMediaRepo(wdb, rdb)

	s3, err := garage.NewS3(garage.Config{
		Endpoint:   mustGetEnv("S3_ENDPOINT"),
		IsSecure:   mustGetBoolEnv("S3_ENDPOINT_IS_SECURE"),
		Region:     "garage",
		Bucket:     "media",
		CredId:     mustGetSecret("GK_ACCESS_KEY"),
		CredSecret: mustGetSecret("GK_SECRET_KEY"),
	})
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	a := app.New(
		&pkg.NoOpTxManager{},
		postRepo,
		postRepo,
		mediaRepo,
		s3,
	)

	mediaWorker := workers.NewMediaCleaner(a, logger, mustGetEnvDurationOr("MEDIA_WORKER_INTERVAL", time.Hour))
	go mediaWorker.Run(ctx)

	mux := chi.NewMux()
	h := transport.NewHTTP(a, logger, mustGetEnv("S3_URL"))
	transport.Route(mux, h)

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

func mustGetSecret(key string) string {
	if file, ok := os.LookupEnv(key + "_FILE"); ok && len(file) != 0 {
		secret, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		return strings.TrimSpace(string(secret))
	}
	// else fallback to env
	return mustGetEnv(key)
}

func mustGetEnv(key string) string {
	key, ok := os.LookupEnv(key)
	if !ok {
		panic(key + " env not found")
	}
	return key

}

func mustGetBoolEnv(key string) bool {
	s := mustGetEnv(key)
	if l := strings.ToLower(s); l == "true" {
		return true
	} else if l == "false" {
		return false
	}

	panic(fmt.Sprintf("key (%s) is not bool: %s", key, s))
}

func mustGetEnvDurationOr(key string, def time.Duration) time.Duration {
	if ds, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(ds)
		if err != nil {
			panic(err)
		}
		return d
	}
	return def
}
