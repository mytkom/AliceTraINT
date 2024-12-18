package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"github.com/stretchr/testify/assert"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
)

func TestUserHandler_Index(t *testing.T) {
	env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := handler.NewUserHandler(env)

	user := &models.User{
		CernPersonId: "12345",
		Username:     "johndoe",
		FirstName:    "John",
		FamilyName:   "Doe",
		Email:        "johndoe@example.com",
	}
	assert.NoError(t, env.User.Create(user))

	req, err := http.NewRequest("GET", "/", strings.NewReader(""))
	assert.NoError(t, err)
	rr := addSessionCookie(t, env, req, user.ID)

	handler.Index(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	assert.Contains(t, rr.Body.String(), "John (johndoe@example.com)")
}

func TestUserHandler_CreateUser(t *testing.T) {
	env, cleanup := setupIntegrationTest(t)
	defer cleanup()

	handler := handler.NewUserHandler(env)

	form := strings.NewReader("cern-person-id=67890&username=janedoe&first-name=Jane&family-name=Doe&email=janedoe@example.com")
	req, err := http.NewRequest("POST", "/users", form)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.CreateUser(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	users, err := handler.User.GetAll()
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Jane Doe", users[0].FirstName+" "+users[0].FamilyName)
	assert.Equal(t, "janedoe@example.com", users[0].Email)
	assert.Equal(t, "janedoe", users[0].Username)
}
