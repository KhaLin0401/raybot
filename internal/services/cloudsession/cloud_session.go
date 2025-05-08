package cloudsession

import (
	"context"

	"github.com/tbe-team/raybot/pkg/xerror"
)

var (
	ErrSessionNotFound = xerror.NotFound(nil, "cloudsession.notFound", "session not found")
	ErrSessionExpired  = xerror.Unauthorized(nil, "cloudsession.expired", "session expired")
)

type Service interface {
	StartSession(ctx context.Context) (Session, error)
	HeartbeatSession(ctx context.Context, sessionID string) error
}

type Repository interface {
	GetSession(ctx context.Context, sessionID string) (Session, error)
	CreateSession(ctx context.Context, session Session) error
	UpdateSession(ctx context.Context, session Session) error
}
