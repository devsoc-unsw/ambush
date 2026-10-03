package memberships

import (
	"backend/internal/httpx"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	mux.HandleFunc("POST /societies/{society_id}/memberships",
		h.assignMemberships)
}

// caps how many rows one request can insert
const maxAssignBatch = 500

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

func (h *Handler) assignMemberships(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must be a SocietyAdmin of society_id
	societyID, err := strconv.ParseInt(r.PathValue("society_id"), 10, 32)
	if err != nil || societyID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid society id")
		return
	}

	var body struct {
		Memberships []struct {
			RoleID int32     `json:"role_id"`
			UserID uuid.UUID `json:"user_id"`
		} `json:"memberships"`
	}
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
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
		if m.RoleID <= 0 || m.UserID == uuid.Nil {
			httpx.WriteError(w, http.StatusBadRequest,
				fmt.Sprintf("memberships[%d]: role_id and user_id required", i))
			return
		}
		pairs[i] = MembershipPair{RoleID: m.RoleID, UserID: m.UserID}
	}

	res, err := h.svc.AssignMemberships(r.Context(), int32(societyID), pairs)
	if errors.Is(err, ErrSocietyNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "society not found")
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
