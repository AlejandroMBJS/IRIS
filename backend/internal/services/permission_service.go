/*
Package services - Permission Matrix Service

==============================================================================
FILE: internal/services/permission_service.go
==============================================================================

DESCRIPTION:
    Manages the permission matrix for role-based access control. Provides
    CRUD operations for permissions and seed default permissions for all roles.

USER PERSPECTIVE:
    - Admin configures permissions for each role through the permission matrix UI
    - Roles automatically get default permissions on system startup
    - Permissions control what features and data each role can access

DEVELOPER GUIDELINES:
    ✅  OK to modify: Default permissions, add new resources
    ⚠️  CAUTION: Changing defaults affects existing installations
    ❌  DO NOT modify: Core permission check logic
    📝  Update SeedDefaultPermissions when adding new roles or resources

==============================================================================
*/
package services

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"backend/internal/models"
)

// PermissionService handles permission matrix operations
type PermissionService struct {
	db *gorm.DB
}

// NewPermissionService creates a new permission service
func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// GetAllPermissions retrieves all permission configurations
func (s *PermissionService) GetAllPermissions() ([]models.Permission, error) {
	var permissions []models.Permission
	if err := s.db.Order("role ASC, resource ASC").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %w", err)
	}
	return permissions, nil
}

// GetPermissionsByRole retrieves all permissions for a specific role
func (s *PermissionService) GetPermissionsByRole(role string) ([]models.Permission, error) {
	var permissions []models.Permission
	if err := s.db.Where("role = ?", role).Order("resource ASC").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions for role %s: %w", role, err)
	}
	return permissions, nil
}

// GetPermission retrieves a specific permission by role and resource
func (s *PermissionService) GetPermission(role, resource string) (*models.Permission, error) {
	var permission models.Permission
	if err := s.db.Where("role = ? AND resource = ?", role, resource).First(&permission).Error; err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}
	return &permission, nil
}

// CreatePermission creates a new permission entry
func (s *PermissionService) CreatePermission(permission *models.Permission) error {
	// Check if permission already exists
	var existing models.Permission
	err := s.db.Where("role = ? AND resource = ?", permission.Role, permission.Resource).First(&existing).Error
	if err == nil {
		return fmt.Errorf("permission already exists for role '%s' and resource '%s'", permission.Role, permission.Resource)
	}

	if err := s.db.Create(permission).Error; err != nil {
		return fmt.Errorf("failed to create permission: %w", err)
	}
	return nil
}

// UpdatePermission updates an existing permission
func (s *PermissionService) UpdatePermission(id string, updates *models.Permission) (*models.Permission, error) {
	var permission models.Permission
	if err := s.db.First(&permission, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}

	// Update fields
	permission.Permissions = updates.Permissions
	permission.Description = updates.Description
	permission.IsActive = updates.IsActive

	if err := s.db.Save(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return &permission, nil
}

// DeletePermission deletes a permission entry
func (s *PermissionService) DeletePermission(id string) error {
	result := s.db.Delete(&models.Permission{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete permission: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("permission not found")
	}
	return nil
}

// SeedDefaultPermissions creates default permission matrix for all roles
func (s *PermissionService) SeedDefaultPermissions() error {
	defaultPermissions := getDefaultPermissions()

	for _, perm := range defaultPermissions {
		// Check if permission already exists
		var existing models.Permission
		err := s.db.Where("role = ? AND resource = ?", perm.Role, perm.Resource).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// Permission doesn't exist, create it
			if err := s.db.Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to seed permission for role %s, resource %s: %w", perm.Role, perm.Resource, err)
			}
		}
		// If permission exists, skip it (don't overwrite custom configurations)
	}

	return nil
}

// getDefaultPermissions returns the default permission matrix for all roles
func getDefaultPermissions() []models.Permission {
	permissions := []models.Permission{}

	// Admin - Full access to everything
	adminPerms := models.PermissionSet{
		CanView:    true,
		CanCreate:  true,
		CanEdit:    true,
		CanDelete:  true,
		CanExport:  true,
		CanApprove: true,
	}
	resources := models.ValidResources()
	for _, resource := range resources {
		perm := createPermission("admin", resource, &adminPerms, "Full access to all resources")
		permissions = append(permissions, perm)
	}

	// HR - Full access to employees, incidences, approvals
	hrPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: true, CanDelete: true, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range hrPerms {
		permissions = append(permissions, createPermission("hr", resource, &perms, "HR access to "+resource))
	}

	// Supervisor - View team, approve requests
	supervisorPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"payroll":       {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range supervisorPerms {
		permissions = append(permissions, createPermission("supervisor", resource, &perms, "Supervisor access to "+resource))
	}

	// Manager - View team + subordinates, approve requests
	managerPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range managerPerms {
		permissions = append(permissions, createPermission("manager", resource, &perms, "Manager access to "+resource))
	}

	// Payroll Staff - Payroll processing and reports
	payrollPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: true},
		"reports":       {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"incidences":    {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range payrollPerms {
		permissions = append(permissions, createPermission("payroll_staff", resource, &perms, "Payroll staff access to "+resource))
	}

	// Employee - Basic view access
	employeePerms := map[string]models.PermissionSet{
		"employees":     {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"reports":       {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"incidences":    {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range employeePerms {
		permissions = append(permissions, createPermission("employee", resource, &perms, "Employee access to "+resource))
	}

	// Accountant - Financial and payroll access
	accountantPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: true},
		"reports":       {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range accountantPerms {
		permissions = append(permissions, createPermission("accountant", resource, &perms, "Accountant access to "+resource))
	}

	// HR and Payroll combined - Full HR + Payroll access
	hrPrPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: true},
		"reports":       {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: true},
		"users":         {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: true, CanDelete: true, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range hrPrPerms {
		permissions = append(permissions, createPermission("hr_and_pr", resource, &perms, "HR & Payroll access to "+resource))
	}

	// HR Blue/Gray collar - HR for blue and gray collar employees only
	hrBlueGrayPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: true, CanDelete: true, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range hrBlueGrayPerms {
		permissions = append(permissions, createPermission("hr_blue_gray", resource, &perms, "HR Blue/Gray access to "+resource))
	}

	// HR White collar - HR for white collar employees only
	hrWhitePerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: true, CanApprove: false},
		"users":         {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: true, CanDelete: true, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: true, CanEdit: true, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range hrWhitePerms {
		permissions = append(permissions, createPermission("hr_white", resource, &perms, "HR White access to "+resource))
	}

	// Supervisor and General Manager combined
	supGmPerms := map[string]models.PermissionSet{
		"employees":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"payroll":       {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"reports":       {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: false},
		"configuration": {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"approvals":     {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: true},
		"incidences":    {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: true, CanApprove: true},
		"users":         {CanView: false, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"announcements": {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"calendar":      {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"shifts":        {CanView: true, CanCreate: false, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
		"inbox":         {CanView: true, CanCreate: true, CanEdit: false, CanDelete: false, CanExport: false, CanApprove: false},
	}
	for resource, perms := range supGmPerms {
		permissions = append(permissions, createPermission("sup_and_gm", resource, &perms, "Supervisor & GM access to "+resource))
	}

	return permissions
}

// createPermission helper function to create a permission with JSONB data
func createPermission(role, resource string, perms *models.PermissionSet, description string) models.Permission {
	permJSON, _ := json.Marshal(perms)
	return models.Permission{
		Role:        role,
		Resource:    resource,
		Permissions: permJSON,
		Description: description,
		IsActive:    true,
	}
}
