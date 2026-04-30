package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/iwandp/community-waste-collection-go/internal/middleware"
)

type PaymentHandler struct {
	svc domain.PaymentService
}

func NewPaymentHandler(svc domain.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

func (h *PaymentHandler) Create(c *gin.Context) {
	var req struct {
		HouseholdID string  `json:"household_id" binding:"required"`
		WasteID     string  `json:"waste_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	payment, err := h.svc.Create(req.HouseholdID, req.WasteID, req.Amount)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Created(c, payment)
}

func (h *PaymentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := domain.PaymentFilter{
		HouseholdID: c.Query("household_id"),
		Status:      domain.PaymentStatus(c.Query("status")),
		Page:        page,
		Limit:       limit,
	}

	if from := c.Query("date_from"); from != "" {
		t, err := time.Parse(time.RFC3339, from)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "date_from must be RFC3339"})
			return
		}
		filter.DateFrom = &t
	}
	if to := c.Query("date_to"); to != "" {
		t, err := time.Parse(time.RFC3339, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "date_to must be RFC3339"})
			return
		}
		filter.DateTo = &t
	}

	list, total, err := h.svc.List(filter)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Paginated(c, list, middleware.Meta{Page: page, Limit: limit, Total: total})
}

func (h *PaymentHandler) Confirm(c *gin.Context) {
	file, fileHeader, err := c.Request.FormFile("proof")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "proof file is required"})
		return
	}
	defer file.Close()

	payment, err := h.svc.Confirm(c.Request.Context(), c.Param("id"), file, fileHeader)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, payment)
}
