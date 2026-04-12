package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/mine/shorturl/internal/links"
)

const sessionUserKey = "uid"

type authChecker interface {
	CheckPassword(ctx context.Context, username string, password string) (bool, error)
}

type apiResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewRouter(
	logger *slog.Logger,
	sessionStore sessions.Store,
	adminStaticDir string,
	linkService *links.Service,
	auth authChecker,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(logger))
	router.Use(sessions.Sessions("shorturl_session", sessionStore))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	registerAdminRoutes(router, adminStaticDir, linkService, auth)

	router.GET("/:code", func(c *gin.Context) {
		targetURL, err := linkService.Resolve(c.Request.Context(), c.Param("code"), links.VisitMeta{
			VisitedAt:   time.Now().UTC(),
			IP:          c.ClientIP(),
			Referer:     c.Request.Referer(),
			RefererHost: "",
			UserAgent:   c.Request.UserAgent(),
		})
		if err != nil {
			if errors.Is(err, links.ErrLinkNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			logger.Error("resolve link failed", "error", err, "code", c.Param("code"))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Redirect(http.StatusFound, targetURL)
	})

	return router
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		logger.Info(
			"request handled",
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(started).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func registerAdminSPARoutes(router *gin.Engine, staticDir string) {
	indexPath := filepath.Join(staticDir, "index.html")
	assetsDir := filepath.Join(staticDir, "assets")

	router.GET("/admin", func(c *gin.Context) {
		serveAdminIndex(c, indexPath)
	})

	if fileInfo, err := os.Stat(assetsDir); err == nil && fileInfo.IsDir() {
		router.StaticFS("/admin/assets", gin.Dir(assetsDir, false))
	}

	router.NoRoute(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/admin") || strings.HasPrefix(c.Request.URL.Path, "/admin/api/") {
			c.Status(http.StatusNotFound)
			return
		}

		serveAdminIndex(c, indexPath)
	})
}

func serveAdminIndex(c *gin.Context, indexPath string) {
	if _, err := os.Stat(indexPath); err == nil {
		c.File(indexPath)
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusServiceUnavailable, "admin frontend not built yet. Run `make admin-install admin-build` first.")
}

func requireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if currentUsername(c) == "" {
			writeJSONError(c, http.StatusUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		c.Next()
	}
}

func currentUsername(c *gin.Context) string {
	session := sessions.Default(c)
	username, _ := session.Get(sessionUserKey).(string)
	return strings.TrimSpace(username)
}

func writeJSONError(c *gin.Context, status int, code string) {
	c.JSON(status, apiResponse{
		Success: false,
		Error:   code,
	})
}
