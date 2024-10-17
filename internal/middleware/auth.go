package middleware

import (
	"net/http"

	"github.com/mytkom/AliceTraINT/internal/auth"
)

func NewAuthMw(auth *auth.Auth) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess := auth.GlobalSessions.SessionStart(w, r)
			loggedUserId := sess.Get("loggedUserId")
			if loggedUserId == nil {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}
