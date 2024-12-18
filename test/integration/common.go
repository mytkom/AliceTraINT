package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/migrate"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
}

func setupIntegrationTest(t *testing.T) (*environment.Env, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	migrate.MigrateDB(db)
	cfg := config.LoadConfig()
	tmpl := utils.BaseTemplate()
	repoContext := repository.NewRepositoryContext(db)
	auth := auth.MockAuth()

	env := environment.NewEnv(repoContext, auth, tmpl, cfg)

	cleanup := func() {
		dbSQL, err := db.DB()
		if err == nil {
			dbSQL.Close()
		}
	}

	return env, cleanup
}

func addSessionCookie(t *testing.T, env *environment.Env, req *http.Request, userId uint) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	sess := env.GlobalSessions.SessionStart(rr, req)
	assert.NoError(t, sess.Set("loggedUserId", userId))

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: sess.SessionID(),
	}
	req.AddCookie(cookie)

	return rr
}
