package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nghiaduong/finai/internal/domain"
	"github.com/nghiaduong/finai/internal/repository/postgres/generated"
)

type CategoryRepo struct {
	q *generated.Queries
}

func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{q: generated.New(pool)}
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	cats := make([]domain.Category, len(rows))
	for i, row := range rows {
		cats[i] = rowToCategory(row)
	}
	return cats, nil
}

func (r *CategoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	row, err := r.q.GetCategoryBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	cat := rowToCategory(row)
	return &cat, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	cat := rowToCategory(row)
	return &cat, nil
}

func rowToCategory(row generated.Category) domain.Category {
	return domain.Category{
		ID:       row.ID,
		Name:     row.Name,
		Slug:     row.Slug,
		Icon:     row.Icon,
		Color:    row.Color,
		ParentID: row.ParentID,
		IsSystem: row.IsSystem,
	}
}
