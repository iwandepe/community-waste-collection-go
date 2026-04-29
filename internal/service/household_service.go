package service

import (
	"fmt"

	"github.com/iwandp/community-waste-collection-go/internal/domain"
)

type householdService struct {
	repo domain.HouseholdRepository
}

func NewHouseholdService(repo domain.HouseholdRepository) domain.HouseholdService {
	return &householdService{repo: repo}
}

func (s *householdService) Create(ownerName, address string) (*domain.Household, error) {
	if ownerName == "" || address == "" {
		return nil, fmt.Errorf("%w: owner_name and address are required", domain.ErrValidation)
	}
	h := &domain.Household{OwnerName: ownerName, Address: address}
	if err := s.repo.Create(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *householdService) GetByID(id string) (*domain.Household, error) {
	return s.repo.FindByID(id)
}

func (s *householdService) List(page, limit int) ([]*domain.Household, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.FindAll(page, limit)
}

func (s *householdService) Delete(id string) error {
	return s.repo.Delete(id)
}
