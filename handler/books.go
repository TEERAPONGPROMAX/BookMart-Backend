package handler

import (
	"errors"
	"net/http"
	"strings"

	"bookmart/backend/store"
)

func (h *Handler) ListBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.Store.ListBooks(r.Context(), r.URL.Query().Get("q"), r.URL.Query().Get("category"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "โหลดรายการหนังสือไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, books)
}
func (h *Handler) GetBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "รหัสหนังสือไม่ถูกต้อง")
		return
	}
	book, err := h.Store.GetBook(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบหนังสือ")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "โหลดหนังสือไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, book)
}

func decodeBook(w http.ResponseWriter, r *http.Request, id int64) (store.Book, error) {
	var b store.Book
	if err := decodeJSON(w, r, &b); err != nil {
		return b, err
	}
	b.ID = id
	b.Title = strings.TrimSpace(b.Title)
	b.Author = strings.TrimSpace(b.Author)
	b.Category = strings.TrimSpace(b.Category)
	b.CoverURL = strings.TrimSpace(b.CoverURL)
	b.Description = strings.TrimSpace(b.Description)
	if b.Title == "" || b.Author == "" || b.Category == "" || b.Price < 0 || b.Stock < 0 {
		return b, errors.New("ข้อมูลหนังสือไม่ถูกต้อง")
	}
	return b, nil
}
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
	b, err := decodeBook(w, r, 0)
	if err != nil {
		writeError(w, http.StatusBadRequest, "กรุณาตรวจสอบข้อมูลหนังสือ")
		return
	}
	created, err := h.Store.CreateBook(r.Context(), b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "เพิ่มหนังสือไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
func (h *Handler) UpdateBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "รหัสหนังสือไม่ถูกต้อง")
		return
	}
	b, err := decodeBook(w, r, id)
	if err != nil {
		writeError(w, http.StatusBadRequest, "กรุณาตรวจสอบข้อมูลหนังสือ")
		return
	}
	updated, err := h.Store.UpdateBook(r.Context(), b)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบหนังสือ")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "แก้ไขหนังสือไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
func (h *Handler) DeleteBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "รหัสหนังสือไม่ถูกต้อง")
		return
	}
	err = h.Store.DeleteBook(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบหนังสือ")
		return
	}
	if err != nil {
		writeError(w, http.StatusConflict, "ลบหนังสือไม่ได้ อาจมีหนังสือเล่มนี้อยู่ในคำสั่งซื้อแล้ว")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
