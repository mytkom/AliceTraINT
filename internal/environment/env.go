package environment

import (
	"html/template"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
)

type Env struct {
	*repository.RepositoryContext
	auth.IAuthService
	*template.Template
	*config.Config
}

func NewEnv(repoContext *repository.RepositoryContext, auth auth.IAuthService, baseTemp *template.Template, cfg *config.Config) *Env {
	return &Env{
		RepositoryContext: repoContext,
		IAuthService:      auth,
		Template:          baseTemp,
		Config:            cfg,
	}
}
