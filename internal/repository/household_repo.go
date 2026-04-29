package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/jmoiron/sqlx"
)

type householdRepo struct {
	db *sqlx.DB
}

func NewHouseholdRepository(db *sqlx.DB) domain.HouseholdRepository {
	return &householdRepo{db: db}
}

func (r *householdRepo) Create(h *domain.Household) error {
	h.ID = uuid.NewString()
	query := `INSERT INTO households (id, owner_name, address, created_at, updated_at)
	          VALUES (:id, :owner_name, :address, NOW(), NOW())`
	_, err := r.db.NamedExec(query, h)
	if err != nil {
		return err
	}
	return r.db.Get(h, `SELECT * FROM households WHERE id = $1`, h.ID)
}

func (r *householdRepo) FindByID(id string) (*domain.Household, error) {
	var h domain.Household
	err := r.db.Get(&h, `SELECT * FROM households WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &h, err
}

func (r *householdRepo) FindAll(page, limit int) ([]*domain.Household, int, error) {
	offset := (page - 1) * limit
	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM households`); err != nil {
		return nil, 0, err
	}
	var list []*domain.Household
	err := r.db.Select(&list, `SELECT * FROM households ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return list, total, err
}

func (r *householdRepo) Delete(id string) error {
	res, err := r.db.Exec(`DELETE FROM households WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
