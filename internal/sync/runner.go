package sync

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github-star-sync/internal/classify"
	"github-star-sync/internal/config"
	gh "github-star-sync/internal/github"
	"github-star-sync/internal/model"
	"github-star-sync/internal/render"
)

// Runner orchestrates fetch → classify → write.
type Runner struct {
	cfg    *config.Config
	client *gh.Client
	logger *slog.Logger
}

// New creates a Runner.
func New(cfg *config.Config, client *gh.Client, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{cfg: cfg, client: client, logger: logger}
}

// Run fetches all sources and writes outputs.
func (r *Runner) Run(ctx context.Context) error {
	report := model.Report{
		Title:     r.cfg.Title,
		Generated: time.Now(),
	}

	for _, src := range r.cfg.Sources {
		r.logger.Info("fetching stars", "user", src.Username)
		repos, err := r.client.ListStarred(ctx, src.Username)
		u := model.UserStars{
			Username: src.Username,
			Label:    src.Label,
		}
		if err != nil {
			u.Err = err.Error()
			r.logger.Error("fetch failed", "user", src.Username, "err", err)
			report.Users = append(report.Users, u)
			continue
		}
		u.Total = len(repos)
		u.Categories = classify.Group(repos, r.cfg.Classify)
		report.Users = append(report.Users, u)
		report.Total += u.Total
		r.logger.Info("classified", "user", src.Username, "repos", u.Total, "categories", len(u.Categories))
	}

	if r.cfg.OutputMD != "" {
		if err := writeFile(r.cfg.OutputMD, func(f *os.File) error {
			return render.Markdown(f, report)
		}); err != nil {
			return fmt.Errorf("write markdown: %w", err)
		}
		r.logger.Info("wrote markdown", "path", r.cfg.OutputMD)
	}
	if r.cfg.OutputHTML != "" {
		if err := writeFile(r.cfg.OutputHTML, func(f *os.File) error {
			return render.HTML(f, report)
		}); err != nil {
			return fmt.Errorf("write html: %w", err)
		}
		r.logger.Info("wrote html", "path", r.cfg.OutputHTML)
	}
	return nil
}

func writeFile(path string, fn func(*os.File) error) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}
