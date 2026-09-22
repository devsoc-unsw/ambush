package societies

import (
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
		writeError(w, http.StatusBadRequest, "malformed json body")
		return
	}
	if body.Name == "" || body.Slug == "" {
		writeError(w, http.StatusBadRequest, "name and slug required")
		return
	}

	id, err := h.svc.CreateSociety(r.Context(), NewSociety{
		Name: body.Name,
		Slug: body.Slug,
	})
	if err != nil {
		log.Printf("create society: %v", err)
		writeError(w, http.StatusInternalServerError, "could not create society")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]int32{"society_id": id})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("json write: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
