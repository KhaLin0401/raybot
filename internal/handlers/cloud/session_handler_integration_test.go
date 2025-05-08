package cloud_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/tbe-team/raybot/internal/handlers/cloud/cloudtest"
	sessionv1 "github.com/tbe-team/raybot/internal/handlers/cloud/gen/session/v1"
	"github.com/tbe-team/raybot/internal/services/cloudsession/cloudsessionimpl"
)

func TestIntegrationSessionHandler_StartSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testEnv := cloudtest.SetupTunnelTestEnv(t)
	client := sessionv1.NewSessionServiceClient(testEnv.TunnelChannel)

	t.Run("Start session successfully", func(t *testing.T) {
		res, err := client.StartSession(context.Background(), &sessionv1.StartSessionRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.SessionId)
		require.NotNil(t, res.HeartbeatInterval)
	})
}

func TestIntegrationSessionHandler_HeartbeatSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	testEnv := cloudtest.SetupTunnelTestEnv(t)
	client := sessionv1.NewSessionServiceClient(testEnv.TunnelChannel)

	t.Run("Heartbeat session successfully", func(t *testing.T) {
		res, err := client.StartSession(context.Background(), &sessionv1.StartSessionRequest{})
		require.NoError(t, err)

		_, err = client.SendSessionHeartbeat(context.Background(), &sessionv1.SendSessionHeartbeatRequest{
			SessionId: res.SessionId,
		})
		require.NoError(t, err)
	})

	t.Run("Heartbeat session expired", func(t *testing.T) {
		res, err := client.StartSession(context.Background(), &sessionv1.StartSessionRequest{})
		require.NoError(t, err)

		time.Sleep(cloudsessionimpl.HeartbeatInterval + 1*time.Nanosecond)

		_, err = client.SendSessionHeartbeat(context.Background(), &sessionv1.SendSessionHeartbeatRequest{
			SessionId: res.SessionId,
		})
		require.Error(t, err)
	})
}
