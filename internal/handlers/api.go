package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ItserX/rest/internal/logger"
	"github.com/ItserX/rest/internal/storage"
	"github.com/ItserX/rest/internal/types"
)

type Handler struct {
	Repo storage.PostRepository
}

func (h *Handler) logStart(c *gin.Context) {
	logger.Logger.Infow("Request started",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)
}

func (h *Handler) logSuccess(c *gin.Context, msg string, code int, keysAndValues ...interface{}) {
	logger.Logger.Infow(msg,
		append([]interface{}{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"code", code,
		}, keysAndValues...)...,
	)
}

func (h *Handler) logError(c *gin.Context, err error, code int, keysAndValues ...interface{}) {
	logger.Logger.Errorw(err.Error(),
		append([]interface{}{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"code", code,
			"error", err,
		}, keysAndValues...)...,
	)
}

// @Summary Получить подписку по ID
// @Description Получить информацию о подписке по её идентификатору
// @Tags Подписки
// @Accept json
// @Produce json
// @Param id path string true "ID подписки" example("8d05c8f6-8a7e-4e07-8dc6-07e1b7bafef0")
// @Success 200 {object} types.Subscription
// @Failure 400 {object} types.InvalidIDErrorResponse
// @Failure 404 {object} types.NotFoundErrorResponse
// @Failure 500 {object} types.FailedToGetSubErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSub(c *gin.Context) {
	h.logStart(c)

	id, err := getID(c)
	if err != nil {
		h.logError(c, err, http.StatusBadRequest, "operation", "getID")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	sub, err := h.Repo.Get(id)
	if errors.Is(err, storage.ErrNotFound) {
		h.logError(c, err, http.StatusNotFound, "operation", "GetByID", "id", id)
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Subscription not found"})
		return
	}
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "GetByID", "id", id)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to get subscription"})
		return
	}

	h.logSuccess(c, "Subscription retrieved", http.StatusOK, "id", id)
	c.JSON(http.StatusOK, sub)
}

// @Summary Создать новую подписку
// @Description Создать новую подписку с указанными параметрами
// @Tags Подписки
// @Accept json
// @Produce json
// @Param subscription body types.Subscription true "Данные подписки"
// @Success 201 {object} types.CreatedResponse
// @Failure 400 {object} types.InvalidRequestBodyErrorResponse
// @Failure 500 {object} types.FailedToCreateErrorResponse
// @Router /subscriptions [post]
func (h *Handler) CreateSub(c *gin.Context) {
	h.logStart(c)

	var sub types.Subscription
	err := c.ShouldBindJSON(&sub)
	if err != nil {
		h.logError(c, err, http.StatusBadRequest, "operation", "ShouldBindJSON")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request body"})
		return
	}

	subID, err := h.Repo.Create(sub)
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "Create", "subscription", sub)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to create subscription"})
		return
	}

	subIDStr := fmt.Sprintf("%v", subID)
	h.logSuccess(c, "Subscription created", http.StatusCreated, "sub_id", subIDStr, "subscription", sub)
	c.JSON(http.StatusCreated, types.CreatedResponse{SubID: subIDStr})
}

// @Summary Обновить подписку
// @Description Обновить существующую подписку по ID
// @Tags Подписки
// @Accept json
// @Produce json
// @Param id path string true "ID подписки" example("8d05c8f6-8a7e-4e07-8dc6-07e1b7bafef0")
// @Param subscription body types.Subscription true "Обновленные данные подписки"
// @Success 200 {object} types.IDResponse
// @Failure 400 {object} types.InvalidIDErrorResponse
// @Failure 400 {object} types.InvalidRequestBodyErrorResponse
// @Failure 404 {object} types.NotFoundErrorResponse
// @Failure 500 {object} types.FailedToUpdateSub
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSub(c *gin.Context) {
	h.logStart(c)

	id, err := getID(c)
	if err != nil {
		h.logError(c, err, http.StatusBadRequest, "operation", "getID")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var sub types.Subscription
	err = c.ShouldBindJSON(&sub)
	if err != nil {
		h.logError(c, err, http.StatusBadRequest, "operation", "ShouldBindJSON")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid request body"})
		return
	}

	err = h.Repo.Update(id, sub)
	if errors.Is(err, storage.ErrNotFound) {
		h.logError(c, err, http.StatusNotFound, "operation", "Update", "id", id, "subscription", sub)
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Subscription not found"})
		return
	}
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "Update", "id", id, "subscription", sub)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to update subscription"})
		return
	}

	h.logSuccess(c, "Subscription updated", http.StatusOK, "id", id)
	c.JSON(http.StatusOK, types.IDResponse{ID: id.String()})
}

// @Summary Удалить подписку
// @Description Удалить подписку по ID
// @Tags Подписки
// @Accept json
// @Produce json
// @Param id path string true "ID подписки" example("8d05c8f6-8a7e-4e07-8dc6-07e1b7bafef0")
// @Success 200 {object} types.IDResponse
// @Failure 400 {object} types.InvalidIDErrorResponse
// @Failure 404 {object} types.NotFoundErrorResponse
// @Failure 500 {object} types.FailedToDeleteErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSub(c *gin.Context) {
	h.logStart(c)

	id, err := getID(c)
	if err != nil {
		h.logError(c, err, http.StatusBadRequest, "operation", "getID")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	err = h.Repo.Delete(id)
	if errors.Is(err, storage.ErrNotFound) {
		h.logError(c, err, http.StatusNotFound, "operation", "Delete", "id", id)
		c.JSON(http.StatusNotFound, types.ErrorResponse{Error: "Subscription not found"})
		return
	}
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "Delete", "id", id)
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to delete subscription"})
		return
	}

	h.logSuccess(c, "Subscription deleted", http.StatusOK, "id", id)
	c.JSON(http.StatusOK, types.IDResponse{ID: id.String()})
}

// @Summary Получить список подписок
// @Description Получить список всех подписок
// @Tags Подписки
// @Accept json
// @Produce json
// @Success 200 {object} types.ListSubscriptionsResponse
// @Failure 500 {object} types.FailedToListSubsErrorResponse
// @Router /subscriptions [get]
func (h *Handler) ListSubs(c *gin.Context) {
	h.logStart(c)

	subs, err := h.Repo.List()
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "List")
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to list subscriptions"})
		return
	}

	h.logSuccess(c, "All subscriptions listed", http.StatusOK, "count", len(subs))
	c.JSON(http.StatusOK, types.ListSubscriptionsResponse{
		Subscriptions: subs,
		Count:         len(subs),
	})
}

// @Summary Рассчитать общую стоимость
// @Description Рассчитать общую стоимость подписок с возможностью фильтрации
// @Tags Подписки
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя для фильтрации" example("60601fee-2bf1-4721-ae6f-7636e79a0cba")
// @Param service_name query string false "Название сервиса для фильтрации" example("Yandex")
// @Param period_start query string true "Начальный период в формате ММ-ГГГГ" example("07-2025")
// @Param period_end query string false "Конечный период в формате ММ-ГГГГ" example("12-2025")
// @Success 200 {object} types.TotalCostResponse
// @Failure 400 {object} types.PeriodStartRequiredErrorResponse
// @Failure 400 {object} types.InvalidUserIDErrorResponse
// @Failure 500 {object} types.FailedToCalculateErrorResponse
// @Router /subscriptions/totalCost [get]
func (h *Handler) GetTotalCost(c *gin.Context) {
	h.logStart(c)

	userIDStr := c.Query("user_id")
	serviceName := c.Query("service_name")
	periodStart := c.Query("period_start")
	periodEnd := c.Query("period_end")
	if periodEnd == "" {
		periodEnd = "12-2100"
	}
	if periodStart == "" {
		err := fmt.Errorf("period_start are required")
		h.logError(c, err, http.StatusBadRequest, "operation", "parameter validation")
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "period_start are required"})
		return
	}

	var userID uuid.UUID
	var err error
	if userIDStr != "" {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			h.logError(c, err, http.StatusBadRequest, "operation", "parse user_id")
			c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "Invalid user_id format"})
			return
		}
	}

	total, err := h.Repo.GetTotalCost(userID, serviceName, periodStart, periodEnd)
	if err != nil {
		h.logError(c, err, http.StatusInternalServerError, "operation", "CalculateTotalCost")
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{Error: "Failed to calculate total cost"})
		return
	}

	h.logSuccess(c, "Total cost calculated", http.StatusOK,
		"userID", userID.String(),
		"serviceName", serviceName,
		"periodStart", periodStart,
		"periodEnd", periodEnd,
		"total", total,
	)
	c.JSON(http.StatusOK, types.TotalCostResponse{TotalCost: total})
}

func getID(c *gin.Context) (uuid.UUID, error) {
	paramID := c.Param("id")
	id, err := uuid.Parse(paramID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}
