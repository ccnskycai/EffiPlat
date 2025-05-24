package handler_test

import (
	"EffiPlat/backend/internal/model"
	"EffiPlat/backend/internal/models"
	"EffiPlat/backend/internal/router"
	"EffiPlat/backend/internal/service"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to get a pointer to a string value
func stringPtr(s string) *string {
	return &s
}

func TestBusinessHandler_CreateBusiness(t *testing.T) {
	components := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, components.Router, components.DB)

	t.Run("Successful creation", func(t *testing.T) {
		uniqueName := fmt.Sprintf("Test Business %d", time.Now().UnixNano())
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			Name:        stringPtr(uniqueName),
			Description: stringPtr("A test business description."),
			Owner:       stringPtr("test.owner@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, err := json.Marshal(input)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var responseBody struct {
			Code    int                       `json:"code"`
			Message string                    `json:"message"`
			Data    service.BusinessOutputDTO `json:"data"`
		}
		err = json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		assert.Equal(t, 0, responseBody.Code)
		assert.Equal(t, "Business created successfully", responseBody.Message)
		assert.NotZero(t, responseBody.Data.ID)
		assert.Equal(t, *input.Name, responseBody.Data.Name)
		assert.Equal(t, *input.Description, responseBody.Data.Description)
		assert.Equal(t, *input.Owner, responseBody.Data.Owner)
		assert.Equal(t, *input.Status, responseBody.Data.Status)
	})

	t.Run("Invalid request payload - Bind error", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Validation error - Missing name", func(t *testing.T) {
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			// Name is intentionally nil
			Description: stringPtr("A test business description."),
			Owner:       stringPtr("test.owner@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, err := json.Marshal(input)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Name already exists", func(t *testing.T) {
		existingName := fmt.Sprintf("Existing Business %d", time.Now().UnixNano())
		statusActive1 := model.BusinessStatusActive
		input1 := service.BusinessInputDTO{
			Name:        stringPtr(existingName),
			Description: stringPtr("First business."),
			Owner:       stringPtr("owner1@example.com"),
			Status:      &statusActive1,
		}
		payloadBytes1, _ := json.Marshal(input1)
		req1, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+token)
		w1 := httptest.NewRecorder()
		components.Router.ServeHTTP(w1, req1)
		require.Equal(t, http.StatusCreated, w1.Code)

		statusActive2 := model.BusinessStatusActive
		input2 := service.BusinessInputDTO{
			Name:        stringPtr(existingName),
			Description: stringPtr("Second business."),
			Owner:       stringPtr("owner2@example.com"),
			Status:      &statusActive2,
		}
		payloadBytes2, _ := json.Marshal(input2)
		req2, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		components.Router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
	})
}

func TestBusinessHandler_GetBusinessByID(t *testing.T) {
	components := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, components.Router, components.DB)

	createTestBusiness := func(nameSuffix string) service.BusinessOutputDTO {
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			Name:        stringPtr(fmt.Sprintf("Test Get Business %s %d", nameSuffix, time.Now().UnixNano())),
			Description: stringPtr("Description for get test."),
			Owner:       stringPtr("get.owner@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Helper failed to create business for GetByID test")
		var createdBusinessResponse struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &createdBusinessResponse)
		require.NoError(t, err)
		return createdBusinessResponse.Data
	}

	t.Run("Successful get", func(t *testing.T) {
		createdBiz := createTestBusiness("Success")

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/businesses/%d", createdBiz.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		assert.Equal(t, createdBiz.ID, responseBody.Data.ID)
		assert.Equal(t, createdBiz.Name, responseBody.Data.Name)
		assert.Equal(t, createdBiz.Description, responseBody.Data.Description)
		assert.Equal(t, createdBiz.Owner, responseBody.Data.Owner)
		assert.Equal(t, createdBiz.Status, responseBody.Data.Status)
	})

	t.Run("Not Found", func(t *testing.T) {
		nonExistentID := uint(999999)
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/businesses/%d", nonExistentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBusinessHandler_ListBusinesses(t *testing.T) {
	components := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, components.Router, components.DB)

	setupDataForListTest := func(prefix string, count int, statusToUse model.BusinessStatusType) []service.BusinessOutputDTO {
		createdBusinesses := make([]service.BusinessOutputDTO, count)
		for i := 0; i < count; i++ {
			currentStatus := statusToUse
			input := service.BusinessInputDTO{
				Name:        stringPtr(fmt.Sprintf("%s Test List Business %d %d", prefix, i, time.Now().UnixNano())),
				Description: stringPtr(fmt.Sprintf("Description for list test %d", i)),
				Owner:       stringPtr(fmt.Sprintf("list.owner%d@example.com", i)),
				Status:      &currentStatus,
			}
			payloadBytes, err := json.Marshal(input)
			require.NoError(t, err)

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			components.Router.ServeHTTP(w, req)

			require.Equal(t, http.StatusCreated, w.Code, "Helper setupDataForListTest failed to create business: %s", w.Body.String())
			var createdBusinessResponse struct {
				Data service.BusinessOutputDTO `json:"data"`
			}
			err = json.Unmarshal(w.Body.Bytes(), &createdBusinessResponse)
			require.NoError(t, err)
			createdBusinesses[i] = createdBusinessResponse.Data
		}
		return createdBusinesses
	}

	_ = setupDataForListTest("InitialList", 5, model.BusinessStatusActive)
	_ = setupDataForListTest("InactiveList", 3, model.BusinessStatusInactive)

	t.Run("List without filters - default pagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		assert.Equal(t, 0, responseBody.Code)

		var paginatedData models.PaginatedData
		err = json.Unmarshal(responseBody.Data, &paginatedData)
		require.NoError(t, err)

		var businesses []service.BusinessOutputDTO
		itemsBytes, _ := json.Marshal(paginatedData.Items)
		err = json.Unmarshal(itemsBytes, &businesses)
		require.NoError(t, err)

		assert.True(t, len(businesses) <= 10)
		assert.True(t, paginatedData.Total >= 8) // 5 active + 3 inactive
	})

	t.Run("List with pagination", func(t *testing.T) {
		setupDataForListTest("PaginationList", 15, model.BusinessStatusActive) // Create more businesses to ensure pagination is testable

		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses?page=2&pageSize=5", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var responseBody struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		var paginatedData models.PaginatedData
		err = json.Unmarshal(responseBody.Data, &paginatedData)
		require.NoError(t, err)

		var businesses []service.BusinessOutputDTO
		itemsBytes, _ := json.Marshal(paginatedData.Items)
		err = json.Unmarshal(itemsBytes, &businesses)
		require.NoError(t, err)

		assert.Equal(t, 2, paginatedData.Page)
		assert.Equal(t, 5, paginatedData.PageSize)
		assert.True(t, len(businesses) <= 5)
		assert.True(t, paginatedData.Total >= 23)
	})

	t.Run("List with name filter", func(t *testing.T) {
		uniqueFilterName := fmt.Sprintf("Filterable Name %d", time.Now().UnixNano())
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			Name:        stringPtr(uniqueFilterName),
			Description: stringPtr("Filter test desc"),
			Owner:       stringPtr("filter.owner@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, _ := json.Marshal(input)
		reqPost, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		reqPost.Header.Set("Content-Type", "application/json")
		reqPost.Header.Set("Authorization", "Bearer "+token)
		wPost := httptest.NewRecorder()
		components.Router.ServeHTTP(wPost, reqPost)
		require.Equal(t, http.StatusCreated, wPost.Code)

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/businesses?name=%s", uniqueFilterName), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var responseBody struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		var paginatedData models.PaginatedData
		err = json.Unmarshal(responseBody.Data, &paginatedData)
		require.NoError(t, err)

		var businesses []service.BusinessOutputDTO
		itemsBytes, _ := json.Marshal(paginatedData.Items)
		err = json.Unmarshal(itemsBytes, &businesses)
		require.NoError(t, err)

		assert.Equal(t, int64(1), paginatedData.Total)
		require.Len(t, businesses, 1)
		assert.Equal(t, *input.Name, businesses[0].Name)
	})

	t.Run("List with status filter", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses?status=active", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var responseBody struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		var paginatedData models.PaginatedData
		err = json.Unmarshal(responseBody.Data, &paginatedData)
		require.NoError(t, err)

		var businesses []service.BusinessOutputDTO
		itemsBytes, _ := json.Marshal(paginatedData.Items)
		err = json.Unmarshal(itemsBytes, &businesses)
		require.NoError(t, err)

		assert.True(t, paginatedData.Total > 0)
		for _, biz := range businesses {
			assert.Equal(t, model.BusinessStatusActive, biz.Status)
		}
	})

	t.Run("List with sorting by name ascending", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses?sortBy=name&order=asc&pageSize=50", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		var responseBody struct {
			Code    int             `json:"code"`
			Message string          `json:"message"`
			Data    json.RawMessage `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		var paginatedData models.PaginatedData
		err = json.Unmarshal(responseBody.Data, &paginatedData)
		require.NoError(t, err)

		var businesses []service.BusinessOutputDTO
		itemsBytes, _ := json.Marshal(paginatedData.Items)
		err = json.Unmarshal(itemsBytes, &businesses)
		require.NoError(t, err)

		require.True(t, len(businesses) > 1, "Need at least two items to check sort order")
		for i := 0; i < len(businesses)-1; i++ {
			assert.True(t, businesses[i].Name <= businesses[i+1].Name, "Businesses should be sorted by name ascending")
		}
	})

	t.Run("List with invalid sort order parameter", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/businesses?order=invalid_order", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestBusinessHandler_UpdateBusiness(t *testing.T) {
	components := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, components.Router, components.DB)

	createTestBusinessForUpdate := func(nameSuffix string) service.BusinessOutputDTO {
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			Name:        stringPtr(fmt.Sprintf("Update Test Business %s %d", nameSuffix, time.Now().UnixNano())),
			Description: stringPtr("Initial Description for update."),
			Owner:       stringPtr("initial.owner.update@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Helper failed to create business for update test")
		var createdBusinessResponse struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &createdBusinessResponse)
		require.NoError(t, err)
		return createdBusinessResponse.Data
	}

	t.Run("Successful update", func(t *testing.T) {
		bizToUpdate := createTestBusinessForUpdate("FullSuccess")

		updatedName := fmt.Sprintf("Updated Business Full %d", time.Now().UnixNano())
		updatedStatus := model.BusinessStatusInactive
		updatePayload := service.BusinessInputDTO{
			Name:        stringPtr(updatedName),
			Description: stringPtr("Updated Description."),
			Owner:       stringPtr("updated.owner@example.com"),
			Status:      &updatedStatus,
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/businesses/%d", bizToUpdate.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var responseBody struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		assert.Equal(t, bizToUpdate.ID, responseBody.Data.ID)
		assert.Equal(t, *updatePayload.Name, responseBody.Data.Name)
		assert.Equal(t, *updatePayload.Description, responseBody.Data.Description)
		assert.Equal(t, *updatePayload.Owner, responseBody.Data.Owner)
		assert.Equal(t, *updatePayload.Status, responseBody.Data.Status)
	})

	t.Run("Update with partial data - only name", func(t *testing.T) {
		bizToUpdate := createTestBusinessForUpdate("PartialName")
		originalDesc := bizToUpdate.Description
		originalOwner := bizToUpdate.Owner
		originalStatus := bizToUpdate.Status

		updatedNameOnly := fmt.Sprintf("Updated Name Only %d", time.Now().UnixNano())
		updatePayload := service.BusinessInputDTO{
			Name: stringPtr(updatedNameOnly),
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/businesses/%d", bizToUpdate.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, "Update request failed: %s", w.Body.String())

		var responseBody struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		assert.Equal(t, bizToUpdate.ID, responseBody.Data.ID)
		assert.Equal(t, *updatePayload.Name, responseBody.Data.Name)
		assert.Equal(t, originalDesc, responseBody.Data.Description, "Description should not change")
		assert.Equal(t, originalOwner, responseBody.Data.Owner, "Owner should not change")
		assert.Equal(t, originalStatus, responseBody.Data.Status, "Status should not change")
	})

	t.Run("Record not found", func(t *testing.T) {
		updatePayloadNonExistentID := service.BusinessInputDTO{
			Name: stringPtr("Any Name For Non Existent ID Test"),
		}
		payloadBytesNonExistentID, err := json.Marshal(updatePayloadNonExistentID)
		require.NoError(t, err)

		reqNonExistentID, _ := http.NewRequest(http.MethodPut, "/api/v1/businesses/999999", bytes.NewBuffer(payloadBytesNonExistentID))
		reqNonExistentID.Header.Set("Content-Type", "application/json")
		reqNonExistentID.Header.Set("Authorization", "Bearer "+token)
		wNonExistentID := httptest.NewRecorder()
		components.Router.ServeHTTP(wNonExistentID, reqNonExistentID)
		assert.Equal(t, http.StatusNotFound, wNonExistentID.Code)

		var errorRsp models.ErrorResponse
		err = json.Unmarshal(wNonExistentID.Body.Bytes(), &errorRsp)
		require.NoError(t, err) // It's okay if unmarshal fails for non-JSON error, but good to check if it's structured
		assert.Contains(t, errorRsp.Message, "not found")
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		updatePayloadInvalidID := service.BusinessInputDTO{
			Name: stringPtr("Any Name For Invalid ID Test"),
		}
		payloadBytesInvalidID, err := json.Marshal(updatePayloadInvalidID)
		require.NoError(t, err)

		reqInvalidID, _ := http.NewRequest(http.MethodPut, "/api/v1/businesses/invalidID", bytes.NewBuffer(payloadBytesInvalidID))
		reqInvalidID.Header.Set("Content-Type", "application/json")
		reqInvalidID.Header.Set("Authorization", "Bearer "+token)
		wInvalidID := httptest.NewRecorder()
		components.Router.ServeHTTP(wInvalidID, reqInvalidID)
		assert.Equal(t, http.StatusBadRequest, wInvalidID.Code)
	})

	t.Run("Validation error - Name too short", func(t *testing.T) {
		bizToUpdate := createTestBusinessForUpdate("ValidationShortName")
		statusActive := model.BusinessStatusActive
		updateInput := service.BusinessInputDTO{
			Name:   stringPtr("a"),
			Status: &statusActive,
		}
		payloadBytes, _ := json.Marshal(updateInput)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/businesses/%d", bizToUpdate.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Name conflict on update", func(t *testing.T) {
		biz1 := createTestBusinessForUpdate("Conflict1")
		biz2 := createTestBusinessForUpdate("Conflict2")

		statusActive := model.BusinessStatusActive
		updatePayload := service.BusinessInputDTO{
			Name:   stringPtr(biz1.Name),
			Status: &statusActive,
		}
		payloadBytes, _ := json.Marshal(updatePayload)

		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/businesses/%d", biz2.ID), bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestBusinessHandler_DeleteBusiness(t *testing.T) {
	components := router.SetupTestApp(t)
	token := router.GetAuthTokenForTest(t, components.Router, components.DB)

	createTestBusinessForDelete := func(nameSuffix string) uint {
		statusActive := model.BusinessStatusActive
		input := service.BusinessInputDTO{
			Name:        stringPtr(fmt.Sprintf("Delete Test Business %s %d", nameSuffix, time.Now().UnixNano())),
			Description: stringPtr("Description for delete test."),
			Owner:       stringPtr("delete.owner@example.com"),
			Status:      &statusActive,
		}
		payloadBytes, _ := json.Marshal(input)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/businesses", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Helper failed to create business for delete test")
		var createdBusinessResponse struct {
			Data service.BusinessOutputDTO `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &createdBusinessResponse)
		require.NoError(t, err)
		return createdBusinessResponse.Data.ID
	}

	t.Run("Successful delete", func(t *testing.T) {
		bizIDToDelete := createTestBusinessForDelete("Success")

		reqDelete, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/businesses/%d", bizIDToDelete), nil)
		reqDelete.Header.Set("Authorization", "Bearer "+token)
		wDelete := httptest.NewRecorder()
		components.Router.ServeHTTP(wDelete, reqDelete)
		assert.Equal(t, http.StatusNoContent, wDelete.Code)

		reqGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/businesses/%d", bizIDToDelete), nil)
		reqGet.Header.Set("Authorization", "Bearer "+token)
		wGet := httptest.NewRecorder()
		components.Router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusNotFound, wGet.Code)
	})

	t.Run("Not Found", func(t *testing.T) {
		nonExistentID := uint(999997)
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/businesses/%d", nonExistentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Invalid ID format", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/businesses/invalid-id-for-delete", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
