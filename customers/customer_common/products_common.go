package customer_common

import (
	"context"
	"sort"
	"strings"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type ModifierGroupIDAndMaxSelection struct {
	ModifierGroupID uuid.UUID
	MaxSelection    int
}

type SyncModifierOptionToOutletRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	OutletID   uuid.UUID `json:"outlet_id"`
}

// ========================================
// CUSTOMERS
// ========================================

// sort models.Product by sort_order
func SortProductsBySortOrder(products []models.Product, isAscending bool) []models.Product {
	if len(products) == 0 {
		return products
	}
	sort.Slice(products, func(i, j int) bool {
		if isAscending {
			return products[i].SortOrder < products[j].SortOrder
		} else {
			return products[i].SortOrder > products[j].SortOrder
		}
	})
	return products
}

// sort models.OutletProduct by sort_order
func SortOutletProductsBySortOrder(products []models.OutletProduct, isAscending bool) []models.OutletProduct {
	if len(products) == 0 {
		return products
	}
	sort.Slice(products, func(i, j int) bool {
		if isAscending {
			return products[i].Product.SortOrder < products[j].Product.SortOrder
		} else {
			return products[i].Product.SortOrder > products[j].Product.SortOrder
		}
	})
	return products
}

// func to get all products from business
func GetAllProductsFromBusiness(
	trx *gorm.DB,
	business_id uuid.UUID,
	is_active bool,
	is_store_outlet bool,
	is_grab_food bool,
	is_shopee_food bool,
) ([]models.Product, error) {
	var products []models.Product
	result := trx.Model(&models.Product{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", is_active).
		Where("is_store_outlet = ?", is_store_outlet).
		Where("is_grab_food = ?", is_grab_food).
		Where("is_shopee_food = ?", is_shopee_food).
		Order("sort_order ASC").
		Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}

// func to get all products from Outlet Product
func GetAllProductsFromOutletProduct(
	trx *gorm.DB,
	outlet_id uuid.UUID,
	is_active_in_outlet *bool,
	is_active_in_business bool,
	is_store_outlet bool,
	is_grab_food bool,
	is_shopee_food bool,
) ([]models.OutletProduct, error) {
	var outletProducts []models.OutletProduct
	query := trx.Model(&models.OutletProduct{}).
		Where("outlet_id = ?", outlet_id).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", is_active_in_business).
				Where("is_store_outlet = ?", is_store_outlet).
				Where("is_grab_food = ?", is_grab_food).
				Where("is_shopee_food = ?", is_shopee_food)
		})
	if is_active_in_outlet != nil {
		query = query.Where("is_active = ?", is_active_in_outlet)
	}
	result := query.Find(&outletProducts)
	if result.Error != nil {
		return nil, result.Error
	}
	/* result := trx.Model(&models.OutletProduct{}).
		Where("outlet_id = ?", outlet_id).
		Where("is_active = ?", is_active_in_outlet).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_active = ?", is_active_in_business).
				Where("is_store_outlet = ?", is_store_outlet).
				Where("is_grab_food = ?", is_grab_food).
				Where("is_shopee_food = ?", is_shopee_food)
		}).
		Find(&outletProducts)
	if result.Error != nil {
		return nil, result.Error
	} */
	outletProducts = SortOutletProductsBySortOrder(outletProducts, true)
	return outletProducts, nil
}

// func to get product ids from outlet products
func GetProductIDsFromOutletProducts(trx *gorm.DB, outlet_id uuid.UUID, is_active bool) ([]uuid.UUID, error) {
	var productIDs []uuid.UUID
	result := trx.Model(&models.OutletProduct{}).
		Where("outlet_id = ? AND is_active = ?", outlet_id, is_active).
		Select("product_id").
		Find(&productIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return productIDs, nil

}

// func to get product modifier ids
func GetModifierGroupIDsFromProductModifierMapping(trx *gorm.DB, product_id uuid.UUID) ([]uuid.UUID, error) {
	var productModifierGroupIDs []uuid.UUID
	result := trx.Model(&models.ProductModifierMapping{}).
		Where("product_id = ?", product_id).
		Select("modifier_group_id").
		Find(&productModifierGroupIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return productModifierGroupIDs, nil
}

// func to get product by id
func GetOutletProductByProductID(trx *gorm.DB, product_id uuid.UUID) (*models.OutletProduct, error) {
	var product models.OutletProduct
	result := trx.Model(&models.OutletProduct{}).
		Where("product_id = ? AND is_active = ?", product_id, true).
		Preload("Product").
		First(&product)
	if result.Error != nil {
		return nil, result.Error
	}
	return &product, nil
}

// check modifier option is active at outlet level
func CheckModifierOptionIsActiveAtOutletLevel(trx *gorm.DB, outlet_id uuid.UUID, modifier_options []models.ModifierOptions) ([]models.ModifierOptions, error) {
	for i := range modifier_options {
		modifier_option := &modifier_options[i]
		// means already inactive in business level
		if !modifier_option.IsActive {
			continue
		}
		var outletModifierOption models.OutletModifierOption
		result := trx.Model(&models.OutletModifierOption{}).
			Where("outlet_id = ? AND modifier_options_id = ?", outlet_id, modifier_option.ID).
			First(&outletModifierOption)
		if result.Error != nil {
			outletModifierOption = models.OutletModifierOption{
				OutletID:          outlet_id,
				ModifierOptionsID: modifier_option.ID,
				IsActive:          true,
			}
			trx.Create(&outletModifierOption)
		}
		// mean is inactive at outlet level
		if !outletModifierOption.IsActive {
			modifier_option.IsActive = false
		}
	}
	return modifier_options, nil
}

// func to get max selection from product modifier mapping
func GetMaxSelectionFromProductModifierMapping(trx *gorm.DB, productID uuid.UUID, modifierGroupIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(modifierGroupIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	var modifierGroupIDAndMaxSelections []ModifierGroupIDAndMaxSelection
	result := trx.Model(&models.ProductModifierMapping{}).
		Select("modifier_group_id, max_selection").
		Where("product_id = ? AND modifier_group_id IN (?)", productID, modifierGroupIDs).
		Find(&modifierGroupIDAndMaxSelections)
	if result.Error != nil {
		return map[uuid.UUID]int{}, result.Error
	}
	out := make(map[uuid.UUID]int, len(modifierGroupIDAndMaxSelections))
	for _, modifierGroupIDAndMaxSelection := range modifierGroupIDAndMaxSelections {
		out[modifierGroupIDAndMaxSelection.ModifierGroupID] = modifierGroupIDAndMaxSelection.MaxSelection
	}
	return out, nil
}

// sync modifier option to outlet level (currently not used in membership and pos only)
func SyncModifierOptionToOutlet(ctx context.Context, req *SyncModifierOptionToOutletRequest, trx *gorm.DB) error {

	var modifierOptions []models.ModifierOptions
	err := trx.Model(&models.ModifierOptions{}).Joins("JOIN modifier_groups ON modifier_options.modifier_group_id = modifier_groups.id").
		Where("modifier_groups.business_id = ?", req.BusinessID).
		Where("modifier_options.is_active = ?", true).
		Find(&modifierOptions).Error
	if err != nil {
		return err
	}

	for _, modifierOption := range modifierOptions {
		var outletModifierOption models.OutletModifierOption
		err = trx.Model(&models.OutletModifierOption{}).Where("outlet_id = ? AND modifier_options_id = ?", req.OutletID, modifierOption.ID).First(&outletModifierOption).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			// Product doesn't exist in outlet, create it
			outletModifierOption = models.OutletModifierOption{
				OutletID:          req.OutletID,
				ModifierOptionsID: modifierOption.ID,
				IsActive:          true,
			}
			err = trx.Create(&outletModifierOption).Error
			if err != nil {
				return err
			}
		}

	}

	return nil
}

// func to get modifier group by id
func GetProductModifierMappingsByProductIDs(trx *gorm.DB, product_ids []uuid.UUID, preloadProduct bool, preloadModifierGroup bool) ([]models.ProductModifierMapping, error) {
	var productModifierMappings []models.ProductModifierMapping
	query := trx.Model(&models.ProductModifierMapping{}).
		Where("product_id IN (?)", product_ids)
	if preloadProduct {
		query = query.Preload("Product")
	}
	if preloadModifierGroup {
		query = query.Preload("ModifierGroup")
	}

	result := query.Find(&productModifierMappings)

	if result.Error != nil {
		return nil, result.Error
	}
	return productModifierMappings, nil
}

func CombineModifiers(modifiers []models.SelectedModifierGroup) string {
	modifierOptionsString := ""
	for _, modifier := range modifiers {
		modifierOptionsString += modifier.ModifierOptions.Name + ", "
	}
	modifierOptionsString = strings.TrimSuffix(modifierOptionsString, ", ")
	return modifierOptionsString
}

// func to get all products by ids
func GetProductsByIDs(trx *gorm.DB, product_ids []uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	result := trx.Model(&models.Product{}).Where("id IN (?)", product_ids).Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}
	return products, nil
}
