package commandimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"golang.org/x/sync/errgroup"

	"github.com/tbe-team/raybot/internal/services/command"
	"github.com/tbe-team/raybot/internal/storage/db"
	"github.com/tbe-team/raybot/internal/storage/db/sqlc"
	"github.com/tbe-team/raybot/pkg/paging"
	"github.com/tbe-team/raybot/pkg/ptr"
)

type repository struct {
	db      db.DB
	queries *sqlc.Queries
}

func NewCommandRepository(db db.DB, queries *sqlc.Queries) command.Repository {
	return &repository{
		db:      db,
		queries: queries,
	}
}

func (r repository) ListCommands(ctx context.Context, params command.ListCommandsParams) (paging.List[command.Command], error) {
	query := sq.
		Select("*").
		From("commands").
		Limit(uint64(params.PagingParams.Limit())).
		Offset(uint64(params.PagingParams.Offset()))

	for _, s := range params.Sorts {
		query = s.Attach(query)
	}

	statuses := []string{}
	for _, s := range params.Statuses {
		statuses = append(statuses, s.String())
	}
	if len(statuses) > 0 {
		query = query.Where(sq.Eq{"status": statuses})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return paging.List[command.Command]{}, fmt.Errorf("failed to build query: %w", err)
	}

	countQuery := sq.
		Select("COUNT(*)").
		From("commands")

	if len(statuses) > 0 {
		countQuery = countQuery.Where(sq.Eq{"status": statuses})
	}

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return paging.List[command.Command]{}, fmt.Errorf("failed to build count query: %w", err)
	}

	var ret paging.List[command.Command]
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		rows, err := r.db.QueryContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("query commands: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var row sqlc.Command
			if err := rows.Scan(
				&row.ID,
				&row.Type,
				&row.Status,
				&row.Source,
				&row.Inputs,
				&row.Error,
				&row.CompletedAt,
				&row.CreatedAt,
				&row.UpdatedAt,
				&row.StartedAt,
				&row.Outputs,
			); err != nil {
				return fmt.Errorf("scan command: %w", err)
			}

			cmd, err := r.convertRowToCommand(row)
			if err != nil {
				return fmt.Errorf("convert row to command: %w", err)
			}
			ret.Items = append(ret.Items, cmd)
		}

		return nil
	})

	g.Go(func() error {
		if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&ret.TotalItems); err != nil {
			return fmt.Errorf("scan count row: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return paging.List[command.Command]{}, fmt.Errorf("errgroup wait: %w", err)
	}

	return ret, nil
}

func (r repository) GetNextExecutableCommand(ctx context.Context) (command.Command, error) {
	row, err := r.queries.CommandGetNextExecutable(ctx, r.db)
	if err != nil {
		if db.IsNoRowsError(err) {
			return command.Command{}, command.ErrNoNextExecutableCommand
		}
		return command.Command{}, fmt.Errorf("failed to get next executable command: %w", err)
	}
	return r.convertRowToCommand(row)
}

func (r repository) GetCurrentProcessingCommand(ctx context.Context) (command.Command, error) {
	row, err := r.queries.CommandGetCurrentProcessing(ctx, r.db)
	if err != nil {
		if db.IsNoRowsError(err) {
			return command.Command{}, command.ErrCommandNotFound
		}
		return command.Command{}, fmt.Errorf("failed to get current processing command: %w", err)
	}
	return r.convertRowToCommand(row)
}

func (r repository) GetCommandByID(ctx context.Context, id int64) (command.Command, error) {
	row, err := r.queries.CommandGetByID(ctx, r.db, id)
	if err != nil {
		if db.IsNoRowsError(err) {
			return command.Command{}, command.ErrCommandNotFound
		}
		return command.Command{}, fmt.Errorf("failed to get command by id: %w", err)
	}
	return r.convertRowToCommand(row)
}

func (r repository) CreateCommand(ctx context.Context, commandArg command.Command) (command.Command, error) {
	inputsBytes, err := json.Marshal(commandArg.Inputs)
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to marshal inputs: %w", err)
	}

	var completedAt *string
	if commandArg.CompletedAt != nil {
		completedAt = ptr.New(commandArg.CompletedAt.Format(time.RFC3339Nano))
	}

	var startedAt *string
	if commandArg.StartedAt != nil {
		startedAt = ptr.New(commandArg.StartedAt.Format(time.RFC3339Nano))
	}

	row, err := r.queries.CommandCreate(ctx, r.db, sqlc.CommandCreateParams{
		Type:        commandArg.Type.String(),
		Status:      commandArg.Status.String(),
		Source:      commandArg.Source.String(),
		Inputs:      string(inputsBytes),
		Error:       commandArg.Error,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		CreatedAt:   commandArg.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:   commandArg.UpdatedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return command.Command{}, fmt.Errorf("queries create command: %w", err)
	}

	commandArg.ID = row.ID
	commandArg.Outputs, err = command.UnmarshalOutputs(commandArg.Type, []byte(row.Outputs))
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	return commandArg, nil
}

func (r repository) UpdateCommand(ctx context.Context, params command.UpdateCommandParams) (command.Command, error) {
	var completedAt *string
	if params.CompletedAt != nil {
		completedAt = ptr.New(params.CompletedAt.Format(time.RFC3339Nano))
	}

	var startedAt *string
	if params.StartedAt != nil {
		startedAt = ptr.New(params.StartedAt.Format(time.RFC3339Nano))
	}

	var outputs string
	if params.Outputs != nil {
		outputsBytes, err := json.Marshal(params.Outputs)
		if err != nil {
			return command.Command{}, fmt.Errorf("failed to marshal outputs: %w", err)
		}
		outputs = string(outputsBytes)
	}

	row, err := r.queries.CommandUpdate(ctx, r.db, sqlc.CommandUpdateParams{
		ID:             params.ID,
		Status:         params.Status.String(),
		SetStatus:      params.SetStatus,
		Outputs:        outputs,
		SetOutputs:     params.SetOutputs,
		Error:          params.Error,
		SetError:       params.SetError,
		StartedAt:      startedAt,
		SetStartedAt:   params.SetStartedAt,
		CompletedAt:    completedAt,
		SetCompletedAt: params.SetCompletedAt,
		UpdatedAt:      params.UpdatedAt.Format(time.RFC3339Nano),
	})
	if err != nil {
		return command.Command{}, fmt.Errorf("queries update command: %w", err)
	}

	return r.convertRowToCommand(row)
}

func (r repository) CancelPendingCommands(ctx context.Context) error {
	err := r.queries.CommandCancelByStatusQueuedAndProcessingAndCanceling(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed to cancel queued and processing commands: %w", err)
	}

	return nil
}

func (r repository) CancelQueuedAndProcessingCommandsCreatedByCloud(ctx context.Context) error {
	err := r.queries.CommandCancelByStatusQueuedAndProcessingAndCreatedByCloud(ctx, r.db)
	if err != nil {
		return fmt.Errorf("failed to cancel queued and processing commands created by cloud: %w", err)
	}
	return nil
}

func (r repository) DeleteCommandByID(ctx context.Context, id int64) error {
	affected, err := r.queries.CommandDeleteByID(ctx, r.db, id)
	if err != nil {
		return fmt.Errorf("failed to delete command by id and not processing: %w", err)
	}
	if affected == 0 {
		return command.ErrCommandInProcessingCanNotBeDeleted
	}
	return nil
}

func (r repository) DeleteOldCommands(ctx context.Context, cutoffTime time.Time) error {
	_, err := r.queries.CommandDeleteOldCommands(ctx, r.db, cutoffTime.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to delete old commands: %w", err)
	}
	return nil
}

func (repository) convertRowToCommand(row sqlc.Command) (command.Command, error) {
	ret := command.Command{
		ID:     row.ID,
		Type:   command.CommandType(row.Type),
		Status: command.Status(row.Status),
		Source: command.Source(row.Source),
		Error:  row.Error,
	}
	var err error

	ret.Inputs, err = command.UnmarshalInputs(command.CommandType(row.Type), []byte(row.Inputs))
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}

	ret.Outputs, err = command.UnmarshalOutputs(command.CommandType(row.Type), []byte(row.Outputs))
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	ret.CreatedAt, err = time.Parse(time.RFC3339Nano, row.CreatedAt)
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to parse created at: %w", err)
	}

	ret.UpdatedAt, err = time.Parse(time.RFC3339Nano, row.UpdatedAt)
	if err != nil {
		return command.Command{}, fmt.Errorf("failed to parse updated at: %w", err)
	}

	if row.StartedAt != nil {
		startedAt, err := time.Parse(time.RFC3339Nano, *row.StartedAt)
		if err != nil {
			return command.Command{}, fmt.Errorf("failed to parse started at: %w", err)
		}
		ret.StartedAt = &startedAt
	}

	if row.CompletedAt != nil {
		completedAt, err := time.Parse(time.RFC3339Nano, *row.CompletedAt)
		if err != nil {
			return command.Command{}, fmt.Errorf("failed to parse completed at: %w", err)
		}
		ret.CompletedAt = &completedAt
	}

	return ret, nil
}
