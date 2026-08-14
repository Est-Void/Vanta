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

func loadClients() map[string][]api.AuthScope {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultClients()
	}

	path := filepath.Join(home, ".config", "vanta", "config.toml")
	f, err := os.Open(path)
	if err != nil {
		return defaultClients()
	}
	defer f.Close()

	var cfg struct {
		Clients []clientConfig `toml:"clients"`
	}
	if _, err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		return defaultClients()
	}

	clients := make(map[string][]api.AuthScope, len(cfg.Clients))
	for _, c := range cfg.Clients {
		clients[c.Token] = c.Scopes
	}
	if len(clients) == 0 {
		return defaultClients()
	}
	return clients
}

func defaultClients() map[string][]api.AuthScope {
	return map[string][]api.AuthScope{
		"dev-token": {api.ScopeScreen, api.ScopeTerminal, api.ScopeDevice, api.ScopeInput, api.ScopeVoice, api.ScopeAgent},
	}
}
