package service

import (
	"EffiPlat/backend/internal/model"
	"context"
)

// BugService defines the interface for bug business logic.
type BugService interface {
	CreateBug(ctx context.Context, req *model.CreateBugRequest) (*model.BugResponse, error)
	GetBugByID(ctx context.Context, id uint) (*model.BugResponse, error)
	UpdateBug(ctx context.Context, id uint, req *model.UpdateBugRequest) (*model.BugResponse, error)
	DeleteBug(ctx context.Context, id uint) error
	ListBugs(ctx context.Context, params *model.BugListParams) ([]*model.BugResponse, int64, error)
}
