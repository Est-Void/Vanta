package daemon

import (
	"context"
	"crypto/subtle"
	"net/http"
	"slices"
	"strings"

	"github.com/Est-Void/Vanta/api"
)

type contextKey string

const scopesKey contextKey = "scopes"

func withScopes(ctx context.Context, scopes []api.AuthScope) context.Context {
	return context.WithValue(ctx, scopesKey, scopes)
}

func scopesFrom(ctx context.Context) []api.AuthScope {
	scopes, _ := ctx.Value(scopesKey).([]api.AuthScope)
	return scopes
}

func matchToken(clients map[string][]api.AuthScope, token string) ([]api.AuthScope, bool) {
	for want, scopes := range clients {
		if subtle.ConstantTimeCompare([]byte(token), []byte(want)) == 1 {
			return scopes, true
		}
	}
	return nil, false
}

func requireAuth(clients map[string][]api.AuthScope, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(h, "Bearer ")
		if !ok {
			writeError(w, api.ErrUnauthorized, "missing token")
			return
		}
		scopes, ok := matchToken(clients, token)
		if !ok {
			writeError(w, api.ErrUnauthorized, "invalid token")
			return
		}

		ctx := withScopes(r.Context(), scopes)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requireScope(scope api.AuthScope, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(scopesFrom(r.Context()), scope) {
			next.ServeHTTP(w, r)
			return
		}
		writeError(w, api.ErrDenied, "scope not granted")
	})
}
