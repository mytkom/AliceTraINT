package handler

import (
	"errors"
	"net/http"
	"os"

	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/service"
)

type DocsHandler struct {
	*environment.Env
	Service service.IDocsService
}

func NewDocsHandler(env *environment.Env, svc service.IDocsService) *DocsHandler {
	return &DocsHandler{
		Env:     env,
		Service: svc,
	}
}

func (h *DocsHandler) Index(w http.ResponseWriter, r *http.Request) {
	docs, err := h.Service.ListDocs()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}

	if len(docs) == 0 {
		// No docs yet – fall back to the simple index page.
		type TemplateData struct {
			Title string
			Docs  []service.DocMeta
		}

		err = h.ExecuteTemplate(w, "docs_index", TemplateData{
			Title: "Documentation",
			Docs:  docs,
		})
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		}
		return
	}

	// Redirect to the first doc in alphabetical order.
	http.Redirect(w, r, "/docs/"+docs[0].Slug, http.StatusTemporaryRedirect)
}

func (h *DocsHandler) Show(w http.ResponseWriter, r *http.Request) {
	type TemplateData struct {
		Title string
		Doc   *service.Doc
		Docs  []service.DocMeta
	}

	slug := r.PathValue("slug")
	if slug == "" {
		http.Redirect(w, r, "/docs", http.StatusTemporaryRedirect)
		return
	}

	doc, err := h.Service.GetDoc(slug)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = h.ExecuteTemplate(w, "not-found", nil)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
				return
			}
			return
		} else {
			writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		}
		return
	}

	docs, err := h.Service.ListDocs()
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}

	err = h.ExecuteTemplate(w, "docs_show", TemplateData{
		Title: doc.Title,
		Doc:   doc,
		Docs:  docs,
	})
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "unexpected internal server error", err)
		return
	}
}

func InitDocsRoutes(mux *http.ServeMux, env *environment.Env) {
	dh := &DocsHandler{
		Env:     env,
		Service: service.NewDocsService(env.Config.DocsDirPath),
	}

	blockHtmxMw := middleware.NewBlockHTMXMw()

	mux.Handle("GET /docs", middleware.Chain(
		http.HandlerFunc(dh.Index),
		blockHtmxMw,
	))

	mux.Handle("GET /docs/{slug}", middleware.Chain(
		http.HandlerFunc(dh.Show),
		blockHtmxMw,
	))
}
