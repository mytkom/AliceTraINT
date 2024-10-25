package handler

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
)

type LandingHandler struct {
	Template *template.Template
	Auth     *auth.Auth
}

func (h *LandingHandler) Index(w http.ResponseWriter, r *http.Request) {
	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		http.Redirect(w, r, "/training-datasets", http.StatusTemporaryRedirect)
	}

	err := h.Template.ExecuteTemplate(w, "landing", map[string]interface{}{
		"Title": "AliceTraINT",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InitLandingRoutes(mux *http.ServeMux, baseTemplate *template.Template, auth *auth.Auth) {
	lh := &LandingHandler{
		Auth:     auth,
		Template: baseTemplate,
	}

	mux.HandleFunc("GET /", lh.Index)
}
