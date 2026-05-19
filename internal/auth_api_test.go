package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()

	loginResp := s.doJSON(http.MethodPost, "/api/iam/auth/login", map[string]any{"username": "admin", "password": "123456"}, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", loginResp.Code, loginResp.Body.String())
	}
	var loginEnvelope apiEnvelope
	decodeAPIJSON(t, loginResp.Body.Bytes(), &loginEnvelope)
	if loginEnvelope.Code != 0 {
		t.Fatalf("login code=%d", loginEnvelope.Code)
	}
	var loginData apiLoginData
	decodeAPIJSON(t, loginEnvelope.Data, &loginData)
	if loginData.AccessToken == "" {
		t.Fatal("expected access token")
	}

	meResp := s.doJSON(http.MethodGet, "/api/iam/auth/me", nil, loginData.AccessToken)
	if meResp.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meResp.Code, meResp.Body.String())
	}
	var meEnvelope apiEnvelope
	decodeAPIJSON(t, meResp.Body.Bytes(), &meEnvelope)
	var meData struct {
		OpenID      *string `json:"openid"`
		Email       *string `json:"email"`
		Mobile      *string `json:"mobile"`
		Remark      string  `json:"remark"`
		LastLoginAt *string `json:"last_login_at"`
	}
	decodeAPIJSON(t, meEnvelope.Data, &meData)
	if meData.OpenID == nil || *meData.OpenID != "ou_admin" {
		t.Fatalf("expected /auth/me openid=ou_admin, got %#v", meData.OpenID)
	}
	if meData.Email == nil || *meData.Email != "admin@example.com" || meData.Mobile == nil || *meData.Mobile != "13900000000" {
		t.Fatalf("expected /auth/me contact fields, got email=%#v mobile=%#v", meData.Email, meData.Mobile)
	}
	if meData.Remark == "" || meData.LastLoginAt == nil {
		t.Fatalf("expected /auth/me remark and last_login_at, got remark=%q last_login_at=%#v", meData.Remark, meData.LastLoginAt)
	}

	logoutResp := s.doJSON(http.MethodPost, "/api/iam/auth/logout", nil, loginData.AccessToken)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logoutResp.Code, logoutResp.Body.String())
	}

	blockedResp := s.doJSON(http.MethodGet, "/api/iam/auth/me", nil, loginData.AccessToken)
	if blockedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to be 401, got %d", blockedResp.Code)
	}

	badLoginResp := s.doJSON(http.MethodPost, "/api/iam/auth/login", map[string]any{"username": "admin", "password": "wrong"}, "")
	if badLoginResp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad login to be 400, got %d", badLoginResp.Code)
	}
}

func TestCORSPreflight(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()

	req := httptest.NewRequest(http.MethodOptions, "/api/iam/users", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	s.engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected CORS preflight status 204, got %d body=%s", w.Code, w.Body.String())
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("unexpected allow origin: %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatalf("expected credentials allowed")
	}
}
