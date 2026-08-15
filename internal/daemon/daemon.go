package daemon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/Est-Void/Vanta/api"
	"github.com/Est-Void/Vanta/internal/llm"
	"github.com/Est-Void/Vanta/internal/transport"
)

type deps struct {
	clients  map[string][]api.AuthScope
	sessions *SessionStore
	llm      llm.Provider
	model    string
}

func Run(ctx context.Context) error {
	cfg := loadConfig()

	d := &deps{
		clients:  clientsFromConfig(cfg),
		sessions: NewSessionStore(),
		llm:      llm.NewOllama(cfg.LLM.OllamaURL),
		model:    cfg.LLM.Model,
	}

	listener, err := listenSocket()
	if err != nil {
		return err
	}
	defer listener.Close()

	server := &http.Server{
		Handler:           newRouter(d),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdown)
	}()

	return server.Serve(listener)
}

func listenSocket() (net.Listener, error) {
	sock := transport.DefaultSocket()
	os.Remove(sock)

	l, err := net.Listen("unix", sock)
	if err != nil {
		return nil, fmt.Errorf("listen %s: %w", sock, err)
	}
	return l, nil
}

func runtimeDir(appName string) (string, error) {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = fmt.Sprintf("/tmp/%s-%d", appName, os.Getuid())
	}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return "", fmt.Errorf("create fallback dir: %w", err)
			}
			return dir, nil
		}
		return "", fmt.Errorf("stat runtime dir: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	if info.Mode().Perm() != 0700 {
		return "", fmt.Errorf("insecure permissions on %s: expected 0700, got %o", dir, info.Mode().Perm())
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", errors.New("failed to cast file info to syscall.Stat_t")
	}
	if int(stat.Uid) != os.Getuid() {
		return "", fmt.Errorf("runtime dir %s is owned by UID %d, current UID is %d", dir, stat.Uid, os.Getuid())
	}

	return dir, nil
}
