package middleware

import (
	"context"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
)

type contextKey string

const userContextKey contextKey = "userId"

func NewAuthMw(auth auth.IAuthService, redirect bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := auth.GetAuthorizedUser(w, r)
			if err != nil {
				if redirect {
					http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
					return
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			}

			r = SetUserContext(r, user)

			next.ServeHTTP(w, r)
		})
	}
}

func SetUserContext(r *http.Request, user *models.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func GetLoggedUser(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(userContextKey).(*models.User)
	return user, ok
}
