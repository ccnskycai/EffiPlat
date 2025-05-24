package router

import (
	"EffiPlat/backend/internal/model"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEnvironment(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	t.Run("Success", func(t *testing.T) {
		createReq := model.CreateEnvironmentRequest{
			Name:        "Production Env",
			Description: "Main production environment",
			Slug:        "prod-env",
		}
		payload, _ := json.Marshal(createReq)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var resp struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, createReq.Name, resp.Data.Name)
		assert.Equal(t, createReq.Slug, resp.Data.Slug)
		assert.NotEmpty(t, resp.Data.ID)
	})

	t.Run("Failure_MissingName", func(t *testing.T) {
		createReq := model.CreateEnvironmentRequest{
			Description: "Missing name test",
			Slug:        "missing-name-env",
		}
		payload, _ := json.Marshal(createReq)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure_MissingSlug", func(t *testing.T) {
		createReq := model.CreateEnvironmentRequest{
			Name:        "Missing Slug Env",
			Description: "Missing slug test",
		}
		payload, _ := json.Marshal(createReq)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Failure_DuplicateSlug", func(t *testing.T) {
		// First, create an environment
		firstCreateReq := model.CreateEnvironmentRequest{
			Name:        "Unique Name For Slug Test",
			Description: "Testing duplicate slug",
			Slug:        "duplicate-slug-test",
		}
		payload1, _ := json.Marshal(firstCreateReq)
		req1, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+token)
		w1 := httptest.NewRecorder()
		components.Router.ServeHTTP(w1, req1)
		require.Equal(t, http.StatusCreated, w1.Code) // Ensure first one is created

		// Attempt to create another with the same slug
		secondCreateReq := model.CreateEnvironmentRequest{
			Name:        "Another Name",
			Description: "Attempting duplicate slug",
			Slug:        "duplicate-slug-test", // Same slug
		}
		payload2, _ := json.Marshal(secondCreateReq)
		req2, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+token)
		w2 := httptest.NewRecorder()
		components.Router.ServeHTTP(w2, req2)

		assert.Equal(t, http.StatusConflict, w2.Code)
	})
}

func TestGetEnvironments(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	// Helper to create an environment for list testing
	createEnv := func(name, slug string) model.EnvironmentResponse {
		createReq := model.CreateEnvironmentRequest{Name: name, Description: name + " desc", Slug: slug}
		payload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create env for GetEnvironments test")
		var apiResp struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &apiResp)
		require.NoError(t, err, "Failed to unmarshal createEnv response in GetEnvironments")
		return apiResp.Data
	}

	env1 := createEnv("List Test Env 1", "list-test-env-1")
	_ = createEnv("List Test Env 2", "list-test-env-2-another") // Deliberately different for filtering
	_ = createEnv("Filter Env Alpha", "filter-slug-alpha")

	t.Run("Success_DefaultPagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/environments", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items      []model.EnvironmentResponse `json:"items"`
				Total      int64                        `json:"total"`
				Page       int                          `json:"page"`
				PageSize   int                          `json:"pageSize"`
				TotalPages int                          `json:"totalPages"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &listResp)
		require.NoError(t, err, "Failed to unmarshal response for DefaultPagination")
		assert.Equal(t, 0, listResp.Code)
		assert.True(t, len(listResp.Data.Items) >= 3, "Expected at least 3 items for DefaultPagination") 
		assert.True(t, listResp.Data.Total >= 3, "Expected total to be at least 3 for DefaultPagination")
		assert.Equal(t, 1, listResp.Data.Page, "DefaultPagination page check")
		assert.Equal(t, 10, listResp.Data.PageSize, "DefaultPagination pageSize check") // Default page size
	})

	t.Run("Success_CustomPagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/environments?page=1&pageSize=1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items      []model.EnvironmentResponse `json:"items"`
				Total      int64                        `json:"total"`
				Page       int                          `json:"page"`
				PageSize   int                          `json:"pageSize"`
				TotalPages int                          `json:"totalPages"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &listResp)
		require.NoError(t, err, "Failed to unmarshal response for CustomPagination")
		assert.Equal(t, 0, listResp.Code)
		assert.Len(t, listResp.Data.Items, 1, "CustomPagination items length check")
		assert.True(t, listResp.Data.Total >= 3, "CustomPagination total check")
		assert.Equal(t, 1, listResp.Data.Page, "CustomPagination page check")
		assert.Equal(t, 1, listResp.Data.PageSize, "CustomPagination pageSize check")
	})

	t.Run("Success_FilterByName", func(t *testing.T) {
		filterName := "List Test Env 1"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments?name=%s", filterName), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items      []model.EnvironmentResponse `json:"items"`
				Total      int64                        `json:"total"`
				Page       int                          `json:"page"`
				PageSize   int                          `json:"pageSize"`
				TotalPages int                          `json:"totalPages"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &listResp)
		require.NoError(t, err, "Failed to unmarshal response for FilterByName")
		assert.Equal(t, 0, listResp.Code)
		assert.True(t, len(listResp.Data.Items) > 0, "Expected at least one result for name filter")
		// Ensure the first created env is in the list. The loop for this is below.
		found := false
		for _, item := range listResp.Data.Items {
			if item.ID == env1.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "The specific environment created for the test should be found by name filter")
	})

	t.Run("Success_FilterBySlug", func(t *testing.T) {
		filterSlug := "list-test-env-1"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments?slug=%s", filterSlug), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var listResp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items      []model.EnvironmentResponse `json:"items"`
				Total      int64                        `json:"total"`
				Page       int                          `json:"page"`
				PageSize   int                          `json:"pageSize"`
				TotalPages int                          `json:"totalPages"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &listResp)
		require.NoError(t, err, "Failed to unmarshal response for FilterBySlug")
		assert.Equal(t, 0, listResp.Code)
		assert.True(t, len(listResp.Data.Items) > 0, "Expected at least one result for slug filter")
		// Ensure the first created env is in the list. The loop for this is below.
		found := false
		for _, item := range listResp.Data.Items {
			if item.ID == env1.ID {
				found = true
				break
			}
		}
		assert.True(t, found, "The specific environment created for the test should be found by slug filter")
	})
}

func TestGetEnvironmentByID(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	// Create an environment to fetch
	createReq := model.CreateEnvironmentRequest{Name: "GetByID Test", Slug: "get-by-id-test"}
	payload, _ := json.Marshal(createReq)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	wCreate := httptest.NewRecorder()
	components.Router.ServeHTTP(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)
	var apiRespSetupByID struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Data    model.EnvironmentResponse `json:"data"`
	}
	t.Logf("TestGetEnvironmentByID Setup Raw Response: %s", wCreate.Body.String())
	errSetupByID := json.Unmarshal(wCreate.Body.Bytes(), &apiRespSetupByID)
	require.NoError(t, errSetupByID, "Failed to unmarshal createdEnv response in TestGetEnvironmentByID setup")
	t.Logf("TestGetEnvironmentByID Setup Unmarshalled apiResp.Data: %+v", apiRespSetupByID.Data)
	createdEnv := apiRespSetupByID.Data
	require.NotEqual(t, uint(0), createdEnv.ID, "TestGetEnvironmentByID: createdEnv.ID should not be zero after creation")
	require.NotEmpty(t, createdEnv.Slug, "TestGetEnvironmentByID: createdEnv.Slug should not be empty after creation")

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments/%d", createdEnv.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respWrapperByID struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		errUnmarshalByID := json.Unmarshal(w.Body.Bytes(), &respWrapperByID)
		require.NoError(t, errUnmarshalByID, "Failed to unmarshal response in TestGetEnvironmentByID/Success")
		resp := respWrapperByID.Data
		assert.Equal(t, createdEnv.ID, resp.ID)
		assert.Equal(t, createdEnv.Name, resp.Name)
		assert.Equal(t, createdEnv.Slug, resp.Slug)
	})

	t.Run("Failure_NotFound", func(t *testing.T) {
		nonExistentID := 99999
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments/%d", nonExistentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Failure_InvalidID", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/environments/invalid-id", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetEnvironmentBySlug(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	// Create an environment to fetch
	createReq := model.CreateEnvironmentRequest{Name: "GetBySlug Test", Slug: "get-by-slug-test"}
	payload, _ := json.Marshal(createReq)
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer "+token)
	wCreate := httptest.NewRecorder()
	components.Router.ServeHTTP(wCreate, reqCreate)
	require.Equal(t, http.StatusCreated, wCreate.Code)
	var apiRespSetupBySlug struct {
		Code    int                        `json:"code"`
		Message string                     `json:"message"`
		Data    model.EnvironmentResponse `json:"data"`
	}
	t.Logf("TestGetEnvironmentBySlug Setup Raw Response: %s", wCreate.Body.String())
	errSetupBySlug := json.Unmarshal(wCreate.Body.Bytes(), &apiRespSetupBySlug)
	require.NoError(t, errSetupBySlug, "Failed to unmarshal createdEnv response in TestGetEnvironmentBySlug setup")
	t.Logf("TestGetEnvironmentBySlug Setup Unmarshalled apiResp.Data: %+v", apiRespSetupBySlug.Data)
	createdEnv := apiRespSetupBySlug.Data
	require.NotEqual(t, uint(0), createdEnv.ID, "TestGetEnvironmentBySlug: createdEnv.ID should not be zero after creation")
	require.NotEmpty(t, createdEnv.Slug, "TestGetEnvironmentBySlug: createdEnv.Slug should not be empty after creation")

	t.Run("Success", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments/slug/%s", createdEnv.Slug), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respWrapperBySlug struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		errUnmarshalBySlug := json.Unmarshal(w.Body.Bytes(), &respWrapperBySlug)
		require.NoError(t, errUnmarshalBySlug, "Failed to unmarshal response in TestGetEnvironmentBySlug/Success")
		resp := respWrapperBySlug.Data
		assert.Equal(t, createdEnv.ID, resp.ID)
		assert.Equal(t, createdEnv.Name, resp.Name)
		assert.Equal(t, createdEnv.Slug, resp.Slug)
	})

	t.Run("Failure_NotFound", func(t *testing.T) {
		nonExistentSlug := "non-existent-slug-for-testing"
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments/slug/%s", nonExistentSlug), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestUpdateEnvironment(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	// Helper to create an environment for updating
	createInitialEnv := func(name, slug string) model.EnvironmentResponse {
		createReq := model.CreateEnvironmentRequest{Name: name, Description: name + " initial desc", Slug: slug}
		payload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create env for UpdateEnvironment test")
		var apiResp struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		t.Logf("createInitialEnv Raw Response for %s: %s", slug, w.Body.String())
		err := json.Unmarshal(w.Body.Bytes(), &apiResp)
		require.NoError(t, err, "Failed to unmarshal response in createInitialEnv helper for TestUpdateEnvironment")
		t.Logf("createInitialEnv Unmarshalled apiResp.Data for %s: %+v", slug, apiResp.Data)
		require.NotEqual(t, uint(0), apiResp.Data.ID, "createInitialEnv: apiResp.Data.ID should not be zero for slug %s", slug)
		require.NotEmpty(t, apiResp.Data.Slug, "createInitialEnv: apiResp.Data.Slug should not be empty for slug %s", slug)
		return apiResp.Data
	}

	initialEnv := createInitialEnv("Update Test Original", "update-test-original")
	// Create another env to test conflict scenarios
	conflictingEnv := createInitialEnv("Conflicting Name", "conflicting-slug")

	t.Run("Success_UpdateNameAndDescription", func(t *testing.T) {
		updateReq := model.UpdateEnvironmentRequest{
			Name:        model.StringPtr("Updated Name"),
			Description: model.StringPtr("Updated Description"),
			// Slug remains unchanged
		}
		payload, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", initialEnv.ID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respWrapperUpdateNameDesc struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		errUnmarshalUpdateNameDesc := json.Unmarshal(w.Body.Bytes(), &respWrapperUpdateNameDesc)
		require.NoError(t, errUnmarshalUpdateNameDesc, "Failed to unmarshal response in TestUpdateEnvironment/Success_UpdateNameAndDescription")
		resp := respWrapperUpdateNameDesc.Data
		assert.Equal(t, *updateReq.Name, resp.Name)
		assert.Equal(t, *updateReq.Description, resp.Description)
		assert.Equal(t, initialEnv.Slug, resp.Slug) // Slug should not have changed
	})

	t.Run("Success_UpdateSlug", func(t *testing.T) {
		// Create a fresh env for this specific slug update test to avoid conflicts from previous sub-tests
		slugUpdateEnv := createInitialEnv("Slug Update Test Original", "slug-update-original")
		updateReq := model.UpdateEnvironmentRequest{
			Slug: model.StringPtr("slug-update-new"),
		}
		payload, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", slugUpdateEnv.ID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var respWrapperUpdateSlug struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		errUnmarshalUpdateSlug := json.Unmarshal(w.Body.Bytes(), &respWrapperUpdateSlug)
		require.NoError(t, errUnmarshalUpdateSlug, "Failed to unmarshal response in TestUpdateEnvironment/Success_UpdateSlug")
		resp := respWrapperUpdateSlug.Data
		assert.Equal(t, *updateReq.Slug, resp.Slug)
		assert.Equal(t, slugUpdateEnv.Name, resp.Name) // Name should not have changed
	})

	t.Run("Failure_UpdateToExistingSlug", func(t *testing.T) {
		updateReq := model.UpdateEnvironmentRequest{
			Slug: model.StringPtr(conflictingEnv.Slug), // Try to update to conflictingEnv's slug
		}
		payload, _ := json.Marshal(updateReq)
		// Use initialEnv's ID for the update attempt
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", initialEnv.ID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("Failure_UpdateToExistingName", func(t *testing.T) {
		updateReq := model.UpdateEnvironmentRequest{
			Name: model.StringPtr(conflictingEnv.Name), // Try to update to conflictingEnv's name
		}
		payload, _ := json.Marshal(updateReq)
		// Use initialEnv's ID for the update attempt
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", initialEnv.ID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code) // Assuming service layer prevents duplicate names too
	})

	t.Run("Failure_NotFound", func(t *testing.T) {
		nonExistentID := 99999
		updateReq := model.UpdateEnvironmentRequest{Name: model.StringPtr("No Matter Name")}
		payload, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", nonExistentID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Failure_InvalidPayload", func(t *testing.T) {
		// e.g. sending a slug that's too short (if min length is defined in model binding)
		// For this test, we will send an empty name which should be caught by binding for non-partial update
		// or if your UpdateEnvironmentRequest has binding for name even if it's a pointer
		// Let's assume empty name is invalid for update. If only non-empty fields are updated, this might pass.
		// Better to test with specific validation like too short slug if such validation exists.
		// For simplicity, if *Name is empty string it might be rejected by `binding:"omitempty,min=2,max=100"` on model if applied to pointers.
		// Our current UpdateEnvironmentRequest doesn't have direct binding tags, relies on model for full updates.
		// Let's try with a name that's too short to trigger model validation if service re-validates.
		
		// Test with a name that would violate constraints if it were a Create request.
		// Service should ideally prevent setting an invalid name.
		shortName := "a"
		updateReq := model.UpdateEnvironmentRequest{Name: &shortName}
		payload, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/environments/%d", initialEnv.ID), bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		// This depends on whether service layer re-validates with model constraints.
		// If it does, and 'min=2' is on Environment.Name, this should be a BadRequest.
		// If it only checks for conflicts, this might pass then fail on DB if DB also has length constraints.
		// For now, assuming service layer validates against model-level constraints for fields being updated.
		assert.Equal(t, http.StatusBadRequest, w.Code, "Expected BadRequest for invalid name update (e.g., too short)")
	})
}

func TestDeleteEnvironment(t *testing.T) {
	components := SetupTestApp(t)
	token := GetAuthTokenForTest(t, components.Router, components.DB)

	// Helper to create an environment for deleting
	createEnvForDelete := func(name, slug string) model.EnvironmentResponse {
		createReq := model.CreateEnvironmentRequest{Name: name, Description: name + " desc", Slug: slug}
		payload, _ := json.Marshal(createReq)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/environments", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code, "Failed to create env for DeleteEnvironment test")
		var apiRespDeleteHelper struct {
			Code    int                        `json:"code"`
			Message string                     `json:"message"`
			Data    model.EnvironmentResponse `json:"data"`
		}
		t.Logf("createEnvForDelete Raw Response for %s: %s", slug, w.Body.String())
		errDeleteHelper := json.Unmarshal(w.Body.Bytes(), &apiRespDeleteHelper)
		require.NoError(t, errDeleteHelper, "Failed to unmarshal response in createEnvForDelete helper for TestDeleteEnvironment")
		t.Logf("createEnvForDelete Unmarshalled apiResp.Data for %s: %+v", slug, apiRespDeleteHelper.Data)
		require.NotEqual(t, uint(0), apiRespDeleteHelper.Data.ID, "createEnvForDelete: apiResp.Data.ID should not be zero for slug %s", slug)
		require.NotEmpty(t, apiRespDeleteHelper.Data.Slug, "createEnvForDelete: apiResp.Data.Slug should not be empty for slug %s", slug)
		return apiRespDeleteHelper.Data
	}

	envToDelete := createEnvForDelete("Env To Delete", "env-to-delete")

	t.Run("Success_DeleteExistingEnvironment", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/environments/%d", envToDelete.ID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)

		// Verify it's soft-deleted (GET by ID should now be 404)
		reqGet, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/environments/%d", envToDelete.ID), nil)
		reqGet.Header.Set("Authorization", "Bearer "+token)
		wGet := httptest.NewRecorder()
		components.Router.ServeHTTP(wGet, reqGet)
		assert.Equal(t, http.StatusNotFound, wGet.Code, "Environment should not be found after deletion")
	})

	t.Run("Failure_DeleteNonExistentEnvironment", func(t *testing.T) {
		nonExistentID := 99988
		req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/environments/%d", nonExistentID), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("Failure_InvalidIDFormat", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/environments/invalid-id-format", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		components.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
