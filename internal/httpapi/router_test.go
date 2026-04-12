package httpapi

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/mine/shorturl/internal/links"
	"github.com/mine/shorturl/internal/store/sqlite"
)

func TestHealthz(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestAdminListUnauthorized(t *testing.T) {
	router := newTestRouter(t)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/links", nil)
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", recorder.Code)
	}
}

func TestAdminLoginAndSession(t *testing.T) {
	router := newTestRouter(t)

	loginRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/auth/login", `{"username":"admin","password":"change-me"}`, "")
	if loginRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", loginRecorder.Code, loginRecorder.Body.String())
	}

	sessionCookie := loginRecorder.Header().Get("Set-Cookie")
	if !strings.Contains(sessionCookie, "shorturl_session=") {
		t.Fatalf("expected session cookie, got %q", sessionCookie)
	}

	sessionRecorder := performJSONRequest(router, http.MethodGet, "/admin/api/v1/auth/session", "", sessionCookie)
	if sessionRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", sessionRecorder.Code, sessionRecorder.Body.String())
	}
	if !strings.Contains(sessionRecorder.Body.String(), `"username":"admin"`) {
		t.Fatalf("expected username in session response, got %s", sessionRecorder.Body.String())
	}
}

func TestAdminLoginRejectsBadPassword(t *testing.T) {
	router := newTestRouter(t)

	recorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/auth/login", `{"username":"admin","password":"wrong"}`, "")
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateLinkFlow(t *testing.T) {
	router := newTestRouter(t)
	sessionCookie := login(t, router)

	badRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"bad","target_url":"ftp://example.com"}`, sessionCookie)
	if badRecorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid url, got %d body=%s", badRecorder.Code, badRecorder.Body.String())
	}

	createRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"launch","target_url":"https://example.com/start","remark":"启动页","tags":["marketing","spring"]}`, sessionCookie)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}
	if !strings.Contains(createRecorder.Body.String(), `"remark":"启动页"`) || !strings.Contains(createRecorder.Body.String(), `"tags":["marketing","spring"]`) {
		t.Fatalf("expected remark and tags in create response, got %s", createRecorder.Body.String())
	}

	conflictRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"launch","target_url":"https://example.com/again"}`, sessionCookie)
	if conflictRecorder.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d body=%s", conflictRecorder.Code, conflictRecorder.Body.String())
	}
}

func TestUpdateLinkFlow(t *testing.T) {
	router := newTestRouter(t)
	sessionCookie := login(t, router)

	createRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"edit-me","target_url":"https://example.com/original","remark":"初始备注","tags":["ops"]}`, sessionCookie)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	listRecorder := performJSONRequest(router, http.MethodGet, "/admin/api/v1/links", "", sessionCookie)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"code":"edit-me"`) {
		t.Fatalf("expected created link in list, got %s", listRecorder.Body.String())
	}

	updateRecorder := performJSONRequest(router, http.MethodPut, "/admin/api/v1/links/1", `{"code":"edited","target_url":"https://example.com/updated","remark":"改过的备注","tags":["ops","done"],"enabled":false}`, sessionCookie)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}
	if !strings.Contains(updateRecorder.Body.String(), `"enabled":false`) || !strings.Contains(updateRecorder.Body.String(), `"remark":"改过的备注"`) {
		t.Fatalf("expected disabled response, got %s", updateRecorder.Body.String())
	}
}

func TestRedirectIncrementsClicksAndHonorsDisabled(t *testing.T) {
	router := newTestRouter(t)
	sessionCookie := login(t, router)

	createRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"go-now","target_url":"https://example.com/live"}`, sessionCookie)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	redirectRecorder := httptest.NewRecorder()
	redirectReq := httptest.NewRequest(http.MethodGet, "/go-now", nil)
	router.ServeHTTP(redirectRecorder, redirectReq)

	if redirectRecorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d body=%s", redirectRecorder.Code, redirectRecorder.Body.String())
	}
	if location := redirectRecorder.Header().Get("Location"); location != "https://example.com/live" {
		t.Fatalf("expected redirect location, got %q", location)
	}

	updateRecorder := performJSONRequest(router, http.MethodPut, "/admin/api/v1/links/1", `{"code":"go-now","target_url":"https://example.com/live","enabled":false}`, sessionCookie)
	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", updateRecorder.Code, updateRecorder.Body.String())
	}

	disabledRecorder := httptest.NewRecorder()
	disabledReq := httptest.NewRequest(http.MethodGet, "/go-now", nil)
	router.ServeHTTP(disabledRecorder, disabledReq)

	if disabledRecorder.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", disabledRecorder.Code)
	}
}

func TestLinkAnalyticsIncludesVisitContext(t *testing.T) {
	router := newTestRouter(t)
	sessionCookie := login(t, router)

	createRecorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/links", `{"code":"track-me","target_url":"https://example.com/live"}`, sessionCookie)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", createRecorder.Code, createRecorder.Body.String())
	}

	redirectRecorder := httptest.NewRecorder()
	redirectReq := httptest.NewRequest(http.MethodGet, "/track-me", nil)
	redirectReq.Header.Set("Referer", "https://mp.weixin.qq.com/s/demo")
	redirectReq.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148 MicroMessenger/8.0.49")
	redirectReq.RemoteAddr = "10.2.3.4:12345"
	router.ServeHTTP(redirectRecorder, redirectReq)

	if redirectRecorder.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d body=%s", redirectRecorder.Code, redirectRecorder.Body.String())
	}

	analyticsRecorder := performJSONRequest(router, http.MethodGet, "/admin/api/v1/links/1/analytics?days=7", "", sessionCookie)
	if analyticsRecorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", analyticsRecorder.Code, analyticsRecorder.Body.String())
	}

	body := analyticsRecorder.Body.String()
	if !strings.Contains(body, `"recent_clicks":1`) {
		t.Fatalf("expected recent_clicks in analytics response, got %s", body)
	}
	if !strings.Contains(body, `"referer_host":"mp.weixin.qq.com"`) {
		t.Fatalf("expected referer host in analytics response, got %s", body)
	}
	if !strings.Contains(body, `"client_name":"微信"`) {
		t.Fatalf("expected client name in analytics response, got %s", body)
	}
	if !strings.Contains(body, `"ip_masked":"10.2.3.*"`) {
		t.Fatalf("expected masked ip in analytics response, got %s", body)
	}
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dbPath := filepath.Join(t.TempDir(), "shorturl-test.db")
	database, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	if err := sqlite.Init(ctx, database); err != nil {
		t.Fatalf("init db: %v", err)
	}

	users := sqlite.NewUserRepository(database)
	if err := users.EnsureAdmin(ctx, "admin", "change-me"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}

	linkRepo := sqlite.NewLinkRepository(database)
	linkService := links.NewService(linkRepo)

	store := cookie.NewStore([]byte("01234567890123456789012345678901"))
	store.Options(sessionsOptions())

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return NewRouter(logger, store, t.TempDir(), linkService, users)
}

func sessionsOptions() sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   int((14 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func login(t *testing.T, router *gin.Engine) string {
	t.Helper()

	recorder := performJSONRequest(router, http.MethodPost, "/admin/api/v1/auth/login", `{"username":"admin","password":"change-me"}`, "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("login failed: %d body=%s", recorder.Code, recorder.Body.String())
	}

	return recorder.Header().Get("Set-Cookie")
}

func performJSONRequest(router *gin.Engine, method string, path string, body string, cookieHeader string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	var bodyReader *bytes.Reader
	if body == "" {
		bodyReader = bytes.NewReader(nil)
	} else {
		bodyReader = bytes.NewReader([]byte(body))
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookieHeader != "" {
		req.Header.Set("Cookie", strings.Split(cookieHeader, ";")[0])
	}

	router.ServeHTTP(recorder, req)
	return recorder
}
