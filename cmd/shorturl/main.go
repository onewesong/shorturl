package main

import (
	"context"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/mine/shorturl/internal/config"
	"github.com/mine/shorturl/internal/db"
	"github.com/mine/shorturl/internal/shortcode"
)

//go:embed web/templates/*.gohtml web/static/*
var embeddedAssets embed.FS

func main() {
	cfg := config.FromEnv()

	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	storeKey := cfg.SessionSecret
	if storeKey == "" {
		storeKey = shortcode.MustRandomString(32)
		log.Printf("WARN: SESSION_SECRET 未设置，已生成临时密钥（重启会使登录失效）")
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	if err := db.EnsureAdmin(database, cfg.AdminUsername, cfg.AdminPassword); err != nil {
		log.Fatalf("ensure admin: %v", err)
	}

	tplFS, err := fs.Sub(embeddedAssets, "web/templates")
	if err != nil {
		log.Fatalf("sub templates: %v", err)
	}
	templates, err := template.ParseFS(tplFS, "*.gohtml")
	if err != nil {
		log.Fatalf("parse templates: %v", err)
	}

	staticFS, err := fs.Sub(embeddedAssets, "web/static")
	if err != nil {
		log.Fatalf("sub static: %v", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	sessionStore := cookie.NewStore([]byte(storeKey))
	sessionStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int((14 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   cfg.CookieSecure,
	})
	router.Use(sessions.Sessions("shorturl_session", sessionStore))

	router.SetHTMLTemplate(templates)
	router.StaticFS("/static", http.FS(staticFS))

	app := newApp(cfg, database)
	app.registerRoutes(router)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on http://%s", cfg.Addr())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
