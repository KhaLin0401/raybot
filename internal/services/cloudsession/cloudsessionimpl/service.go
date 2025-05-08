package cloudsessionimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/tbe-team/raybot/internal/services/cloudsession"
)

const HeartbeatInterval = 1 * time.Second

type service struct {
	sessionRepository cloudsession.Repository
}

func NewService(sessionRepository cloudsession.Repository) cloudsession.Service {
	return &service{
		sessionRepository: sessionRepository,
	}
}

func (s service) StartSession(ctx context.Context) (cloudsession.Session, error) {
	session := cloudsession.NewSession(HeartbeatInterval)
	err := s.sessionRepository.CreateSession(ctx, session)
	if err != nil {
		return cloudsession.Session{}, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func (s service) HeartbeatSession(ctx context.Context, sessionID string) error {
	session, err := s.sessionRepository.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	if !session.Active(time.Now()) {
		return cloudsession.ErrSessionExpired
	}

	session.Heartbeat()
	return s.sessionRepository.UpdateSession(ctx, session)
}
