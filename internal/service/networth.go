package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	apperr "github.com/nghiaduong/finai/internal/errors"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

// NetWorthSnapshot represents a point-in-time net worth calculation.
type NetWorthSnapshot struct {
	ID               uuid.UUID       `json:"id"`
	TotalAssets      decimal.Decimal `json:"total_assets"`
	TotalLiabilities decimal.Decimal `json:"total_liabilities"`
	NetWorth         decimal.Decimal `json:"net_worth"`
	SnapshotDate     time.Time       `json:"snapshot_date"`
	Breakdown        json.RawMessage `json:"breakdown,omitempty"`
	CreatedAt        *time.Time      `json:"created_at,omitempty"`
}

// NetWorthService calculates and manages net worth snapshots.
type NetWorthService struct {
	queries generated.Querier
}

// NewNetWorthService creates a new net worth service.
func NewNetWorthService(queries generated.Querier) *NetWorthService {
	return &NetWorthService{queries: queries}
}

// CalculateSnapshot sums asset and liability balances, stores a snapshot, and returns it.
func (s *NetWorthService) CalculateSnapshot(ctx context.Context, userID uuid.UUID) (*NetWorthSnapshot, error) {
	assets, err := s.queries.SumAssetBalances(ctx, userID)
	if err != nil {
		return nil, apperr.NewInternalError("failed to sum assets")
	}

	liabilities, err := s.queries.SumLiabilityBalances(ctx, userID)
	if err != nil {
		return nil, apperr.NewInternalError("failed to sum liabilities")
	}

	netWorth := assets.Sub(liabilities)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	breakdown, _ := json.Marshal(map[string]string{
		"assets":      assets.StringFixed(2),
		"liabilities": liabilities.StringFixed(2),
	})
	raw := json.RawMessage(breakdown)

	row, err := s.queries.CreateNetworthSnapshot(ctx, generated.CreateNetworthSnapshotParams{
		UserID:           userID,
		TotalAssets:      assets,
		TotalLiabilities: liabilities,
		NetWorth:         netWorth,
		SnapshotDate:     today,
		Breakdown:        &raw,
	})
	if err != nil {
		return nil, apperr.NewInternalError("failed to save snapshot")
	}

	return mapSnapshot(row), nil
}

// GetLatest returns the most recent net worth snapshot.
func (s *NetWorthService) GetLatest(ctx context.Context, userID uuid.UUID) (*NetWorthSnapshot, error) {
	row, err := s.queries.GetLatestNetworthSnapshot(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.NewNotFoundError("net worth snapshot")
		}
		return nil, apperr.NewInternalError("failed to get latest snapshot")
	}
	return mapSnapshot(row), nil
}

// ListHistory returns net worth snapshots ordered by date descending.
func (s *NetWorthService) ListHistory(ctx context.Context, userID uuid.UUID, limit int) ([]NetWorthSnapshot, error) {
	if limit <= 0 || limit > 365 {
		limit = 30
	}
	rows, err := s.queries.ListNetworthSnapshots(ctx, generated.ListNetworthSnapshotsParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, apperr.NewInternalError("failed to list snapshots")
	}
	result := make([]NetWorthSnapshot, 0, len(rows))
	for _, r := range rows {
		result = append(result, *mapSnapshot(r))
	}
	return result, nil
}

func mapSnapshot(row generated.NetworthSnapshot) *NetWorthSnapshot {
	snap := &NetWorthSnapshot{
		ID:               row.ID,
		TotalAssets:      row.TotalAssets,
		TotalLiabilities: row.TotalLiabilities,
		NetWorth:         row.NetWorth,
		SnapshotDate:     row.SnapshotDate,
		CreatedAt:        row.CreatedAt,
	}
	if row.Breakdown != nil {
		snap.Breakdown = *row.Breakdown
	}
	return snap
}
