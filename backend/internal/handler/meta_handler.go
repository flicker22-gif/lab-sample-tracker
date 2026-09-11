package handler

import (
	"net/http"

	"sample-tracker/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MetaHandler struct{ db *gorm.DB }

func NewMetaHandler(db *gorm.DB) *MetaHandler { return &MetaHandler{db: db} }

// ListUsers GET /api/v1/users
func (h *MetaHandler) ListUsers(c *gin.Context) {
	var users []model.User
	if err := h.db.Order("id").Find(&users).Error; err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(users))
}

// ListLocations GET /api/v1/locations
func (h *MetaHandler) ListLocations(c *gin.Context) {
	q := h.db.Model(&model.Location{})
	if t := c.Query("type"); t != "" {
		q = q.Where("type = ?", t)
	}
	var locs []model.Location
	if err := q.Order("type, id").Find(&locs).Error; err != nil {
		fail(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, ok(locs))
}
