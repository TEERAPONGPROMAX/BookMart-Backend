package mw

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"bookmart/backend/auth"
)

type contextKey struct{}
type Who struct {
	ID    int64
	Role  string
	Email string
}

func WhoFrom(ctx context.Context) Who { who, _ := ctx.Value(contextKey{}).(Who); return who }

func RequireLogin(issuer *auth.Issuer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			deny(w, http.StatusUnauthorized, "กรุณาเข้าสู่ระบบก่อน")
			return
		}
		id, role, email, err := issuer.Parse(token)
		if err != nil {
			deny(w, http.StatusUnauthorized, "ตั๋วใช้ไม่ได้หรือหมดอายุ กรุณาเข้าสู่ระบบใหม่")
			return
		}
		ctx := context.WithValue(r.Context(), contextKey{}, Who{ID: id, Role: role, Email: email})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if WhoFrom(r.Context()).Role != role {
			deny(w, http.StatusForbidden, "บัญชีของคุณไม่มีสิทธิ์ทำรายการนี้")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func deny(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
