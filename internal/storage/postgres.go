package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ItserX/rest/internal/logger"
	"github.com/ItserX/rest/internal/types"
)

type PostgresRepository struct {
	db *sql.DB
}

var ErrNotFound = errors.New("subscription not found")

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) Create(sub types.Subscription) (uuid.UUID, error) {
	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		logger.Logger.Errorw("Invalid start_date format",
			"error", err,
			"start_date", sub.StartDate,
		)
		return uuid.Nil, fmt.Errorf("invalid start_date format, expected MM-YYYY: %w", err)
	}

	var endDate *time.Time
	if sub.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", sub.EndDate)
		if err != nil {
			logger.Logger.Errorw("Invalid end_date format",
				"error", err,
				"end_date", sub.EndDate,
			)
			return uuid.Nil, fmt.Errorf("invalid end_date format, expected MM-YYYY: %w", err)
		}
		endDate = &parsedEndDate
	}

	query := `
        INSERT INTO subscriptions (sub_id, user_id, service_name, price, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

	subID := uuid.New()
	logger.Logger.Debugw("Creating new subscription",
		"subscriptionID", subID,
		"userID", sub.UserID,
		"serviceName", sub.ServiceName,
	)

	_, err = r.db.Exec(
		query,
		subID,
		sub.UserID,
		sub.ServiceName,
		sub.Price,
		startDate,
		endDate,
	)

	if err != nil {
		logger.Logger.Errorw("Failed to create subscription",
			"error", err,
			"subscriptionID", subID,
		)
		return uuid.Nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	logger.Logger.Infow("Successfully created subscription",
		"subscriptionID", subID,
	)
	return subID, nil
}

func (r *PostgresRepository) Get(id uuid.UUID) (*types.Subscription, error) {
	query := `
        SELECT sub_id, user_id, service_name, price, start_date, end_date
        FROM subscriptions
        WHERE sub_id = $1
    `

	logger.Logger.Debugw("Getting subscription",
		"subscriptionID", id,
	)

	var (
		dbSubID       uuid.UUID
		dbUserID      uuid.UUID
		dbServiceName string
		dbPrice       int
		dbStartDate   time.Time
		dbEndDate     sql.NullTime
	)

	err := r.db.QueryRow(query, id).Scan(
		&dbSubID,
		&dbUserID,
		&dbServiceName,
		&dbPrice,
		&dbStartDate,
		&dbEndDate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.Logger.Warnw("Subscription not found",
				"subscriptionID", id,
			)
			return nil, ErrNotFound
		}
		logger.Logger.Errorw("Failed to get subscription",
			"error", err,
			"subscriptionID", id,
		)
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	sub := &types.Subscription{
		ServiceName: dbServiceName,
		Price:       dbPrice,
		UserID:      dbUserID,
		StartDate:   dbStartDate.Format("01-2006"),
	}

	if dbEndDate.Valid {
		sub.EndDate = dbEndDate.Time.Format("01-2006")
	}

	logger.Logger.Debugw("Successfully retrieved subscription",
		"subscriptionID", id,
	)
	return sub, nil
}

func (r *PostgresRepository) Update(id uuid.UUID, sub types.Subscription) error {
	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		logger.Logger.Errorw("Invalid start_date format",
			"error", err,
			"start_date", sub.StartDate,
		)
		return fmt.Errorf("invalid start_date format, expected MM-YYYY: %w", err)
	}

	var endDate *time.Time
	if sub.EndDate != "" {
		parsedEndDate, err := time.Parse("01-2006", sub.EndDate)
		if err != nil {
			logger.Logger.Errorw("Invalid end_date format",
				"error", err,
				"end_date", sub.EndDate,
			)
			return fmt.Errorf("invalid end_date format, expected MM-YYYY: %w", err)
		}
		endDate = &parsedEndDate
	}

	query := `
        UPDATE subscriptions
        SET 
            service_name = $1,
            price = $2,
            start_date = $3,
            end_date = $4
        WHERE sub_id = $5
    `

	logger.Logger.Debugw("Updating subscription",
		"subscriptionID", id,
		"updateData", sub,
	)

	var result sql.Result
	if endDate != nil {
		result, err = r.db.Exec(
			query,
			sub.ServiceName,
			sub.Price,
			startDate,
			endDate,
			id,
		)
	} else {
		result, err = r.db.Exec(
			query,
			sub.ServiceName,
			sub.Price,
			startDate,
			nil,
			id,
		)
	}

	if err != nil {
		logger.Logger.Errorw("Failed to update subscription",
			"error", err,
			"subscriptionID", id,
		)
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Logger.Errorw("Failed to check rows affected",
			"error", err,
			"subscriptionID", id,
		)
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Logger.Warnw("Subscription not found for update",
			"subscriptionID", id,
		)
		return ErrNotFound
	}

	logger.Logger.Infow("Successfully updated subscription",
		"subscriptionID", id,
		"rowsAffected", rowsAffected,
	)
	return nil
}

func (r *PostgresRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE sub_id = $1`

	logger.Logger.Debugw("Deleting subscription",
		"subscriptionID", id,
	)

	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.Logger.Errorw("Failed to delete subscription",
			"error", err,
			"subscriptionID", id,
		)
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Logger.Errorw("Failed to check rows affected",
			"error", err,
			"subscriptionID", id,
		)
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		logger.Logger.Warnw("Subscription not found for deletion",
			"subscriptionID", id,
		)
		return ErrNotFound
	}

	logger.Logger.Infow("Successfully deleted subscription",
		"subscriptionID", id,
		"rowsAffected", rowsAffected,
	)
	return nil
}

func (r *PostgresRepository) GetTotalCost(userID uuid.UUID, serviceName, periodStart, periodEnd string) (int, error) {
	startTime, err := time.Parse("01-2006", periodStart)
	if err != nil {
		logger.Logger.Errorw("Invalid period_start format",
			"error", err,
			"period_start", periodStart,
		)
		return 0, fmt.Errorf("invalid period_start format: %w", err)
	}

	endTime, err := time.Parse("01-2006", periodEnd)
	if err != nil {
		logger.Logger.Errorw("Invalid period_end format",
			"error", err,
			"period_end", periodEnd,
		)
		return 0, fmt.Errorf("invalid period_end format: %w", err)
	}

	query := `
        SELECT COALESCE(SUM(price), 0)
        FROM subscriptions
        WHERE 
            start_date <= $1 AND 
            (end_date >= $2 OR end_date IS NULL)
    `
	args := []interface{}{endTime, startTime}

	if userID != uuid.Nil {
		query += " AND user_id = $3"
		args = append(args, userID)
	}
	if serviceName != "" {
		query += " AND service_name = $4"
		args = append(args, serviceName)
	}

	logger.Logger.Debugw("Calculating total cost",
		"userID", userID,
		"serviceName", serviceName,
		"periodStart", periodStart,
		"periodEnd", periodEnd,
	)

	var total int
	err = r.db.QueryRow(query, args...).Scan(&total)
	if err != nil {
		logger.Logger.Errorw("Failed to calculate total cost",
			"error", err,
		)
		return 0, fmt.Errorf("failed to calculate total cost: %w", err)
	}

	logger.Logger.Infow("Successfully calculated total cost",
		"total", total,
	)
	return total, nil
}

func (r *PostgresRepository) List() ([]types.Subscription, error) {
	query := `
        SELECT sub_id, user_id, service_name, price, start_date, end_date
        FROM subscriptions
        ORDER BY start_date DESC
    `

	logger.Logger.Debugw("Listing all subscriptions")

	rows, err := r.db.Query(query)
	if err != nil {
		logger.Logger.Errorw("Failed to list subscriptions",
			"error", err,
		)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []types.Subscription
	for rows.Next() {
		var (
			dbSubID       uuid.UUID
			dbUserID      uuid.UUID
			dbServiceName string
			dbPrice       int
			dbStartDate   time.Time
			dbEndDate     sql.NullTime
		)

		if err := rows.Scan(
			&dbSubID,
			&dbUserID,
			&dbServiceName,
			&dbPrice,
			&dbStartDate,
			&dbEndDate,
		); err != nil {
			logger.Logger.Errorw("Failed to scan subscription row",
				"error", err,
			)
			return nil, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		sub := types.Subscription{
			ServiceName: dbServiceName,
			Price:       dbPrice,
			UserID:      dbUserID,
			StartDate:   dbStartDate.Format("01-2006"),
		}

		if dbEndDate.Valid {
			sub.EndDate = dbEndDate.Time.Format("01-2006")
		}

		subscriptions = append(subscriptions, sub)
	}

	if err := rows.Err(); err != nil {
		logger.Logger.Errorw("Error after scanning rows",
			"error", err,
		)
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}

	logger.Logger.Infow("Successfully listed all subscriptions",
		"count", len(subscriptions),
	)
	return subscriptions, nil
}
