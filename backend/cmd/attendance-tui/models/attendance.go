package models

import (
	"time"

	"github.com/google/uuid"

	mainModels "backend/internal/models"
)

type AttendanceCard struct {
	ID         uuid.UUID            `gorm:"type:text;primaryKey" json:"id" `
	CardUID    string               `gorm:"type:varchar(20);uniqueIndex;not null" json:"card_uuid"`
	EmployeeID uuid.UUID            `gorm:"type:text;not null" json:"employee_id"`
	IsActive   bool                 `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time            `gorm:"autoCreateTime" json:"created_at"`
	Employee   *mainModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

type AttendanceRecord struct {
	ID         uuid.UUID            `gorm:"type:text;primaryKey" json:"id" `
	EmployeeID uuid.UUID            `gorm:"type:text;not null;uniqueIndex:idx_emp_date" json:"employee_id" `
	CardUID    string               `gorm:"type:varchar(20);not null" json:"card_uuid"`
	Date       time.Time            `gorm:"type:date;not null;uniqueIndex:idx_emp_date" json:"date"`
	CheckIn    time.Time            `gorm:"not null" json:"check_in"`
	CheckOut   *time.Time           `json:"check_out"`
	Employee   *mainModels.Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}
