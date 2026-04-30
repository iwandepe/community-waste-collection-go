package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
)

type paymentService struct {
	paymentRepo domain.PaymentRepository
	pickupRepo  domain.PickupRepository
	storage     domain.StorageService
}

func NewPaymentService(paymentRepo domain.PaymentRepository, pickupRepo domain.PickupRepository, storage domain.StorageService) domain.PaymentService {
	return &paymentService{paymentRepo: paymentRepo, pickupRepo: pickupRepo, storage: storage}
}

func (s *paymentService) Create(householdID, wasteID string, amount float64) (*domain.Payment, error) {
	if householdID == "" || wasteID == "" {
		return nil, fmt.Errorf("%w: household_id and waste_id are required", domain.ErrValidation)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", domain.ErrValidation)
	}

	// ensure the pickup exists and belongs to the household
	pickup, err := s.pickupRepo.FindByID(wasteID)
	if err != nil {
		return nil, err
	}
	if pickup.HouseholdID != householdID {
		return nil, fmt.Errorf("%w: pickup does not belong to this household", domain.ErrValidation)
	}

	p := &domain.Payment{
		HouseholdID: householdID,
		WasteID:     wasteID,
		Amount:      amount,
		Status:      domain.PaymentStatusPending,
	}
	if err := s.paymentRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *paymentService) List(filter domain.PaymentFilter) ([]*domain.Payment, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}
	return s.paymentRepo.FindAll(filter)
}

func (s *paymentService) Confirm(ctx context.Context, id string, file multipart.File, fileHeader *multipart.FileHeader) (*domain.Payment, error) {
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if payment.Status != domain.PaymentStatusPending {
		return nil, fmt.Errorf("%w: only pending payments can be confirmed (current: %s)", domain.ErrInvalidStatus, payment.Status)
	}

	ext := filepath.Ext(fileHeader.Filename)
	objectName := fmt.Sprintf("payments/%s%s", uuid.NewString(), ext)

	proofURL, err := s.storage.UploadFile(ctx, objectName, file, fileHeader)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.paymentRepo.Confirm(id, proofURL, now); err != nil {
		return nil, err
	}
	payment.Status = domain.PaymentStatusPaid
	payment.ProofFileURL = &proofURL
	payment.PaymentDate = &now
	return payment, nil
}
