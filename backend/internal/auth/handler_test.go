package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type loginStub struct {
	request AuthLoginRequest
	calls   int
	err     error
}

func (s *loginStub) Login(_ context.Context, req AuthLoginRequest) (*AuthResponse, error) {
	s.request = req
	s.calls++
	return &AuthResponse{ID: "application-id", Email: "viewer@example.com"}, s.err
}

func TestLoginRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name, path, body string
		serviceErr       error
		status, calls    int
	}{
		{name: "token alone is sufficient", path: "/api/auth/login", body: `{"id_token":"firebase-token"}`, status: 200, calls: 1},
		{name: "token required", path: "/api/auth/login", body: `{"email":"viewer@example.com"}`, status: 400},
		{name: "invalid token", path: "/api/auth/login", body: `{"id_token":"firebase-token"}`, serviceErr: ErrInvalidCredentials, status: 401, calls: 1},
		{name: "Firebase unavailable", path: "/api/auth/login", body: `{"id_token":"firebase-token"}`, serviceErr: ErrNetworkError, status: 503, calls: 1},
		{name: "database failure", path: "/api/auth/login", body: `{"id_token":"firebase-token"}`, serviceErr: fmt.Errorf("%w: private details", ErrDatabase), status: 500, calls: 1},
		{name: "public signup removed", path: "/api/auth/signup", body: `{"uid":"unverified","email":"viewer@example.com"}`, status: 404},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &loginStub{err: tt.serviceErr}
			router := gin.New()
			NewHandler(service).RegisterRoutes(router.Group("/api"))
			request := httptest.NewRequest("POST", tt.path, strings.NewReader(tt.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != tt.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.status, response.Body.String())
			}
			if service.calls != tt.calls {
				t.Fatalf("service called %d times, want %d", service.calls, tt.calls)
			}
			if tt.calls > 0 && service.request.IDToken != "firebase-token" {
				t.Fatalf("token not forwarded: %+v", service.request)
			}
			if tt.status == 200 {
				var profile AuthResponse
				if err := json.Unmarshal(response.Body.Bytes(), &profile); err != nil {
					t.Fatal(err)
				}
				if profile.ID != "application-id" || profile.Email != "viewer@example.com" {
					t.Fatalf("unexpected profile: %+v", profile)
				}
			}
			if tt.status == 500 && strings.Contains(response.Body.String(), "private details") {
				t.Fatal("database details exposed")
			}
		})
	}
}
