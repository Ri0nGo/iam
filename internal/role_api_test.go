package internal_test

import (
	"net/http"
	"testing"
)

func TestRoleAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()
	admin := s.loginAsAdmin()

	createResp := s.doJSON(http.MethodPost, "/api/iam/roles", map[string]any{
		"code":   "operator",
		"name":   "运营角色",
		"remark": "业务运营",
	}, admin.AccessToken)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create role status=%d body=%s", createResp.Code, createResp.Body.String())
	}

	listResp := s.doJSON(http.MethodGet, "/api/iam/roles", nil, admin.AccessToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list roles status=%d body=%s", listResp.Code, listResp.Body.String())
	}

	unauthorizedResp := s.doJSON(http.MethodGet, "/api/iam/roles", nil, "")
	if unauthorizedResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized roles request to be 401, got %d", unauthorizedResp.Code)
	}
}
