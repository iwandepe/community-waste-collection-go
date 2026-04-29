package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/iwandp/community-waste-collection-go/internal/middleware"
)

type PickupHandler struct {
	svc domain.PickupService
}

func NewPickupHandler(svc domain.PickupService) *PickupHandler {
	return &PickupHandler{svc: svc}
}

func (h *PickupHandler) Create(c *gin.Context) {
	var req struct {
		HouseholdID string            `json:"household_id" binding:"required"`
		Type        domain.PickupType `json:"type" binding:"required"`
		SafetyCheck *bool             `json:"safety_check"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	pickup, err := h.svc.Create(req.HouseholdID, req.Type, req.SafetyCheck)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Created(c, pickup)
}

func (h *PickupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filter := domain.PickupFilter{
		HouseholdID: c.Query("household_id"),
		Status:      domain.PickupStatus(c.Query("status")),
		Page:        page,
		Limit:       limit,
	}
	list, total, err := h.svc.List(filter)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Paginated(c, list, middleware.Meta{Page: page, Limit: limit, Total: total})
}

func (h *PickupHandler) Schedule(c *gin.Context) {
	var req struct {
		PickupDate time.Time `json:"pickup_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	pickup, err := h.svc.Schedule(c.Param("id"), req.PickupDate)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, pickup)
}

func (h *PickupHandler) Complete(c *gin.Context) {
	pickup, err := h.svc.Complete(c.Param("id"))
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, pickup)
}

func (h *PickupHandler) Cancel(c *gin.Context) {
	pickup, err := h.svc.Cancel(c.Param("id"))
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, pickup)
}
