package common_operations

import (
	"encore.app/database/models"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

func GetMerchantSecret(db *gorm.DB, outlet_id uuid.UUID) (*models.MerchantSecret, error) {
	var merchantSecret models.MerchantSecret
	err := db.Where("outlet_id = ?", outlet_id).First(&merchantSecret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Merchant secret not found for this outlet",
			}
		}
		return nil, err
	}

	return &merchantSecret, nil
}
