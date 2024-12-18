package environment

import (
	"html/template"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
)

type Env struct {
	*repository.RepositoryContext
	*auth.Auth
	*template.Template
	*config.Config
}

func NewEnv(repoContext *repository.RepositoryContext, auth *auth.Auth, baseTemp *template.Template, cfg *config.Config) *Env {
	return &Env{
		RepositoryContext: repoContext,
		Auth:              auth,
		Template:          baseTemp,
		Config:            cfg,
	}
}
