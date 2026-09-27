package societies

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
	mux.HandleFunc("POST /societies", h.createSociety)
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
	if err != nil {
		log.Printf("create society: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "could not create society")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]int32{"society_id": id})
}
