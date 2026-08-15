package daemon

import (
	"testing"

	"github.com/Est-Void/Vanta/api"
)

func TestSessionStore_Create(t *testing.T) {
	store := NewSessionStore()
	sess := store.Create("s1", "Test Session")

	if sess.ID != "s1" {
		t.Errorf("expected id 's1', got %q", sess.ID)
	}
	if sess.Title != "Test Session" {
		t.Errorf("expected title 'Test Session', got %q", sess.Title)
	}
}

func TestSessionStore_Get(t *testing.T) {
	store := NewSessionStore()
	store.Create("s1", "Test")

	sess, ok := store.Get("s1")
	if !ok {
		t.Fatal("expected session to exist")
	}
	if sess.ID != "s1" {
		t.Errorf("expected id 's1', got %q", sess.ID)
	}
}

func TestSessionStore_Get_NotFound(t *testing.T) {
	store := NewSessionStore()

	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("expected session not to exist")
	}
}

func TestSessionStore_GetOrCreate_Existing(t *testing.T) {
	store := NewSessionStore()
	store.Create("s1", "Original")

	sess := store.GetOrCreate("s1", "New Title")
	if sess.Title != "Original" {
		t.Errorf("expected original title, got %q", sess.Title)
	}
}

func TestSessionStore_GetOrCreate_New(t *testing.T) {
	store := NewSessionStore()

	sess := store.GetOrCreate("s1", "New Session")
	if sess.Title != "New Session" {
		t.Errorf("expected 'New Session', got %q", sess.Title)
	}
}

func TestSessionStore_AddMessage(t *testing.T) {
	store := NewSessionStore()
	store.Create("s1", "Test")

	err := store.AddMessage("s1", api.RoleUser, "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := store.GetMessages("s1")
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("expected 'hello', got %q", msgs[0].Content)
	}
}

func TestSessionStore_AddMessage_NotFound(t *testing.T) {
	store := NewSessionStore()

	err := store.AddMessage("nonexistent", api.RoleUser, "hello")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestSessionStore_GetMessages_NotFound(t *testing.T) {
	store := NewSessionStore()

	msgs := store.GetMessages("nonexistent")
	if msgs != nil {
		t.Errorf("expected nil, got %v", msgs)
	}
}

func TestSessionStore_MultipleMessages(t *testing.T) {
	store := NewSessionStore()
	store.Create("s1", "Test")

	store.AddMessage("s1", api.RoleUser, "hello")
	store.AddMessage("s1", api.RoleAgent, "hi there")
	store.AddMessage("s1", api.RoleUser, "how are you?")

	msgs := store.GetMessages("s1")
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Errorf("expected 'hello', got %q", msgs[0].Content)
	}
	if msgs[1].Content != "hi there" {
		t.Errorf("expected 'hi there', got %q", msgs[1].Content)
	}
	if msgs[2].Content != "how are you?" {
		t.Errorf("expected 'how are you?', got %q", msgs[2].Content)
	}
}
