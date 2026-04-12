package customer_common

import (
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// BUSINESS MANAGEMENT & UTILITIES
// ========================================

// get tax configuration by business id
func GetTaxConfigurationByBusinessID(trx *gorm.DB, business_id uuid.UUID) (*models.BusinessConfiguration, error) {
	var businessConfiguration models.BusinessConfiguration
	query := trx.Model(&models.BusinessConfiguration{}).
		Where("business_id = ?", business_id).
		First(&businessConfiguration)
	if query.Error != nil {
		return nil, query.Error
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
	is_active bool,
	is_maintenance bool,
	is_visible bool,
	valid_from *time.Time,
	valid_until *time.Time,
	compliance_status *models.ComplianceStatus,
) (*[]models.PaymentMethodConfiguration, error) {
	var paymentMethods []models.PaymentMethodConfiguration

	query := trx.Model(&models.PaymentMethodConfiguration{}).Where("business_id = ?", business_id)

	if payment_method != nil {
		query = query.Where("payment_method = ?", *payment_method)
	}
	if payment_channel != nil {
		query = query.Where("payment_channel = ?", *payment_channel)
	}
	if payment_platform != nil {
		query = query.Where("payment_platform = ?", *payment_platform)
	}

	// Always apply boolean filters - don't check if they're true/false
	query = query.Where("is_active = ?", is_active)
	query = query.Where("is_maintenance = ?", is_maintenance)
	query = query.Where("is_visible = ?", is_visible)

	if valid_from != nil {
		query = query.Where("valid_from <= ?", *valid_from)
	}
	if valid_until != nil {
		query = query.Where("valid_until >= ?", *valid_until)
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

// get latest app version by platform and app package name
func GetLastestAppVersionByPlatformAndAppPackageName(trx *gorm.DB, platform string, app_package_name string) (*models.AppVersion, error) {
	var appVersion models.AppVersion
	query := trx.Model(&models.AppVersion{}).
		Where("platform = ?", platform).
		Where("app_package_name = ?", app_package_name).
		Where("is_active = ?", true).
		Order("version_name DESC , version_code DESC").
		First(&appVersion)
	if query.Error != nil {
		return nil, query.Error
	}
	return &appVersion, nil
}
