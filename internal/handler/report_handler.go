package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/iwandp/community-waste-collection-go/internal/middleware"
)

type ReportHandler struct {
	pickupRepo  domain.PickupRepository
	paymentRepo domain.PaymentRepository
}

func NewReportHandler(pickupRepo domain.PickupRepository, paymentRepo domain.PaymentRepository) *ReportHandler {
	return &ReportHandler{pickupRepo: pickupRepo, paymentRepo: paymentRepo}
}

func (h *ReportHandler) WasteSummary(c *gin.Context) {
	rows, err := h.pickupRepo.SummaryByTypeAndStatus()
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, rows)
}

func (h *ReportHandler) PaymentSummary(c *gin.Context) {
	rows, err := h.paymentRepo.SummaryByStatus()
	if err != nil {
		middleware.Error(c, err)
		return
	}

	var totalRevenue float64
	for _, r := range rows {
		if r.Status == domain.PaymentStatusPaid {
			totalRevenue += r.TotalAmount
		}
	}

	middleware.Success(c, gin.H{
		"by_status":     rows,
		"total_revenue": totalRevenue,
	})
}

func (h *ReportHandler) HouseholdHistory(c *gin.Context) {
	id := c.Param("id")

	pickups, err := h.pickupRepo.FindByHousehold(id)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	payments, err := h.paymentRepo.FindByHousehold(id)
	if err != nil {
		middleware.Error(c, err)
		return
	}

	middleware.Success(c, gin.H{
		"household_id": id,
		"pickups":      pickups,
		"payments":     payments,
	})
}
