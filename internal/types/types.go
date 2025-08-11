package types

import "github.com/google/uuid"

// @Description Информация о подписке
type Subscription struct {
	// Название сервиса
	ServiceName string `json:"service_name" binding:"required"`
	// Месячная стоимость в рублях
	Price int `json:"price" binding:"required,min=0"`
	// ID пользователя-владельца подписки
	UserID uuid.UUID `json:"user_id" binding:"required,uuid4"`
	// Дата начала подписки в формате ММ-ГГГГ
	StartDate string `json:"start_date" binding:"required"`
	// Опциональная дата окончания подписки в формате ММ-ГГГГ
	EndDate string `json:"end_date,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid ID format"`
}

type IDResponse struct {
	ID string `json:"id" example:"8d05c8f6-8a7e-4e07-8dc6-07e1b7bafef0"`
}

type CreatedResponse struct {
	SubID string `json:"sub_id" example:"d79c4c83-b0e4-4cc7-a6b1-3f2c5b8c9b76"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost" example:"2997"`
}

type ListSubscriptionsResponse struct {
	Subscriptions []Subscription `json:"subscriptions"`
	Count         int            `json:"count" example:"1"`
}

type InvalidIDErrorResponse struct {
	Error string `json:"error" example:"Invalid ID format"`
}

type NotFoundErrorResponse struct {
	Error string `json:"error" example:"Subscription not found"`
}

type InvalidRequestBodyErrorResponse struct {
	Error string `json:"error" example:"Invalid request body"`
}

type InternalServerErrorResponse struct {
	Error string `json:"error" example:"Failed to process request"`
}

type PeriodStartRequiredErrorResponse struct {
	Error string `json:"error" example:"period_start are required"`
}

type InvalidUserIDErrorResponse struct {
	Error string `json:"error" example:"Invalid user_id format"`
}

type FailedToCalculateErrorResponse struct {
	Error string `json:"error" example:"Failed to calculate total cost"`
}

type FailedToGetSubErrorResponse struct {
	Error string `json:"error" example:"Failed to get subscription"`
}

type FailedToCreateErrorResponse struct {
	Error string `json:"error" example:"Failed to create subscription"`
}

type FailedToUpdateSub struct {
	Error string `json:"error" example:"Failed to update subscription"`
}

type FailedToDeleteErrorResponse struct {
	Error string `json:"error" example:"Failed to delete subscription"`
}

type FailedToListSubsErrorResponse struct {
	Error string `json:"error" example:"Failed to list subscriptions"`
}
