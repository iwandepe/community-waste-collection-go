package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/jmoiron/sqlx"
)

type paymentRepo struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) domain.PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(p *domain.Payment) error {
	p.ID = uuid.NewString()
	query := `INSERT INTO payments (id, household_id, waste_id, amount, status, created_at, updated_at)
	          VALUES (:id, :household_id, :waste_id, :amount, :status, NOW(), NOW())`
	_, err := r.db.NamedExec(query, p)
	if err != nil {
		return err
	}
	return r.db.Get(p, `SELECT * FROM payments WHERE id = $1`, p.ID)
}

func (r *paymentRepo) FindByID(id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.Get(&p, `SELECT * FROM payments WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *paymentRepo) FindAll(f domain.PaymentFilter) ([]*domain.Payment, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	i := 1

	if f.HouseholdID != "" {
		where += fmt.Sprintf(" AND household_id = $%d", i)
		args = append(args, f.HouseholdID)
		i++
	}
	if f.Status != "" {
		where += fmt.Sprintf(" AND status = $%d", i)
		args = append(args, f.Status)
		i++
	}
	if f.DateFrom != nil {
		where += fmt.Sprintf(" AND payment_date >= $%d", i)
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != nil {
		where += fmt.Sprintf(" AND payment_date <= $%d", i)
		args = append(args, f.DateTo)
		i++
	}

	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM payments `+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`SELECT * FROM payments %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, i, i+1)

	var list []*domain.Payment
	err := r.db.Select(&list, query, args...)
	return list, total, err
}

func (r *paymentRepo) HasPendingByHousehold(householdID string) (bool, error) {
	var count int
	err := r.db.Get(&count,
		`SELECT COUNT(*) FROM payments WHERE household_id = $1 AND status = 'pending'`,
		householdID,
	)
	return count > 0, err
}

func (r *paymentRepo) Confirm(id string, proofURL string, paymentDate time.Time) error {
	res, err := r.db.Exec(
		`UPDATE payments SET status = 'paid', proof_file_url = $1, payment_date = $2, updated_at = NOW() WHERE id = $3`,
		proofURL, paymentDate, id,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *paymentRepo) SummaryByStatus() ([]*domain.PaymentSummaryRow, error) {
	var rows []*domain.PaymentSummaryRow
	err := r.db.Select(&rows,
		`SELECT status, COUNT(*) AS count, COALESCE(SUM(amount), 0) AS total_amount FROM payments GROUP BY status ORDER BY status`,
	)
	return rows, err
}

func (r *paymentRepo) FindByHousehold(householdID string) ([]*domain.Payment, error) {
	var list []*domain.Payment
	err := r.db.Select(&list,
		`SELECT * FROM payments WHERE household_id = $1 ORDER BY created_at DESC`,
		householdID,
	)
	return list, err
}
