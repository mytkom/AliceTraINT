package handler

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/middleware"
)

type LandingHandler struct {
	Template *template.Template
	Auth     *auth.Auth
}

func (h *LandingHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	if r.URL.Path != "/" {
		err := h.Template.ExecuteTemplate(w, "not-found", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	templateData := TemplateData{
		Title: "AliceTraINT",
	}

	sess := h.Auth.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get("loggedUserId")
	if loggedUserId != nil {
		http.Redirect(w, r, "/training-datasets", http.StatusTemporaryRedirect)
		return
	}

	err := h.Template.ExecuteTemplate(w, "landing", templateData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InitLandingRoutes(mux *http.ServeMux, baseTemplate *template.Template, auth *auth.Auth) {
	lh := &LandingHandler{
		Auth:     auth,
		Template: baseTemplate,
	}

	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle("GET /", middleware.Chain(
		http.HandlerFunc(lh.Index),
		blockHtmxMw,
	))
}
