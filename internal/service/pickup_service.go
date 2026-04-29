package service

import (
	"fmt"
	"time"

	"github.com/iwandp/community-waste-collection-go/internal/domain"
)

type pickupService struct {
	pickupRepo  domain.PickupRepository
	paymentRepo domain.PaymentRepository
}

func NewPickupService(pickupRepo domain.PickupRepository, paymentRepo domain.PaymentRepository) domain.PickupService {
	return &pickupService{pickupRepo: pickupRepo, paymentRepo: paymentRepo}
}

func (s *pickupService) Create(householdID string, pickupType domain.PickupType, safetyCheck *bool) (*domain.WastePickup, error) {
	if householdID == "" {
		return nil, fmt.Errorf("%w: household_id is required", domain.ErrValidation)
	}
	switch pickupType {
	case domain.PickupTypeOrganic, domain.PickupTypePlastic, domain.PickupTypePaper, domain.PickupTypeElectronic:
	default:
		return nil, fmt.Errorf("%w: invalid pickup type", domain.ErrValidation)
	}

	// Business rule 1: block if household has any pending payment
	hasPending, err := s.paymentRepo.HasPendingByHousehold(householdID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, domain.ErrPendingPayment
	}

	p := &domain.WastePickup{
		HouseholdID: householdID,
		Type:        pickupType,
		Status:      domain.PickupStatusPending,
		SafetyCheck: safetyCheck,
	}
	if err := s.pickupRepo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *pickupService) List(filter domain.PickupFilter) ([]*domain.WastePickup, int, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}
	return s.pickupRepo.FindAll(filter)
}

func (s *pickupService) Schedule(id string, pickupDate time.Time) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Business rule 2: can only schedule if currently pending
	if p.Status != domain.PickupStatusPending {
		return nil, fmt.Errorf("%w: pickup must be pending to schedule (current: %s)", domain.ErrInvalidStatus, p.Status)
	}

	// Business rule 3: electronic requires safety_check = true
	if p.Type == domain.PickupTypeElectronic {
		if p.SafetyCheck == nil || !*p.SafetyCheck {
			return nil, domain.ErrSafetyCheckRequired
		}
	}

	if err := s.pickupRepo.UpdateSchedule(id, &pickupDate, domain.PickupStatusScheduled); err != nil {
		return nil, err
	}
	p.Status = domain.PickupStatusScheduled
	p.PickupDate = &pickupDate
	return p, nil
}

func (s *pickupService) Complete(id string) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status != domain.PickupStatusScheduled {
		return nil, fmt.Errorf("%w: pickup must be scheduled to complete (current: %s)", domain.ErrInvalidStatus, p.Status)
	}
	if err := s.pickupRepo.UpdateStatus(id, domain.PickupStatusCompleted); err != nil {
		return nil, err
	}

	// Business rule 5: auto-generate payment on completion
	amount := domain.PickupAmounts[p.Type]
	payment := &domain.Payment{
		HouseholdID: p.HouseholdID,
		WasteID:     p.ID,
		Amount:      amount,
		Status:      domain.PaymentStatusPending,
	}
	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, err
	}

	p.Status = domain.PickupStatusCompleted
	return p, nil
}

func (s *pickupService) Cancel(id string) (*domain.WastePickup, error) {
	p, err := s.pickupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p.Status == domain.PickupStatusCompleted || p.Status == domain.PickupStatusCanceled {
		return nil, fmt.Errorf("%w: cannot cancel a %s pickup", domain.ErrInvalidStatus, p.Status)
	}
	if err := s.pickupRepo.UpdateStatus(id, domain.PickupStatusCanceled); err != nil {
		return nil, err
	}
	p.Status = domain.PickupStatusCanceled
	return p, nil
}
