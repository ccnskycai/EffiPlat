package repository

import (
	"EffiPlat/backend/internal/model"
	"context"
)

// BugRepository defines the interface for bug data operations.
type BugRepository interface {
	Create(ctx context.Context, bug *model.Bug) error
	GetByID(ctx context.Context, id uint) (*model.Bug, error)
	Update(ctx context.Context, bug *model.Bug) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, params *model.BugListParams) ([]*model.Bug, int64, error) // Returns bugs and total count
	
	// Additional query methods
	CountBugsByAssigneeID(ctx context.Context, assigneeID uint) (int64, error)
	CountBugsByEnvironmentID(ctx context.Context, environmentID uint) (int64, error)
	GetBugsByStatus(ctx context.Context, status model.BugStatusType, params *model.BugListParams) ([]*model.Bug, int64, error)
}
