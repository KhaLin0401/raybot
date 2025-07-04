package command

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

//nolint:revive
type CommandType string

func (c CommandType) String() string {
	return string(c)
}

func (c CommandType) Validate() error {
	switch c {
	case CommandTypeStopMovement, CommandTypeMoveForward, CommandTypeMoveBackward,
		CommandTypeMoveTo, CommandTypeCargoOpen, CommandTypeCargoClose,
		CommandTypeCargoLift, CommandTypeCargoLower, CommandTypeCargoCheckQR,
		CommandTypeScanLocation, CommandTypeWait:
		return nil
	}
	return fmt.Errorf("invalid command type: %s", c)
}

const (
	CommandTypeStopMovement CommandType = "STOP_MOVEMENT"
	CommandTypeMoveForward  CommandType = "MOVE_FORWARD"
	CommandTypeMoveBackward CommandType = "MOVE_BACKWARD"
	CommandTypeMoveTo       CommandType = "MOVE_TO"

	CommandTypeCargoOpen    CommandType = "CARGO_OPEN"
	CommandTypeCargoClose   CommandType = "CARGO_CLOSE"
	CommandTypeCargoLift    CommandType = "CARGO_LIFT"
	CommandTypeCargoLower   CommandType = "CARGO_LOWER"
	CommandTypeCargoCheckQR CommandType = "CARGO_CHECK_QR"

	CommandTypeScanLocation CommandType = "SCAN_LOCATION"
	CommandTypeWait         CommandType = "WAIT"
)

type Source string

func (s Source) String() string {
	return string(s)
}

func (s Source) Validate() error {
	switch s {
	case SourceApp, SourceCloud:
		return nil
	}
	return fmt.Errorf("invalid source: %s", s)
}

const (
	SourceApp   Source = "APP"
	SourceCloud Source = "CLOUD"
)

type Status string

func (s Status) Validate() error {
	switch s {
	case StatusQueued, StatusProcessing, StatusCanceling, StatusSucceeded, StatusFailed, StatusCanceled:
		return nil
	}
	return fmt.Errorf("invalid status: %s", s)
}

func (s Status) String() string {
	return string(s)
}

const (
	StatusQueued     Status = "QUEUED"
	StatusProcessing Status = "PROCESSING"
	StatusCanceling  Status = "CANCELING"
	StatusSucceeded  Status = "SUCCEEDED"
	StatusFailed     Status = "FAILED"
	StatusCanceled   Status = "CANCELED"
)

type Command struct {
	ID          int64
	Type        CommandType
	Status      Status
	Source      Source
	Inputs      Inputs
	Outputs     Outputs
	Error       *string
	StartedAt   *time.Time
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RequestID   *string
}

func NewCommand(source Source, inputs Inputs, requestID *string) Command {
	var reqID string
	if requestID == nil {
		reqID = uuid.NewString()
	}

	now := time.Now()
	return Command{
		Type:      inputs.CommandType(),
		Status:    StatusQueued,
		Source:    source,
		Inputs:    inputs,
		CreatedAt: now,
		UpdatedAt: now,
		RequestID: &reqID,
	}
}

// CancelableCommand is a command that can be canceled.
type CancelableCommand struct {
	Command

	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewCancelableCommand(ctx context.Context, cmd Command) CancelableCommand {
	ctx, cancel := context.WithCancel(ctx)
	return CancelableCommand{
		Command:    cmd,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (c *CancelableCommand) Cancel() {
	c.Status = StatusCanceling
	c.UpdatedAt = time.Now()
	c.cancelFunc()
}

func (c *CancelableCommand) CanBeCanceled() bool {
	return c.Status == StatusProcessing
}

func (c *CancelableCommand) Context() context.Context {
	return c.ctx
}
