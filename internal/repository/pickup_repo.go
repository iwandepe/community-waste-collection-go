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

type pickupRepo struct {
	db *sqlx.DB
}

func NewPickupRepository(db *sqlx.DB) domain.PickupRepository {
	return &pickupRepo{db: db}
}

func (r *pickupRepo) Create(p *domain.WastePickup) error {
	p.ID = uuid.NewString()
	query := `INSERT INTO waste_pickups (id, household_id, type, status, pickup_date, safety_check, created_at, updated_at)
	          VALUES (:id, :household_id, :type, :status, :pickup_date, :safety_check, NOW(), NOW())`
	_, err := r.db.NamedExec(query, p)
	if err != nil {
		return err
	}
	return r.db.Get(p, `SELECT * FROM waste_pickups WHERE id = $1`, p.ID)
}

func (r *pickupRepo) FindByID(id string) (*domain.WastePickup, error) {
	var p domain.WastePickup
	err := r.db.Get(&p, `SELECT * FROM waste_pickups WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &p, err
}

func (r *pickupRepo) FindAll(f domain.PickupFilter) ([]*domain.WastePickup, int, error) {
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

	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM waste_pickups `+where, args...); err != nil {
		return nil, 0, err
	}

	offset := (f.Page - 1) * f.Limit
	args = append(args, f.Limit, offset)
	query := fmt.Sprintf(`SELECT * FROM waste_pickups %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, where, i, i+1)

	var list []*domain.WastePickup
	err := r.db.Select(&list, query, args...)
	return list, total, err
}

func (r *pickupRepo) UpdateStatus(id string, status domain.PickupStatus) error {
	res, err := r.db.Exec(
		`UPDATE waste_pickups SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
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

func (r *pickupRepo) UpdateSchedule(id string, pickupDate *time.Time, status domain.PickupStatus) error {
	res, err := r.db.Exec(
		`UPDATE waste_pickups SET status = $1, pickup_date = $2, updated_at = NOW() WHERE id = $3`,
		status, pickupDate, id,
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

func (r *pickupRepo) CancelOrganicExpired(before time.Time) (int64, error) {
	res, err := r.db.Exec(
		`UPDATE waste_pickups SET status = 'canceled', updated_at = NOW()
		 WHERE type = 'organic' AND status = 'pending' AND created_at < $1`,
		before,
	)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
