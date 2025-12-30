package main

import (
	"database/sql"
	"net/http"
	"net/url"
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
	admin := r.Group("/admin")
	admin.GET("/login", a.showLogin)
	admin.POST("/login", a.doLogin)
	admin.POST("/logout", a.doLogout)

	protected := admin.Group("/")
	protected.Use(a.requireLogin)
	protected.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/admin/links") })
	protected.GET("/links", a.listLinks)
	protected.GET("/links/new", a.newLinkForm)
	protected.POST("/links", a.createLink)
	protected.GET("/links/:id/edit", a.editLinkForm)
	protected.POST("/links/:id", a.updateLink)

	r.GET("/:code", a.handleRedirect)
}

func (a *app) requireLogin(c *gin.Context) {
	sess := sessions.Default(c)
	if sess.Get("uid") == nil {
		c.Redirect(http.StatusFound, "/admin/login")
		c.Abort()
		return
	}
	c.Next()
}

func (a *app) showLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{
		"Title": "登录",
		"Err":   c.Query("err"),
	})
}

func (a *app) doLogin(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	ok, err := db.CheckPassword(a.db, username, password)
	if err != nil {
		c.Redirect(http.StatusFound, "/admin/login?err=系统错误")
		return
	}
	if !ok {
		c.Redirect(http.StatusFound, "/admin/login?err=用户名或密码错误")
		return
	}

	sess := sessions.Default(c)
	sess.Set("uid", username)
	_ = sess.Save()
	c.Redirect(http.StatusFound, "/admin/links")
}

func (a *app) doLogout(c *gin.Context) {
	sess := sessions.Default(c)
	sess.Clear()
	_ = sess.Save()
	c.Redirect(http.StatusFound, "/admin/login")
}

func (a *app) listLinks(c *gin.Context) {
	links, err := db.ListLinks(a.db, 200)
	if err != nil {
		c.String(http.StatusInternalServerError, "db error")
		return
	}

	c.HTML(http.StatusOK, "links_list", gin.H{
		"Title": "短链管理",
		"Links": links,
	})
}

func (a *app) newLinkForm(c *gin.Context) {
	c.HTML(http.StatusOK, "links_form", gin.H{
		"Title": "新建短链",
		"Mode":  "new",
		"Err":   c.Query("err"),
	})
}

func (a *app) createLink(c *gin.Context) {
	code := strings.TrimSpace(c.PostForm("code"))
	target := strings.TrimSpace(c.PostForm("target_url"))

	if target == "" || !isValidURL(target) {
		c.Redirect(http.StatusFound, "/admin/links/new?err=目标URL无效")
		return
	}

	if code != "" && !shortcode.IsValidCustom(code) {
		c.Redirect(http.StatusFound, "/admin/links/new?err=短码格式不合法")
		return
	}

	var err error
	if code == "" {
		code, err = db.GenerateUniqueCode(a.db, 6)
		if err != nil {
			c.Redirect(http.StatusFound, "/admin/links/new?err=生成短码失败")
			return
		}
	}

	if err := db.CreateLink(a.db, code, target); err != nil {
		c.Redirect(http.StatusFound, "/admin/links/new?err=创建失败(可能短码重复)")
		return
	}
	c.Redirect(http.StatusFound, "/admin/links")
}

func (a *app) editLinkForm(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "bad id")
		return
	}

	link, err := db.GetLinkByID(a.db, id)
	if err != nil {
		c.String(http.StatusNotFound, "not found")
		return
	}

	c.HTML(http.StatusOK, "links_form", gin.H{
		"Title": "编辑短链",
		"Mode":  "edit",
		"Err":   c.Query("err"),
		"Link":  link,
	})
}

func (a *app) updateLink(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "bad id")
		return
	}

	target := strings.TrimSpace(c.PostForm("target_url"))
	enabled := c.PostForm("enabled") == "on"

	if target == "" || !isValidURL(target) {
		c.Redirect(http.StatusFound, "/admin/links/"+c.Param("id")+"/edit?err=目标URL无效")
		return
	}

	if err := db.UpdateLink(a.db, id, target, enabled); err != nil {
		c.Redirect(http.StatusFound, "/admin/links/"+c.Param("id")+"/edit?err=更新失败")
		return
	}
	c.Redirect(http.StatusFound, "/admin/links")
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
