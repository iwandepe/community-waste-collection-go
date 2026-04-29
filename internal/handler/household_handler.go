package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
	"github.com/iwandp/community-waste-collection-go/internal/middleware"
)

type HouseholdHandler struct {
	svc domain.HouseholdService
}

func NewHouseholdHandler(svc domain.HouseholdService) *HouseholdHandler {
	return &HouseholdHandler{svc: svc}
}

func (h *HouseholdHandler) Create(c *gin.Context) {
	var req struct {
		OwnerName string `json:"owner_name" binding:"required"`
		Address   string `json:"address" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	household, err := h.svc.Create(req.OwnerName, req.Address)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Created(c, household)
}

func (h *HouseholdHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	list, total, err := h.svc.List(page, limit)
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Paginated(c, list, middleware.Meta{Page: page, Limit: limit, Total: total})
}

func (h *HouseholdHandler) GetByID(c *gin.Context) {
	household, err := h.svc.GetByID(c.Param("id"))
	if err != nil {
		middleware.Error(c, err)
		return
	}
	middleware.Success(c, household)
}

func (h *HouseholdHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Param("id")); err != nil {
		middleware.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "household deleted"})
}
