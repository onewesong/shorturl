package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/mine/shorturl/internal/links"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func registerAdminRoutes(router *gin.Engine, adminStaticDir string, linkService *links.Service, auth authChecker) {
	adminAPI := router.Group("/admin/api/v1")
	adminAPI.POST("/auth/login", loginHandler(auth))
	adminAPI.POST("/auth/logout", logoutHandler())
	adminAPI.GET("/auth/session", sessionHandler())

	protected := adminAPI.Group("/")
	protected.Use(requireLogin())
	protected.GET("/links", listLinksHandler(linkService))
	protected.GET("/links/:id/analytics", getLinkAnalyticsHandler(linkService))
	protected.POST("/links", createLinkHandler(linkService))
	protected.PUT("/links/:id", updateLinkHandler(linkService))

	registerAdminSPARoutes(router, adminStaticDir)
}

func loginHandler(auth authChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request")
			return
		}

		username := strings.TrimSpace(request.Username)
		ok, err := auth.CheckPassword(c.Request.Context(), username, request.Password)
		if err != nil {
			writeJSONError(c, http.StatusInternalServerError, "internal_error")
			return
		}
		if !ok {
			writeJSONError(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		session := sessions.Default(c)
		session.Set(sessionUserKey, username)
		if err := session.Save(); err != nil {
			writeJSONError(c, http.StatusInternalServerError, "internal_error")
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data: gin.H{
				"authenticated": true,
				"username":      username,
			},
		})
	}
}

func logoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		if err := session.Save(); err != nil {
			writeJSONError(c, http.StatusInternalServerError, "internal_error")
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    gin.H{"ok": true},
		})
	}
}

func sessionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := currentUsername(c)
		if username == "" {
			writeJSONError(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data: gin.H{
				"authenticated": true,
				"username":      username,
			},
		})
	}
}

func listLinksHandler(linkService *links.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := linkService.List(c.Request.Context())
		if err != nil {
			writeJSONError(c, http.StatusInternalServerError, "internal_error")
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    result,
		})
	}
}

func createLinkHandler(linkService *links.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request links.CreateLinkInput
		if err := c.ShouldBindJSON(&request); err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request")
			return
		}

		link, err := linkService.Create(c.Request.Context(), request)
		if err != nil {
			writeLinkError(c, err)
			return
		}

		c.JSON(http.StatusCreated, apiResponse{
			Success: true,
			Data:    link,
		})
	}
}

func getLinkAnalyticsHandler(linkService *links.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request")
			return
		}

		days := 7
		if rawDays := strings.TrimSpace(c.Query("days")); rawDays != "" {
			parsedDays, parseErr := strconv.Atoi(rawDays)
			if parseErr != nil {
				writeJSONError(c, http.StatusBadRequest, "invalid_request")
				return
			}
			days = parsedDays
		}

		analytics, err := linkService.Analytics(c.Request.Context(), id, days)
		if err != nil {
			writeLinkError(c, err)
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    analytics,
		})
	}
}

func updateLinkHandler(linkService *links.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request")
			return
		}

		var request links.UpdateLinkInput
		if err := c.ShouldBindJSON(&request); err != nil {
			writeJSONError(c, http.StatusBadRequest, "invalid_request")
			return
		}

		link, err := linkService.Update(c.Request.Context(), id, request)
		if err != nil {
			writeLinkError(c, err)
			return
		}

		c.JSON(http.StatusOK, apiResponse{
			Success: true,
			Data:    link,
		})
	}
}

func writeLinkError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	code := "internal_error"

	switch {
	case errors.Is(err, links.ErrValidation):
		status = http.StatusBadRequest
		code = "invalid_request"
	case errors.Is(err, links.ErrLinkExists):
		status = http.StatusConflict
		code = "conflict"
	case errors.Is(err, links.ErrLinkNotFound):
		status = http.StatusNotFound
		code = "not_found"
	}

	writeJSONError(c, status, code)
}
