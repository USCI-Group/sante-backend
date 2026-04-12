package common_operations

import (
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// get expenses by outlet id
func GetExpensesByID(trx *gorm.DB, expensesID uuid.UUID, preloadOutlet bool) (*models.ExpensesOutlet, error) {
	var expenses models.ExpensesOutlet
	query := trx.Model(&models.ExpensesOutlet{}).Where("id = ?", expensesID)
	if preloadOutlet {
		query = query.Preload("Outlet")
	}
	result := query.First(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}
	return &expenses, nil
}
