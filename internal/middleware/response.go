package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iwandp/community-waste-collection-go/internal/domain"
)

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func Paginated(c *gin.Context, data any, meta Meta) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "meta": meta})
}

func Error(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
	case errors.Is(err, domain.ErrValidation):
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
	case errors.Is(err, domain.ErrConflict),
		errors.Is(err, domain.ErrPendingPayment),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrSafetyCheckRequired):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "internal server error"})
	}
}
