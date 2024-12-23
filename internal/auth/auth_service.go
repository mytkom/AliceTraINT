package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	oidc "github.com/coreos/go-oidc"
	"github.com/google/uuid"
	"github.com/mytkom/AliceTraINT/internal/db/models"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/thomasdarimont/go-kc-example/session"
	_ "github.com/thomasdarimont/go-kc-example/session_memory"
	"golang.org/x/oauth2"
)

const (
	loggedUserIdQuery string = "loggedUserId"
)

type UserInfo struct {
	CernPersonId      string `json:"cern_person_id"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Email             string `json:"email"`
}

type IAuthService interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	GetAuthorizedUser(w http.ResponseWriter, r *http.Request) (*models.User, error)
	LogUser(sess session.Session, userId uint) error
}

type AuthService struct {
	config         *oauth2.Config
	verifier       *oidc.IDTokenVerifier
	userRepo       repository.UserRepository
	state          string
	GlobalSessions *session.Manager
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	globalSessions, err := session.NewManager("memory", "gosessionid", 3600)
	if err != nil {
		log.Fatal(err)
	}
	go globalSessions.GC()

	configURL := os.Getenv("CERN_REALM_URL")
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, configURL)
	if err != nil {
		panic(err)
	}

	oidcConfig := &oidc.Config{
		ClientID: os.Getenv("CERN_CLIENT_ID"),
	}
	verifier := provider.Verifier(oidcConfig)

	return &AuthService{
		config: &oauth2.Config{
			ClientID:     os.Getenv("CERN_CLIENT_ID"),
			ClientSecret: os.Getenv("CERN_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("CERN_REDIRECT_URL"),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
			Endpoint:     provider.Endpoint(),
		},
		verifier:       verifier,
		userRepo:       userRepo,
		state:          "auth-state",
		GlobalSessions: globalSessions,
	}
}

func (a *AuthService) GetAuthorizedUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	sess := a.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get(loggedUserIdQuery)
	if loggedUserId != nil {
		loggedUser, err := a.userRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			return nil, err
		}
		return loggedUser, nil
	}

	return nil, fmt.Errorf("user not logged in")
}

func (a *AuthService) LogUser(sess session.Session, userId uint) error {
	return sess.Set(loggedUserIdQuery, userId)
}

func (a *AuthService) LoginHandler(w http.ResponseWriter, r *http.Request) {
	sess := a.GlobalSessions.SessionStart(w, r)
	oauthState := uuid.New().String()
	err := sess.Set(a.state, oauthState)
	if err != nil {
		http.Error(w, "Failed to set session", http.StatusInternalServerError)
	}

	//checking the userinfo in the session. If it is nil, then the user is not authenticated yet
	userInfo := sess.Get("userinfo")
	if userInfo == nil {
		http.Redirect(w, r, a.config.AuthCodeURL(oauthState), http.StatusFound)
		return
	}

	//just redirect the user to any other page
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *AuthService) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	sess := a.GlobalSessions.SessionStart(w, r)

	state := sess.Get(a.state)

	if state == nil {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}
	if r.URL.Query().Get("state") != state.(string) {
		http.Error(w, "state did not match", http.StatusBadRequest)
		return
	}
	ctx := context.Background()

	//exchanging the code for a token
	oauth2Token, err := a.config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
		return
	}
	idToken, err := a.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := struct {
		OAuth2Token   *oauth2.Token
		IDTokenClaims *json.RawMessage
	}{oauth2Token, new(json.RawMessage)}

	err = idToken.Claims(&resp.IDTokenClaims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenClaims := &UserInfo{}

	err = json.Unmarshal(*resp.IDTokenClaims, tokenClaims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var user *models.User
	user, err = a.userRepo.GetByCernPersonId(tokenClaims.CernPersonId)
	if err != nil {
		user = &models.User{
			CernPersonId: tokenClaims.CernPersonId,
			Username:     tokenClaims.PreferredUsername,
			FirstName:    tokenClaims.GivenName,
			FamilyName:   tokenClaims.FamilyName,
			Email:        tokenClaims.Email,
		}

		err = a.userRepo.Create(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	//storing the token and the info of the user in session memory
	err = sess.Set("rawIDToken", rawIDToken)
	if err != nil {
		http.Error(w, "Failed to set session", http.StatusInternalServerError)
	}
	err = sess.Set("userinfo", resp.IDTokenClaims)
	if err != nil {
		http.Error(w, "Failed to set session", http.StatusInternalServerError)
	}
	err = a.LogUser(sess, user.ID)
	if err != nil {
		http.Error(w, "Failed to set session", http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// AuthService without external SSO (for testing purposes)
type AuthServiceMock struct {
	GlobalSessions *session.Manager
	userRepo       repository.UserRepository
}

func MockAuthService(userRepo repository.UserRepository) *AuthServiceMock {
	globalSessions, err := session.NewManager("memory", "gosessionid", 3600)
	if err != nil {
		log.Fatal(err)
	}
	go globalSessions.GC()

	return &AuthServiceMock{
		GlobalSessions: globalSessions,
		userRepo:       userRepo,
	}
}

func (a *AuthServiceMock) GetAuthorizedUser(w http.ResponseWriter, r *http.Request) (*models.User, error) {
	sess := a.GlobalSessions.SessionStart(w, r)
	loggedUserId := sess.Get(loggedUserIdQuery)
	if loggedUserId != nil {
		loggedUser, err := a.userRepo.GetByID(loggedUserId.(uint))
		if err != nil {
			return nil, err
		}
		return loggedUser, nil
	}

	return nil, fmt.Errorf("user not logged in")
}

func (a *AuthServiceMock) LogUser(sess session.Session, userId uint) error {
	return sess.Set(loggedUserIdQuery, userId)
}

func (a *AuthServiceMock) LoginHandler(w http.ResponseWriter, r *http.Request) {
	sess := a.GlobalSessions.SessionStart(w, r)

	var userId uint
	var err error

	// get user id to log in from path
	userIdStr := r.URL.Query().Get("userId")
	if userIdStr != "" {
		uId, err := strconv.ParseUint(userIdStr, 10, 32)
		if err != nil {
			log.Panic(err.Error())
		}
		userId = uint(uId)
	} else {
		// default to user with id 1
		userId = uint(1)
	}

	// log in in session
	err = a.LogUser(sess, userId)
	if err != nil {
		log.Panic(err.Error())
	}

	// redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// for interface completeness
func (a *AuthServiceMock) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	a.LoginHandler(w, r)
}
