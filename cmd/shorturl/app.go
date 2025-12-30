package main

import (
	"database/sql"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/mine/shorturl/internal/config"
	"github.com/mine/shorturl/internal/db"
	"github.com/mine/shorturl/internal/shortcode"
)

type app struct {
	cfg *config.Config
	db  *sql.DB
}

func newApp(cfg *config.Config, database *sql.DB) *app {
	return &app{cfg: cfg, db: database}
}

func (a *app) registerRoutes(r *gin.Engine) {
	r.GET("/admin", func(c *gin.Context) { c.Redirect(http.StatusFound, "/admin/") })
	a.registerAdminSPA(r)

	api := r.Group("/api")
	api.POST("/login", a.apiLogin)
	api.POST("/logout", a.apiLogout)
	api.GET("/me", a.apiMe)

	protected := api.Group("/")
	protected.Use(a.requireLoginAPI)
	protected.GET("/links", a.apiListLinks)
	protected.POST("/links", a.apiCreateLink)
	protected.GET("/links/:id", a.apiGetLink)
	protected.PUT("/links/:id", a.apiUpdateLink)

	r.GET("/:code", a.handleRedirect)
}

func (a *app) requireLoginAPI(c *gin.Context) {
	sess := sessions.Default(c)
	if sess.Get("uid") == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		c.Abort()
	}
	c.Next()
}

func (a *app) apiMe(c *gin.Context) {
	sess := sessions.Default(c)
	uid := sess.Get("uid")
	if uid == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": uid})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *app) apiLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	username := strings.TrimSpace(req.Username)
	password := req.Password

	ok, err := db.CheckPassword(a.db, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "系统错误"})
		return
	}
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	sess := sessions.Default(c)
	sess.Set("uid", username)
	_ = sess.Save()
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (a *app) apiLogout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	c.Status(http.StatusNoContent)
}

func (a *app) apiListLinks(c *gin.Context) {
	links, err := db.ListLinks(a.db, 500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	c.JSON(http.StatusOK, links)
}

type createLinkRequest struct {
	Code      string `json:"code"`
	TargetURL string `json:"target_url"`
}

func (a *app) apiCreateLink(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	code := strings.TrimSpace(req.Code)
	target := strings.TrimSpace(req.TargetURL)

	if target == "" || !isValidURL(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标URL无效"})
		return
	}

	if code != "" && !shortcode.IsValidCustom(code) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "短码格式不合法"})
		return
	}

	var err error
	if code == "" {
		code, err = db.GenerateUniqueCode(a.db, 6)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成短码失败"})
			return
		}
	}

	if err := db.CreateLink(a.db, code, target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建失败(可能短码重复)"})
		return
	}
	link, err := db.GetLinkByCode(a.db, code)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, link)
}

func (a *app) apiGetLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}

	link, err := db.GetLinkByID(a.db, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, link)
}

type updateLinkRequest struct {
	TargetURL string `json:"target_url"`
	Enabled   *bool  `json:"enabled"`
}

func (a *app) apiUpdateLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad id"})
		return
	}

	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	target := strings.TrimSpace(req.TargetURL)
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if target == "" || !isValidURL(target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标URL无效"})
		return
	}

	if err := db.UpdateLink(a.db, id, target, enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}
	link, err := db.GetLinkByID(a.db, id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	c.JSON(http.StatusOK, link)
}

func (a *app) handleRedirect(c *gin.Context) {
	code := c.Param("code")
	code = strings.TrimPrefix(code, "/")
	if code == "" || strings.Contains(code, "/") {
		c.Status(http.StatusNotFound)
		return
	}

	link, err := db.GetLinkByCode(a.db, code)
	if err != nil || link == nil || !link.Enabled {
		c.Status(http.StatusNotFound)
		return
	}

	_ = db.IncrementClick(a.db, link.ID)
	c.Redirect(http.StatusFound, link.TargetURL)
}

func isValidURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func (a *app) registerAdminSPA(r *gin.Engine) {
	distFS, err := fs.Sub(embeddedAssets, "web/dist")
	if err != nil {
		// embed 失败时直接 404，避免启动崩溃影响短链跳转
		r.GET("/admin", func(c *gin.Context) { c.Status(http.StatusNotFound) })
		r.GET("/admin/*path", func(c *gin.Context) { c.Status(http.StatusNotFound) })
		return
	}

	serveFile := func(c *gin.Context, rel string) bool {
		b, err := fs.ReadFile(distFS, rel)
		if err != nil {
			return false
		}
		ct := mime.TypeByExtension(path.Ext(rel))
		if ct == "" {
			ct = http.DetectContentType(b)
		}
		if strings.HasPrefix(rel, "assets/") {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			c.Header("Cache-Control", "no-cache")
		}
		c.Data(http.StatusOK, ct, b)
		return true
	}

	serve := func(c *gin.Context) {
		// filepath 示例：/admin/  /admin/assets/xx.js  /admin/links
		rel := strings.TrimPrefix(c.Param("filepath"), "/")
		if rel == "" {
			rel = "index.html"
		}

		rel = path.Clean(rel)
		if rel == "." || strings.HasPrefix(rel, "..") {
			c.Status(http.StatusBadRequest)
			return
		}

		if serveFile(c, rel) {
			return
		}
		_ = serveFile(c, "index.html")
	}

	r.GET("/admin/*filepath", serve)
}
