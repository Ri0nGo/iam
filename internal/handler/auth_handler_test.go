package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthLoginBindError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(nil)
	r := gin.New()
	r.POST("/login", h.Login)

	body, _ := json.Marshal(map[string]any{"username": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
