package mw

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSAllowsFrontendPreflightAndRejectsUnknownOrigin(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusOK) })
	handler := CORS([]string{"http://localhost:3000"}, next)

	allowed := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("Access-Control-Request-Method", "POST")
	request.Header.Set("Access-Control-Request-Headers", "authorization,content-type")
	handler.ServeHTTP(allowed, request)
	if allowed.Code != http.StatusNoContent {
		t.Fatalf("allowed preflight status = %d", allowed.Code)
	}
	if allowed.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("allowed origin header missing")
	}
	if allowed.Header().Get("Access-Control-Allow-Headers") != "Authorization, Content-Type" {
		t.Fatal("authorization header not allowed")
	}
	if called {
		t.Fatal("preflight should return before calling the route")
	}

	blocked := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	request.Header.Set("Origin", "https://not-bookmart.example")
	handler.ServeHTTP(blocked, request)
	if blocked.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unknown origin was allowed")
	}
}
