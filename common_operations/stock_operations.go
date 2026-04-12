package common_operations

import (
	"context"

	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// CreateStock creates a new stock entry for an ingredient at an outlet
func CreateStock(db *gorm.DB, ctx context.Context, outletID uuid.UUID, ingredient models.Ingredient) (models.Stock, error) {
	stock := models.Stock{
		Name:               ingredient.Name,
		Description:        ingredient.Description,
		IngredientID:       ingredient.ID,
		OutletID:           outletID,
		SmallScaleUnit:     constants.GetSmallUnitFromLarge(ingredient.Unit),
		LargeScaleUnit:     constants.GetLargeUnitFromSmall(ingredient.Unit),
		SmallScaleQuantity: 0,
		LargeScaleQuantity: 0,
	}
	result := db.Create(&stock)
	if result.Error != nil {
		return models.Stock{}, result.Error
	}
	// Fetch the newly created stock with ingredient preloaded
	var stockWithIngredient models.Stock
	if err := db.Where("id = ?", stock.ID).Preload("Ingredient").First(&stockWithIngredient).Error; err != nil {
		return models.Stock{}, err
	}

	// Return the stock with ingredient preloaded
	return stockWithIngredient, nil
}

func GetStockReportByIngredientForToday(db *gorm.DB, ctx context.Context, ingredient_id uuid.UUID, outlet_id uuid.UUID) (*models.StockReport, error) {
	var existingReport models.StockReport
	result := db.Where("ingredient_id = ? AND outlet_id = ?",
		ingredient_id, outlet_id).
		Where("DATE(created_at) = CURRENT_DATE").
		First(&existingReport)

	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		var outlet models.Outlet
		result = db.Where("id = ?", outlet_id).First(&outlet)
		if result.Error != nil {
			return nil, result.Error
		}

		createNewReport := models.StockReport{
			IngredientID:    ingredient_id,
			OutletID:        outlet_id,
			BusinessID:      outlet.BusinessID,
			Opening:         &decimal.Zero,
			OpeningBySystem: &decimal.Zero,
			Sales:           0,
			Purchases:       0,
			TransferIn:      &decimal.Zero,
			TransferOut:     &decimal.Zero,
			Wastage:         0,
			Closing:         &decimal.Zero,
			ClosingBySystem: &decimal.Zero,
			Variance:        &decimal.Zero,
		}
		result = db.Create(&createNewReport)
		if result.Error != nil {
			return nil, result.Error
		}
		return &createNewReport, nil
	} else if result.Error != nil {
		return nil, result.Error
	}

	return &existingReport, nil
}
