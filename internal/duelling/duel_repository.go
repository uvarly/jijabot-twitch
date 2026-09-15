package duelling

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"jijabot/internal/store"
)

var ErrDuelNotPending = errors.New("duelling: duel is no longer pending")

type DuelStatus string

const (
	DuelStatusPending   DuelStatus = "pending"
	DuelStatusResolved  DuelStatus = "resolved"
	DuelStatusDeclined  DuelStatus = "declined"
	DuelStatusCancelled DuelStatus = "cancelled"
	DuelStatusExpired   DuelStatus = "expired"
)

type DuelResult string

const (
	DuelResultChallengerWon DuelResult = "challenger_won"
	DuelResultOpponentWon   DuelResult = "opponent_won"
	DuelResultDraw          DuelResult = "draw"
)

type Duel struct {
	ID             int64
	ChallengerID   int64
	OpponentID     int64
	Stake          int64
	Status         DuelStatus
	ChallengerRoll *int
	OpponentRoll   *int
	Result         *DuelResult
	CreatedAt      time.Time
	ExpiresAt      time.Time
	ResolvedAt     *time.Time
}

type DuelRepository interface {
	WithExecutor(executor store.Executor) DuelRepository
	Create(ctx context.Context, challengerID, opponentID int64, stake int64, expiresAt time.Time) (Duel, error)
	FindPendingDuelByChallenger(ctx context.Context, challengerID int64) (Duel, bool, error)
	FindPendingDuelByOpponent(ctx context.Context, opponentID int64) (Duel, bool, error)
	Resolve(ctx context.Context, duelID int64, challengerRoll, opponentRoll int, result DuelResult) error
	SetStatus(ctx context.Context, duelID int64, status DuelStatus) error
	ExpirePending(ctx context.Context, now time.Time) ([]Duel, error)
}

type SQLiteDuelRepository struct {
	executor store.Executor
}

func NewSQLiteDuelRepository(executor store.Executor) *SQLiteDuelRepository {
	return &SQLiteDuelRepository{executor: executor}
}

func (r *SQLiteDuelRepository) WithExecutor(executor store.Executor) DuelRepository {
	return &SQLiteDuelRepository{executor: executor}
}

func (r *SQLiteDuelRepository) Create(ctx context.Context, challengerID, opponentID int64, stake int64, expiresAt time.Time) (Duel, error) {
	const query = `
		INSERT INTO arena_duels (challenger_id, opponent_id, stake, status, expires_at)
		VALUES (?, ?, ?, 'pending', ?)
		RETURNING id, created_at
	`

	var (
		id        int64
		createdAt time.Time
	)

	if err := r.executor.QueryRowContext(ctx, query, challengerID, opponentID, stake, expiresAt).Scan(&id, &createdAt); err != nil {
		return Duel{}, fmt.Errorf("failed to create duel: %w", err)
	}

	return Duel{
		ID:           id,
		ChallengerID: challengerID,
		OpponentID:   opponentID,
		Stake:        stake,
		Status:       DuelStatusPending,
		CreatedAt:    createdAt,
		ExpiresAt:    expiresAt,
	}, nil
}

const findPendingDuelQuery = `
	SELECT id, challenger_id, opponent_id, stake, status, challenger_roll, opponent_roll, result, created_at, expires_at, resolved_at
	FROM arena_duels
	WHERE status = 'pending' AND %s = ?
`

func (r *SQLiteDuelRepository) FindPendingDuelByChallenger(ctx context.Context, challengerID int64) (Duel, bool, error) {
	return r.findPendingDuelBy(ctx, "challenger_id", challengerID)
}

func (r *SQLiteDuelRepository) FindPendingDuelByOpponent(ctx context.Context, opponentID int64) (Duel, bool, error) {
	return r.findPendingDuelBy(ctx, "opponent_id", opponentID)
}

func (r SQLiteDuelRepository) findPendingDuelBy(ctx context.Context, column string, userID int64) (Duel, bool, error) {
	query := fmt.Sprintf(findPendingDuelQuery, column)

	var (
		duel           Duel
		status         string
		challengerRoll *int
		opponentRoll   *int
		result         *string
		resolvedAt     *time.Time
	)

	if err := r.executor.QueryRowContext(ctx, query, userID).Scan(
		&duel.ID, &duel.ChallengerID, &duel.OpponentID, &duel.Stake,
		&status, &challengerRoll, &opponentRoll, &result,
		&duel.CreatedAt, &duel.ExpiresAt, &resolvedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Duel{}, false, nil
		}

		return Duel{}, false, fmt.Errorf("failed to scan pending duel: %w", err)
	}

	duel.Status = DuelStatus(status)
	duel.ChallengerRoll = challengerRoll
	duel.OpponentRoll = opponentRoll
	duel.ResolvedAt = resolvedAt

	if result != nil {
		r := DuelResult(*result)
		duel.Result = &r
	}

	return duel, true, nil
}

func (r *SQLiteDuelRepository) Resolve(ctx context.Context, duelID int64, challengerRoll, opponentRoll int, duelResult DuelResult) error {
	const query = `
		UPDATE arena_duels
		SET status = 'resolved', challenger_roll = ?, opponent_roll = ?, result = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'pending'
	`

	result, err := r.executor.ExecContext(ctx, query, challengerRoll, opponentRoll, string(duelResult), duelID)
	if err != nil {
		return fmt.Errorf("failed to resolve duel: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if n == 0 {
		return ErrDuelNotPending
	}

	return nil
}

func (r *SQLiteDuelRepository) SetStatus(ctx context.Context, duelID int64, duelStatus DuelStatus) error {
	const query = `
		UPDATE arena_duels
		SET status = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = 'pending'
	`

	result, err := r.executor.ExecContext(ctx, query, string(duelStatus), duelID)
	if err != nil {
		return fmt.Errorf("failed to set duel status: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if n == 0 {
		return ErrDuelNotPending
	}

	return nil
}

func (r *SQLiteDuelRepository) ExpirePending(ctx context.Context, now time.Time) ([]Duel, error) {
	const query = `
		UPDATE arena_duels
		SET status = 'expired', resolved_at = CURRENT_TIMESTAMP
		WHERE status = 'pending' AND expires_at <= ?
		RETURNING id, challenger_id, stake
	`

	rows, err := r.executor.QueryContext(ctx, query, now)
	if err != nil {
		return nil, fmt.Errorf("failed to expire pending duels: %w", err)
	}
	defer rows.Close()

	var expiredDuels []Duel

	for rows.Next() {
		var expiredDuel Duel

		if err := rows.Scan(&expiredDuel.ID, &expiredDuel.ChallengerID, &expiredDuel.Stake); err != nil {
			return nil, fmt.Errorf("failed to scan expired duel: %w", err)
		}

		expiredDuels = append(expiredDuels, expiredDuel)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate expired duels: %w", err)
	}

	return expiredDuels, nil
}
