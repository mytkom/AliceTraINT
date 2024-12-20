package handler

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/middleware"
)

type LandingHandler struct {
	*environment.Env
}

func (h *LandingHandler) Index(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
	}

	if r.URL.Path != "/" {
		err := h.ExecuteTemplate(w, "not-found", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_, err := h.Auth.GetAuthorizedUser(w, r)
	if err == nil {
		http.Redirect(w, r, "/training-datasets", http.StatusTemporaryRedirect)
		return
	}

	err = h.ExecuteTemplate(w, "landing", TemplateData{
		Title: "AliceTraINT",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func InitLandingRoutes(mux *http.ServeMux, env *environment.Env) {
	lh := &LandingHandler{
		Env: env,
	}

	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle("GET /", middleware.Chain(
		http.HandlerFunc(lh.Index),
		blockHtmxMw,
	))
}
