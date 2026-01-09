package service

import (
	"backend/cmd/attendance-tui/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errors"
	"time"
)

type Service struct {
	db *gorm.DB
}

type SwipeResult struct {
	EmployeeName string
	Type         string
	Time         time.Time
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterCard(cardUID string, employeeID uuid.UUID) error {
	card := models.AttendanceCard{
		ID:         uuid.New(),
		CardUID:    cardUID,
		EmployeeID: employeeID,
		IsActive:   true,
	}
	return s.db.Create(&card).Error
}

func (s *Service) ProcessSwipe(cardUID string) (*SwipeResult, error) {
	var card models.AttendanceCard
	err := s.db.Preload("Employee").Where("card_uid = ?", cardUID).First(&card).Error
	if err != nil {
		return nil, errors.New("Tarjeta no registrada")
	}

	if !card.IsActive {
		return nil, errors.New("tarjeta desactivada")
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0,0,0,0,now.Location())

	var record models.AttendanceRecord
	err = s.db.Where("employee_id = ? AND date = ?", card.EmployeeID, today).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		record = models.AttendanceRecord{
			ID:	uuid.New(),
			EmployeeID: card.EmployeeID,
			CardUID: cardUID,
			Date: today,
			CheckIn: time.Now(),
			Employee: card.Employee,
			
		}
		err = s.db.Create(&record).Error
		if err != nil {
			return nil, err
		}
		return &SwipeResult{
			EmployeeName: card.Employee.FirstName + "" + card.Employee.LastName,
			Type: "ENTRADA",
			Time: now,
		}, nil
	} else if err != nil {
		return nil, err 
	}
	if record.CheckOut != nil {
		return nil, errors.New("Ya completo su jornada")
	}
	record.CheckOut = &now
	err = s.db.Save(&record).Error
	if err != nil {
		return nil, err
	}
	return &SwipeResult{
		
			EmployeeName: card.Employee.FirstName + " " + card.Employee.LastName,
			Type: "SALIDA",
			Time: now,
	}, nil
}
