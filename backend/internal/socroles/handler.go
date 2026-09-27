package socroles

import (
	"backend/internal/httpx"
	"encoding/json"
	"log"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /socroles", h.createSocRole)
}

func (h *Handler) createSocRole(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SocietyID    int32  `json:"society_id"`
		ParentRoleID *int32 `json:"parent_role_id"`
		Name         string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.SocietyID <= 0 || body.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest,
			"society id, name required")
		return
	}
	if body.ParentRoleID != nil && *body.ParentRoleID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid parent role id")
		return
	}
	id, err := h.svc.CreateSocRole(r.Context(), NewSocRole{
		SocietyID:    body.SocietyID,
		ParentRoleID: body.ParentRoleID,
		Name:         body.Name,
	})
	if err != nil {
		log.Printf("create socrole: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not create socrole")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]int32{"socrole_id": id})
}
