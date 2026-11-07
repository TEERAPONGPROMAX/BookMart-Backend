package handler

import (
	"errors"
	"net/http"
	"strings"

	"bookmart/backend/mw"
	"bookmart/backend/store"
)

func (h *Handler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "รหัสสมาชิกไม่ถูกต้อง")
		return
	}
	var input struct {
		Role string `json:"role"`
	}
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "ข้อมูลบทบาทไม่ถูกต้อง")
		return
	}
	input.Role = strings.TrimSpace(input.Role)
	if input.Role != "member" && input.Role != "admin" {
		writeError(w, http.StatusBadRequest, "บทบาทต้องเป็น member หรือ admin")
		return
	}
	who := mw.WhoFrom(r.Context())
	if id == who.ID && input.Role != "admin" {
		writeError(w, http.StatusConflict, "ไม่สามารถลดสิทธิ์บัญชีที่กำลังใช้งานอยู่ได้")
		return
	}
	member, err := h.Store.UpdateMemberRole(r.Context(), id, input.Role)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบบัญชีสมาชิก")
		return
	}
	if errors.Is(err, store.ErrLastAdmin) {
		writeError(w, http.StatusConflict, "ต้องมีผู้ดูแลระบบอย่างน้อยหนึ่งบัญชี")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "เปลี่ยนสิทธิ์สมาชิกไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, publicMember(member))
}
