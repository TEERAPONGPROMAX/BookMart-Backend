package handler

import (
	"errors"
	"net/http"
	"strings"

	"bookmart/backend/mw"
	"bookmart/backend/store"
)

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Items []store.OrderRequestItem `json:"items"`
	}
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "รายการสั่งซื้อไม่ถูกต้อง")
		return
	}
	who := mw.WhoFrom(r.Context())
	order, err := h.Store.CreateOrder(r.Context(), who.ID, input.Items)
	if errors.Is(err, store.ErrOutOfStock) {
		writeError(w, http.StatusConflict, "มีหนังสือบางรายการสินค้าไม่เพียงพอ")
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบหนังสือที่เลือก")
		return
	}
	if errors.Is(err, store.ErrInvalidOrder) {
		writeError(w, http.StatusBadRequest, "ข้อมูลรายการสั่งซื้อไม่ถูกต้อง")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "บันทึกคำสั่งซื้อไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusCreated, order)
}

func (h *Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	orders, err := h.Store.ListOrders(r.Context(), &who.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "โหลดประวัติไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}
func (h *Handler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.Store.ListOrders(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "โหลดคำสั่งซื้อไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, orders)
}

func (h *Handler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "รหัสคำสั่งซื้อไม่ถูกต้อง")
		return
	}
	var input struct {
		Status string `json:"status"`
	}
	if decodeJSON(w, r, &input) != nil {
		writeError(w, http.StatusBadRequest, "สถานะไม่ถูกต้อง")
		return
	}
	input.Status = strings.TrimSpace(input.Status)
	valid := map[string]bool{"pending": true, "paid": true, "shipped": true, "cancelled": true}
	if !valid[input.Status] {
		writeError(w, http.StatusBadRequest, "สถานะไม่ถูกต้อง")
		return
	}
	order, err := h.Store.UpdateOrderStatus(r.Context(), id, input.Status)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ไม่พบคำสั่งซื้อ")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "แก้ไขคำสั่งซื้อไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, order)
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	members, err := h.Store.ListMembers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "โหลดสมาชิกไม่สำเร็จ")
		return
	}
	writeJSON(w, http.StatusOK, members)
}
