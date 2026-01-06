package api

import (
	"backend/internal/models"
	"backend/internal/services"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupPermissionTest creates test database and services
func setupPermissionTest(t *testing.T) (*gorm.DB, *PermissionHandler) {
	gin.SetMode(gin.TestMode)

	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate tables
	err = db.AutoMigrate(
		&models.Company{},
		&models.User{},
		&models.Permission{},
	)
	require.NoError(t, err)

	// Create test company
	company := &models.Company{
		Name:     "Test Company",
		RFC:      "TEST123456789",
		IsActive: true,
	}
	company.ID = uuid.New()
	require.NoError(t, db.Create(company).Error)

	// Create services
	permissionService := services.NewPermissionService(db)
	handler := NewPermissionHandler(permissionService)

	// Seed default permissions
	err = permissionService.SeedDefaultPermissions()
	require.NoError(t, err)

	return db, handler
}

// createAdminUser creates an admin user for testing
func createAdminUser(t *testing.T, db *gorm.DB, companyID uuid.UUID) *models.User {
	user := &models.User{
		Email:        "admin@test.com",
		PasswordHash: "hashed_password",
		Role:         "admin",
		FullName:     "Admin User",
		IsActive:     true,
		CompanyID:    companyID,
	}
	user.ID = uuid.New()
	require.NoError(t, db.Create(user).Error)
	return user
}

// createNonAdminUser creates a non-admin user for testing
func createNonAdminUser(t *testing.T, db *gorm.DB, companyID uuid.UUID) *models.User {
	user := &models.User{
		Email:        "employee@test.com",
		PasswordHash: "hashed_password",
		Role:         "employee",
		FullName:     "Employee User",
		IsActive:     true,
		CompanyID:    companyID,
	}
	user.ID = uuid.New()
	require.NoError(t, db.Create(user).Error)
	return user
}

// TestListPermissions_AdminAccess tests that admin can list all permissions
func TestListPermissions_AdminAccess(t *testing.T) {
	_, handler := setupPermissionTest(t)

	// Create test router
	router := gin.New()
	router.GET("/permissions", handler.ListPermissions)

	// Create request
	req, _ := http.NewRequest("GET", "/permissions", nil)
	w := httptest.NewRecorder()

	// Execute request
	router.ServeHTTP(w, req)

	// Assert success
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response
	var permissions []models.Permission
	err := json.Unmarshal(w.Body.Bytes(), &permissions)
	require.NoError(t, err)

	// Should have permissions from seed
	assert.Greater(t, len(permissions), 0)

	// Verify we have permissions for standard roles
	roleMap := make(map[string]bool)
	for _, perm := range permissions {
		roleMap[perm.Role] = true
	}

	// Check for standard roles
	expectedRoles := []string{"admin", "supervisor", "manager", "hr", "employee"}
	for _, role := range expectedRoles {
		assert.True(t, roleMap[role], "Should have permissions for role: %s", role)
	}
}

// TestGetPermissionsByRole_Success tests getting permissions for a specific role
func TestGetPermissionsByRole_Success(t *testing.T) {
	_, handler := setupPermissionTest(t)

	router := gin.New()
	router.GET("/permissions/role/:role", handler.GetPermissionsByRole)

	req, _ := http.NewRequest("GET", "/permissions/role/admin", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var permissions []models.Permission
	err := json.Unmarshal(w.Body.Bytes(), &permissions)
	require.NoError(t, err)

	// Admin should have permissions for multiple resources
	assert.Greater(t, len(permissions), 0)

	// Verify all permissions are for admin role
	for _, perm := range permissions {
		assert.Equal(t, "admin", perm.Role)
	}
}

// TestCreatePermission_Success tests creating a new permission
func TestCreatePermission_Success(t *testing.T) {
	_, handler := setupPermissionTest(t)

	router := gin.New()
	router.POST("/permissions", handler.CreatePermission)

	// Create new permission
	newPerm := models.Permission{
		Role:        "custom_role",
		Resource:    "custom_resource",
		Description: "Custom permission for testing",
		IsActive:    true,
	}

	// Set permissions
	permSet := &models.PermissionSet{
		CanView:   true,
		CanCreate: true,
		CanEdit:   false,
		CanDelete: false,
		CanExport: true,
		CanApprove: false,
	}
	err := newPerm.SetPermissions(permSet)
	require.NoError(t, err)

	body, _ := json.Marshal(newPerm)
	req, _ := http.NewRequest("POST", "/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var created models.Permission
	err = json.Unmarshal(w.Body.Bytes(), &created)
	require.NoError(t, err)

	assert.Equal(t, "custom_role", created.Role)
	assert.Equal(t, "custom_resource", created.Resource)
	assert.NotEqual(t, uuid.Nil, created.ID)
}

// TestCreatePermission_Duplicate tests that duplicate permissions are rejected
func TestCreatePermission_Duplicate(t *testing.T) {
	_, handler := setupPermissionTest(t)

	router := gin.New()
	router.POST("/permissions", handler.CreatePermission)

	// Try to create permission that already exists (from seed)
	newPerm := models.Permission{
		Role:     "admin",
		Resource: "employees",
	}

	permSet := &models.PermissionSet{
		CanView:   true,
		CanCreate: true,
	}
	err := newPerm.SetPermissions(permSet)
	require.NoError(t, err)

	body, _ := json.Marshal(newPerm)
	req, _ := http.NewRequest("POST", "/permissions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail with conflict
	assert.Equal(t, http.StatusConflict, w.Code)
}

// TestUpdatePermission_Success tests updating an existing permission
func TestUpdatePermission_Success(t *testing.T) {
	db, handler := setupPermissionTest(t)

	// Get an existing permission to update
	var existingPerm models.Permission
	err := db.Where("role = ? AND resource = ?", "supervisor", "employees").First(&existingPerm).Error
	require.NoError(t, err)

	router := gin.New()
	router.PUT("/permissions/:id", handler.UpdatePermission)

	// Update the permission
	updates := models.Permission{
		Description: "Updated description",
		IsActive:    false,
	}

	permSet := &models.PermissionSet{
		CanView:   true,
		CanCreate: false,
		CanEdit:   false,
		CanDelete: false,
		CanExport: false,
		CanApprove: false,
	}
	err = updates.SetPermissions(permSet)
	require.NoError(t, err)

	body, _ := json.Marshal(updates)
	req, _ := http.NewRequest("PUT", "/permissions/"+existingPerm.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated models.Permission
	err = json.Unmarshal(w.Body.Bytes(), &updated)
	require.NoError(t, err)

	assert.Equal(t, "Updated description", updated.Description)
	assert.False(t, updated.IsActive)
}

// TestDeletePermission_Success tests deleting a permission
func TestDeletePermission_Success(t *testing.T) {
	db, handler := setupPermissionTest(t)

	// Create a custom permission to delete
	customPerm := &models.Permission{
		Role:     "test_role",
		Resource: "test_resource",
		IsActive: true,
	}
	customPerm.ID = uuid.New()
	permSet := &models.PermissionSet{CanView: true}
	err := customPerm.SetPermissions(permSet)
	require.NoError(t, err)
	require.NoError(t, db.Create(customPerm).Error)

	router := gin.New()
	router.DELETE("/permissions/:id", handler.DeletePermission)

	req, _ := http.NewRequest("DELETE", "/permissions/"+customPerm.ID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's deleted
	var found models.Permission
	err = db.First(&found, "id = ?", customPerm.ID).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

// TestListRoles tests listing all valid roles
func TestListRoles(t *testing.T) {
	_, handler := setupPermissionTest(t)

	router := gin.New()
	router.GET("/permissions/roles", handler.ListRoles)

	req, _ := http.NewRequest("GET", "/permissions/roles", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	roles, ok := response["roles"]
	require.True(t, ok, "Response should have 'roles' key")

	// Should include standard roles
	expectedRoles := []string{"admin", "supervisor", "manager", "hr", "employee"}
	for _, expected := range expectedRoles {
		assert.Contains(t, roles, expected)
	}
}

// TestListResources tests listing all valid resources
func TestListResources(t *testing.T) {
	_, handler := setupPermissionTest(t)

	router := gin.New()
	router.GET("/permissions/resources", handler.ListResources)

	req, _ := http.NewRequest("GET", "/permissions/resources", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string][]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	resources, ok := response["resources"]
	require.True(t, ok, "Response should have 'resources' key")

	// Should include standard resources
	expectedResources := []string{"employees", "payroll", "reports", "configuration"}
	for _, expected := range expectedResources {
		assert.Contains(t, resources, expected)
	}
}

// TestPermissionMatrix_AdminHasFullAccess tests that admin has all permissions
func TestPermissionMatrix_AdminHasFullAccess(t *testing.T) {
	db, _ := setupPermissionTest(t)

	// Get all admin permissions
	var adminPerms []models.Permission
	err := db.Where("role = ?", "admin").Find(&adminPerms).Error
	require.NoError(t, err)

	// Admin should have permissions for all resources
	assert.Greater(t, len(adminPerms), 0)

	// Verify admin has full permissions for each resource
	for _, perm := range adminPerms {
		permSet, err := perm.GetPermissions()
		require.NoError(t, err)

		// Admin should have all permissions
		assert.True(t, permSet.CanView, "Admin should have view access to "+perm.Resource)
		assert.True(t, permSet.CanCreate, "Admin should have create access to "+perm.Resource)
		assert.True(t, permSet.CanEdit, "Admin should have edit access to "+perm.Resource)
		assert.True(t, permSet.CanDelete, "Admin should have delete access to "+perm.Resource)
	}
}

// TestPermissionMatrix_EmployeeHasLimitedAccess tests employee has restricted access
func TestPermissionMatrix_EmployeeHasLimitedAccess(t *testing.T) {
	db, _ := setupPermissionTest(t)

	// Get employee permissions
	var employeePerms []models.Permission
	err := db.Where("role = ?", "employee").Find(&employeePerms).Error
	require.NoError(t, err)

	// Verify employee cannot delete employees
	var empPerm models.Permission
	err = db.Where("role = ? AND resource = ?", "employee", "employees").First(&empPerm).Error
	require.NoError(t, err)

	permSet, err := empPerm.GetPermissions()
	require.NoError(t, err)

	// Employee should not be able to delete (limited permissions)
	assert.False(t, permSet.CanDelete, "Employee should not be able to delete")
	assert.False(t, permSet.CanCreate, "Employee should not be able to create")

	// Note: Actual view permissions depend on seed configuration
	// This test verifies employee has restricted permissions compared to admin
}
