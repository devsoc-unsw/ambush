package orgs

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
	mux.HandleFunc("POST /orgs", h.createOrg)
	mux.HandleFunc("PATCH /orgs/{org_id}", h.updateOrg)
	mux.HandleFunc("POST /orgs/{org_id}/archive", h.archiveOrg)
}

func (h *Handler) createOrg(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string `json:"name"`
		Slug         string `json:"slug"`
		RootTeamName string `json:"root_team_name"` // optional
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == "" || body.Slug == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name and slug required")
		return
	}

	created, err := h.svc.CreateOrg(r.Context(), NewOrg{
		Name:         body.Name,
		Slug:         body.Slug,
		RootTeamName: body.RootTeamName,
	})
	if errors.Is(err, ErrSlugTaken) {
		httpx.WriteError(w, http.StatusConflict, "slug already taken")
		return
	}
	if err != nil {
		log.Printf("create org: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create org")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]uuid.UUID{
		"org_id":       created.OrgID,
		"root_team_id": created.RootTeamID,
	})
}

func (h *Handler) updateOrg(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must be an admin of org_id
	id, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	// pointers so a missing field (keep as is) differs from "" (invalid)
	var body struct {
		Name *string `json:"name"`
		Slug *string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == nil && body.Slug == nil {
		httpx.WriteError(w, http.StatusBadRequest, "name or slug required")
		return
	}
	if (body.Name != nil && *body.Name == "") ||
		(body.Slug != nil && *body.Slug == "") {
		httpx.WriteError(w, http.StatusBadRequest, "name and slug can't be empty")
		return
	}

	org, err := h.svc.UpdateOrg(r.Context(), id,
		OrgUpdate{Name: body.Name, Slug: body.Slug})
	switch {
	case errors.Is(err, ErrOrgNotFound):
		httpx.WriteError(w, http.StatusNotFound, "org not found")
	case errors.Is(err, ErrSlugTaken):
		httpx.WriteError(w, http.StatusConflict, "slug already taken")
	case err != nil:
		log.Printf("update org: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not update org")
	default:
		// same fields as Org, so a conversion just adds the json tags
		httpx.WriteJSON(w, http.StatusOK, orgJSON(org))
	}
}

func (h *Handler) archiveOrg(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must be an admin of org_id
	id, ok := httpx.PathUUID(r, "org_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid org id")
		return
	}

	org, err := h.svc.ArchiveOrg(r.Context(), id)
	if errors.Is(err, ErrOrgNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "org not found")
		return
	}
	if err != nil {
		log.Printf("archive org: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not archive org")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, orgJSON(org))
}

type orgJSON struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at"`
}
