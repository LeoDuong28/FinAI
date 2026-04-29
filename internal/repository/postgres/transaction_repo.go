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

// TransactionRepo implements both TransactionRepository and TransactionSyncRepository.
type TransactionRepo struct {
	q *generated.Queries
}

func NewTransactionRepo(pool *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{q: generated.New(pool)}
}

// ── TransactionSyncRepository ────────────────────────────────────

func (r *TransactionRepo) UpsertByPlaidID(ctx context.Context, params domain.UpsertTransactionParams) (uuid.UUID, error) {
	row, err := r.q.UpsertTransactionByPlaidID(ctx, generated.UpsertTransactionByPlaidIDParams{
		UserID:       params.UserID,
		AccountID:    params.AccountID,
		PlaidTxnID:   params.PlaidTxnID,
		Amount:       params.Amount,
		CurrencyCode: params.CurrencyCode,
		Date:         params.Date,
		Name:         params.Name,
		MerchantName: params.MerchantName,
		Pending:      params.Pending,
		Type:         params.Type,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return row.ID, nil
}

func (r *TransactionRepo) DeleteByPlaidID(ctx context.Context, userID uuid.UUID, plaidTxnID string) error {
	return r.q.DeleteTransactionByPlaidID(ctx, generated.DeleteTransactionByPlaidIDParams{
		PlaidTxnID: &plaidTxnID,
		UserID:     userID,
	})
}

// ── TransactionRepository ────────────────────────────────────────

func (r *TransactionRepo) Create(ctx context.Context, txn *domain.Transaction) error {
	row, err := r.q.CreateTransaction(ctx, generated.CreateTransactionParams{
		UserID:               txn.UserID,
		AccountID:            txn.AccountID,
		CategoryID:           txn.CategoryID,
		PlaidTxnID:           txn.PlaidTxnID,
		Amount:               txn.Amount,
		CurrencyCode:         txn.CurrencyCode,
		Date:                 txn.Date,
		Name:                 txn.Name,
		MerchantName:         txn.MerchantName,
		Pending:              txn.Pending,
		Type:                 txn.Type,
		Notes:                txn.Notes,
		IsRecurring:          txn.IsRecurring,
		AiCategoryConfidence: txn.AICategoryConfidence,
	})
	if err != nil {
		return err
	}
	txn.ID = row.ID
	txn.CreatedAt = row.CreatedAt
	txn.UpdatedAt = row.UpdatedAt
	return nil
}

func (r *TransactionRepo) GetByID(ctx context.Context, userID, txnID uuid.UUID) (*domain.Transaction, error) {
	row, err := r.q.GetTransactionByID(ctx, generated.GetTransactionByIDParams{
		ID:     txnID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return txnRowToDomain(row), nil
}

func (r *TransactionRepo) List(ctx context.Context, userID uuid.UUID, filter domain.TransactionFilter) ([]domain.Transaction, *domain.TransactionCursor, error) {
	limit := int32(filter.Limit)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Request one extra to determine if there's a next page.
	fetchLimit := limit + 1

	var rows []generated.ListTransactionsRow
	var err error

	if filter.Cursor != nil {
		cursorRows, cursorErr := r.q.ListTransactionsWithCursor(ctx, generated.ListTransactionsWithCursorParams{
			UserID:     userID,
			Limit:      fetchLimit,
			CursorDate: filter.Cursor.Date,
			CursorID:   filter.Cursor.ID,
		})
		if cursorErr != nil {
			return nil, nil, cursorErr
		}
		// Convert cursor rows to standard rows (same struct shape).
		rows = make([]generated.ListTransactionsRow, len(cursorRows))
		for i, cr := range cursorRows {
			rows[i] = generated.ListTransactionsRow{
				ID: cr.ID, UserID: cr.UserID, AccountID: cr.AccountID, CategoryID: cr.CategoryID,
				PlaidTxnID: cr.PlaidTxnID, Amount: cr.Amount, CurrencyCode: cr.CurrencyCode,
				Date: cr.Date, Name: cr.Name, MerchantName: cr.MerchantName, Pending: cr.Pending,
				Type: cr.Type, Notes: cr.Notes, IsExcluded: cr.IsExcluded, IsRecurring: cr.IsRecurring,
				AiCategoryConfidence: cr.AiCategoryConfidence, CreatedAt: cr.CreatedAt, UpdatedAt: cr.UpdatedAt,
				CategoryName: cr.CategoryName, CategoryIcon: cr.CategoryIcon, CategoryColor: cr.CategoryColor,
			}
		}
	} else if filter.CategoryID != nil {
		catRows, catErr := r.q.ListTransactionsByCategory(ctx, generated.ListTransactionsByCategoryParams{
			UserID:     userID,
			CategoryID: filter.CategoryID,
			Limit:      fetchLimit,
		})
		if catErr != nil {
			return nil, nil, catErr
		}
		rows = make([]generated.ListTransactionsRow, len(catRows))
		for i, cr := range catRows {
			rows[i] = generated.ListTransactionsRow{
				ID: cr.ID, UserID: cr.UserID, AccountID: cr.AccountID, CategoryID: cr.CategoryID,
				PlaidTxnID: cr.PlaidTxnID, Amount: cr.Amount, CurrencyCode: cr.CurrencyCode,
				Date: cr.Date, Name: cr.Name, MerchantName: cr.MerchantName, Pending: cr.Pending,
				Type: cr.Type, Notes: cr.Notes, IsExcluded: cr.IsExcluded, IsRecurring: cr.IsRecurring,
				AiCategoryConfidence: cr.AiCategoryConfidence, CreatedAt: cr.CreatedAt, UpdatedAt: cr.UpdatedAt,
				CategoryName: cr.CategoryName, CategoryIcon: cr.CategoryIcon, CategoryColor: cr.CategoryColor,
			}
		}
	} else if filter.DateFrom != nil && filter.DateTo != nil {
		dateRows, dateErr := r.q.ListTransactionsByDateRange(ctx, generated.ListTransactionsByDateRangeParams{
			UserID: userID,
			Date:   *filter.DateFrom,
			Date_2: *filter.DateTo,
			Limit:  fetchLimit,
		})
		if dateErr != nil {
			return nil, nil, dateErr
		}
		rows = make([]generated.ListTransactionsRow, len(dateRows))
		for i, dr := range dateRows {
			rows[i] = generated.ListTransactionsRow{
				ID: dr.ID, UserID: dr.UserID, AccountID: dr.AccountID, CategoryID: dr.CategoryID,
				PlaidTxnID: dr.PlaidTxnID, Amount: dr.Amount, CurrencyCode: dr.CurrencyCode,
				Date: dr.Date, Name: dr.Name, MerchantName: dr.MerchantName, Pending: dr.Pending,
				Type: dr.Type, Notes: dr.Notes, IsExcluded: dr.IsExcluded, IsRecurring: dr.IsRecurring,
				AiCategoryConfidence: dr.AiCategoryConfidence, CreatedAt: dr.CreatedAt, UpdatedAt: dr.UpdatedAt,
				CategoryName: dr.CategoryName, CategoryIcon: dr.CategoryIcon, CategoryColor: dr.CategoryColor,
			}
		}
	} else {
		rows, err = r.q.ListTransactions(ctx, generated.ListTransactionsParams{
			UserID: userID,
			Limit:  fetchLimit,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	// Determine cursor for next page.
	var nextCursor *domain.TransactionCursor
	if int32(len(rows)) > limit {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		nextCursor = &domain.TransactionCursor{Date: last.Date, ID: last.ID}
	}

	txns := make([]domain.Transaction, len(rows))
	for i, row := range rows {
		txns[i] = domain.Transaction{
			ID: row.ID, UserID: row.UserID, AccountID: row.AccountID, CategoryID: row.CategoryID,
			PlaidTxnID: row.PlaidTxnID, Amount: row.Amount, CurrencyCode: row.CurrencyCode,
			Date: row.Date, Name: row.Name, MerchantName: row.MerchantName, Pending: row.Pending,
			Type: row.Type, Notes: row.Notes, IsExcluded: row.IsExcluded, IsRecurring: row.IsRecurring,
			AICategoryConfidence: row.AiCategoryConfidence, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			txns[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				txns[i].Category.ID = *row.CategoryID
			}
		}
	}
	return txns, nextCursor, nil
}

func (r *TransactionRepo) Update(ctx context.Context, txn *domain.Transaction) error {
	return r.q.UpdateTransactionCategory(ctx, generated.UpdateTransactionCategoryParams{
		ID:                   txn.ID,
		UserID:               txn.UserID,
		CategoryID:           txn.CategoryID,
		AiCategoryConfidence: txn.AICategoryConfidence,
	})
}

func (r *TransactionRepo) UpdateNotes(ctx context.Context, userID, txnID uuid.UUID, notes *string) error {
	return r.q.UpdateTransactionNotes(ctx, generated.UpdateTransactionNotesParams{
		ID:     txnID,
		UserID: userID,
		Notes:  notes,
	})
}

func (r *TransactionRepo) Delete(ctx context.Context, userID, txnID uuid.UUID) error {
	return r.q.DeleteTransaction(ctx, generated.DeleteTransactionParams{
		ID:     txnID,
		UserID: userID,
	})
}

func (r *TransactionRepo) Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]domain.Transaction, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := r.q.SearchTransactions(ctx, generated.SearchTransactionsParams{
		UserID:     userID,
		Similarity: query,
		Limit:      int32(limit),
	})
	if err != nil {
		return nil, err
	}
	txns := make([]domain.Transaction, len(rows))
	for i, row := range rows {
		txns[i] = domain.Transaction{
			ID: row.ID, UserID: row.UserID, AccountID: row.AccountID, CategoryID: row.CategoryID,
			Amount: row.Amount, CurrencyCode: row.CurrencyCode, Date: row.Date, Name: row.Name,
			MerchantName: row.MerchantName, Pending: row.Pending, Type: row.Type, Notes: row.Notes,
			IsExcluded: row.IsExcluded, IsRecurring: row.IsRecurring,
			AICategoryConfidence: row.AiCategoryConfidence, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
		if row.CategoryName != nil {
			txns[i].Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
			if row.CategoryID != nil {
				txns[i].Category.ID = *row.CategoryID
			}
		}
	}
	return txns, nil
}

func txnRowToDomain(row generated.GetTransactionByIDRow) *domain.Transaction {
	txn := &domain.Transaction{
		ID: row.ID, UserID: row.UserID, AccountID: row.AccountID, CategoryID: row.CategoryID,
		PlaidTxnID: row.PlaidTxnID, Amount: row.Amount, CurrencyCode: row.CurrencyCode,
		Date: row.Date, Name: row.Name, MerchantName: row.MerchantName, Pending: row.Pending,
		Type: row.Type, Notes: row.Notes, IsExcluded: row.IsExcluded, IsRecurring: row.IsRecurring,
		AICategoryConfidence: row.AiCategoryConfidence, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.CategoryName != nil {
		txn.Category = &domain.Category{Name: *row.CategoryName, Icon: row.CategoryIcon, Color: row.CategoryColor}
		if row.CategoryID != nil {
			txn.Category.ID = *row.CategoryID
		}
	}
	return txn
}
