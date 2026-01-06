/*
Package repositories - User Authentication Data Access Layer

==============================================================================
FILE: internal/repositories/user_repository.go
==============================================================================

DESCRIPTION:
    Manages user authentication and authorization data including credentials,
    roles, company associations, and session tracking. This repository provides
    full CRUD operations, user lookups by email, role-based filtering, and
    activity tracking (last login). Supports multi-tenant user management with
    company-scoped queries.

USER PERSPECTIVE:
    - When users log in, their credentials are validated using this repository
    - User management (create, update, deactivate) flows through these methods
    - Role-based access control relies on user role data from this layer
    - Last login tracking helps administrators monitor system usage

DEVELOPER GUIDELINES:
    ✅  OK to modify: Adding new query methods for user analytics, implementing
        password reset tracking, adding session management
    ⚠️  CAUTION: User data is security-sensitive - never log passwords or tokens;
        implement proper password hashing before storage; validate email uniqueness
    ❌  DO NOT modify: Authentication logic without security review; never expose
        password hashes in API responses; don't disable soft deletes for compliance
    📝  Best practices: Always use ExistsByEmail() before creating users;
        implement password complexity validation; use parameterized queries to
        prevent SQL injection; consider rate limiting for authentication attempts

SYNTAX EXPLANATION:
    - UserRepository: Main struct holding the GORM database connection
    - Create(user *models.User): Inserts new user record (password should be
      pre-hashed by service layer)
    - FindByEmail(email string): Primary lookup method for authentication
    - List(limit, offset int, filters map[string]interface{}): Paginated user
      list with dynamic filtering (role, active status, search term)
    - query.Where("email ILIKE ?"): Case-insensitive search using PostgreSQL ILIKE
    - ExistsByEmail(): Uniqueness check returning boolean
    - UpdateLastLogin(userID uuid.UUID): Updates last_login_at using GORM's
      Expr("NOW()") for database-side timestamp
    - FindByCompanyID(): Multi-tenant query for company-scoped user management

==============================================================================
*/

package repositories

import (
    "github.com/google/uuid"
    "gorm.io/gorm"

    "backend/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
    db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, "id = ?", id).Error
    return &user, err
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, "email = ?", email).Error
    return &user, err
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
    return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// List lists users with pagination
func (r *UserRepository) List(limit, offset int, filters map[string]interface{}) ([]models.User, int64, error) {
    var users []models.User
    var total int64
    
    query := r.db.Model(&models.User{})
    
    // Apply filters
    if role, ok := filters["role"]; ok {
        query = query.Where("role = ?", role)
    }
    if isActive, ok := filters["is_active"]; ok {
        query = query.Where("is_active = ?", isActive)
    }
    if search, ok := filters["search"]; ok {
        query = query.Where("email ILIKE ? OR full_name ILIKE ?", 
            "%"+search.(string)+"%", "%"+search.(string)+"%")
    }
    
    // Count total
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Apply pagination
    if limit > 0 {
        query = query.Limit(limit)
    }
    if offset > 0 {
        query = query.Offset(offset)
    }
    
    // Execute query
    err := query.Order("created_at DESC").Find(&users).Error
    return users, total, err
}

// ExistsByEmail checks if a user exists by email
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
    var count int64
    err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}

// ExistsByEmailIncludingDeleted checks if a user exists by email including soft-deleted
func (r *UserRepository) ExistsByEmailIncludingDeleted(email string) (bool, error) {
    var count int64
    err := r.db.Unscoped().Model(&models.User{}).Where("email = ?", email).Count(&count).Error
    return count > 0, err
}

// HardDeleteByEmail permanently deletes a user by email (including soft-deleted)
func (r *UserRepository) HardDeleteByEmail(email string) error {
    return r.db.Unscoped().Where("email = ?", email).Delete(&models.User{}).Error
}

// UpdateLastLogin updates user's last login time
func (r *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
    return r.db.Model(&models.User{}).Where("id = ?", userID).
        Update("last_login_at", gorm.Expr("NOW()")).Error
}

// FindByCompanyID finds all users by company ID
func (r *UserRepository) FindByCompanyID(companyID uuid.UUID) ([]models.User, error) {
    var users []models.User
    err := r.db.Where("company_id = ?", companyID).Order("created_at DESC").Find(&users).Error
    return users, err
}
