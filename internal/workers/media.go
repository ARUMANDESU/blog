package workers

import (
	"context"
	"log/slog"
	"time"

	"github.com/arumandesu/blog/internal/app"
)

type MediaCleaner struct {
	app      *app.App
	logger   *slog.Logger
	interval time.Duration
}

func NewMediaCleaner(app *app.App, logger *slog.Logger, interval time.Duration) *MediaCleaner {
	return &MediaCleaner{app: app, logger: logger, interval: interval}
}

func (mc *MediaCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(mc.interval)
	defer ticker.Stop()

	mc.cleanup(ctx)
	for {
		select {
		case <-ticker.C:
			mc.cleanup(ctx)
		case <-ctx.Done():
			mc.logger.Debug("media cleaner stopped")
			return
		}
	}

}

func (mc *MediaCleaner) cleanup(ctx context.Context) {
	err := mc.app.DeleteUnusedMedia(ctx)
	if err != nil {
		mc.logger.Error("failed to delete unused media", "err", err.Error())
	}
}
