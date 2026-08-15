package daemon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Est-Void/Vanta/api"
)

func TestMatchToken_Found(t *testing.T) {
	clients := map[string][]api.AuthScope{
		"tok1": {api.ScopeTerminal},
		"tok2": {api.ScopeScreen, api.ScopeDevice},
	}

	scopes, ok := matchToken(clients, "tok1")
	if !ok {
		t.Fatal("expected token to match")
	}
	if len(scopes) != 1 || scopes[0] != api.ScopeTerminal {
		t.Errorf("unexpected scopes: %v", scopes)
	}

	scopes, ok = matchToken(clients, "tok2")
	if !ok {
		t.Fatal("expected token to match")
	}
	if len(scopes) != 2 {
		t.Errorf("unexpected scopes: %v", scopes)
	}
}

func TestMatchToken_NotFound(t *testing.T) {
	clients := map[string][]api.AuthScope{
		"tok1": {api.ScopeTerminal},
	}

	_, ok := matchToken(clients, "wrong-token")
	if ok {
		t.Error("expected token not to match")
	}
}

func TestMatchToken_Empty(t *testing.T) {
	clients := map[string][]api.AuthScope{}

	_, ok := matchToken(clients, "anything")
	if ok {
		t.Error("expected no match on empty clients")
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	clients := map[string][]api.AuthScope{
		"good-token": {api.ScopeTerminal},
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		scopes := scopesFrom(r.Context())
		if len(scopes) != 1 || scopes[0] != api.ScopeTerminal {
			t.Errorf("unexpected scopes in context: %v", scopes)
		}
		w.WriteHeader(http.StatusOK)
	})

	handler := requireAuth(clients, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer good-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	clients := map[string][]api.AuthScope{}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := requireAuth(clients, inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	clients := map[string][]api.AuthScope{
		"real-token": {api.ScopeTerminal},
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := requireAuth(clients, inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestRequireAuth_BearerPrefix(t *testing.T) {
	clients := map[string][]api.AuthScope{
		"tok": {api.ScopeTerminal},
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := requireAuth(clients, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token tok")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for non-Bearer auth, got %d", rr.Code)
	}
}

func TestRequireScope_HasScope(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := requireScope(api.ScopeTerminal, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := withScopes(req.Context(), []api.AuthScope{api.ScopeTerminal, api.ScopeScreen})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestRequireScope_MissingScope(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := requireScope(api.ScopeTerminal, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := withScopes(req.Context(), []api.AuthScope{api.ScopeScreen})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestRequireScope_NoScopes(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	handler := requireScope(api.ScopeTerminal, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestContextScopes_RoundTrip(t *testing.T) {
	ctx := context.Background()
	scopes := []api.AuthScope{api.ScopeTerminal, api.ScopeDevice}

	ctx = withScopes(ctx, scopes)
	got := scopesFrom(ctx)

	if len(got) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(got))
	}
	if got[0] != api.ScopeTerminal || got[1] != api.ScopeDevice {
		t.Errorf("unexpected scopes: %v", got)
	}
}

func TestContextScopes_Empty(t *testing.T) {
	got := scopesFrom(context.Background())
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}
