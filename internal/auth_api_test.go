package internal_test

import (
	"net/http"
	"testing"
)

func TestAuthAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()

	loginResp := s.doJSON(http.MethodPost, "/api/v1/auth/login", map[string]any{"username": "admin", "password": "123456"}, "")
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

	meResp := s.doJSON(http.MethodGet, "/api/v1/auth/me", nil, loginData.AccessToken)
	if meResp.Code != http.StatusOK {
		t.Fatalf("me status=%d body=%s", meResp.Code, meResp.Body.String())
	}

	logoutResp := s.doJSON(http.MethodPost, "/api/v1/auth/logout", nil, loginData.AccessToken)
	if logoutResp.Code != http.StatusOK {
		t.Fatalf("logout status=%d body=%s", logoutResp.Code, logoutResp.Body.String())
	}

	blockedResp := s.doJSON(http.MethodGet, "/api/v1/auth/me", nil, loginData.AccessToken)
	if blockedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to be 401, got %d", blockedResp.Code)
	}

	badLoginResp := s.doJSON(http.MethodPost, "/api/v1/auth/login", map[string]any{"username": "admin", "password": "wrong"}, "")
	if badLoginResp.Code != http.StatusBadRequest {
		t.Fatalf("expected bad login to be 400, got %d", badLoginResp.Code)
	}
}
