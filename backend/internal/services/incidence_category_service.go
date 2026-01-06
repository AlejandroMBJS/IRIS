/*
Package services - Incidence Category Service

==============================================================================
FILE: internal/services/incidence_category_service.go
==============================================================================

DESCRIPTION:
    Manages incidence categories - the top-level grouping for incidence types.
    Provides CRUD operations and handles seeding of default categories.

USER PERSPECTIVE:
    - Admins can create/edit/delete categories from the UI
    - System categories (vacation, sick) are protected from deletion
    - Categories organize incidence types into logical groups

DEVELOPER GUIDELINES:
    OK to modify: Add new category operations
    CAUTION: IsSystem flag protects categories from deletion
    DO NOT modify: Default category codes without migration

==============================================================================
*/
package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"backend/internal/models"
)

// IncidenceCategoryService handles category business logic
type IncidenceCategoryService struct {
	db *gorm.DB
}

// NewIncidenceCategoryService creates a new category service
func NewIncidenceCategoryService(db *gorm.DB) *IncidenceCategoryService {
	return &IncidenceCategoryService{db: db}
}

// CategoryRequest represents request for creating/updating categories
type CategoryRequest struct {
	Name          string `json:"name" binding:"required"`
	Code          string `json:"code" binding:"required"`
	Description   string `json:"description"`
	Color         string `json:"color"`
	Icon          string `json:"icon"`
	DisplayOrder  int    `json:"display_order"`
	IsRequestable bool   `json:"is_requestable"`
	IsActive      bool   `json:"is_active"`
}

// GetAllCategories returns all categories with their incidence types
func (s *IncidenceCategoryService) GetAllCategories() ([]models.IncidenceCategory, error) {
	var categories []models.IncidenceCategory
	err := s.db.Preload("IncidenceTypes").Order("display_order, name").Find(&categories).Error
	return categories, err
}

// GetCategoryByID returns a single category by ID
func (s *IncidenceCategoryService) GetCategoryByID(id uuid.UUID) (*models.IncidenceCategory, error) {
	var category models.IncidenceCategory
	if err := s.db.Preload("IncidenceTypes").First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}
	return &category, nil
}

// GetCategoryByCode returns a category by its code
func (s *IncidenceCategoryService) GetCategoryByCode(code string) (*models.IncidenceCategory, error) {
	var category models.IncidenceCategory
	if err := s.db.First(&category, "code = ?", code).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Not found, but not an error
		}
		return nil, err
	}
	return &category, nil
}

// CreateCategory creates a new category
func (s *IncidenceCategoryService) CreateCategory(req CategoryRequest) (*models.IncidenceCategory, error) {
	// Check if code already exists
	existing, err := s.GetCategoryByCode(req.Code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("category with this code already exists")
	}

	category := &models.IncidenceCategory{
		Name:          req.Name,
		Code:          req.Code,
		Description:   req.Description,
		Color:         req.Color,
		Icon:          req.Icon,
		DisplayOrder:  req.DisplayOrder,
		IsRequestable: req.IsRequestable,
		IsSystem:      false, // New categories are never system categories
		IsActive:      req.IsActive,
	}

	if err := s.db.Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

// UpdateCategory updates an existing category
func (s *IncidenceCategoryService) UpdateCategory(id uuid.UUID, req CategoryRequest) (*models.IncidenceCategory, error) {
	var category models.IncidenceCategory
	if err := s.db.First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("category not found")
		}
		return nil, err
	}

	// Check if new code conflicts with another category
	if req.Code != category.Code {
		existing, err := s.GetCategoryByCode(req.Code)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			return nil, errors.New("category with this code already exists")
		}
	}

	// Update fields
	category.Name = req.Name
	category.Code = req.Code
	category.Description = req.Description
	category.Color = req.Color
	category.Icon = req.Icon
	category.DisplayOrder = req.DisplayOrder
	category.IsRequestable = req.IsRequestable
	category.IsActive = req.IsActive
	// Note: IsSystem cannot be changed through update

	if err := s.db.Save(&category).Error; err != nil {
		return nil, err
	}

	return &category, nil
}

// DeleteCategory deletes a category (if allowed)
func (s *IncidenceCategoryService) DeleteCategory(id uuid.UUID) error {
	var category models.IncidenceCategory
	if err := s.db.First(&category, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("category not found")
		}
		return err
	}

	// Check if system category
	if category.IsSystem {
		return errors.New("cannot delete system category")
	}

	// Check if category has incidence types
	var count int64
	if err := s.db.Model(&models.IncidenceType{}).Where("category_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete category with existing incidence types")
	}

	// Soft delete
	return s.db.Delete(&category).Error
}

// UpdateDisplayOrder updates the display order of a category
func (s *IncidenceCategoryService) UpdateDisplayOrder(id uuid.UUID, order int) error {
	return s.db.Model(&models.IncidenceCategory{}).Where("id = ?", id).Update("display_order", order).Error
}

// SeedDefaultCategories seeds the database with default categories if they don't exist
func (s *IncidenceCategoryService) SeedDefaultCategories() error {
	defaults := models.DefaultIncidenceCategories()

	for _, cat := range defaults {
		existing, err := s.GetCategoryByCode(cat.Code)
		if err != nil {
			return err
		}
		if existing == nil {
			if err := s.db.Create(&cat).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// MigrateLegacyCategories updates existing IncidenceTypes to use category_id
func (s *IncidenceCategoryService) MigrateLegacyCategories() error {
	// Get all categories
	var categories []models.IncidenceCategory
	if err := s.db.Find(&categories).Error; err != nil {
		return err
	}

	// Create map of code -> ID
	categoryMap := make(map[string]uuid.UUID)
	for _, cat := range categories {
		categoryMap[cat.Code] = cat.ID
	}

	// Update incidence types that don't have category_id set
	var types []models.IncidenceType
	if err := s.db.Where("category_id IS NULL").Find(&types).Error; err != nil {
		return err
	}

	for _, t := range types {
		if catID, ok := categoryMap[t.Category]; ok {
			if err := s.db.Model(&t).Update("category_id", catID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
