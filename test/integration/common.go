package integration_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/config"
	"github.com/mytkom/AliceTraINT/internal/db/migrate"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/environment"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/mytkom/AliceTraINT/internal/service"
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

type MockedServices struct {
	CCDB        *service.MockCCDBService
	JAliEn      *service.MockJAliEnService
	FileService *service.LocalFileService
}

type IntegrationTestUtils struct {
	*environment.Env
	Router *http.ServeMux
	*MockedServices
}

func mockRouter(db *gorm.DB, cfg *config.Config) *IntegrationTestUtils {
	mux := http.NewServeMux()

	baseTemplate := utils.BaseTemplate()
	repoContext := repository.NewRepositoryContext(db)
	auth := auth.MockAuth(repoContext.User)

	env := environment.NewEnv(repoContext, auth, baseTemplate, cfg)

	// services
	hasher := service.NewArgon2Hasher()
	ccdbService := service.NewMockCCDBService()
	jalienService := service.NewMockJAliEnService()
	nnArch := service.NewNNArchService(cfg.NNArchPath)
	fileService := service.NewLocalFileService(cfg.DataDirPath)

	// handlers' routes
	handler.InitLandingRoutes(mux, env)
	handler.InitTrainingDatasetRoutes(mux, env, jalienService)
	handler.InitTrainingTaskRoutes(mux, env, ccdbService, fileService, nnArch)
	handler.InitTrainingMachineRoutes(mux, env, hasher)
	handler.InitQueueRoutes(mux, env, fileService, hasher)

	return &IntegrationTestUtils{
		Env:    env,
		Router: mux,
		MockedServices: &MockedServices{
			CCDB:        ccdbService,
			JAliEn:      jalienService,
			FileService: fileService,
		},
	}
}

func setupIntegrationTest(t *testing.T) (*IntegrationTestUtils, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	migrate.MigrateDB(db)
	cfg := config.LoadConfig()
	testUtils := mockRouter(db, cfg)

	cleanup := func() {
		dbSQL, err := db.DB()
		if err == nil {
			dbSQL.Close()
		}
	}

	return testUtils, cleanup
}

func addSessionCookie(t *testing.T, env *environment.Env, req *http.Request, userId uint) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	sess := env.GlobalSessions.SessionStart(rr, req)
	assert.NoError(t, env.LogUser(sess, userId))

	cookie := &http.Cookie{
		Name:  "gosessionid",
		Value: sess.SessionID(),
	}
	req.AddCookie(cookie)

	return rr
}

func HTMXReq(r *http.Request) {
	r.Header.Set("HX-Request", "true")
}

func testUnauthorized(t *testing.T, method, url string, body []byte) {
	t.Helper()

	ut, cleanup := setupIntegrationTest(t)
	defer cleanup()

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	assert.NoError(t, err)

	rr := httptest.NewRecorder() // No session cookie added
	ut.Router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
}
