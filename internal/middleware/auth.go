package middleware

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
)

func NewAuthMw(auth *auth.Auth, redirect bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess := auth.GlobalSessions.SessionStart(w, r)
			loggedUserId := sess.Get("loggedUserId")
			if loggedUserId == nil {
				if redirect {
					http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				} else {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
