package cloud

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/durationpb"

	sessionv1 "github.com/tbe-team/raybot/internal/handlers/cloud/gen/session/v1"
	"github.com/tbe-team/raybot/internal/services/cloudsession"
)

type sessionHandler struct {
	sessionv1.UnimplementedSessionServiceServer
	sessionService cloudsession.Service
}

func newSessionHandler(sessionService cloudsession.Service) sessionv1.SessionServiceServer {
	return &sessionHandler{
		sessionService: sessionService,
	}
}

func (h sessionHandler) StartSession(ctx context.Context, _ *sessionv1.StartSessionRequest) (*sessionv1.StartSessionResponse, error) {
	session, err := h.sessionService.StartSession(ctx)
	if err != nil {
		return nil, fmt.Errorf("start session: %w", err)
	}

	return &sessionv1.StartSessionResponse{
		SessionId:         session.ID(),
		HeartbeatInterval: durationpb.New(session.HeartbeatInterval()),
	}, nil
}

func (h sessionHandler) SendSessionHeartbeat(ctx context.Context, req *sessionv1.SendSessionHeartbeatRequest) (*sessionv1.SendSessionHeartbeatResponse, error) {
	err := h.sessionService.HeartbeatSession(ctx, req.SessionId)
	if err != nil {
		return nil, fmt.Errorf("heartbeat session: %w", err)
	}

	return &sessionv1.SendSessionHeartbeatResponse{}, nil
}
