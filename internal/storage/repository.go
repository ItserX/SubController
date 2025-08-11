package storage

import (
	"github.com/google/uuid"

	"github.com/ItserX/rest/internal/types"
)

type PostRepository interface {
	Create(sub types.Subscription) (uuid.UUID, error)
	Get(id uuid.UUID) (*types.Subscription, error)
	Update(id uuid.UUID, sub types.Subscription) error
	Delete(id uuid.UUID) error
	List() ([]types.Subscription, error)
	GetTotalCost(id uuid.UUID, serviceName, periodStart, periodEnd string) (int, error)
}
