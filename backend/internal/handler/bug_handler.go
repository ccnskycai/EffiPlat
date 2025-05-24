package handler

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/repository"
	"EffiPlat/backend/internal/service"
	"EffiPlat/backend/internal/utils"
	"net/http"
	"strconv"

	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BugHandler handles HTTP requests for bugs.
type BugHandler struct {
	bugService service.BugService
	validate   *validator.Validate
}

// NewBugHandler creates a new BugHandler instance.
func NewBugHandler(bugService service.BugService) *BugHandler {
	return &BugHandler{
		bugService: bugService,
		validate:   validator.New(),
	}
}

// CreateBug godoc
// @Summary Create a new bug
// @Description Create a new bug with the input payload
// @Tags bugs
// @Accept  json
// @Produce  json
// @Param   bug_request body model.CreateBugRequest true "Create Bug Request"
// @Success 201 {object} model.BugResponse
// @Failure 400 {object} model.ErrorResponse "Invalid input"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /bugs [post]
func (h *BugHandler) CreateBug(c *gin.Context) {
	var req model.CreateBugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	bugResp, err := h.bugService.CreateBug(c.Request.Context(), &req)
	if err != nil {
		// TODO: Differentiate between client errors (e.g., duplicate) and server errors
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to create bug: "+err.Error())
		return
	}

	c.JSON(http.StatusCreated, bugResp)
}

// GetBugByID godoc
// @Summary Get a bug by ID
// @Description Get details of a bug by its ID
// @Tags bugs
// @Produce json
// @Param id path int true "Bug ID"
// @Success 200 {object} model.BugResponse
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Bug not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /bugs/{id} [get]
func (h *BugHandler) GetBugByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid bug ID format.")
		return
	}

	bugResp, err := h.bugService.GetBugByID(c.Request.Context(), uint(id))
	if err != nil {
		// Check if the error is a known 'not found' error from the service/repository
		if errors.Is(err, repository.ErrBugNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Bug not found.")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get bug: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, bugResp)
}

// UpdateBug godoc
// @Summary Update an existing bug
// @Description Update details of an existing bug by its ID
// @Tags bugs
// @Accept  json
// @Produce  json
// @Param id path int true "Bug ID"
// @Param   bug_request body model.UpdateBugRequest true "Update Bug Request"
// @Success 200 {object} model.BugResponse
// @Failure 400 {object} model.ErrorResponse "Invalid ID format or input"
// @Failure 404 {object} model.ErrorResponse "Bug not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /bugs/{id} [put]
func (h *BugHandler) UpdateBug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid bug ID format.")
		return
	}

	var req model.UpdateBugRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	if err := h.validate.Struct(req); err != nil {
		// Note: Validator only validates non-nil fields in UpdateBugRequest due to pointer types.
		// Custom validation might be needed if at least one field must be present.
		utils.SendErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	bugResp, err := h.bugService.UpdateBug(c.Request.Context(), uint(id), &req)
	if err != nil {
		if errors.Is(err, repository.ErrBugNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Bug not found.")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update bug: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, bugResp)
}

// DeleteBug godoc
// @Summary Delete a bug by ID
// @Description Delete a bug by its ID
// @Tags bugs
// @Produce  json
// @Param id path int true "Bug ID"
// @Success 204 "No Content"
// @Failure 400 {object} model.ErrorResponse "Invalid ID format"
// @Failure 404 {object} model.ErrorResponse "Bug not found"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /bugs/{id} [delete]
func (h *BugHandler) DeleteBug(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid bug ID format.")
		return
	}

	if err := h.bugService.DeleteBug(c.Request.Context(), uint(id)); err != nil {
		if errors.Is(err, repository.ErrBugNotFound) {
			utils.SendErrorResponse(c, http.StatusNotFound, "Bug not found.")
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to delete bug: "+err.Error())
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ListBugs godoc
// @Summary List bugs
// @Description Get a list of bugs with optional filters and pagination
// @Tags bugs
// @Produce  json
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Param title query string false "Filter by title (substring match)"
// @Param status query string false "Filter by status (e.g., OPEN, RESOLVED)" Enums(OPEN, IN_PROGRESS, RESOLVED, CLOSED, REOPENED)
// @Param priority query string false "Filter by priority (e.g., LOW, HIGH)" Enums(LOW, MEDIUM, HIGH, URGENT)
// @Param assigneeId query int false "Filter by assignee ID"
// @Param reporterId query int false "Filter by reporter ID"
// @Success 200 {object} model.SuccessResponse{data=model.PaginatedData{items=[]model.BugResponse}} "List of bugs"
// @Failure 400 {object} model.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} model.ErrorResponse "Internal server error"
// @Router /bugs [get]
func (h *BugHandler) ListBugs(c *gin.Context) {
	var params model.BugListParams
	// Bind query parameters
	if err := c.ShouldBindQuery(&params); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid query parameters: "+err.Error())
		return
	}

	// Set defaults if not provided or invalid (validator might handle some of this based on struct tags)
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	// Max PageSize can be enforced here too
	// e.g., if params.PageSize > 100 { params.PageSize = 100 }

	bugs, totalCount, err := h.bugService.ListBugs(c.Request.Context(), &params)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list bugs: "+err.Error())
		return
	}

	// Use the SendPaginatedSuccessResponse helper from utils package
	utils.SendPaginatedSuccessResponse(c, http.StatusOK, "Bugs listed successfully", bugs, params.Page, params.PageSize, totalCount)
}
