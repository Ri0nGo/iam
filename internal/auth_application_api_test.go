package internal_test

import (
	"fmt"
	"net/http"
	"testing"

	"iam/internal/dto"
)

func TestAuthApplicationAPI(t *testing.T) {
	s := newAPISuite(t)
	defer s.close()
	admin := s.loginAsAdmin()

	createResp := s.doJSON(http.MethodPost, "/api/v1/auth-applications", map[string]any{
		"name":          "系统 C OAuth2 认证",
		"code":          "system-c-oauth2",
		"client_id":     "system-c",
		"secret_key":    "system-c-secret",
		"response_type": "code",
		"redirect_uri":  "http://system-c.local/callback",
		"status":        1,
		"remark":        "system c client",
	}, admin.AccessToken)
	if createResp.Code != http.StatusOK {
		t.Fatalf("create auth application status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var createEnvelope apiEnvelope
	decodeAPIJSON(t, createResp.Body.Bytes(), &createEnvelope)
	var created dto.AuthApplicationResponse
	decodeAPIJSON(t, createEnvelope.Data, &created)
	if created.ID == 0 || created.ClientID != "system-c" || created.SecretKey != "system-c-secret" {
		t.Fatalf("unexpected created auth application: %+v", created)
	}

	listResp := s.doJSON(http.MethodGet, "/api/v1/auth-applications?keyword=system-c", nil, admin.AccessToken)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list auth applications status=%d body=%s", listResp.Code, listResp.Body.String())
	}

	getResp := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/auth-applications/%d", created.ID), nil, admin.AccessToken)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get auth application status=%d body=%s", getResp.Code, getResp.Body.String())
	}

	updateResp := s.doJSON(http.MethodPut, fmt.Sprintf("/api/v1/auth-applications/%d", created.ID), map[string]any{
		"name":          "系统 C OAuth2 认证 Updated",
		"code":          "system-c-oauth2-updated",
		"client_id":     "system-c-updated",
		"secret_key":    "system-c-secret-updated",
		"response_type": "code",
		"redirect_uri":  "http://system-c.local/oauth/callback",
		"status":        2,
		"remark":        "updated",
	}, admin.AccessToken)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update auth application status=%d body=%s", updateResp.Code, updateResp.Body.String())
	}
	var updateEnvelope apiEnvelope
	decodeAPIJSON(t, updateResp.Body.Bytes(), &updateEnvelope)
	var updated dto.AuthApplicationResponse
	decodeAPIJSON(t, updateEnvelope.Data, &updated)
	if updated.ClientID != "system-c-updated" || updated.Status != 2 {
		t.Fatalf("unexpected updated auth application: %+v", updated)
	}

	deleteResp := s.doJSON(http.MethodDelete, fmt.Sprintf("/api/v1/auth-applications/%d", created.ID), nil, admin.AccessToken)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete auth application status=%d body=%s", deleteResp.Code, deleteResp.Body.String())
	}

	getDeletedResp := s.doJSON(http.MethodGet, fmt.Sprintf("/api/v1/auth-applications/%d", created.ID), nil, admin.AccessToken)
	if getDeletedResp.Code != http.StatusNotFound {
		t.Fatalf("expected deleted auth application get to be 404, got %d", getDeletedResp.Code)
	}
}
