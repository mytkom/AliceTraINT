package handler

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/jalien"
	"github.com/mytkom/AliceTraINT/internal/middleware"
	"github.com/mytkom/AliceTraINT/internal/utils"
)

type TrainJobHandler struct {
	NewTemplate              *template.Template
	ExploreDirectoryTemplate *template.Template
	FindAODsTemplate         *template.Template
}

func (h *TrainJobHandler) New(w http.ResponseWriter, r *http.Request) {
	err := h.NewTemplate.Execute(w, map[string]interface{}{
		"Title": "Create New Train Job!",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type exploreDirectoryTemplateData struct {
	Path      string
	Subdirs   []jalien.Dir
	AODFiles  []jalien.AODFile
	ParentDir string
}

func (h *TrainJobHandler) ExploreDirectory(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	dirContents, err := jalien.ListAndParseDirectory(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	parentDir := "/"
	if path != "/" {
		parentDir = filepath.Dir(strings.TrimSuffix(path, "/"))
		if parentDir != "/" {
			parentDir += "/"
		}
	}

	data := exploreDirectoryTemplateData{
		Path:      path,
		AODFiles:  dirContents.AODFiles,
		Subdirs:   dirContents.Subdirs,
		ParentDir: parentDir,
	}

	err = h.ExploreDirectoryTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type findAODsTemplateData struct {
	AODFiles []jalien.AODFile
}

func (h *TrainJobHandler) FindAods(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	aods, err := jalien.FindAODFiles(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	data := &findAODsTemplateData{
		AODFiles: aods,
	}

	err = h.FindAODsTemplate.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func InitTrainJobRoutes(mux *http.ServeMux, baseTemplate *template.Template, auth *auth.Auth) {
	prefix := "train-jobs"
	base := template.Must(baseTemplate.Clone())

	tjh := &TrainJobHandler{
		NewTemplate:              template.Must(base.ParseFiles("web/templates/selector.html")),
		ExploreDirectoryTemplate: template.Must(template.ParseFiles("web/templates/tree_browser.html")),
		FindAODsTemplate:         template.Must(template.ParseFiles("web/templates/file_list.html")),
	}

	cache := utils.NewCache(60 * time.Minute)

	authMw := middleware.NewAuthMw(auth)
	cacheMw := middleware.NewCacheMw(cache)

	mux.Handle(fmt.Sprintf("GET /%s/new", prefix), middleware.Chain(
		http.HandlerFunc(tjh.New),
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/explore-directory", prefix), middleware.Chain(
		http.HandlerFunc(tjh.ExploreDirectory),
		cacheMw,
		authMw,
	))

	mux.Handle(fmt.Sprintf("GET /%s/find-aods", prefix), middleware.Chain(
		http.HandlerFunc(tjh.FindAods),
		cacheMw,
		authMw,
	))
}
