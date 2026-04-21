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

type BillRepo struct {
	q *generated.Queries
}

func NewBillRepo(pool *pgxpool.Pool) *BillRepo {
	return &BillRepo{q: generated.New(pool)}
}

func (r *BillRepo) Create(ctx context.Context, bill *domain.Bill) error {
	row, err := r.q.CreateBill(ctx, generated.CreateBillParams{
		UserID:       bill.UserID,
		Name:         bill.Name,
		Amount:       bill.Amount,
		DueDate:      bill.DueDate,
		Frequency:    bill.Frequency,
		CategoryID:   bill.CategoryID,
		IsAutopay:    bill.IsAutopay,
		ReminderDays: int32(bill.ReminderDays),
	})
	if err != nil {
		return err
	}
	bill.ID = row.ID
	bill.Status = row.Status
	bill.CreatedAt = row.CreatedAt
	bill.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *BillRepo) GetByID(ctx context.Context, userID, billID uuid.UUID) (*domain.Bill, error) {
	row, err := r.q.GetBillByID(ctx, generated.GetBillByIDParams{
		ID:     billID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return billRowToDomain(row), nil
}

func (r *BillRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Bill, error) {
	rows, err := r.q.ListBillsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return billListToDomain(rows), nil
}

func (r *BillRepo) ListUpcoming(ctx context.Context, userID uuid.UUID, days int) ([]domain.Bill, error) {
	rows, err := r.q.ListUpcomingBills(ctx, generated.ListUpcomingBillsParams{
		UserID: userID,
		Column2: int32(days),
	})
	if err != nil {
		return nil, err
	}
	bills := make([]domain.Bill, len(rows))
	for i, row := range rows {
		bills[i] = domain.Bill{
			ID: row.ID, UserID: row.UserID, Name: row.Name, Amount: row.Amount,
			DueDate: row.DueDate, Frequency: row.Frequency, CategoryID: row.CategoryID,
			IsAutopay: row.IsAutopay, Status: row.Status, ReminderDays: int(row.ReminderDays),
			NegotiationTip: row.NegotiationTip, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			bills[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				bills[i].Category.ID = *row.CategoryID
			}
		}
	}
	return bills, nil
}

func (r *BillRepo) ListOverdue(ctx context.Context, userID uuid.UUID) ([]domain.Bill, error) {
	rows, err := r.q.ListOverdueBills(ctx, userID)
	if err != nil {
		return nil, err
	}
	bills := make([]domain.Bill, len(rows))
	for i, row := range rows {
		bills[i] = domain.Bill{
			ID: row.ID, UserID: row.UserID, Name: row.Name, Amount: row.Amount,
			DueDate: row.DueDate, Frequency: row.Frequency, CategoryID: row.CategoryID,
			IsAutopay: row.IsAutopay, Status: row.Status, ReminderDays: int(row.ReminderDays),
			NegotiationTip: row.NegotiationTip, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			bills[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				bills[i].Category.ID = *row.CategoryID
			}
		}
	}
	return bills, nil
}

func (r *BillRepo) Update(ctx context.Context, bill *domain.Bill) error {
	_, err := r.q.UpdateBill(ctx, generated.UpdateBillParams{
		ID:           bill.ID,
		UserID:       bill.UserID,
		Name:         bill.Name,
		Amount:       bill.Amount,
		DueDate:      bill.DueDate,
		Frequency:    bill.Frequency,
		CategoryID:   bill.CategoryID,
		IsAutopay:    bill.IsAutopay,
		ReminderDays: int32(bill.ReminderDays),
	})
	return err
}

func (r *BillRepo) UpdateStatus(ctx context.Context, userID, billID uuid.UUID, status string) error {
	return r.q.UpdateBillStatus(ctx, generated.UpdateBillStatusParams{
		ID:     billID,
		UserID: userID,
		Status: status,
	})
}

func (r *BillRepo) Delete(ctx context.Context, userID, billID uuid.UUID) error {
	return r.q.DeleteBill(ctx, generated.DeleteBillParams{
		ID:     billID,
		UserID: userID,
	})
}

func billRowToDomain(row generated.GetBillByIDRow) *domain.Bill {
	b := &domain.Bill{
		ID: row.ID, UserID: row.UserID, Name: row.Name, Amount: row.Amount,
		DueDate: row.DueDate, Frequency: row.Frequency, CategoryID: row.CategoryID,
		IsAutopay: row.IsAutopay, Status: row.Status, ReminderDays: int(row.ReminderDays),
		NegotiationTip: row.NegotiationTip, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.CategoryName != nil {
		b.Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
		if row.CategoryID != nil {
			b.Category.ID = *row.CategoryID
		}
	}
	return b
}

func billListToDomain(rows []generated.ListBillsByUserIDRow) []domain.Bill {
	bills := make([]domain.Bill, len(rows))
	for i, row := range rows {
		bills[i] = domain.Bill{
			ID: row.ID, UserID: row.UserID, Name: row.Name, Amount: row.Amount,
			DueDate: row.DueDate, Frequency: row.Frequency, CategoryID: row.CategoryID,
			IsAutopay: row.IsAutopay, Status: row.Status, ReminderDays: int(row.ReminderDays),
			NegotiationTip: row.NegotiationTip, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			bills[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				bills[i].Category.ID = *row.CategoryID
			}
		}
	}
	return bills
}
