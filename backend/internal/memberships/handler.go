package memberships

import (
	"backend/internal/httpx"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

type membershipJSON struct {
	ID       uuid.UUID  `json:"id"`
	RoleID   uuid.UUID  `json:"role_id"`
	UserID   uuid.UUID  `json:"user_id"`
	Term     *string    `json:"term"`
	JoinedAt time.Time  `json:"joined_at"`
	EndedAt  *time.Time `json:"ended_at"`
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /memberships", h.createMembership)
	mux.HandleFunc("POST /orgs/{org_id}/memberships", h.assignMemberships)
	mux.HandleFunc("POST /orgs/{org_id}/roles/{role_id}/handover",
		h.handoverRole)
	mux.HandleFunc("POST /orgs/{org_id}/memberships/{membership_id}/archive",
		h.archiveMembership)
}

// caps how many rows one request can insert
const maxAssignBatch = 500

func (h *Handler) createMembership(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrgID  uuid.UUID `json:"org_id"`
		RoleID uuid.UUID `json:"role_id"`
		UserID uuid.UUID `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.OrgID == uuid.Nil || body.RoleID == uuid.Nil || body.UserID == uuid.Nil {
		httpx.WriteError(w, http.StatusBadRequest,
			"org_id, role_id and user_id required")
		return
	}

	id, err := h.svc.CreateMembership(r.Context(), NewMembership{
		OrgID:  body.OrgID,
		RoleID: body.RoleID,
		UserID: body.UserID,
	})
	if errors.Is(err, ErrRoleNotInOrg) {
		httpx.WriteError(w, http.StatusNotFound, "role not found in org")
		return
	}
	if errors.Is(err, ErrAlreadyMember) {
		httpx.WriteError(w, http.StatusConflict, "user already holds this role")
		return
	}
	if err != nil {
		log.Printf("create membership: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not create membership")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated,
		map[string]uuid.UUID{"membership_id": id})
}

func (h *Handler) assignMemberships(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead the role's team or one of its ancestors
	orgID, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	var body struct {
		Memberships []struct {
			RoleID uuid.UUID `json:"role_id"`
			UserID uuid.UUID `json:"user_id"`
		} `json:"memberships"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if len(body.Memberships) == 0 || len(body.Memberships) > maxAssignBatch {
		httpx.WriteError(w, http.StatusBadRequest,
			fmt.Sprintf("memberships must have 1 to %d items", maxAssignBatch))
		return
	}

	pairs := make([]MembershipPair, len(body.Memberships))
	for i, m := range body.Memberships {
		if m.RoleID == uuid.Nil || m.UserID == uuid.Nil {
			httpx.WriteError(w, http.StatusBadRequest,
				fmt.Sprintf("memberships[%d]: role_id and user_id required", i))
			return
		}
		pairs[i] = MembershipPair{RoleID: m.RoleID, UserID: m.UserID}
	}

	res, err := h.svc.AssignMemberships(r.Context(), orgID, pairs)
	if errors.Is(err, ErrOrgNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "org not found")
		return
	}
	if err != nil {
		log.Printf("assign memberships: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not assign memberships")
		return
	}

	httpx.WriteJSON(w, http.StatusOK,
		map[string]int{"created": res.Created, "skipped": res.Skipped})
}

func (h *Handler) handoverRole(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead the role's team or one of its ancestors
	orgID, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	roleID, ok := httpx.PathUUID(r, "role_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var body struct {
		NewUserID uuid.UUID `json:"new_user_id"`
		// optional, leave out to end every current holder of the role
		PreviousUserID *uuid.UUID `json:"previous_user_id"`
		Term           string     `json:"term"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.NewUserID == uuid.Nil || body.Term == "" ||
		(body.PreviousUserID != nil && *body.PreviousUserID == uuid.Nil) {
		httpx.WriteError(w, http.StatusBadRequest,
			"new_user_id and term required, previous_user_id must be valid if set")
		return
	}

	res, err := h.svc.HandoverRole(r.Context(), Handover{
		OrgID:          orgID,
		RoleID:         roleID,
		NewUserID:      body.NewUserID,
		PreviousUserID: body.PreviousUserID,
		Term:           body.Term,
	})
	switch {
	case errors.Is(err, ErrRoleNotInOrg):
		httpx.WriteError(w, http.StatusNotFound, "role not found in org")
	case errors.Is(err, ErrPreviousHolderNotFound):
		httpx.WriteError(w, http.StatusNotFound,
			"previous user does not hold this role")
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, http.StatusConflict, "new user already holds this role")
	case err != nil:
		log.Printf("handover role: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not hand over role")
	default:
		httpx.WriteJSON(w, http.StatusOK, map[string]any{
			"membership_id": res.MembershipID,
			"ended":         res.Ended,
		})
	}
}

func (h *Handler) archiveMembership(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must lead the role's team or one of its ancestors
	orgID, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}
	membershipID, ok := httpx.PathUUID(r, "membership_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid membership id")
		return
	}

	m, err := h.svc.ArchiveMembership(r.Context(), orgID, membershipID)
	if errors.Is(err, ErrMembershipNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "membership not found")
		return
	}
	if err != nil {
		log.Printf("archive membership: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not archive membership")
		return
	}

	// same fields as Membership, so a conversion just adds the json tags
	httpx.WriteJSON(w, http.StatusOK, membershipJSON(m))
}
