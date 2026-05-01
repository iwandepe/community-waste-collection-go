package domain

import (
	"context"
	"mime/multipart"
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

var PickupAmounts = map[PickupType]float64{
	PickupTypeOrganic:    50,
	PickupTypePlastic:    50,
	PickupTypePaper:      50,
	PickupTypeElectronic: 100,
}

type Payment struct {
	ID           string        `db:"id" json:"id"`
	HouseholdID  string        `db:"household_id" json:"household_id"`
	WasteID      string        `db:"waste_id" json:"waste_id"`
	Amount       float64       `db:"amount" json:"amount"`
	PaymentDate  *time.Time    `db:"payment_date" json:"payment_date"`
	Status       PaymentStatus `db:"status" json:"status"`
	ProofFileURL *string       `db:"proof_file_url" json:"proof_file_url"`
	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time     `db:"updated_at" json:"updated_at"`
}

type PaymentFilter struct {
	HouseholdID string
	Status      PaymentStatus
	DateFrom    *time.Time
	DateTo      *time.Time
	Page        int
	Limit       int
}

type PaymentSummaryRow struct {
	Status       PaymentStatus `db:"status" json:"status"`
	Count        int           `db:"count" json:"count"`
	TotalAmount  float64       `db:"total_amount" json:"total_amount"`
}

type PaymentRepository interface {
	Create(p *Payment) error
	FindByID(id string) (*Payment, error)
	FindAll(filter PaymentFilter) ([]*Payment, int, error)
	HasPendingByHousehold(householdID string) (bool, error)
	Confirm(id string, proofURL string, paymentDate time.Time) error
	SummaryByStatus() ([]*PaymentSummaryRow, error)
	FindByHousehold(householdID string) ([]*Payment, error)
}

type StorageService interface {
	UploadFile(ctx context.Context, objectName string, file multipart.File, fileHeader *multipart.FileHeader) (string, error)
}

type PaymentService interface {
	Create(householdID, wasteID string, amount float64) (*Payment, error)
	List(filter PaymentFilter) ([]*Payment, int, error)
	Confirm(ctx context.Context, id string, file multipart.File, fileHeader *multipart.FileHeader) (*Payment, error)
}
