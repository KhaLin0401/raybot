package cloudsession

import (
	"time"

	"github.com/lithammer/shortuuid/v4"
)

type Session struct {
	id                string
	deadline          time.Time
	heartbeatInterval time.Duration
}

// NewSession creates a new session with a random ID and the given heartbeat interval.
func NewSession(heartbeatInterval time.Duration) Session {
	return NewSessionWithID(shortuuid.New(), heartbeatInterval)
}

// NewSessionWithID creates a new session with the given ID and heartbeat interval.
func NewSessionWithID(id string, heartbeatInterval time.Duration) Session {
	sess := Session{
		id:                id,
		heartbeatInterval: heartbeatInterval,
	}
	sess.Heartbeat()
	return sess
}

// ID returns the ID of the session.
func (s *Session) ID() string {
	return s.id
}

// Heartbeat updates the deadline of the session to the current time plus the heartbeat interval.
func (s *Session) Heartbeat() {
	s.deadline = time.Now().Add(s.heartbeatInterval)
}

// Active returns true if the session is active, false otherwise.
func (s *Session) Active(at time.Time) bool {
	return s.deadline.After(at)
}

// HeartbeatInterval returns the heartbeat interval of the session.
func (s *Session) HeartbeatInterval() time.Duration {
	return s.heartbeatInterval
}

// Deadline returns the deadline of the session.
func (s *Session) Deadline() time.Time {
	return s.deadline
}
