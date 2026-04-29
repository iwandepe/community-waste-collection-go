package domain

import "time"

type Household struct {
	ID        string    `db:"id" json:"id"`
	OwnerName string    `db:"owner_name" json:"owner_name"`
	Address   string    `db:"address" json:"address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type HouseholdRepository interface {
	Create(h *Household) error
	FindByID(id string) (*Household, error)
	FindAll(page, limit int) ([]*Household, int, error)
	Delete(id string) error
}

type HouseholdService interface {
	Create(ownerName, address string) (*Household, error)
	GetByID(id string) (*Household, error)
	List(page, limit int) ([]*Household, int, error)
	Delete(id string) error
}
