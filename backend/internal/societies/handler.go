package societies

import (
	"backend/internal/httpx"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /societies", h.createSociety)
	mux.HandleFunc("PATCH /societies/{society_id}", h.updateSociety)
}

func (h *Handler) createSociety(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == "" || body.Slug == "" {
		httpx.WriteError(w, http.StatusBadRequest, "name and slug required")
		return
	}

	id, err := h.svc.CreateSociety(r.Context(), NewSociety{
		Name: body.Name,
		Slug: body.Slug,
	})
	if errors.Is(err, ErrSlugTaken) {
		httpx.WriteError(w, http.StatusConflict, "slug already taken")
		return
	}
	if err != nil {
		log.Printf("create society: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create society")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]int32{"society_id": id})
}

func (h *Handler) updateSociety(w http.ResponseWriter, r *http.Request) {
	// TODO: authz, caller must be a SocietyAdmin of society_id
	id, ok := httpx.PathID(r, "society_id")
	if !ok {
		httpx.WriteError(w, http.StatusBadRequest, "invalid society id")
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

	soc, err := h.svc.UpdateSociety(r.Context(), id,
		SocietyUpdate{Name: body.Name, Slug: body.Slug})
	switch {
	case errors.Is(err, ErrSocietyNotFound):
		httpx.WriteError(w, http.StatusNotFound, "society not found")
	case errors.Is(err, ErrSlugTaken):
		httpx.WriteError(w, http.StatusConflict, "slug already taken")
	case err != nil:
		log.Printf("update society: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError,
			"could not update society")
	default:
		// same fields as Society, so a conversion just adds the json tags
		httpx.WriteJSON(w, http.StatusOK, societyJSON(soc))
	}
}

type societyJSON struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
