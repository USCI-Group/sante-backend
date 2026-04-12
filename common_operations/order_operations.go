package common_operations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type CreateOrderRequest struct {
	OutletID                  uuid.UUID  `json:"outlet_id"`
	CustomerID                *uuid.UUID `json:"customer_id"`
	UserID                    uuid.UUID  `json:"user_id"`
	OrderType                 string     `json:"order_type" valid:"required~Order type is required"`
	InvoiceNumber             string     `json:"invoice_number" valid:"required~Invoice number is required"`
	OrderStatus               string     `json:"order_status" valid:"required~Order status is required"`
	OrderDate                 *time.Time `json:"order_date"`
	GrossTotal                float32    `json:"gross_total" valid:"required~Gross total is required"`
	NetTotal                  float32    `json:"net_total" valid:"required~Net total is required"`
	RoundedAmount             float32    `json:"rounded_amount" `
	RoundedNetTotal           float32    `json:"rounded_net_total" `
	TaxCharge                 float32    `json:"tax_charge" `
	TaxPercentage             float32    `json:"tax_percentage" `
	ServiceCharge             float32    `json:"service_charge" `
	ServiceChargePercentage   float32    `json:"service_charge_percentage" `
	DiscountType              string     `json:"discount_type"`
	DiscountAmount            float32    `json:"discount_amount" `
	DiscountPercentage        float32    `json:"discount_percentage" `
	VoucherDiscountAmount     *float32   `json:"voucher_discount_amount"`
	VoucherDiscountType       *string    `json:"voucher_discount_type"`
	VoucherDiscountPercentage *float32   `json:"voucher_discount_percentage"`
	PaymentMethod             string     `json:"payment_method"`
	Notes                     string     `json:"notes"`
	TableNumber               string     `json:"table_number"`
	// order items
	CartItems         []CartItem `json:"cart_items" valid:"required~Cart items are required"`
	PickupAt          *time.Time `json:"pickup_at"`
	VoucherID         *uuid.UUID `json:"voucher_id"`
	CustomerVoucherID *uuid.UUID `json:"customer_voucher_id"`
	// platform info, grabfood
	Platform                *models.Platform `json:"platform"`
	PlatformOrderID         *string          `json:"platform_order_id"`
	PlatformState           *string          `json:"platform_state"`
	CustomerName            *string          `json:"customer_name"`
	CustomerPhone           *string          `json:"customer_phone"`
	CustomerAddress         *string          `json:"customer_address"`
	CustomerLatitude        *float32         `json:"customer_latitude"`
	CustomerLongitude       *float32         `json:"customer_longitude"`
	EstimatedOrderReadyTime *time.Time       `json:"estimated_order_ready_time"`
	MaxOrderReadyTime       *time.Time       `json:"max_order_ready_time"`
	NewOrderReadyTime       *time.Time       `json:"new_order_ready_time"`
	GrabShortOrderNum       *string          `json:"grab_short_order_num"`
	ShopeeFoodShortOrderNum *string          `json:"shopeefood_short_order_num"`
}

type CartItem struct {
	ID uuid.UUID `json:"id" `
	// need switch back to this after fixing the bug
	//Quantity               int                     `json:"quantity" valid:"required~Quantity is required"`
	Quantity  int     `json:"quantity"`
	UnitPrice float32 `json:"unit_price" valid:"required~Unit price is required"`
	// need switch back to this after fixing the bug
	//SubTotal               float32                 `json:"sub_total" valid:"required~Sub total is required"`
	SubTotal               float32                 `json:"sub_total"`
	ItemNotes              string                  `json:"item_notes"`
	SelectedModifierGroups []SelectedModifierGroup `json:"modifier_groups"`
}

type SelectedModifierGroup struct {
	ID           uuid.UUID        `json:"group_id"`
	Name         string           `json:"name" valid:"required~Name is required"`
	InputType    string           `json:"input_type" valid:"required~Input type is required"`
	MaxSelection int              `json:"max_selection"`
	Options      []ModifierOption `json:"options" valid:"required~Options are required"`
}

type ModifierOption struct {
	ID               uuid.UUID `json:"option_id"`
	Name             string    `json:"name" valid:"required~Name is required"`
	PriceAdjustment  float32   `json:"price_adjustment"`
	SelectedQuantity int       `json:"selected_quantity"`
	//Description     string    `json:"description" valid:"required~Description is required"`
}

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

type CompleteCashPaymentRequest struct {
	OrderID           uuid.UUID `json:"order_id"`
	AmountReceived    float32   `json:"amount_received" valid:"required"`
	AutoCompleteOrder bool      `json:"auto_complete_order"`
}

// use in apply order filter
type ApplyOrderFilterRequest struct {
	OutletID                       uuid.UUID  `json:"outlet_id"`
	PageNumber                     int        `json:"page_number" valid:"required~Page number is required"`
	PageSize                       int        `json:"page_size" valid:"required~Page size is required"`
	OrderStatus                    *string    `json:"order_status"` // if order status is specified, the include attributes for order will not included.
	OrderType                      *string    `json:"order_type"`
	IncludeOrderCompleted          *bool      `json:"include_order_completed" default:"true"`
	IncludeOrderCancelled          *bool      `json:"include_order_cancelled" default:"true"`
	IncludeOrderPending            *bool      `json:"include_order_pending" default:"true"`
	IncludeOrderReady              *bool      `json:"include_order_ready" default:"true"`
	IncludeOrderCollected          *bool      `json:"include_order_collected" default:"true"`
	IncludeOrderOnTheWay           *bool      `json:"include_order_on_the_way" default:"true"`
	IncludeOrderDelivered          *bool      `json:"include_order_delivered" default:"true"`
	PaymentStatus                  *string    `json:"payment_status"` // if payment status is specified, the include attributes for payment will not included.
	IncludePaymentCompleted        *bool      `json:"include_payment_completed" default:"true"`
	IncludePaymentFailed           *bool      `json:"include_payment_failed" default:"true"`
	IncludePaymentRefunded         *bool      `json:"include_payment_refunded" default:"true"`
	IncludePaymentVoided           *bool      `json:"include_payment_voided" default:"true"`
	IncludePaymentPending          *bool      `json:"include_payment_pending" default:"true"`
	IncludePaymentPendingAuthorize *bool      `json:"include_payment_pending_authorize" default:"true"`
	SearchKey                      *string    `json:"search_key"`
	StartDate                      *time.Time `json:"start_date"`
	EndDate                        *time.Time `json:"end_date"`
}

type CustomOrder struct {
	ID                        uuid.UUID                   `json:"id"`
	OrderNumber               string                      `json:"order_number"`
	OrderDate                 time.Time                   `json:"order_date"`
	OrderType                 string                      `json:"order_type"`
	InvoiceNumber             string                      `json:"invoice_number"`
	InvoiceDate               time.Time                   `json:"invoice_date"`
	OrderStatus               string                      `json:"order_status"`
	GrossTotal                float32                     `json:"gross_total"`
	NetTotal                  float32                     `json:"net_total"`
	ServiceCharge             float32                     `json:"service_charge"`
	ServiceChargePercentage   float32                     `json:"service_charge_percentage"`
	TaxCharge                 float32                     `json:"tax_charge"`
	TaxPercentage             float32                     `json:"tax_percentage"`
	DiscountType              string                      `json:"discount_type"`
	DiscountAmount            float32                     `json:"discount_amount"`
	DiscountPercentage        float32                     `json:"discount_percentage"`
	PaymentMethod             string                      `json:"payment_method"`
	PaymentStatus             string                      `json:"payment_status"`
	Notes                     string                      `json:"notes"`
	TableNumber               string                      `json:"table_number"`
	Products                  []OrderItem                 `json:"products"`
	Platform                  string                      `json:"platform"`
	PlatformOrderID           string                      `json:"platform_order_id"`
	PlatformOrderState        string                      `json:"platform_order_state"`
	OrderDetails              models.OrderDetails         `json:"order_details"`
	AmountReceived            float32                     `json:"amount_received"`
	RoundedAmount             float32                     `json:"rounded_amount"`
	RoundedNetTotal           float32                     `json:"rounded_net_total"`
	VoucherDiscountAmount     *float32                    `json:"voucher_discount_amount"`
	VoucherDiscountType       *models.VoucherDiscountType `json:"voucher_discount_type"`
	VoucherDiscountPercentage *float32                    `json:"voucher_discount_percentage"`
	OutletName                string                      `json:"outlet_name"`
	PickupAt                  *time.Time                  `json:"pickup_at"`
	//Products                []CustomOrderItem `json:"products"`
}

type OrderItem struct {
	ID                   uuid.UUID             `json:"order_item_id"`
	ProductID            uuid.UUID             `json:"product_id"`
	Product              models.Product        `json:"product" gorm:"foreignKey:ProductID"`
	Quantity             int                   `json:"quantity"`
	UnitPrice            float32               `json:"unit_price"`
	SubTotal             float32               `json:"sub_total"`
	ItemNotes            string                `json:"item_notes"`
	CustomModifierGroups []CustomModifierGroup `json:"custom_modifier_groups"`
	//CustomCustomizationGroups []CustomCustomizationGroup `json:"customization_groups"`
}

type CustomModifierGroup struct {
	ID                      uuid.UUID                `json:"id"`
	Name                    string                   `json:"modifier_group_name"`
	InputType               models.InputType         `json:"input_type"`
	SelectedModifierOptions []SelectedModifierOption `json:"selected_modifier_options"`
}

type SelectedModifierOption struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	ModifierOptionQuantity int       `json:"modifier_option_quantity"`
	//Description     string    `json:"description"`
	PriceAdjustment float32 `json:"price_adjustment"`
}

// convertStringToVoucherDiscountType converts *string to *models.VoucherDiscountType
func convertStringToVoucherDiscountType(s *string) *models.VoucherDiscountType {
	if s == nil {
		return nil
	}
	discountType := models.VoucherDiscountType(*s)
	return &discountType
}

func DetermineOrderType(orderType string) string {
	switch strings.ToLower(orderType) {
	case "pickup", "pick up":
		return models.OrderTypePickup
	case "pickup_later", "pickup later":
		return models.OrderTypePickupLater
	case "delivery":
		return models.OrderTypeDelivery
	case "dine_in", "dine in":
		return models.OrderTypeDineIn
	case "take_away", "take away":
		return models.OrderTypeTakeAway
	case "other":
		return models.OrderTypeOther
	default:
		return models.OrderTypeOther
	}
}

// create order from POS/GRABFOOD
func CreateOrder(trx *gorm.DB, req *CreateOrderRequest) (*models.Order, error) {
	var outlet models.Outlet
	result := trx.Where("id=?", req.OutletID).First(&outlet)
	if result.Error != nil {
		return nil, result.Error
	}

	// Generate order number
	todayFull := time.Now()
	orderDate := todayFull
	if req.OrderDate != nil {
		orderDate = *req.OrderDate
	}

	orderNumber, err := GenerateOrderNumber(trx, req.OutletID)
	if err != nil {
		return nil, err
	}

	pickupAt := req.PickupAt

	orderType := DetermineOrderType(req.OrderType)

	// 1. Create order
	order := &models.Order{
		BusinessID:                outlet.BusinessID,
		CustomerID:                req.CustomerID,
		OutletID:                  req.OutletID,
		OrderNumber:               orderNumber,
		OrderDate:                 orderDate,
		OrderType:                 orderType,
		InvoiceNumber:             req.InvoiceNumber,
		InvoiceDate:               todayFull,
		OrderStatus:               models.OrderStatusPending,
		GrossTotal:                req.GrossTotal,
		NetTotal:                  req.NetTotal,
		RoundedAmount:             req.RoundedAmount,
		RoundedNetTotal:           req.RoundedNetTotal,
		ServiceCharge:             req.ServiceCharge,
		ServiceChargePercentage:   req.ServiceChargePercentage,
		TaxCharge:                 req.TaxCharge,
		TaxPercentage:             req.TaxPercentage,
		DiscountType:              &req.DiscountType,
		DiscountAmount:            req.DiscountAmount,
		DiscountPercentage:        req.DiscountPercentage,
		PaymentMethod:             req.PaymentMethod,
		PaymentStatus:             models.PaymentStatusPending,
		Notes:                     req.Notes,
		TableNumber:               req.TableNumber,
		Platform:                  req.Platform,
		PlatformOrderID:           req.PlatformOrderID,
		PlatformState:             req.PlatformState,
		VoucherDiscountAmount:     req.VoucherDiscountAmount,
		VoucherDiscountType:       convertStringToVoucherDiscountType(req.VoucherDiscountType),
		VoucherDiscountPercentage: req.VoucherDiscountPercentage,
		PickupAt:                  pickupAt,
		CreatedAt:                 todayFull,
	}

	if req.UserID != uuid.Nil {
		order.UserID = &req.UserID
	}

	if req.CustomerID != nil {
		order.CustomerID = req.CustomerID
	}

	result = trx.Create(order)
	if result.Error != nil {
		return nil, result.Error
	}

	// Normalize zero UUIDs to nil for foreign key constraints
	voucherID := req.VoucherID
	if voucherID != nil && *voucherID == uuid.Nil {
		voucherID = nil
	}
	customerVoucherID := req.CustomerVoucherID
	if customerVoucherID != nil && *customerVoucherID == uuid.Nil {
		customerVoucherID = nil
	}

	// 2. create order details
	orderDetails := &models.OrderDetails{
		OrderID:                 order.ID,
		VoucherID:               req.VoucherID,
		CustomerVoucherID:       req.CustomerVoucherID,
		CustomerName:            req.CustomerName,
		CustomerPhone:           req.CustomerPhone,
		CustomerAddress:         req.CustomerAddress,
		CustomerLatitude:        req.CustomerLatitude,
		CustomerLongitude:       req.CustomerLongitude,
		EstimatedOrderReadyTime: req.EstimatedOrderReadyTime,
		MaxOrderReadyTime:       req.MaxOrderReadyTime,
		NewOrderReadyTime:       req.NewOrderReadyTime,
		GrabShortOrderNum:       req.GrabShortOrderNum,
		ShopeeFoodShortOrderNum: req.ShopeeFoodShortOrderNum,
		CreatedAt:               todayFull,
		UpdatedAt:               nil,
		DeletedAt:               gorm.DeletedAt{},
	}

	result = trx.Create(orderDetails)
	if result.Error != nil {
		return nil, result.Error
	}

	// 3. Create order item
	for _, product := range req.CartItems {
		orderItem := &models.OrderItem{
			OrderID:   order.ID,
			ProductID: product.ID,
			Quantity:  product.Quantity,
			UnitPrice: product.UnitPrice,
			SubTotal:  product.SubTotal,
			CreatedAt: todayFull,
			UpdatedAt: nil,
			DeletedAt: gorm.DeletedAt{},
		}

		result = trx.Create(orderItem)
		if result.Error != nil {
			return nil, result.Error
		}

		if len(product.SelectedModifierGroups) > 0 {
			for _, modifierGroup := range product.SelectedModifierGroups {
				for _, modifierOption := range modifierGroup.Options {
					selectedModifierGroup := &models.SelectedModifierGroup{
						OrderItemID:            orderItem.ID,
						ModifierGroupID:        modifierGroup.ID,
						ModifierOptionsID:      modifierOption.ID,
						ModifierOptionQuantity: modifierOption.SelectedQuantity,
						CreatedAt:              todayFull,
						UpdatedAt:              nil,
						DeletedAt:              gorm.DeletedAt{},
					}
					result = trx.Create(selectedModifierGroup)
					if result.Error != nil {
						return nil, result.Error
					}
				}
			}

		}
	}

	/* transactionNumberInt, err := strconv.Atoi(orderNumberCount)
	if err != nil {
		return nil, err
	}
	transactionNumber := fmt.Sprintf("%s-%s-%06d", "TXN", time.Now().Format("20060102"), transactionNumberInt) */
	// if payment method is cash, use rounded net total, otherwise use net total
	var amount float32
	/* if order.PaymentMethod == string(constants.PaymentMethodCash) {
		amount = order.RoundedNetTotal
	} else {
		amount = order.NetTotal
	} */
	amount = order.RoundedNetTotal

	// 4. Create transaction
	transction := &models.Transaction{
		OrderID:           order.ID,
		TransactionNumber: nil,
		TransactionDate:   todayFull,
		Amount:            amount,
		PaymentMethod:     order.PaymentMethod,
		PaymentStatus:     order.PaymentStatus,
		MolTransactionID:  nil,
		//ErrorCode:         "",
		CreatedAt: todayFull,
		UpdatedAt: nil,
		DeletedAt: gorm.DeletedAt{},
	}
	result = trx.Create(transction)
	if result.Error != nil {
		return nil, result.Error
	}

	// digital_signage removed

	return order, nil
}

// get order by order number
func GetOrderByID(trx *gorm.DB, order_id uuid.UUID, preloadOutlet bool, preloadOrderDetails bool) (*models.Order, error) {
	var order models.Order
	query := trx.Model(&models.Order{}).Where("id = ?", order_id)
	if preloadOutlet {
		query = query.Preload("Outlet")
	}
	if preloadOrderDetails {
		query = query.Preload("OrderDetails")
	}
	result := query.First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

// get transaction by order id
func GetTransactionByOrderID(trx *gorm.DB, order_id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	result := trx.Model(&models.Transaction{}).Where("order_id = ?", order_id).Preload("Order").Preload("Order.OrderDetails").First(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}
	return &transaction, nil
}

func CompleteCashPayment(trx *gorm.DB, ctx context.Context, req *CompleteCashPaymentRequest) (*models.Transaction, error) {
	var transaction models.Transaction
	result := trx.Where("order_id = ?", req.OrderID).Preload("Order").First(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}

	// check if the payment method is cash / static QR
	/* if transaction.Order.PaymentMethod != models.CasePaymentCash && transaction.Order.PaymentMethod != models.CasePaymentStaticQR {
		return nil, errors.New("payment method is not cash or static QR")
	} */

	// check if the amount received is equal to the order total amount
	if float32(req.AmountReceived) < float32(transaction.Order.RoundedNetTotal) && float32(req.AmountReceived) <= float32(transaction.Amount) {
		return nil, errors.New("amount received is smaller than the net total")
	}

	// payment.GenerateTransactionNumber removed - setting to nil or dummy
	transactionNumber := fmt.Sprintf("TXN-%s", transaction.Order.OrderNumber)
	transaction.TransactionNumber = &transactionNumber

	var businessConfig models.BusinessConfiguration
	businessID := transaction.Order.BusinessID
	result = trx.Where("business_id = ?", businessID).First(&businessConfig)
	if result.Error != nil {
		return nil, result.Error
	}

	// determine the order status based on the business configuration
	// whether to auto complete order on payment success
	autoCompleteOrderOnPaymentSuccess := businessConfig.AutoCompleteOrderOnPaymentSuccess
	if autoCompleteOrderOnPaymentSuccess != nil && *autoCompleteOrderOnPaymentSuccess == true {
		transaction.Order.OrderStatus = models.OrderStatusCompleted
	} else {
		transaction.Order.OrderStatus = models.OrderStatusPending
	}

	// update the order payment status to completed
	transaction.Order.PaymentStatus = models.PaymentStatusCompleted

	// update the transaction payment status to completed
	transaction.PaymentStatus = models.PaymentStatusCompleted

	// update the amount received
	transaction.Order.AmountReceived = &req.AmountReceived

	// if auto complete order is true, set the order status to completed
	if req.AutoCompleteOrder {
		transaction.Order.OrderStatus = models.OrderStatusCompleted
	}

	// Save the updated Order object
	result = trx.Save(&transaction.Order)
	if result.Error != nil {
		return nil, result.Error
	}

	// Save the updated Transaction object
	result = trx.Save(&transaction)
	if result.Error != nil {
		return nil, result.Error
	}

	ProcessSale(trx, ctx, transaction.OrderID)

	return &transaction, nil
}

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

// insert notification into database
func InsertNotification(trx *gorm.DB, ctx context.Context, req *models.Notification) (*models.Notification, error) {
	fmt.Println("start insert notification in insert notification")
	notification := models.Notification{
		OutletID:         req.OutletID,
		UserID:           req.UserID,
		FCMDeviceToken:   req.FCMDeviceToken,
		Title:            req.Title,
		Body:             req.Body,
		Data:             nil,
		NotificationType: req.NotificationType,
		IsRead:           req.IsRead,
		ActionURL:        req.ActionURL,
		ImageURL:         req.ImageURL,
		ExpiredAt:        req.ExpiredAt,
	}

	result := trx.Create(&notification)
	if result.Error != nil {
		return nil, result.Error
	}

	return &notification, nil
}

// get all notification for a user
func GetNotificationForUser(trx *gorm.DB, ctx context.Context, userID uuid.UUID, outletID uuid.UUID, offset int, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	result := trx.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications)
	if result.Error != nil {
		return nil, result.Error
	}

	return notifications, nil
}

// condition to determine the order type
// Returns a function that applies the order type filter using parameterized queries
func ApplyOrderTypeFilter(query *gorm.DB, order_type string) *gorm.DB {
	switch order_type {
	case "take_away", "Take Away", "take away":
		return query.Where("order_type IN (?)", []string{"take_away", "take away", "Take Away"})
	case "delivery", "Delivery":
		return query.Where("order_type IN (?)", []string{"delivery", "Delivery"})
	case "pickup", "Pickup", "pickup_later", "Pickup Later", "pick up", "Pick Up":
		return query.Where("order_type IN (?)", []string{"pickup", "Pickup", "pick up", "Pick Up", "pickup_later", "Pickup Later", "pickup later"})
	case "dine_in", "Dine In", "dine in":
		return query.Where("order_type IN (?)", []string{"dine_in", "dine in", "Dine In"})
	case "online_order", "Online Order", "online order":
		return query.Where("platform IN (?)", []models.Platform{models.PlatformGrabFood, models.PlatformShopeeFood, models.PlatformMembershipApp})
	case "other", "Other":
		return query.Where("order_type IN (?)", []string{"other", "Other"})
	default:
		return query
	}
}

// get next order status based on the current order status
// Sequence: pending / preparing -> ready -> collected / on_the_way / delivered / completed / cancelled
func GetNextOrderStatus(currentOrderStatus string) string {
	switch currentOrderStatus {
	case models.OrderStatusPending, models.OrderStatusPreparing:
		return models.OrderStatusReady
	case models.OrderStatusReady:
		// After ready, status depends on order type - default to completed
		return models.OrderStatusCompleted
	case models.OrderStatusOnTheWay:
		return models.OrderStatusDelivered
	case models.OrderStatusDelivered, models.OrderStatusCollected:
		return models.OrderStatusCompleted
	case models.OrderStatusCompleted, models.OrderStatusCancelled:
		// Terminal states - no next status
		return ""
	default:
		return ""
	}
}

// apply filter to the query based on the request
func ApplyOrderFilter(query *gorm.DB, req *ApplyOrderFilterRequest) *gorm.DB {
	log.Println("Start apply order filter")
	// ensure order status is use only if order status is specified (other include attributes will not be used)
	// ensure payment status is use only if payment status is specified (other include attributes will not be used)
	isOrderStatusAvailable := req.OrderStatus != nil && *req.OrderStatus != ""
	isPaymentStatusAvailable := req.PaymentStatus != nil && *req.PaymentStatus != ""

	if isOrderStatusAvailable {
		query = query.Where("order_status = ?", req.OrderStatus)
	}

	// Collect order statuses that should be included
	if !isOrderStatusAvailable {
		var orderStatuses []string
		if req.IncludeOrderCompleted != nil && *req.IncludeOrderCompleted {
			orderStatuses = append(orderStatuses, models.OrderStatusCompleted)
		}
		if req.IncludeOrderCancelled != nil && *req.IncludeOrderCancelled {
			orderStatuses = append(orderStatuses, models.OrderStatusCancelled)
		}
		if req.IncludeOrderPending != nil && *req.IncludeOrderPending {
			orderStatuses = append(orderStatuses, models.OrderStatusPending)
		}
		if req.IncludeOrderReady != nil && *req.IncludeOrderReady {
			orderStatuses = append(orderStatuses, models.OrderStatusReady)
		}
		if req.IncludeOrderCollected != nil && *req.IncludeOrderCollected {
			orderStatuses = append(orderStatuses, models.OrderStatusCollected)
		}
		if req.IncludeOrderOnTheWay != nil && *req.IncludeOrderOnTheWay {
			orderStatuses = append(orderStatuses, models.OrderStatusOnTheWay)
		}
		if req.IncludeOrderDelivered != nil && *req.IncludeOrderDelivered {
			orderStatuses = append(orderStatuses, models.OrderStatusDelivered)
		}
		if len(orderStatuses) > 0 {
			query = query.Where("order_status IN ?", orderStatuses)
		}
	}

	if isPaymentStatusAvailable && !isOrderStatusAvailable {
		query = query.Where("payment_status = ?", req.PaymentStatus)
	}

	// Collect payment statuses that should be included
	if !isPaymentStatusAvailable {
		var paymentStatuses []string
		if req.IncludePaymentCompleted != nil && *req.IncludePaymentCompleted {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusCompleted)
		}
		if req.IncludePaymentFailed != nil && *req.IncludePaymentFailed {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusFailed)
		}
		if req.IncludePaymentRefunded != nil && *req.IncludePaymentRefunded {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusRefunded)
		}
		if req.IncludePaymentVoided != nil && *req.IncludePaymentVoided {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusVoided)
		}
		if req.IncludePaymentPending != nil && *req.IncludePaymentPending {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusPending)
		}
		if req.IncludePaymentPendingAuthorize != nil && *req.IncludePaymentPendingAuthorize {
			paymentStatuses = append(paymentStatuses, models.PaymentStatusPendingAuthorize)
		}
		if len(paymentStatuses) > 0 {
			query = query.Where("payment_status IN ?", paymentStatuses)
		}
	}

	// set the start date and end date to 00:00:00 and 23:59:59
	if req.StartDate != nil {
		startDate := common.GetStartOfDayByTruncate(*req.StartDate)
		req.StartDate = &startDate
	}
	if req.EndDate != nil {
		endDate := common.GetEndOfDayByTruncate(*req.EndDate)
		req.EndDate = &endDate
	}
	// add start date and end date sql query if it is not empty
	if req.StartDate != nil && req.EndDate != nil {
		query = query.Where("order_date BETWEEN ? AND ?", *req.StartDate, *req.EndDate)
	}

	// only add additional order type sql query if it is not empty
	if req.OrderType != nil && *req.OrderType != "" {
		query = ApplyOrderTypeFilter(query, *req.OrderType)
	}

	if req.OutletID != uuid.Nil {
		query = query.Where("outlet_id = ?", req.OutletID)
	}

	// only add search key sql query if it is not empty
	if req.SearchKey != nil && *req.SearchKey != "" {
		searchPattern := "%" + *req.SearchKey + "%"
		query = query.Where("order_number = ? OR SUBSTRING(order_number FROM '[^-]+$') LIKE ?", searchPattern, searchPattern)
	}
	log.Println("query", query.ToSQL(func(tx *gorm.DB) *gorm.DB { return tx }))
	log.Println("End apply order filter")
	return query
}

// mapping order to custom order
func MapOrderToCustomOrder(trx *gorm.DB, orders []models.Order) ([]CustomOrder, error) {
	var customOrders []CustomOrder
	for _, order := range orders {
		// order details (to ensure that return empty object if is nil/ not found)
		var orderDetails models.OrderDetails
		if order.OrderDetails != nil {
			orderDetails = *order.OrderDetails
		} else {
			orderDetails = models.OrderDetails{}
		}

		// based on the each order, get the order items (products)
		var orderItems []models.OrderItem
		result := trx.Where("order_id = ?", order.ID).
			Find(&orderItems)
		if result.Error != nil {
			return nil, fmt.Errorf(
				"failed to retrieve order items: %w",
				result.Error,
			)
		}

		// Create a slice to hold our response items
		productItemsWithCustomizationGroups := make([]OrderItem, 0, len(orderItems))
		var customOrderItems []OrderItem

		// based on the each order item, get the product
		for i, orderItem := range orderItems {
			var product models.Product
			result = trx.Where("id = ?", orderItem.ProductID).First(&product)
			if result.Error != nil {
				return nil, fmt.Errorf(
					"failed to retrieve product: %w",
					result.Error,
				)
			}
			customOrderItems = append(customOrderItems, OrderItem{
				ID:                   orderItem.ProductID,
				Quantity:             orderItem.Quantity,
				UnitPrice:            orderItem.UnitPrice,
				SubTotal:             orderItem.SubTotal,
				Product:              product,
				ItemNotes:            orderItem.ItemNotes,
				CustomModifierGroups: []CustomModifierGroup{},
				//CustomCustomizationGroups: []CustomCustomizationGroup{},
			})

			// Get selected customization groups for this order item
			var selectedGroups []models.SelectedModifierGroup
			result = trx.Where("order_item_id = ?", orderItem.ID).Find(&selectedGroups)
			if result.Error != nil {
				return nil, result.Error
			}

			// Map of product_customization_groups_id -> CustomCustomizationGroup
			modifierGroupMap := make(map[uuid.UUID]CustomModifierGroup)

			// For each selected group
			for _, selectedGroup := range selectedGroups {
				// Get group details if we haven't already
				group, exists := modifierGroupMap[selectedGroup.ModifierGroupID]
				if !exists {
					// Get the group details
					var dbGroup models.ModifierGroups
					result = trx.Where("id = ?", selectedGroup.ModifierGroupID).First(&dbGroup)
					if result.Error != nil {
						continue
					}

					group = CustomModifierGroup{
						ID:                      dbGroup.ID,
						Name:                    dbGroup.Name,
						InputType:               dbGroup.InputType,
						SelectedModifierOptions: []SelectedModifierOption{},
						//SelectedCustomizationOptions: []SelectedCustomizationOption{},
					}
					modifierGroupMap[selectedGroup.ModifierGroupID] = group
				}

				// Get option details
				var option models.ModifierOptions
				result = trx.Where("id = ?", selectedGroup.ModifierOptionsID).First(&option)
				if result.Error != nil {
					continue // Skip if not found
				}

				// Add option to the group
				modifierOption := SelectedModifierOption{
					ID:                     option.ID,
					Name:                   option.Name,
					PriceAdjustment:        option.PriceAdjustment,
					ModifierOptionQuantity: selectedGroup.ModifierOptionQuantity,
				}

				// Update the group with the new option
				group = modifierGroupMap[selectedGroup.ModifierGroupID]
				group.SelectedModifierOptions = append(group.SelectedModifierOptions, modifierOption)
				modifierGroupMap[selectedGroup.ModifierGroupID] = group
			}

			// Convert the map to a slice
			for _, group := range modifierGroupMap {
				customOrderItems[i].CustomModifierGroups = append(customOrderItems[i].CustomModifierGroups, group)
				//orderItem.CustomCustomizationGroups = append(orderItem.CustomCustomizationGroups, group)
			}

			// Add the completed item to our response
			productItemsWithCustomizationGroups = append(productItemsWithCustomizationGroups, customOrderItems[i])

		}

		var amountReceived float32
		if order.AmountReceived != nil {
			amountReceived = *order.AmountReceived
		} else {
			amountReceived = 0
		}

		customOrders = append(customOrders, CustomOrder{
			ID:                        order.ID,
			OrderNumber:               order.OrderNumber,
			OrderDate:                 order.OrderDate,
			OrderType:                 order.OrderType,
			InvoiceNumber:             order.InvoiceNumber,
			InvoiceDate:               order.InvoiceDate,
			OrderStatus:               order.OrderStatus,
			GrossTotal:                order.GrossTotal,
			NetTotal:                  order.NetTotal,
			ServiceCharge:             order.ServiceCharge,
			ServiceChargePercentage:   order.ServiceChargePercentage,
			TaxCharge:                 order.TaxCharge,
			TaxPercentage:             order.TaxPercentage,
			DiscountType:              *order.DiscountType,
			DiscountAmount:            order.DiscountAmount,
			DiscountPercentage:        order.DiscountPercentage,
			PaymentMethod:             order.PaymentMethod,
			PaymentStatus:             order.PaymentStatus,
			Notes:                     order.Notes,
			TableNumber:               order.TableNumber,
			Products:                  productItemsWithCustomizationGroups,
			Platform:                  common.SafeString((*string)(order.Platform)),
			PlatformOrderID:           common.SafeString((*string)(order.PlatformOrderID)),
			PlatformOrderState:        common.SafeString((*string)(order.PlatformState)),
			OrderDetails:              orderDetails,
			AmountReceived:            amountReceived,
			RoundedAmount:             order.RoundedAmount,
			RoundedNetTotal:           order.RoundedNetTotal,
			OutletName:                order.Outlet.Name,
			PickupAt:                  order.PickupAt,
			VoucherDiscountAmount:     order.VoucherDiscountAmount,
			VoucherDiscountType:       order.VoucherDiscountType,
			VoucherDiscountPercentage: order.VoucherDiscountPercentage,
		})
	}
	return customOrders, nil
}
