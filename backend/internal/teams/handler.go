package teams

import (
	"backend/internal/httpx"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /orgs/{org_id}/teams", h.createTeam)
	mux.HandleFunc("PUT /orgs/{org_id}/teams/{team_id}/leader", h.setLeader)
}

func (h *Handler) createTeam(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead parent_id or one of its ancestors
	orgID, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var body struct {
		Name     string    `json:"name"`
		ParentID uuid.UUID `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == "" || body.ParentID == uuid.Nil {
		httpx.WriteError(w, http.StatusBadRequest, "name and parent_id required")
		return
	}

	t, err := h.svc.CreateTeam(r.Context(), NewTeam{
		OrgID:    orgID,
		ParentID: body.ParentID,
		Name:     body.Name,
	})
	switch {
	case errors.Is(err, ErrParentNotFound):
		httpx.WriteError(w, http.StatusNotFound, "parent team not found in org")
	case errors.Is(err, ErrNameTaken):
		httpx.WriteError(w, http.StatusConflict,
			"team name already taken in this org")
	case err != nil:
		log.Printf("create team: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create team")
	default:
		httpx.WriteJSON(w, http.StatusCreated, teamJSON(t))
	}
}

func (h *Handler) setLeader(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead an ancestor of team_id
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

	// a map so a missing role_id (mistake) differs from null (clear the leader)
	var body map[string]*uuid.UUID
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	roleID, present := body["role_id"]
	if !present || (roleID != nil && *roleID == uuid.Nil) {
		httpx.WriteError(w, http.StatusBadRequest,
			"role_id required, null to clear the leader")
		return
	}

	t, err := h.svc.SetLeader(r.Context(), orgID, teamID, roleID)
	switch {
	case errors.Is(err, ErrTeamNotFound):
		httpx.WriteError(w, http.StatusNotFound, "team not found in org")
	case errors.Is(err, ErrRoleNotFound):
		httpx.WriteError(w, http.StatusNotFound, "role not found in org")
	case err != nil:
		log.Printf("set team leader: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not set team leader")
	default:
		httpx.WriteJSON(w, http.StatusOK, teamJSON(t))
	}
}

type teamJSON struct {
	ID           uuid.UUID  `json:"id"`
	OrgID        uuid.UUID  `json:"org_id"`
	ParentID     *uuid.UUID `json:"parent_id"`
	LeaderRoleID *uuid.UUID `json:"leader_role_id"`
	Name         string     `json:"name"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ArchivedAt   *time.Time `json:"archived_at"`
}
