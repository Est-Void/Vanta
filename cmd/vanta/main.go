package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Est-Void/Vanta/api"
	"github.com/Est-Void/Vanta/internal/transport"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "status":
		runStatus(args)
	case "send":
		runSend(args)
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: vanta <command>")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  status       Show daemon status")
	fmt.Fprintln(os.Stderr, "  send <msg>   Send message to agent (streams response)")
	fmt.Fprintln(os.Stderr, "  help         Show this help")
}

func runStatus(args []string) {
	token := loadToken()
	socket := transport.DefaultSocket()

	client := transport.New(socket, token)
	resp, err := client.Get(context.Background(), "/v1/status")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr api.Error
		if json.Unmarshal(body, &apiErr) == nil {
			fmt.Fprintf(os.Stderr, "error: %s: %s\n", apiErr.Code, apiErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	var status api.Status
	if err := json.Unmarshal(body, &status); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Daemon is running\n")
	if status.SessionID != "" {
		fmt.Printf("Session: %s\n", status.SessionID)
	}
	for k, v := range status.Data {
		fmt.Printf("  %s: %v\n", k, v)
	}
}

func runSend(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: vanta send <message>")
		os.Exit(1)
	}

	msg := strings.Join(args, " ")
	token := loadToken()
	socket := transport.DefaultSocket()

	client := transport.New(socket, token)

	reqBody := api.SendMessage{
		ID:      "msg-1",
		Content: msg,
	}
	data, _ := json.Marshal(reqBody)

	resp, err := client.Post(context.Background(), "/v1/session/default/message", "application/json", bytes.NewReader(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiErr api.Error
		if json.Unmarshal(body, &apiErr) == nil {
			fmt.Fprintf(os.Stderr, "error: %s: %s\n", apiErr.Code, apiErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "error: HTTP %d\n", resp.StatusCode)
		}
		os.Exit(1)
	}

	ctx := context.Background()
	events := transport.ReadSSE(ctx, resp.Body)

	for ev := range events {
		switch api.EventType(ev.Event) {
		case api.EvToken:
			var data api.TokenData
			if json.Unmarshal([]byte(ev.Data), &data) == nil {
				fmt.Print(data.Text)
			}
		case api.EvDone:
			fmt.Println()
		case api.EvError:
			var data api.ErrorData
			if json.Unmarshal([]byte(ev.Data), &data) == nil {
				fmt.Fprintf(os.Stderr, "\nerror: %s\n", data.Message)
			}
		}
	}
}

func loadToken() string {
	if tok := os.Getenv("VANTA_TOKEN"); tok != "" {
		return tok
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return defaultToken()
	}

	path := filepath.Join(home, ".config", "vanta", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return defaultToken()
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "token") {
			_, val, _ := strings.Cut(line, "=")
			val = strings.Trim(val, " \"'")
			if val != "" {
				return val
			}
		}
	}
	return defaultToken()
}

func defaultToken() string {
	return "dev-token"
}
