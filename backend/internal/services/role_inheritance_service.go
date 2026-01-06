/*
Package services - IRIS Payroll System Business Logic

==============================================================================
FILE: internal/services/role_inheritance_service.go
==============================================================================

DESCRIPTION:
    Manages role inheritance configuration and resolves inherited permissions.
    Provides CRUD operations and circular dependency detection.

USER PERSPECTIVE:
    - Configure which roles inherit from others
    - Automatically inherit permissions without manual assignment
    - Visual hierarchy shows permission flow

DEVELOPER GUIDELINES:
    ✅  OK to modify: Validation rules, inheritance resolution logic
    ⚠️  CAUTION: Circular dependency detection (critical for system stability)
    ❌  DO NOT modify: Core role names without updating middleware
    📝  Use ResolveInheritedRoles() to get all roles a user has (direct + inherited)

INHERITANCE RESOLUTION:
    1. Start with user's assigned role
    2. Find all parent roles (roles the user's role inherits from)
    3. Recursively find parents of parents (transitive inheritance)
    4. Return complete set of roles (de-duplicated)

CIRCULAR DEPENDENCY DETECTION:
    Before creating A → B inheritance:
    - Check if B already inherits from A (direct or transitive)
    - If yes, reject to prevent infinite loop

==============================================================================
*/
package services

import (
	"backend/internal/models"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleInheritanceService manages role inheritance configuration
type RoleInheritanceService struct {
	db *gorm.DB
}

// NewRoleInheritanceService creates a new role inheritance service
func NewRoleInheritanceService(db *gorm.DB) *RoleInheritanceService {
	return &RoleInheritanceService{db: db}
}

// GetAllInheritances retrieves all role inheritance relationships
func (s *RoleInheritanceService) GetAllInheritances() ([]models.RoleInheritance, error) {
	var inheritances []models.RoleInheritance
	err := s.db.Order("child_role ASC, priority DESC").Find(&inheritances).Error
	return inheritances, err
}

// GetInheritancesByRole retrieves all inheritances for a specific role (as child)
func (s *RoleInheritanceService) GetInheritancesByRole(role string) ([]models.RoleInheritance, error) {
	var inheritances []models.RoleInheritance
	err := s.db.Where("child_role = ? AND is_active = ?", role, true).
		Order("priority DESC").
		Find(&inheritances).Error
	return inheritances, err
}

// GetParentRoles retrieves direct parent roles for a given role
func (s *RoleInheritanceService) GetParentRoles(role string) ([]string, error) {
	var inheritances []models.RoleInheritance
	err := s.db.Where("child_role = ? AND is_active = ?", role, true).
		Order("priority DESC").
		Find(&inheritances).Error
	if err != nil {
		return nil, err
	}

	parents := make([]string, len(inheritances))
	for i, inheritance := range inheritances {
		parents[i] = inheritance.ParentRole
	}
	return parents, nil
}

// ResolveInheritedRoles recursively resolves all roles (direct + inherited) for a given role
// Returns de-duplicated list of all roles a user with the given role effectively has
func (s *RoleInheritanceService) ResolveInheritedRoles(role string) ([]string, error) {
	visited := make(map[string]bool)
	return s.resolveRolesRecursive(role, visited)
}

// resolveRolesRecursive is the recursive helper for ResolveInheritedRoles
func (s *RoleInheritanceService) resolveRolesRecursive(role string, visited map[string]bool) ([]string, error) {
	// Prevent infinite loops
	if visited[role] {
		return []string{}, nil
	}
	visited[role] = true

	// Start with the role itself
	roles := []string{role}

	// Get parent roles
	parents, err := s.GetParentRoles(role)
	if err != nil {
		return nil, err
	}

	// Recursively get inherited roles
	for _, parent := range parents {
		inheritedRoles, err := s.resolveRolesRecursive(parent, visited)
		if err != nil {
			return nil, err
		}
		roles = append(roles, inheritedRoles...)
	}

	return roles, nil
}

// CreateInheritance creates a new role inheritance relationship
func (s *RoleInheritanceService) CreateInheritance(inheritance *models.RoleInheritance) error {
	// Validate role names
	if !models.IsValidRole(inheritance.ChildRole) {
		return fmt.Errorf("invalid child role: %s", inheritance.ChildRole)
	}
	if !models.IsValidRole(inheritance.ParentRole) {
		return fmt.Errorf("invalid parent role: %s", inheritance.ParentRole)
	}

	// Prevent self-inheritance
	if inheritance.ChildRole == inheritance.ParentRole {
		return fmt.Errorf("a role cannot inherit from itself")
	}

	// Check for circular dependency
	hasCircular, err := s.wouldCreateCircularDependency(inheritance.ChildRole, inheritance.ParentRole)
	if err != nil {
		return fmt.Errorf("failed to check circular dependency: %w", err)
	}
	if hasCircular {
		return fmt.Errorf("circular dependency detected: %s already inherits from %s (directly or transitively)",
			inheritance.ParentRole, inheritance.ChildRole)
	}

	// Check if inheritance already exists
	var existing models.RoleInheritance
	err = s.db.Where("child_role = ? AND parent_role = ?", inheritance.ChildRole, inheritance.ParentRole).
		First(&existing).Error
	if err == nil {
		return fmt.Errorf("inheritance already exists between %s and %s", inheritance.ChildRole, inheritance.ParentRole)
	}

	return s.db.Create(inheritance).Error
}

// wouldCreateCircularDependency checks if adding childRole → parentRole would create a circular dependency
// Returns true if parentRole already inherits from childRole (directly or transitively)
func (s *RoleInheritanceService) wouldCreateCircularDependency(childRole, parentRole string) (bool, error) {
	// Get all roles that parentRole inherits from (transitively)
	inheritedRoles, err := s.ResolveInheritedRoles(parentRole)
	if err != nil {
		return false, err
	}

	// Check if childRole is in the inherited roles
	for _, inherited := range inheritedRoles {
		if inherited == childRole {
			return true, nil
		}
	}

	return false, nil
}

// UpdateInheritance updates an existing inheritance relationship
func (s *RoleInheritanceService) UpdateInheritance(id uuid.UUID, updates *models.RoleInheritance) error {
	var existing models.RoleInheritance
	err := s.db.First(&existing, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("inheritance not found")
		}
		return fmt.Errorf("failed to get inheritance: %w", err)
	}

	// Update allowed fields
	if updates.IsActive != existing.IsActive {
		existing.IsActive = updates.IsActive
	}
	if updates.Priority != 0 {
		existing.Priority = updates.Priority
	}
	if updates.Notes != "" {
		existing.Notes = updates.Notes
	}

	return s.db.Save(&existing).Error
}

// DeleteInheritance deletes a role inheritance relationship
func (s *RoleInheritanceService) DeleteInheritance(id uuid.UUID) error {
	result := s.db.Delete(&models.RoleInheritance{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete inheritance: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("inheritance not found")
	}
	return nil
}

// GetInheritanceHierarchy returns a hierarchical representation of all role inheritances
// Useful for visualization in UI
func (s *RoleInheritanceService) GetInheritanceHierarchy() (map[string][]string, error) {
	var inheritances []models.RoleInheritance
	err := s.db.Where("is_active = ?", true).Find(&inheritances).Error
	if err != nil {
		return nil, err
	}

	hierarchy := make(map[string][]string)
	for _, inheritance := range inheritances {
		hierarchy[inheritance.ChildRole] = append(hierarchy[inheritance.ChildRole], inheritance.ParentRole)
	}

	return hierarchy, nil
}

// SeedDefaultInheritances creates common role inheritance relationships
// Should be called during application setup
func (s *RoleInheritanceService) SeedDefaultInheritances() error {
	// Check if any inheritances already exist
	var count int64
	s.db.Model(&models.RoleInheritance{}).Count(&count)
	if count > 0 {
		// Inheritances already exist, skip seeding
		return nil
	}

	// Default inheritance relationships
	defaultInheritances := []models.RoleInheritance{
		// Admin inherits from everyone
		{ChildRole: "admin", ParentRole: "hr", Priority: 10, IsActive: true, Notes: "Admin has all HR permissions"},
		{ChildRole: "admin", ParentRole: "manager", Priority: 9, IsActive: true, Notes: "Admin has all manager permissions"},
		{ChildRole: "admin", ParentRole: "payroll", Priority: 8, IsActive: true, Notes: "Admin has all payroll permissions"},
		{ChildRole: "admin", ParentRole: "supervisor", Priority: 7, IsActive: true, Notes: "Admin has all supervisor permissions"},

		// HR roles inherit from supervisor
		{ChildRole: "hr", ParentRole: "supervisor", Priority: 5, IsActive: true, Notes: "HR can do everything supervisors can"},
		{ChildRole: "hr_blue_gray", ParentRole: "supervisor", Priority: 5, IsActive: true, Notes: "HR Blue/Gray can supervise"},
		{ChildRole: "hr_white", ParentRole: "supervisor", Priority: 5, IsActive: true, Notes: "HR White can supervise"},

		// Manager inherits from supervisor
		{ChildRole: "manager", ParentRole: "supervisor", Priority: 5, IsActive: true, Notes: "Managers can supervise"},

		// GM inherits from manager and hr
		{ChildRole: "gm", ParentRole: "manager", Priority: 6, IsActive: true, Notes: "GM has manager permissions"},
		{ChildRole: "gm", ParentRole: "hr", Priority: 6, IsActive: true, Notes: "GM has HR permissions"},

		// Payroll inherits from accountant
		{ChildRole: "payroll", ParentRole: "accountant", Priority: 5, IsActive: true, Notes: "Payroll has accounting access"},

		// HR and PR combined role
		{ChildRole: "hr_and_pr", ParentRole: "hr", Priority: 5, IsActive: true, Notes: "Combined role has HR permissions"},
		{ChildRole: "hr_and_pr", ParentRole: "payroll", Priority: 5, IsActive: true, Notes: "Combined role has payroll permissions"},
	}

	for _, inheritance := range defaultInheritances {
		// Ignore errors for duplicates (in case some were created manually)
		s.db.Create(&inheritance)
	}

	return nil
}
