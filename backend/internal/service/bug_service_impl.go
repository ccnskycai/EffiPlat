package service

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"context"
	"errors"
	// For errors.Is(err, gorm.ErrRecordNotFound)
)

// bugServiceImpl implements the BugService interface.
type bugServiceImpl struct {
	bugRepo repository.BugRepository
	// userRepo repository.UserRepository // Example: if reporter/assignee validation is needed
}

// NewBugService creates a new instance of bugServiceImpl.
func NewBugService(bugRepo repository.BugRepository) BugService {
	return &bugServiceImpl{bugRepo: bugRepo}
}

// CreateBug creates a new bug.
func (s *bugServiceImpl) CreateBug(ctx context.Context, req *model.CreateBugRequest) (*model.BugResponse, error) {
	// TODO: Add validation for ReporterID and AssigneeID if they should exist in the User table.
	// For example, check if s.userRepo.GetByID(ctx, *req.ReporterID) returns a user.

	bug := &model.Bug{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		ReporterID:  req.ReporterID,
		AssigneeID:  req.AssigneeID,
		// ProjectID:   req.ProjectID,
		// EnvironmentID: req.EnvironmentID,
	}

	if err := s.bugRepo.Create(ctx, bug); err != nil {
		return nil, err
	}

	bugResponseValue := bug.ToBugResponse()
	return &bugResponseValue, nil
}

// GetBugByID retrieves a bug by its ID.
func (s *bugServiceImpl) GetBugByID(ctx context.Context, id uint) (*model.BugResponse, error) {
	bug, err := s.bugRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBugNotFound) { // Check for custom repository error
			return nil, err // Or a service-specific error like ErrBugNotFound
		}
		return nil, err
	}
	bugResponseValue := bug.ToBugResponse()
	return &bugResponseValue, nil
}

// UpdateBug updates an existing bug.
func (s *bugServiceImpl) UpdateBug(ctx context.Context, id uint, req *model.UpdateBugRequest) (*model.BugResponse, error) {
	bug, err := s.bugRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrBugNotFound) {
			return nil, err // Or a service-specific error
		}
		return nil, err
	}

	// Apply updates from request
	if req.Title != nil {
		bug.Title = *req.Title
	}
	if req.Description != nil {
		bug.Description = *req.Description
	}
	if req.Status != nil {
		bug.Status = *req.Status
	}
	if req.Priority != nil {
		bug.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		// TODO: Validate AssigneeID if necessary
		bug.AssigneeID = req.AssigneeID
	}
	// if req.ProjectID != nil { bug.ProjectID = req.ProjectID }
	// if req.EnvironmentID != nil { bug.EnvironmentID = req.EnvironmentID }

	if err := s.bugRepo.Update(ctx, bug); err != nil {
		// Handle potential concurrency issues or other update errors
		return nil, err
	}

	bugResponseValue := bug.ToBugResponse()
	return &bugResponseValue, nil
}

// DeleteBug deletes a bug by its ID.
func (s *bugServiceImpl) DeleteBug(ctx context.Context, id uint) error {
	if err := s.bugRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrBugNotFound) {
			return err // Or a service-specific error
		}
		return err
	}
	return nil
}

// ListBugs retrieves a list of bugs with pagination and filters.
func (s *bugServiceImpl) ListBugs(ctx context.Context, params *model.BugListParams) ([]*model.BugResponse, int64, error) {
	bugs, totalCount, err := s.bugRepo.List(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	bugResponses := make([]*model.BugResponse, len(bugs))
	for i, b := range bugs {
		resp := b.ToBugResponse() // Get the value type response
		bugResponses[i] = &resp   // Store its address in the slice of pointers
	}

	return bugResponses, totalCount, nil
}
