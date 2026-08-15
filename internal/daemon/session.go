package daemon

import (
	"errors"
	"sync"
	"time"

	"github.com/Est-Void/Vanta/api"
	"github.com/Est-Void/Vanta/internal/llm"
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID           string
	Title        string
	Messages     []llm.Message
	CreatedAt    time.Time
	eventCounter int64
}

func (s *Session) NextEventID() int64 {
	s.eventCounter++
	return s.eventCounter
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return sess, ok
}

func (s *SessionStore) Create(id, title string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess := &Session{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
	}
	s.sessions[id] = sess
	return sess
}

func (s *SessionStore) GetOrCreate(id, title string) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess, ok := s.sessions[id]; ok {
		return sess
	}
	sess := &Session{
		ID:        id,
		Title:     title,
		CreatedAt: time.Now(),
	}
	s.sessions[id] = sess
	return sess
}

func (s *SessionStore) AddMessage(sessionID string, role api.Role, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}
	sess.Messages = append(sess.Messages, llm.Message{
		Role:    llm.Role(role),
		Content: content,
	})
	return nil
}

func (s *SessionStore) GetMessages(sessionID string) []llm.Message {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[sessionID]
	if !ok {
		return nil
	}
	out := make([]llm.Message, len(sess.Messages))
	copy(out, sess.Messages)
	return out
}
