package common_operations

import (
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// get customer by ID
func GetCustomerByID(trx *gorm.DB, customer_id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	result := trx.Model(&models.Customer{}).
		Where("id = ?", customer_id).
		First(&customer)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customer, nil
}

// get customer by ID
func GetCustomerTokenByCustomerID(trx *gorm.DB, customer_id uuid.UUID) (*[]models.CustomerToken, error) {
	var customerTokens []models.CustomerToken
	result := trx.Model(&models.CustomerToken{}).
		Select("DISTINCT ON (fcm_token) *").
		Where("customer_id = ?", customer_id).
		Where("fcm_token IS NOT NULL AND fcm_token != ''").
		Find(&customerTokens)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerTokens, nil
}
