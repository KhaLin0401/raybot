package cloudsessionimpl

import (
	"context"
	"sync"

	"github.com/tbe-team/raybot/internal/services/cloudsession"
)

type repository struct {
	mu       sync.RWMutex
	sessions map[string]cloudsession.Session
}

func NewRepository() cloudsession.Repository {
	return &repository{
		sessions: make(map[string]cloudsession.Session),
	}
}

func (r *repository) GetSession(_ context.Context, sessionID string) (cloudsession.Session, error) {
	r.mu.RLock()
	session, ok := r.sessions[sessionID]
	r.mu.RUnlock()

	if !ok {
		return cloudsession.Session{}, cloudsession.ErrSessionNotFound
	}
	return session, nil
}

func (r *repository) CreateSession(_ context.Context, session cloudsession.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID()] = session
	return nil
}

func (r *repository) UpdateSession(_ context.Context, session cloudsession.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID()] = session
	return nil
}
