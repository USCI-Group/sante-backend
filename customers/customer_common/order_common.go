package customer_common

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type ProcessIngredientSaleRequest struct {
	Order     models.Order
	OrderItem models.OrderItem
	Mapping   GeneralIngredientMapping
}
type GeneralIngredientMapping struct {
	Ingredient models.Ingredient
	Unit       constants.UnitMeasurement
	Quantity   float32
}

// ========================================
// ORDER MANAGEMENT & UTILITIES
// ========================================

func GenerateOrderNumber(trx *gorm.DB, outletID uuid.UUID) (string, error) {
	// Generate order number
	today := time.Now().Format("2006-01-02")
	var orders []models.Order
	var orderCount int64
	result := trx.Model(&models.Order{}).
		Where("outlet_id = ?", outletID).
		Where("DATE(order_date) = ?", today).
		Order("order_date DESC").
		Find(&orders)
	if result.Error != nil {
		return "", result.Error
	}
	// get the order number based on DESC order, first index is the latest order created, hence the latest order number can extract from the first index
	if len(orders) > 0 {
		orderNumberTemp := strings.Split(orders[0].OrderNumber, "-")
		orderNumberTempInt, err := strconv.Atoi(orderNumberTemp[3])
		orderCount = int64(orderNumberTempInt)
		if err != nil {
			return "", err
		}
	} else {
		// mostly when start of the day
		// if no order found, set the order count to 0
		orderCount = 0
	}

	// generate order number and use for transaction number too
	orderNumberCount := common.GenerateOrderNumber(int(orderCount))
	orderNumberInt, err := strconv.Atoi(orderNumberCount)
	if err != nil {
		return "", err
	}
	orderNumber := fmt.Sprintf("%s-%s-%s-%06d", "ORDER", time.Now().Format("20060102"), time.Now().Format("150405"), orderNumberInt)
	return orderNumber, nil
}

// get order by order id
func GetOrderByID(trx *gorm.DB, order_id uuid.UUID, isPreload bool) (*models.Order, error) {
	var order models.Order
	query := trx.Model(&models.Order{}).Where("id = ?", order_id)
	if isPreload {
		query = query.
			Preload("OrderDetails").
			Preload("Customer").
			Preload("Outlet").
			Preload("OrderItems").
			Preload("OrderItems.Product").
			Preload("OrderItems.SelectedModifierGroups").
			Preload("OrderItems.SelectedModifierGroups.ModifierGroup").
			Preload("OrderItems.SelectedModifierGroups.ModifierOptions").
			Preload("Customer")
	}
	result := query.First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	//
	return &order, nil
}

// get order by order id
func GetOrderByIDWithCustomerID(trx *gorm.DB, order_id uuid.UUID, isPreload bool, customer_id uuid.UUID) (*models.Order, error) {
	var order models.Order
	query := trx.Model(&models.Order{}).Where("id = ?", order_id)
	if isPreload {
		query = query.Preload("OrderDetails").
			Preload("OrderDetails").
			Preload("Outlet").
			Preload("OrderItems").
			Preload("OrderItems.Product").
			Preload("OrderItems.SelectedModifierGroups").
			Preload("OrderItems.SelectedModifierGroups.ModifierGroup").
			Preload("OrderItems.SelectedModifierGroups.ModifierOptions").
			Preload("Customer")
	}
	if customer_id == uuid.Nil {
		return nil, errors.New("customer_id is required")
	}
	query = query.Where("customer_id = ?", customer_id)
	result := query.First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

// get transaction by order id
func GetTransactionByOrderID(trx *gorm.DB, order_id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	result := trx.Model(&models.Transaction{}).
		Where("order_id = ?", order_id).
		Preload("Order").
		Preload("Order.OrderDetails").
		First(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}
	return &transaction, nil
}

// get order items by order id
func GetOrderItemsByOrderID(trx *gorm.DB, order_id uuid.UUID) ([]models.OrderItem, error) {
	var orderItems []models.OrderItem
	result := trx.Model(&models.OrderItem{}).Where("order_id = ?", order_id).Preload("Product").Find(&orderItems)
	if result.Error != nil {
		return nil, result.Error
	}
	return orderItems, nil
}

// filter order items by products required
func FilterOrderItemsByProductsRequired(orderItems []models.OrderItem, productsRequired []models.MembershipUpgradeRule) ([]models.OrderItem, error) {
	if len(productsRequired) == 0 {
		return []models.OrderItem{}, nil
	}

	// Create a map for O(1) lookup of required product IDs
	requiredProductIDs := make(map[uuid.UUID]bool, len(productsRequired))
	for _, productRequired := range productsRequired {
		requiredProductIDs[*productRequired.ProductID] = true
	}

	// Pre-allocate slice with estimated capacity to reduce memory allocations
	filteredOrderItems := make([]models.OrderItem, 0, len(orderItems))

	for _, item := range orderItems {
		// Check if product ID is not empty (zero UUID check is more efficient)
		if item.Product.ID != uuid.Nil && requiredProductIDs[item.Product.ID] {
			filteredOrderItems = append(filteredOrderItems, item)
		}
	}

	return filteredOrderItems, nil
}

// func to get product ids from order items
func GetProductIDsFromOrderItems(trx *gorm.DB, order_id uuid.UUID) ([]uuid.UUID, error) {
	var productIDs []uuid.UUID
	result := trx.Model(&models.OrderItem{}).
		Select("product_id").
		Where("order_id = ?", order_id).
		Find(&productIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return productIDs, nil
}

// func to extract order item ids from order items
func GetOrderItemIDsFromOrderItems(trx *gorm.DB, order_id uuid.UUID) ([]uuid.UUID, error) {
	var orderItemIDs []uuid.UUID
	result := trx.Model(&models.OrderItem{}).
		Select("id").
		Where("order_id = ?", order_id).
		Find(&orderItemIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return orderItemIDs, nil
}

// func to get modifier group ids from order items
func GetModifierGroupIDsFromOrderItems(trx *gorm.DB, order_item_ids []uuid.UUID) ([]uuid.UUID, error) {
	var modifierGroupIDs []uuid.UUID
	result := trx.Model(&models.SelectedModifierGroup{}).
		Select("modifier_group_id").
		Where("order_item_id IN (?)", order_item_ids).
		Find(&modifierGroupIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return modifierGroupIDs, nil
}

// RoundToTwoDecimals rounds a float32 to 2 decimal places
func RoundToTwoDecimals(value float32) float32 {
	return float32(math.Round(float64(value)*100) / 100)
}

// func to get order that need to be auto cancelled
func GetOrderIDsThatNeedToBeAutoCancelled(trx *gorm.DB, timeLimit time.Duration) ([]uuid.UUID, error) {
	var orderIDs []uuid.UUID
	now := time.Now()
	timeLimitTime := now.Add(-timeLimit) // time limit is the time after which the order will be auto cancelled
	result := trx.Model(&models.Order{}).
		Where("order_status = ?", models.OrderStatusPending).
		Where("payment_status = ?", models.PaymentStatusPending).
		Where("platform = ?", models.PlatformMembershipApp).
		Where("created_at <= ?", timeLimitTime). // Changed from < to <= to include orders exactly at the time limit
		Select("id").
		Find(&orderIDs)
	if result.Error != nil {
		return nil, result.Error
	}
	return orderIDs, nil
}

// func to auto cancel order (pass in order ids)
func AutoCancelOrdersAndTransactions(trx *gorm.DB, order_ids []uuid.UUID) error {
	result := trx.Model(&models.Order{}).Where("id IN (?)", order_ids).Updates(map[string]any{
		"order_status":   models.OrderStatusCancelled,
		"payment_status": models.PaymentStatusVoided,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != int64(len(order_ids)) {
		return errors.New("failed to auto cancel orders")
	}

	result = trx.Model(&models.Transaction{}).Where("order_id IN (?)", order_ids).Updates(map[string]any{
		"payment_status": models.PaymentStatusVoided,
	})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func ProcessSale(trx *gorm.DB, ctx context.Context, order_id uuid.UUID) error {

	// Order
	var order models.Order
	result := trx.Where("id = ?", order_id).Preload("Outlet").First(&order)
	if result.Error != nil {
		return result.Error
	}

	// Order items
	var orderItems []models.OrderItem
	result = trx.Where("order_id = ?", order_id).Preload("Product").Find(&orderItems)
	if result.Error != nil {
		return result.Error
	}

	// Update stocks
	// Process each order item to update stock
	for _, orderItem := range orderItems {
		ProcessProductSale(trx, ctx, order, orderItem)
		ProcessModifierSale(trx, ctx, order, orderItem)
	}

	return nil
}

func ProcessProductSale(trx *gorm.DB, ctx context.Context, order models.Order, orderItem models.OrderItem) error {
	// Find product ingredient mappings for this product
	var productIngredientMappings []models.ProductIngredientMapping
	result := trx.Where("product_id = ?", orderItem.ProductID).Preload("Ingredient").Find(&productIngredientMappings)
	if result.Error != nil {
		return result.Error
	}

	for _, mapping := range productIngredientMappings {
		// Update ingredient stock & stock report for each mapping
		err := ProcessIngredientSale(trx, ctx, &ProcessIngredientSaleRequest{
			Order:     order,
			OrderItem: orderItem,
			Mapping: GeneralIngredientMapping{
				Ingredient: mapping.Ingredient,
				Unit:       mapping.Unit,
				Quantity:   mapping.Quantity,
			},
		})
		if err != nil {
			continue
		}
	}

	return nil
}

func ProcessModifierSale(trx *gorm.DB, ctx context.Context, order models.Order, orderItem models.OrderItem) error {
	// Find Modifier for each Order Item
	var selectedModifierGroups []models.SelectedModifierGroup
	result := trx.Where("order_item_id = ?", orderItem.ID).Find(&selectedModifierGroups)
	if result.Error != nil {
		return result.Error
	}

	for _, selectedModifierGroup := range selectedModifierGroups {
		// Find Modifier Option for each order item
		var modifierOptions []models.ModifierOptions
		result = trx.Where("id = ?", selectedModifierGroup.ModifierOptionsID).Preload("IngredientMappings.Ingredient").Find(&modifierOptions)
		if result.Error != nil {
			continue
		}

		for _, modifierOption := range modifierOptions {
			for _, modifierIngredientMapping := range modifierOption.IngredientMappings {
				// Modify quantity uses from selectedModifierGroup
				orderItem.Quantity = int(selectedModifierGroup.ModifierOptionQuantity)
				err := ProcessIngredientSale(trx, ctx, &ProcessIngredientSaleRequest{
					Order:     order,
					OrderItem: orderItem,
					Mapping: GeneralIngredientMapping{
						Ingredient: modifierIngredientMapping.Ingredient,
						Unit:       modifierIngredientMapping.Unit,
						Quantity:   modifierIngredientMapping.Quantity,
					},
				})
				if err != nil {
					continue
				}
			}
		}
	}

	return nil
}

func ProcessIngredientSale(trx *gorm.DB, ctx context.Context, req *ProcessIngredientSaleRequest) error {
	// Update ingredient stock
	var stock models.Stock
	order := req.Order
	orderItem := req.OrderItem
	mapping := req.Mapping

	result := trx.Where("ingredient_id = ?", mapping.Ingredient.ID).Where("outlet_id = ?", order.OutletID).Preload("Ingredient").First(&stock)
	if result.Error != nil {
		var err error
		stock, err = CreateStock(trx, ctx, order.OutletID, mapping.Ingredient)
		if err != nil {
			return err
		}
	}

	ingredientUsed := float32(orderItem.Quantity) * float32(mapping.Quantity)
	stock.SmallScaleQuantity -= constants.ConvertToSmallUnit(mapping.Unit, ingredientUsed)
	stock.LargeScaleQuantity -= constants.ConvertToLargeUnit(mapping.Unit, ingredientUsed)

	result = trx.Save(&stock)
	if result.Error != nil {
		return result.Error
	}

	// Update stock reports
	ingredientUsedInStockUnit := constants.ConvertToTargetUnit(mapping.Unit, ingredientUsed, stock.Ingredient.Unit)
	stockReport, _ := GetStockReportByIngredientForToday(trx, ctx, mapping.Ingredient.ID, order.OutletID)

	stockReport.Sales += ingredientUsedInStockUnit
	result = trx.Save(&stockReport)
	if result.Error != nil {
		return result.Error
	}

	return nil
}
