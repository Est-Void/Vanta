package daemon

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/Est-Void/Vanta/api"
)

type clientConfig struct {
	Token  string          `toml:"token"`
	Scopes []api.AuthScope `toml:"scopes"`
}

type LLMConfig struct {
	OllamaURL string `toml:"ollama_url"`
	Model     string `toml:"model"`
}

type Config struct {
	Clients []clientConfig `toml:"clients"`
	LLM     LLMConfig      `toml:"llm"`
}

func loadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig()
	}
	return loadConfigFrom(filepath.Join(home, ".config", "vanta", "config.toml"))
}

func loadConfigFrom(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		return defaultConfig()
	}
	defer f.Close()

	var cfg Config
	if _, err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return defaultConfig()
	}

	if len(cfg.Clients) == 0 {
		cfg.Clients = defaultConfig().Clients
	}
	if cfg.LLM.OllamaURL == "" {
		cfg.LLM.OllamaURL = "http://localhost:11434"
	}
	if cfg.LLM.Model == "" {
		cfg.LLM.Model = "llama3"
	}
	return cfg
}

func defaultConfig() Config {
	return Config{
		Clients: []clientConfig{
			{Token: "dev-token", Scopes: []api.AuthScope{
				api.ScopeScreen, api.ScopeTerminal, api.ScopeDevice,
				api.ScopeInput, api.ScopeVoice, api.ScopeAgent,
			}},
		},
		LLM: LLMConfig{
			OllamaURL: "http://localhost:11434",
			Model:     "llama3",
		},
	}
}

func clientsFromConfig(cfg Config) map[string][]api.AuthScope {
	m := make(map[string][]api.AuthScope, len(cfg.Clients))
	for _, c := range cfg.Clients {
		m[c.Token] = c.Scopes
	}
	return m
}
