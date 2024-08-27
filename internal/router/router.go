package router

import (
	"html/template"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/repository"
	"github.com/mytkom/AliceTraINT/internal/handler"
	"gorm.io/gorm"
)

type Middleware struct {
	next        *Middleware
	handlerFunc http.HandlerFunc
}

type MiddlewareChain struct {
	head *Middleware
	tail *Middleware
}

func (m *MiddlewareChain) RegisterMiddleware(middleware *Middleware) {
	if m.tail == nil {
		m.head = middleware
		m.tail = middleware
	} else {
		m.tail.next = middleware
	}

	middleware.next = nil
}

func (m *MiddlewareChain) HandleFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		node := m.head
		for {
			node.handlerFunc(w, r)

			if node.next == nil {
				break
			}

			node = node.next
		}

		handler(w, r)
	}
}

func authMiddleware(auth *auth.Auth) *Middleware {
	return &Middleware{
		handlerFunc: func(w http.ResponseWriter, r *http.Request) {
			sess := auth.GlobalSessions.SessionStart(w, r)
			loggedUserId := sess.Get("loggedUserId")
			if loggedUserId == nil {
				http.Redirect(w, r, "/", http.StatusPermanentRedirect)
			}
		},
	}
}

func NewRouter(db *gorm.DB) *http.ServeMux {
	mux := http.NewServeMux()

	templates := template.Must(template.ParseGlob("web/templates/*.html"))

	userRepo := repository.NewUserRepository(db)
	auth := auth.NewAuth(userRepo)
	userHandler := handler.NewUserHandler(userRepo, templates, auth.GlobalSessions)
	landingHandler := handler.NewLandingHandler(templates)

	var authChain MiddlewareChain
	authChain.RegisterMiddleware(authMiddleware(auth))

	mux.HandleFunc("GET /", landingHandler.Index)
	mux.HandleFunc("GET /users", authChain.HandleFunc(userHandler.Index))
	mux.HandleFunc("POST /users", authChain.HandleFunc(userHandler.CreateUser))
	mux.HandleFunc("GET /login", auth.LoginHandler)
	mux.HandleFunc("GET /callback", auth.CallbackHandler)

	return mux
}
