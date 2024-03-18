package middleware

import (
	"net/http"
	"strings"

	"filmlibrary/pkg/session"
	"filmlibrary/pkg/users"

	"go.uber.org/zap"
)

var (
	noAuthUrls = map[string]struct{}{
		"/api/login":          {},
		"/api/register":       {},
		"/swagger/index.html": {},
	}
)

func Auth(logger *zap.SugaredLogger, next http.Handler, repo *users.UserMemoryRepository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Authentication middleware",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path))

		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}

		myUser, err := session.GetUser(r.Header.Get("Authorization"), repo)
		if err != nil {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}
		if myUser.Role != "admin" && r.Method == http.MethodPost {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}

		ctx := users.ContextWithUser(r.Context(), myUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
