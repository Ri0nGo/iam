package internal_test

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"iam/internal/model"
)

func TestUserAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()
	admin := s.loginAsAdmin()

	roleResp := s.doJSON(http.MethodPost, "/api/v1/roles", map[string]any{
		"code":   "operator",
		"name":   "运营角色",
		"remark": "业务运营",
	}, admin.AccessToken)
	if roleResp.Code != http.StatusOK {
		t.Fatalf("create operator role status=%d body=%s", roleResp.Code, roleResp.Body.String())
	}

	createResp := s.doJSON(http.MethodPost, "/api/v1/users", map[string]any{
		"username":     "alice",
		"password":     "123456",
		"display_name": "Alice",
		"email":        "alice@example.com",
		"mobile":       "13800000000",
		"remark":       "demo user",
		"role_codes":   []string{"operator"},
	}, admin.AccessToken)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create user status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var userEnvelope apiEnvelope
	decodeAPIJSON(t, createResp.Body.Bytes(), &userEnvelope)
	var createdUser model.User
	decodeAPIJSON(t, userEnvelope.Data, &createdUser)
	if createdUser.ID == 0 {
		t.Fatal("expected created user id")
	}
	if createdUser.OpenID == nil || *createdUser.OpenID == "" {
		t.Fatal("expected generated openid")
	}
	if !regexp.MustCompile(`^o[0-9A-Za-z]{27}$`).MatchString(*createdUser.OpenID) {
		t.Fatalf("expected wechat-style openid, got %q", *createdUser.OpenID)
	}

	listResp := s.doJSON(http.MethodGet, "/api/v1/users?keyword=alice", nil, admin.AccessToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list users status=%d body=%s", listResp.Code, listResp.Body.String())
	}

	detailResp := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/users/%d", createdUser.ID), nil, admin.AccessToken)
	if detailResp.Code != http.StatusOK {
		t.Fatalf("user detail status=%d body=%s", detailResp.Code, detailResp.Body.String())
	}

	rolesResp := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/users/%d/roles", createdUser.ID), nil, admin.AccessToken)
	if rolesResp.Code != http.StatusOK {
		t.Fatalf("user roles status=%d body=%s", rolesResp.Code, rolesResp.Body.String())
	}

	bindResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/v1/users/%d/roles", createdUser.ID), map[string]any{"role_codes": []string{"admin", "operator"}}, admin.AccessToken)
	if bindResp.Code != http.StatusOK {
		t.Fatalf("bind user roles status=%d body=%s", bindResp.Code, bindResp.Body.String())
	}

	statusResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/v1/users/%d/status", createdUser.ID), map[string]any{"status": 2}, admin.AccessToken)
	if statusResp.Code != http.StatusOK {
		t.Fatalf("update user status=%d body=%s", statusResp.Code, statusResp.Body.String())
	}

	passwordResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/v1/users/%d/password", createdUser.ID), map[string]any{"password": "new-password"}, admin.AccessToken)
	if passwordResp.Code != http.StatusOK {
		t.Fatalf("reset password status=%d body=%s", passwordResp.Code, passwordResp.Body.String())
	}

	disabledLoginResp := s.doJSON(http.MethodPost, "/api/v1/auth/login", map[string]any{"username": "alice", "password": "new-password"}, "")
	if disabledLoginResp.Code != http.StatusBadRequest {
		t.Fatalf("expected disabled user login to be 400, got %d", disabledLoginResp.Code)
	}
}
