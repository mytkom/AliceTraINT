package handler

import (
	"html/template"
	"net/http"
)

type LandingHandler struct {
	Templates      *template.Template
}

func NewLandingHandler(templates *template.Template) *LandingHandler {
	return &LandingHandler{Templates: templates}
}

func (h *LandingHandler) Index(w http.ResponseWriter, r *http.Request) {
  err := h.Templates.ExecuteTemplate(w, "landing.html", nil)
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
	}
}

