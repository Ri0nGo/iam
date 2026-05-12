package internal_test

import (
	"net/http"
	"net/url"
	"testing"

	"iam/internal/dto"
)

func TestOAuthAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()

	admin := s.loginAsAdmin()

	badPreviewResp := s.doJSON(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=system-a&redirect_uri=http://evil.local/callback", nil, admin.AccessToken)
	if badPreviewResp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid redirect_uri to be 400, got %d", badPreviewResp.Code)
	}

	authorizeResp := s.doJSON(http.MethodGet, "/api/v1/oauth/authorize?response_type=code&client_id=system-a&redirect_uri=http://system-a.local/callback&state=xyz&scope=basic", nil, admin.AccessToken)
	if authorizeResp.Code != http.StatusFound {
		t.Fatalf("oauth authorize status=%d body=%s", authorizeResp.Code, authorizeResp.Body.String())
	}
	redirectURL, err := url.Parse(authorizeResp.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	code := redirectURL.Query().Get("code")
	if code == "" {
		t.Fatal("expected authorization code")
	}

	tokenResp := s.doJSON(http.MethodPost, "/api/v1/oauth/token", map[string]any{
		"grant_type":    "authorization_code",
		"client_id":     "system-a",
		"client_secret": "system-a-secret",
		"code":          code,
		"redirect_uri":  "http://system-a.local/callback",
	}, "")
	if tokenResp.Code != http.StatusOK {
		t.Fatalf("oauth token status=%d body=%s", tokenResp.Code, tokenResp.Body.String())
	}
	var token dto.TokenResponse
	decodeAPIJSON(t, tokenResp.Body.Bytes(), &token)
	if token.AccessToken == "" {
		t.Fatal("expected oauth access token")
	}

	userinfoResp := s.doJSON(http.MethodGet, "/api/v1/oauth/userinfo", nil, token.AccessToken)
	if userinfoResp.Code != http.StatusOK {
		t.Fatalf("oauth userinfo status=%d body=%s", userinfoResp.Code, userinfoResp.Body.String())
	}

	reusedTokenResp := s.doJSON(http.MethodPost, "/api/v1/oauth/token", map[string]any{
		"grant_type":    "authorization_code",
		"client_id":     "system-a",
		"client_secret": "system-a-secret",
		"code":          code,
		"redirect_uri":  "http://system-a.local/callback",
	}, "")
	if reusedTokenResp.Code != http.StatusBadRequest {
		t.Fatalf("expected reused code to be 400, got %d", reusedTokenResp.Code)
	}
}
