package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Est-Void/Vanta/api"
)

func TestLoadConfigFrom_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`
[[clients]]
token = "test-token-1"
scopes = ["terminal", "screen"]

[[clients]]
token = "test-token-2"
scopes = ["device"]

[llm]
ollama_url = "http://custom:8080"
model = "mistral"
`), 0644)

	cfg := loadConfigFrom(path)

	if len(cfg.Clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(cfg.Clients))
	}
	if cfg.Clients[0].Token != "test-token-1" {
		t.Errorf("expected test-token-1, got %s", cfg.Clients[0].Token)
	}
	if cfg.LLM.OllamaURL != "http://custom:8080" {
		t.Errorf("expected custom ollama url, got %s", cfg.LLM.OllamaURL)
	}
	if cfg.LLM.Model != "mistral" {
		t.Errorf("expected mistral model, got %s", cfg.LLM.Model)
	}
}

func TestLoadConfigFrom_EmptyConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(``), 0644)

	cfg := loadConfigFrom(path)
	defaults := defaultConfig()

	if len(cfg.Clients) != len(defaults.Clients) {
		t.Errorf("expected default clients, got %d clients", len(cfg.Clients))
	}
	if cfg.Clients[0].Token != "dev-token" {
		t.Error("expected dev-token in fallback")
	}
}

func TestLoadConfigFrom_NoFile(t *testing.T) {
	cfg := loadConfigFrom("/nonexistent/path/config.toml")
	defaults := defaultConfig()

	if len(cfg.Clients) != len(defaults.Clients) {
		t.Errorf("expected default clients, got %d clients", len(cfg.Clients))
	}
}

func TestLoadConfigFrom_InvalidToml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`not valid toml {{{`), 0644)

	cfg := loadConfigFrom(path)
	if cfg.Clients[0].Token != "dev-token" {
		t.Error("expected dev-token fallback on invalid toml")
	}
}

func TestClientsFromConfig(t *testing.T) {
	cfg := Config{
		Clients: []clientConfig{
			{Token: "tok1", Scopes: []api.AuthScope{api.ScopeTerminal}},
			{Token: "tok2", Scopes: []api.AuthScope{api.ScopeScreen}},
		},
	}

	clients := clientsFromConfig(cfg)
	if len(clients) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(clients))
	}
	if _, ok := clients["tok1"]; !ok {
		t.Error("missing tok1")
	}
	if _, ok := clients["tok2"]; !ok {
		t.Error("missing tok2")
	}
}
