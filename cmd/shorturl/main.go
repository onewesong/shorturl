package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"

	"github.com/mine/shorturl/internal/config"
	"github.com/mine/shorturl/internal/httpapi"
	"github.com/mine/shorturl/internal/links"
	"github.com/mine/shorturl/internal/shortcode"
	sqlitestore "github.com/mine/shorturl/internal/store/sqlite"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.FromEnv()

	storeKey := cfg.SessionSecret
	if storeKey == "" {
		storeKey = shortcode.MustRandomString(32)
		logger.Warn("SESSION_SECRET 未设置，已生成临时密钥（重启会使登录失效）")
	}

	database, err := sqlitestore.Open(cfg.DBPath)
	if err != nil {
		logger.Error("open database failed", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	ctx := context.Background()
	if err := sqlitestore.Init(ctx, database); err != nil {
		logger.Error("init database failed", "error", err)
		os.Exit(1)
	}

	userRepo := sqlitestore.NewUserRepository(database)
	if err := userRepo.EnsureAdmin(ctx, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		logger.Error("ensure admin failed", "error", err)
		os.Exit(1)
	}

	sessionStore := cookie.NewStore([]byte(storeKey))
	sessionStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int((14 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.CookieSecure,
	})

	linkRepo := sqlitestore.NewLinkRepository(database)
	linkService := links.NewService(linkRepo)
	router := httpapi.NewRouter(logger, sessionStore, cfg.AdminStaticDir, linkService, userRepo)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("shorturl server started", "addr", cfg.Addr(), "database", cfg.DBPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen server failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
