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

	tokenURL := "/api/v1/oauth/token?client_id=system-a&secret=system-a-secret&grant_type=authorization_code&code=" + url.QueryEscape(code)
	tokenResp := s.doJSON(http.MethodGet, tokenURL, nil, "")
	if tokenResp.Code != http.StatusOK {
		t.Fatalf("oauth token status=%d body=%s", tokenResp.Code, tokenResp.Body.String())
	}
	var tokenEnvelope apiEnvelope
	decodeAPIJSON(t, tokenResp.Body.Bytes(), &tokenEnvelope)
	var token dto.TokenResponse
	decodeAPIJSON(t, tokenEnvelope.Data, &token)
	if token.AccessToken == "" || token.RefreshToken == "" {
		t.Fatal("expected oauth access token and refresh token")
	}
	if token.OpenID != "ou_admin" {
		t.Fatalf("expected oauth token response openid=ou_admin, got %q", token.OpenID)
	}

	loginTokenUserinfoResp := s.doJSON(http.MethodGet, "/api/v1/oauth/userinfo?access_token="+url.QueryEscape(admin.AccessToken), nil, "")
	if loginTokenUserinfoResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected login access token to be rejected by oauth userinfo, got %d", loginTokenUserinfoResp.Code)
	}

	missingOpenIDResp := s.doJSON(http.MethodGet, "/api/v1/oauth/userinfo?access_token="+url.QueryEscape(token.AccessToken), nil, "")
	if missingOpenIDResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing openid to be rejected by oauth userinfo, got %d", missingOpenIDResp.Code)
	}

	userinfoResp := s.doJSON(http.MethodGet, "/api/v1/oauth/userinfo?access_token="+url.QueryEscape(token.AccessToken)+"&openid="+url.QueryEscape(token.OpenID), nil, "")
	if userinfoResp.Code != http.StatusOK {
		t.Fatalf("oauth userinfo status=%d body=%s", userinfoResp.Code, userinfoResp.Body.String())
	}
	var userinfoEnvelope apiEnvelope
	decodeAPIJSON(t, userinfoResp.Body.Bytes(), &userinfoEnvelope)
	var userinfo map[string]any
	decodeAPIJSON(t, userinfoEnvelope.Data, &userinfo)
	if userinfo["openid"] != "ou_admin" {
		t.Fatalf("expected openid ou_admin, got %v", userinfo["openid"])
	}
	if _, ok := userinfo["id"]; ok {
		t.Fatalf("expected oauth userinfo to omit id, got %s", string(userinfoEnvelope.Data))
	}

	checkResp := s.doJSON(http.MethodGet, "/api/v1/oauth/auth?access_token="+url.QueryEscape(token.AccessToken)+"&openid="+url.QueryEscape(token.OpenID), nil, "")
	if checkResp.Code != http.StatusOK {
		t.Fatalf("oauth auth status=%d body=%s", checkResp.Code, checkResp.Body.String())
	}

	refreshResp := s.doJSON(http.MethodGet, "/api/v1/oauth/refresh_token?client_id=system-a&grant_type=refresh_token&refresh_token="+url.QueryEscape(token.RefreshToken), nil, "")
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("oauth refresh_token status=%d body=%s", refreshResp.Code, refreshResp.Body.String())
	}
	var refreshed dto.TokenResponse
	var refreshEnvelope apiEnvelope
	decodeAPIJSON(t, refreshResp.Body.Bytes(), &refreshEnvelope)
	decodeAPIJSON(t, refreshEnvelope.Data, &refreshed)
	if refreshed.AccessToken == "" || refreshed.OpenID != token.OpenID || refreshed.RefreshToken != token.RefreshToken {
		t.Fatalf("unexpected refresh token response: %+v", refreshed)
	}

	reusedTokenResp := s.doJSON(http.MethodGet, tokenURL, nil, "")
	if reusedTokenResp.Code != http.StatusBadRequest {
		t.Fatalf("expected reused code to be 400, got %d", reusedTokenResp.Code)
	}
}
