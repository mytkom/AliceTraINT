package middleware

import (
	"context"
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
	"github.com/mytkom/AliceTraINT/internal/db/models"
)

type contextKey string

const userContextKey contextKey = "userId"

func NewAuthMw(auth *auth.Auth, redirect bool) Middleware {
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

			ctx := context.WithValue(r.Context(), userContextKey, user)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func GetLoggedUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*models.User)
	return user, ok
}
