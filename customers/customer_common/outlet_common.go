package customer_common

import (
	"strings"

	"encore.app/common"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// CUSTOMERS
// ========================================

// func to get all outlet address by business id
func GetOutletAddress(trx *gorm.DB, business_id uuid.UUID, must_open bool) ([]models.Outlet, error) {
	var outlets []models.Outlet
	query := trx.Model(&models.Outlet{}).Where("business_id = ?", business_id)
	if must_open {
		query = query.Where("outlet_status = ?", models.OutletStatusOpen)
	}
	result := query.Find(&outlets)
	if result.Error != nil {
		return []models.Outlet{}, result.Error
	}
	return outlets, nil
}

// func to combine address to string
func CombineAddress(address common.Address) string {
	streetLine1 := address.StreetLine1
	streetLine2 := address.StreetLine2
	streetLine3 := address.StreetLine3
	city := address.City
	state := address.State
	country := address.Country
	postcode := address.PostalCode
	if streetLine1 != "" {
		streetLine1 += ", "
	}
	if streetLine2 != "" {
		streetLine2 += ", "
	}
	if streetLine3 != "" {
		streetLine3 += ", "
	}
	if postcode != "" {
		postcode += ", "
	}
	if city != "" {
		city += ", "
	}
	if state != "" {
		state += ", "
	}
	if country != "" {
		country += ", "
	}
	return streetLine1 + streetLine2 + streetLine3 + postcode + city + state + country
}

// func to get outlet by search key
func GetOutletBySearchKey(trx *gorm.DB, searchKey string) ([]models.Outlet, error) {
	var outlets []models.Outlet
	query := trx.Model(&models.Outlet{}).
		Where("LOWER(name) LIKE ?", "%"+strings.ToLower(searchKey)+"%")
	result := query.Find(&outlets)
	if result.Error != nil {
		return []models.Outlet{}, result.Error
	}
	return outlets, nil
}

// Get all outlets based on business id
func GetAllOutletsBasedOnBusinessID(
	trx *gorm.DB,
	business_id uuid.UUID,
	outlet_status models.OutletStatus,
) ([]models.Outlet, error) {
	var outlets []models.Outlet
	query := trx.Model(&models.Outlet{}).Where("business_id = ?", business_id)
	if outlet_status != "" {
		query = query.Where("outlet_status = ?", outlet_status)
	}
	result := query.Find(&outlets)
	if result.Error != nil {
		return []models.Outlet{}, result.Error
	}
	return outlets, nil
}

// func to get outlet by id
func GetOutletByID(trx *gorm.DB, outlet_id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	result := trx.Model(&models.Outlet{}).Where("id = ?", outlet_id).First(&outlet)
	if result.Error != nil {
		return nil, result.Error
	}
	return &outlet, nil
}
