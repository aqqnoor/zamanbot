package main

import (
	"net/http"
	"sync"

	"zmb-assistant/services"
)

const sessionCookieName = "zmb_session"

type SessionStore struct {
	mu   sync.RWMutex
	data map[string][]services.ChatMessage
}

func NewSessionStore() *SessionStore {
	return &SessionStore{data: make(map[string][]services.ChatMessage)}
}

func (s *SessionStore) Get(id string) []services.ChatMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]services.ChatMessage(nil), s.data[id]...)
}

func (s *SessionStore) Append(id string, msg services.ChatMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = append(s.data[id], msg)
}

func (s *SessionStore) Reset(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, id)
}

func (s *SessionStore) EnsureSessionID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	id := RandString(24)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}
