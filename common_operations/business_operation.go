package common_operations

import (
	"fmt"
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// func to get customer data from auth data
func GetBusinessDataFromAuthData(authData func() any) (*models.User, error) {
	userData, ok := authData().(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}
	return userData, nil
}

func GetBusinessConfigurationByBusinessID(db *gorm.DB, business_id uuid.UUID) (*models.BusinessConfiguration, error) {
	var businessConfiguration models.BusinessConfiguration
	if err := db.First(&businessConfiguration, "business_id = ?", business_id).Error; err != nil {
		return nil, err
	}
	return &businessConfiguration, nil
}

// get payment methods by business id
// include multiple filters
func GetPaymentMethodsByBusinessID(
	trx *gorm.DB,
	business_id uuid.UUID,
	payment_method *models.PaymentMethod,
	payment_channel *models.PaymentChannel,
	payment_platform *models.PaymentPlatform,
	is_active *bool,
	is_maintenance *bool,
	valid_from *time.Time,
	valid_until *time.Time,
	compliance_status *models.ComplianceStatus,
) (*[]models.PaymentMethodConfiguration, error) {
	var paymentMethods []models.PaymentMethodConfiguration
	query := trx.Model(&models.PaymentMethodConfiguration{}).Where("business_id = ?", business_id)

	// Check for nil AND empty string values
	if payment_method != nil && string(*payment_method) != "" {
		query = query.Where("payment_method = ?", *payment_method)
	}
	if payment_channel != nil && string(*payment_channel) != "" {
		query = query.Where("payment_channel = ?", *payment_channel)
	}
	if payment_platform != nil && string(*payment_platform) != "" {
		query = query.Where("payment_platform = ?", *payment_platform)
	}

	// Always apply boolean filters - don't check if they're true/false
	if is_active != nil {
		query = query.Where("is_active = ?", &is_active)
	}
	if is_maintenance != nil {
		query = query.Where("is_maintenance = ?", &is_maintenance)
	}

	if valid_from != nil {
		query = query.Where("valid_from >= ?", *valid_from)
	}
	if valid_until != nil {
		query = query.Where("valid_until <= ?", *valid_until)
	}
	if compliance_status != nil {
		query = query.Where("compliance_status = ?", *compliance_status)
	}

	query = query.Order("priority ASC")
	result := query.Find(&paymentMethods)
	if result.Error != nil {
		return nil, result.Error
	}
	return &paymentMethods, nil
}

// get payment methods by business id
// include multiple filters
func GetPaymentMethodsPOSByBusinessIDWithoutFilter(
	trx *gorm.DB,
	business_id uuid.UUID,
) (*[]models.PaymentMethodConfiguration, error) {
	var paymentMethods []models.PaymentMethodConfiguration
	query := trx.Model(&models.PaymentMethodConfiguration{}).Where("business_id = ?", business_id)
	query = query.Where("payment_platform = ?", models.PaymentPlatformPOS)
	query = query.Order("priority ASC")
	result := query.Find(&paymentMethods)
	if result.Error != nil {
		return nil, result.Error
	}
	return &paymentMethods, nil
}

// Get Group Role by ID
func GetGroupRoleByID(trx *gorm.DB, id uuid.UUID, preloadRole bool) (*models.GroupRole, error) {
	var groupRole models.GroupRole
	query := trx.Model(&models.GroupRole{}).Where("id = ?", id)
	if preloadRole {
		query = query.Preload("Role")
	}
	result := query.First(&groupRole)
	if result.Error != nil {
		return nil, result.Error
	}
	return &groupRole, nil
}
