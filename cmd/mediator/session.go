package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Session struct {
	Conn   *websocket.Conn
	UserID string

	done    chan struct{}
	manager *SessionManager
}

type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

func newSessions() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
	}
}

func (sm *SessionManager) NewSession(conn *websocket.Conn) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session := &Session{
		Conn:    conn,
		done:    make(chan struct{}),
		manager: sm,
	}

	return session
}

func (sm *SessionManager) GetUserSessions() []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

func (sm *SessionManager) GetUserSession(userID string) *Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return sm.sessions[userID]
}

func (s *Session) Closed() <-chan struct{} {
	return s.done
}

func (s *Session) Close() error {
	s.manager.mutex.Lock()
	defer s.manager.mutex.Unlock()

	if s.UserID != "" {
		delete(s.manager.sessions, s.UserID)
	}

	select {
	case <- s.done:
		// already closed
	default:
		close(s.done)
	}

	return nil
}

func (s *Session) SetUserID(userID string) {
	s.manager.mutex.Lock()
	defer s.manager.mutex.Unlock()

	s.UserID = userID
	s.manager.sessions[s.UserID] = s
}
