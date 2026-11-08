package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

type ActivityLogController interface {
	GetRecent(c *gin.Context)
}

type activityLogController struct {
	usecase model.ActivityLogUsecase
}

func NewActivityLogController(usecase model.ActivityLogUsecase) ActivityLogController {
	return &activityLogController{usecase: usecase}
}

func (ac *activityLogController) GetRecent(c *gin.Context) {
	limit := int64(20)
	if raw := c.Query("limit"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil {
			limit = parsed
		}
	}

	logs, err := ac.usecase.GetRecent(c, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Status{
			Message: common.MessInternalServerError,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response.Status{
		IsSuccessful: true,
		Message:      "Activity logs fetched successfully",
		Data:         logs,
	})
}
