package handler

import (
	"errors"
	"net/http"
	"strings"

	"bookmart/backend/auth"
	"bookmart/backend/mw"
	"bookmart/backend/store"
	"github.com/jackc/pgx/v5/pgconn"
)

var dummyHash, _ = auth.HashPassword("ค่ารหัสผ่านหลอกสำหรับป้องกันการเดาอีเมล")

type authRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

func publicMember(m store.Member) map[string]any {
	return map[string]any{"id": m.ID, "email": m.Email, "name": m.Name, "displayName": m.Name, "role": m.Role}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input authRequest
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "ข้อมูลสมัครสมาชิกไม่ถูกต้อง")
		return
	}
	if input.Name == "" {
		input.Name = input.DisplayName
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if len(input.Name) < 2 || len(input.Name) > 80 || !strings.Contains(input.Email, "@") || len(input.Password) < 8 || len(input.Password) > 72 {
		writeError(w, http.StatusBadRequest, "กรุณากรอกชื่อ อีเมล และรหัสผ่าน 8–72 ไบต์")
		return
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "สมัครสมาชิกไม่สำเร็จ")
		return
	}
	member, err := h.Store.CreateMember(r.Context(), input.Email, input.Name, hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "อีเมลนี้มีบัญชีแล้ว")
			return
		}
		writeError(w, http.StatusInternalServerError, "สมัครสมาชิกไม่สำเร็จ")
		return
	}
	issueSession(w, h, member, http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input authRequest
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "ข้อมูลเข้าสู่ระบบไม่ถูกต้อง")
		return
	}
	member, err := h.Store.MemberByEmail(r.Context(), input.Email)
	if errors.Is(err, store.ErrNotFound) {
		_ = auth.CheckPassword(dummyHash, input.Password)
		writeError(w, http.StatusUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "เข้าสู่ระบบไม่สำเร็จ")
		return
	}
	if !auth.CheckPassword(member.PasswordHash, input.Password) {
		writeError(w, http.StatusUnauthorized, "อีเมลหรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	issueSession(w, h, member, http.StatusOK)
}

func issueSession(w http.ResponseWriter, h *Handler, member store.Member, status int) {
	token, err := h.Tickets.Issue(member.ID, member.Email, member.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ออกตั๋วเข้าสู่ระบบไม่สำเร็จ")
		return
	}
	writeJSON(w, status, map[string]any{"token": token, "user": publicMember(member)})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	member, err := h.Store.MemberByID(r.Context(), who.ID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "ไม่พบบัญชีสมาชิก")
		return
	}
	writeJSON(w, http.StatusOK, publicMember(member))
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "ข้อมูลโปรไฟล์ไม่ถูกต้อง")
		return
	}
	if input.Name == "" {
		input.Name = input.DisplayName
	}
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Name) < 2 || len(input.Name) > 80 {
		writeError(w, http.StatusBadRequest, "ชื่อต้องมี 2–80 ตัวอักษร")
		return
	}
	who := mw.WhoFrom(r.Context())
	member, err := h.Store.UpdateMemberName(r.Context(), who.ID, input.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "บันทึกโปรไฟล์ไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, publicMember(member))
}

func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Password string `json:"password"`
	}
	if decodeJSON(w, r, &input) != nil || len(input.Password) < 8 || len(input.Password) > 72 {
		writeError(w, http.StatusBadRequest, "รหัสผ่านต้องมี 8–72 ไบต์")
		return
	}
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "เปลี่ยนรหัสผ่านไม่สำเร็จ")
		return
	}
	who := mw.WhoFrom(r.Context())
	if err = h.Store.UpdatePassword(r.Context(), who.ID, hash); err != nil {
		writeError(w, http.StatusInternalServerError, "เปลี่ยนรหัสผ่านไม่สำเร็จ")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
