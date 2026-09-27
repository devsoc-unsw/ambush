package memberships

import (
	"backend/internal/httpx"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /memberships", h.createMembership)
}

func (h *Handler) createMembership(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SocietyID int32     `json:"society_id"`
		RoleID    int32     `json:"role_id"`
		UserID    uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.SocietyID <= 0 || body.RoleID <= 0 || body.UserID == uuid.Nil {
		httpx.WriteError(w, http.StatusBadRequest,
			"society_id, role_id and user_id required")
		return
	}

	id, err := h.svc.CreateMembership(r.Context(), NewMembership{
		SocietyID: body.SocietyID,
		RoleID:    body.RoleID,
		UserID:    body.UserID,
	})
	if errors.Is(err, ErrRoleNotInSociety) {
		httpx.WriteError(w, http.StatusNotFound, "role not found in society")
		return
	}
	if err != nil {
		log.Printf("create membership: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not create membership")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated,
		map[string]int32{"membership_id": id})
}
