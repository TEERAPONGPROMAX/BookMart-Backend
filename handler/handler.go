package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"bookmart/backend/auth"
	"bookmart/backend/mw"
	"bookmart/backend/store"
)

type Handler struct {
	Store   store.Repository
	Tickets *auth.Issuer
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "bookmart-api"})
	})
	mux.HandleFunc("GET /api/books", h.ListBooks)
	mux.HandleFunc("GET /api/books/{id}", h.GetBook)
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	login := func(next http.Handler) http.Handler { return mw.RequireLogin(h.Tickets, next) }
	admin := func(next http.Handler) http.Handler { return mw.RequireLogin(h.Tickets, mw.RequireRole("admin", next)) }
	mux.Handle("GET /api/auth/me", login(http.HandlerFunc(h.Me)))
	mux.Handle("PATCH /api/auth/profile", login(http.HandlerFunc(h.UpdateProfile)))
	mux.Handle("PATCH /api/auth/password", login(http.HandlerFunc(h.ChangePassword)))
	mux.Handle("GET /api/orders", login(http.HandlerFunc(h.ListMyOrders)))
	mux.Handle("POST /api/orders", login(http.HandlerFunc(h.CreateOrder)))
	mux.Handle("POST /api/admin/books", admin(http.HandlerFunc(h.CreateBook)))
	mux.Handle("PATCH /api/admin/books/{id}", admin(http.HandlerFunc(h.UpdateBook)))
	mux.Handle("DELETE /api/admin/books/{id}", admin(http.HandlerFunc(h.DeleteBook)))
	mux.Handle("GET /api/admin/orders", admin(http.HandlerFunc(h.ListAllOrders)))
	mux.Handle("PATCH /api/admin/orders/{id}", admin(http.HandlerFunc(h.UpdateOrderStatus)))
	mux.Handle("GET /api/admin/members", admin(http.HandlerFunc(h.ListMembers)))
	return mux
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(dst)
}
func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errInvalidID
	}
	return id, nil
}

type invalidIDError struct{}

func (invalidIDError) Error() string { return "invalid id" }

var errInvalidID error = invalidIDError{}
