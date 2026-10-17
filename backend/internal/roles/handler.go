package roles

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
	mux.HandleFunc("POST /orgs/{org_id}/teams/{team_id}/roles", h.createRole)
}

func (h *Handler) createRole(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead team_id or one of its ancestors
	orgID, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	teamID, ok := httpx.PathUUID(r, "team_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid team id")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name required")
		return
	}

	id, err := h.svc.CreateRole(r.Context(), NewRole{
		OrgID:  orgID,
		TeamID: teamID,
		Name:   body.Name,
	})
	switch {
	case errors.Is(err, ErrTeamNotFound):
		httpx.WriteError(w, http.StatusNotFound, "team not found in org")
	case errors.Is(err, ErrNameTaken):
		httpx.WriteError(w, http.StatusConflict,
			"role name already taken in this team")
	case err != nil:
		log.Printf("create role: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create role")
	default:
		httpx.WriteJSON(w, http.StatusCreated, map[string]uuid.UUID{"role_id": id})
	}
}
