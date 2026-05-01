package domain

import "time"

type PickupType string

const (
	PickupTypeOrganic    PickupType = "organic"
	PickupTypePlastic    PickupType = "plastic"
	PickupTypePaper      PickupType = "paper"
	PickupTypeElectronic PickupType = "electronic"
)

type PickupStatus string

const (
	PickupStatusPending   PickupStatus = "pending"
	PickupStatusScheduled PickupStatus = "scheduled"
	PickupStatusCompleted PickupStatus = "completed"
	PickupStatusCanceled  PickupStatus = "canceled"
)

type WastePickup struct {
	ID          string       `db:"id" json:"id"`
	HouseholdID string       `db:"household_id" json:"household_id"`
	Type        PickupType   `db:"type" json:"type"`
	Status      PickupStatus `db:"status" json:"status"`
	PickupDate  *time.Time   `db:"pickup_date" json:"pickup_date"`
	SafetyCheck *bool        `db:"safety_check" json:"safety_check"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updated_at"`
}

type PickupFilter struct {
	HouseholdID string
	Status      PickupStatus
	Page        int
	Limit       int
}

type PickupSummaryRow struct {
	Type   PickupType   `db:"type" json:"type"`
	Status PickupStatus `db:"status" json:"status"`
	Count  int          `db:"count" json:"count"`
}

type PickupRepository interface {
	Create(p *WastePickup) error
	FindByID(id string) (*WastePickup, error)
	FindAll(filter PickupFilter) ([]*WastePickup, int, error)
	UpdateStatus(id string, status PickupStatus) error
	UpdateSchedule(id string, pickupDate *time.Time, status PickupStatus) error
	CancelOrganicExpired(before time.Time) (int64, error)
	SummaryByTypeAndStatus() ([]*PickupSummaryRow, error)
	FindByHousehold(householdID string) ([]*WastePickup, error)
}

type PickupService interface {
	Create(householdID string, pickupType PickupType, safetyCheck *bool) (*WastePickup, error)
	List(filter PickupFilter) ([]*WastePickup, int, error)
	Schedule(id string, pickupDate time.Time) (*WastePickup, error)
	Complete(id string) (*WastePickup, error)
	Cancel(id string) (*WastePickup, error)
}
