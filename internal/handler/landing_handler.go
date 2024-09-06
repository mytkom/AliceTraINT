package handler

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
)

type LandingHandler struct {
	LandingTemplate *template.Template
	Auth            *auth.Auth
}

func (h *LandingHandler) Index(w http.ResponseWriter, r *http.Request) {
	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		http.Redirect(w, r, "/train-jobs/new", http.StatusTemporaryRedirect)
	}

	err := h.LandingTemplate.Execute(w, map[string]interface{}{
		"Title": "AliceTraINT",
	})
	if err != nil {
		http.Error(w, "Cannot render template", http.StatusInternalServerError)
	}
}

func InitLandingRoutes(mux *http.ServeMux, baseTemplate *template.Template, auth *auth.Auth) {
	base := template.Must(baseTemplate.Clone())

	lh := &LandingHandler{
		LandingTemplate: template.Must(base.ParseFiles("web/templates/landing.html")),
		Auth:            auth,
	}

	mux.HandleFunc("GET /", lh.Index)
}
