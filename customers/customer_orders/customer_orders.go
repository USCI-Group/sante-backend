package customer_orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/common/payment"
	"encore.app/identity"
	"encore.app/sante_admin/products"
	"encore.app/customers/customer_common"
	"encore.app/customers/customer_payments"
	"encore.app/database"
	"encore.app/database/models"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

var secretsKeys struct {
	jwtSecretKey string
}

// initService initializes the user service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	secretsKeys.jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

	// Create the service first
	service := &Service{db: db}

	// Now call the method on the created service
	err = service.StartCustomerOrdersAutoCancelJobWorkers()
	if err != nil {
		log.Println("Error starting customer orders auto cancel job workers:", err)
		return nil, err
	}
	log.Println("Customer orders auto cancel job workers started successfully")
	return service, nil
}

type GetAllOutletsRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	State      string    `json:"state"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

type GetAllOutletsResponse struct {
	Outlets []models.Outlet `json:"outlets"`
}

type GetAllProductsFromOutletResponse struct {
	ProductCategories  []ProductCategory    `json:"product_categories"`
	OutletStatus       *models.OutletStatus `json:"outlet_status"`
	OnlineOrderEnabled *bool                `json:"online_order_enabled"`
}

type ProductCategory struct {
	Category models.ProductCategory `json:"category"`
	Products []ProductWithFavourite `json:"products"`
}

type ProductWithFavourite struct {
	Product            models.Product `json:"product"`
	IsFavourite        bool           `json:"is_favourite"`
	IsActiveInOutlet   *bool          `json:"is_active_in_outlet,omitempty"`
	IsActiveInBusiness *bool          `json:"is_active_in_business,omitempty"`
}

type GetAllFavouritesProductsResponse struct {
	Products []ProductWithFavourite `json:"products"`
}

type GetModifiersBasedOnProductIDResponse struct {
	Modifiers []CustomModifierGroupWithModifiers `json:"modifiers"`
}

type CustomModifierGroupWithModifiers struct {
	ModifierGroup   models.ModifierGroups    `json:"modifier_group"`
	MaxSelection    int                      `json:"max_selection"`
	ModifierOptions []models.ModifierOptions `json:"modifier_options"`
}

type GetTaxConfigurationResponse struct {
	ServiceChargePercentage float32 `json:"service_charge_percentage"`
	TaxPercentage           float32 `json:"tax_percentage"`
	IsTaxIncludedInPrice    bool    `json:"is_tax_included_in_price"`
}

type CreateOrderRequest struct {
	OutletID                  uuid.UUID                   `json:"outlet_id"`
	CustomerID                *uuid.UUID                  `json:"customer_id"`
	UserID                    uuid.UUID                   `json:"user_id"`
	OrderType                 string                      `json:"order_type" valid:"required~Order type is required"`
	InvoiceNumber             string                      `json:"invoice_number" valid:"required~Invoice number is required"`
	OrderStatus               string                      `json:"order_status" valid:"required~Order status is required"`
	OrderDate                 *time.Time                  `json:"order_date"`
	GrossTotal                float32                     `json:"gross_total" valid:"required~Gross total is required"`
	NetTotal                  float32                     `json:"net_total" valid:"required~Net total is required"`
	RoundedAmount             float32                     `json:"rounded_amount" `
	RoundedNetTotal           float32                     `json:"rounded_net_total" `
	DeliveryFee               float32                     `json:"delivery_fee" `
	TaxCharge                 float32                     `json:"tax_charge" `
	TaxPercentage             float32                     `json:"tax_percentage" `
	ServiceCharge             float32                     `json:"service_charge" `
	ServiceChargePercentage   float32                     `json:"service_charge_percentage" `
	DiscountType              string                      `json:"discount_type"`
	DiscountAmount            float32                     `json:"discount_amount" `
	DiscountPercentage        float32                     `json:"discount_percentage" `
	VoucherDiscountAmount     *float32                    `json:"voucher_discount_amount"`
	VoucherDiscountType       *models.VoucherDiscountType `json:"voucher_discount_type"`
	VoucherDiscountPercentage *float32                    `json:"voucher_discount_percentage"`
	PaymentMethod             string                      `json:"payment_method" valid:"required~Payment method is required"`
	PaymentChannel            string                      `json:"payment_channel"`
	Notes                     string                      `json:"notes"`
	TableNumber               string                      `json:"table_number"`
	// order items
	CartItems []CartItem `json:"cart_items" valid:"required~Cart items are required"`
	PickupAt  *time.Time `json:"pickup_at"`
	// platform info, grabfood
	Platform                *models.Platform `json:"platform"`
	PlatformOrderID         *string          `json:"platform_order_id"`
	PlatformState           *string          `json:"platform_state"`
	CustomerName            *string          `json:"customer_name"`
	CustomerPhone           *string          `json:"customer_phone"`
	CustomerAddress         *string          `json:"customer_address"`
	CustomerLatitude        *float32         `json:"customer_latitude"`
	CustomerLongitude       *float32         `json:"customer_longitude"`
	// Student Info for School-Based Outlets
	StudentID               string           `json:"student_id"`
	ParentPhone             string           `json:"parent_phone"`
	EstimatedOrderReadyTime *time.Time       `json:"estimated_order_ready_time"`
	MaxOrderReadyTime       *time.Time       `json:"max_order_ready_time"`
	NewOrderReadyTime       *time.Time       `json:"new_order_ready_time"`
	GrabShortOrderNum       *string          `json:"grab_short_order_num"`
	// voucher
	VoucherApplied  *VoucherApplied `json:"voucher_applied"`
	VoucherDiscount *float32        `json:"voucher_discount"`
}

type CartItem struct {
	// this id is product id
	ID                     uuid.UUID               `json:"id" `
	Quantity               int                     `json:"quantity" valid:"required~Quantity is required"`
	UnitPrice              float32                 `json:"unit_price" valid:"required~Unit price is required"`
	SubTotal               float32                 `json:"sub_total" valid:"required~Sub total is required"`
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

// voucher use in ordering in membership app may using voucher redeemed first or directly apply voucher
type VoucherApplied struct {
	VoucherID            *uuid.UUID `json:"voucher_id"`
	VoucherCode          *string    `json:"voucher_code"`
	CustomerVoucherID    *uuid.UUID `json:"customer_voucher_id"`
	CustomerVoucherCode  *string    `json:"customer_voucher_code"`
	IsDirectApplyVoucher bool       `json:"is_direct_apply_voucher"`
}

type CreateOrderResponse struct {
	Message           string    `json:"message"`
	OrderID           uuid.UUID `json:"order_id"`
	PaymentLink       string    `json:"payment_link"`
	PaymentMethod     string    `json:"payment_method"`
	PaymentURL        string    `json:"payment_url"`
	TransactionID     string    `json:"transaction_id"`
	TransactionNumber string    `json:"transaction_number"`
}

type GetAllOrdersResponse struct {
	Orders  []CustomerOrderResponse `json:"orders"`
	MaxPage int                     `json:"max_page"`
}

type CustomerOrderResponse struct {
	ID                      uuid.UUID           `json:"id"`
	Outlet                  models.Outlet       `json:"outlet"`
	OrderNumber             string              `json:"order_number"`
	OrderDate               time.Time           `json:"order_date"`
	OrderType               string              `json:"order_type"`
	InvoiceNumber           string              `json:"invoice_number"`
	InvoiceDate             time.Time           `json:"invoice_date"`
	OrderStatus             string              `json:"order_status"`
	Platform                *string             `json:"platform"`
	PlatformOrderID         *string             `json:"platform_order_id"`
	PlatformState           *string             `json:"platform_state"`
	GrossTotal              float32             `json:"gross_total"`
	NetTotal                float32             `json:"net_total"`
	RoundedAmount           float32             `json:"rounded_amount"`
	RoundedNetTotal         float32             `json:"rounded_net_total"`
	DeliveryFee             float32             `json:"delivery_fee"`
	AmountReceived          *float32            `json:"amount_received"`
	TaxCharge               float32             `json:"tax_charge"`
	TaxPercentage           float32             `json:"tax_percentage"`
	ServiceCharge           float32             `json:"service_charge"`
	ServiceChargePercentage float32             `json:"service_charge_percentage"`
	DiscountType            string              `json:"discount_type"`
	DiscountAmount          float32             `json:"discount_amount"`
	DiscountPercentage      float32             `json:"discount_percentage"`
	PaymentMethod           string              `json:"payment_method"`
	PaymentStatus           string              `json:"payment_status"`
	Notes                   string              `json:"notes"`
	TableNumber             string              `json:"table_number"`
	EInvoiceSubmissionID    *string             `json:"e_invoice_submission_id"`
	EInvoiceStatus          *string             `json:"e_invoice_status"`
	EInvoiceURL             *string             `json:"e_invoice_url"`
	EInvoiceRejectedReason  *string             `json:"e_invoice_rejected_reason"`
	OrderDetails            models.OrderDetails `json:"order_details"`
	OrderItems              []models.OrderItem  `json:"order_items"`
	Customer                *models.Customer    `json:"customer"`
	PointEarned             *int                `json:"point_earned"`
	PickupAt                *time.Time          `json:"pickup_at"`
	CompletedAt             *time.Time          `json:"completed_at"`
	PointsRewarded          *int                `json:"points_rewarded"`
	PointsRewardedAt        *time.Time          `json:"points_rewarded_at"`
	ExpRewarded             *int                `json:"exp_rewarded"`
	ExpRewardedAt           *time.Time          `json:"exp_rewarded_at"`
	CreatedAt               time.Time           `json:"created_at"`
	UpdatedAt               *time.Time          `json:"updated_at"`
}

type GetOrderResponse struct {
	Message string                `json:"message"`
	Order   CustomerOrderResponse `json:"order"`
}

type OrderItemAndSelectedModifier struct {
	OrderItem              models.OrderItem               `json:"order_item"`
	SelectedModifierGroups []models.SelectedModifierGroup `json:"selected_modifier_groups"`
}

type GetProductAndSelectedModifierResponse struct {
	OrderItems []OrderItemAndSelectedModifier `json:"order_items"`
}

type ReorderOrderDetailsRequest struct {
	OrderID    uuid.UUID `json:"order_id"`
	OutletID   uuid.UUID `json:"outlet_id"`
	BusinessID uuid.UUID `json:"business_id"`
	CustomerID uuid.UUID `json:"customer_id"`
}

type ReorderOrderDetailsResponse struct {
	Message  string      `json:"message"`
	IsChange bool        `json:"is_change"` // if there is any change of information in the order
	IsValid  bool        `json:"is_valid"`  // if the order is valid (if the products list is zero, then it is invalid)
	Order    CustomOrder `json:"order"`
}

type CustomOrder struct {
	ID                      uuid.UUID                 `json:"id"`
	BusinessID              uuid.UUID                 `json:"business_id"`
	Business                *models.Business          `json:"business"`
	OutletID                uuid.UUID                 `json:"outlet_id"`
	Outlet                  *models.Outlet            `json:"outlet"`
	UserID                  *uuid.UUID                `json:"user_id"`
	User                    *models.User              `json:"user"`
	CustomerID              *uuid.UUID                `json:"customer_id"`
	Customer                *models.Customer          `json:"customer"`
	OrderNumber             string                    `json:"order_number"`
	OrderDate               time.Time                 `json:"order_date"`
	OrderType               string                    `json:"order_type"`
	InvoiceNumber           string                    `json:"invoice_number"`
	InvoiceDate             time.Time                 `json:"invoice_date"`
	OrderStatus             string                    `json:"order_status"`      // Order status (pending, payment, completed, cancelled)
	Platform                *models.Platform          `json:"platform"`          // GrabFood Info
	PlatformOrderID         *string                   `json:"platform_order_id"` // GrabFood Info
	PlatformState           *string                   `json:"platform_state"`    // GrabFood Info
	GrossTotal              float32                   `json:"gross_total"`
	NetTotal                float32                   `json:"net_total"`
	RoundedAmount           float32                   `json:"rounded_amount"`
	RoundedNetTotal         float32                   `json:"rounded_net_total"`
	AmountReceived          *float32                  `json:"amount_received"`
	TaxCharge               float32                   `json:"tax_charge" default:"0"`
	TaxPercentage           float32                   `json:"tax_percentage" default:"0"`
	ServiceCharge           float32                   `json:"service_charge" default:"0"`
	ServiceChargePercentage float32                   `json:"service_charge_percentage" default:"0"`
	DiscountType            string                    `json:"discount_type" enum:"fixed,percentage,none"`
	DiscountAmount          float32                   `json:"discount_amount" default:"0"`
	DiscountPercentage      float32                   `json:"discount_percentage" default:"0"`
	PaymentMethod           string                    `json:"payment_method"`
	PaymentChannel          *models.PaymentChannel    `json:"payment_channel"`
	PaymentStatus           string                    `json:"payment_status"`
	Notes                   string                    `json:"notes"`
	TableNumber             string                    `json:"table_number"`
	EInvoiceSubmissionID    *string                   `json:"e_invoice_submission_id"`
	EInvoiceStatus          *constants.EInvoiceStatus `json:"e_invoice_status"`
	EInvoiceURL             *string                   `json:"e_invoice_url"`
	EInvoiceRejectedReason  *string                   `json:"e_invoice_rejected_reason"`
	OrderDetails            *models.OrderDetails      `json:"order_details"`
	OrderItems              []CustomOrderItem         `json:"order_items"`
	CreatedAt               time.Time                 `json:"created_at"`
	UpdatedAt               *time.Time                `json:"updated_at"`
}

type CustomOrderItem struct {
	ID                     uuid.UUID                     `json:"id"`
	OrderID                uuid.UUID                     `json:"order_id"`
	Order                  *models.Order                 `json:"order,omitempty"`
	ProductID              uuid.UUID                     `json:"product_id"`
	Product                models.Product                `json:"product"`
	Quantity               int                           `json:"quantity"`
	UnitPrice              float32                       `json:"unit_price"`
	SubTotal               float32                       `json:"sub_total"`
	ItemNotes              string                        `json:"item_notes"`
	SelectedModifierGroups []CustomSelectedModifierGroup `json:"selected_modifier_groups"`
	CreatedAt              time.Time                     `json:"created_at"`
	UpdatedAt              *time.Time                    `json:"updated_at"`
}

type CustomSelectedModifierGroup struct {
	ID                     uuid.UUID              `json:"id"`
	OrderItemID            uuid.UUID              `json:"order_item_id"`
	OrderItem              *CustomOrderItem       `json:"order_item,omitempty"`
	ModifierGroupID        uuid.UUID              `json:"modifier_group_id"`
	ModifierGroup          models.ModifierGroups  `json:"modifier_group"`
	ModifierOptionsID      uuid.UUID              `json:"modifier_options_id"`
	ModifierOptions        models.ModifierOptions `json:"modifier_options"`
	ModifierOptionQuantity int                    `json:"modifier_option_quantity"`
	MaxSelection           int                    `json:"max_selection"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              *time.Time             `json:"updated_at"`
}

type AllProductModifierMappingsRequest struct {
	ProductIDs []uuid.UUID `json:"product_ids"`
}

type AllProductModifierMappingsResponse struct {
	Message                 string                          `json:"message"`
	ProductModifierMappings []models.ProductModifierMapping `json:"product_modifier_mappings"`
}

type GetProductWithFavouriteResponse struct {
	Message string               `json:"message"`
	Data    ProductWithFavourite `json:"data"`
}

type CancelOrderRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	OrderID    uuid.UUID `json:"order_id"`
}

type CompleteOrderRequest struct {
	OrderID uuid.UUID `json:"order_id"`
}

type SendNotificationToMembershipAppRequest struct {
	Order            *models.Order           `json:"order"`
	Title            *string                 `json:"title,omitempty"`
	Body             *string                 `json:"body,omitempty"`
	ActionURL        *string                 `json:"action_url,omitempty"`
	NotificationType models.NotificationType `json:"notification_type"`
	Data             *map[string]string      `json:"data,omitempty"`
}

// Customer API to get all outlets
//
//encore:api auth method=POST path=/api/customers/outlets/all
func (s *Service) GetAllOutlets(ctx context.Context, req *GetAllOutletsRequest) (*GetAllOutletsResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	var outlets []models.Outlet
	query := s.db.Model(&models.Outlet{}).Where("business_id = ?", req.BusinessID).Find(&outlets)
	if req.State != "" {
		query = query.Where("state = ?", req.State)
	}
	query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&outlets)
	return &GetAllOutletsResponse{
		Outlets: outlets,
	}, nil
}

// Customer API to get all outlets
//
//encore:api public method=GET path=/api/customers/outlets/region/:business_id/:region
func (s *Service) GetOutletByRegion(ctx context.Context, business_id uuid.UUID, region string) (*GetAllOutletsResponse, error) {
	var outlets []models.Outlet
	query := s.db.Model(&models.Outlet{}).
		Where("state = ?", region).
		Where("business_id = ?", business_id).
		Find(&outlets)
	if query.Error != nil {
		return nil, query.Error
	}
	return &GetAllOutletsResponse{
		Outlets: outlets,
	}, nil
}

// Customer API to get all products from outlet.
//
//encore:api auth method=GET path=/api/customers/outlets/products/all/:business_id/:outlet_id
func (s *Service) GetAllProductsFromOutlet(ctx context.Context, business_id uuid.UUID, outlet_id uuid.UUID) (*GetAllProductsFromOutletResponse, error) {
	// sync product to outlet
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	customerData, err := customer_common.GetCustomerDataFromAuthData(auth.Data)
	if err != nil {
		return nil, err
	}

	_, err = products.SyncProductToOutletWithTx(ctx, &products.SyncProductToOutletRequest{
		BusinessID: business_id,
		OutletID:   outlet_id,
		ProductID:  uuid.Nil,
		IsAdd:      false,
	}, trx)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	trx.Commit()
	var responseOfGetAllProductsFromOutlet []ProductCategory
	var productCategories []models.ProductCategory
	query := s.db.Model(&models.ProductCategory{}).
		Where("business_id = ?", business_id).
		Find(&productCategories)
	if query.Error != nil {
		return nil, query.Error
	}

	// get all favourite products
	favProducts, err := customer_common.GetAllFavouriteProductsByCustomerID(s.db, customerData.ID, false)
	if err != nil {
		return nil, err
	}
	favProductIDs := make(map[uuid.UUID]bool)
	for _, favProduct := range favProducts {
		favProductIDs[favProduct.ProductID] = true
	}

	// get all products from outlet and map with product category and favourite
	for _, productCategory := range productCategories {
		var productCategoryMappings []models.ProductCategoryMapping
		var products []ProductWithFavourite
		query := s.db.Model(&models.ProductCategoryMapping{}).Where("product_category_id = ?", productCategory.ID).Preload("Product").Find(&productCategoryMappings)
		if query.Error != nil {
			return nil, query.Error
		}
		for _, productCategoryMapping := range productCategoryMappings {
			// check is it in outlets product table
			var outletProduct models.OutletProduct
			// Skip products that are inactive, or not in store outlet
			if !productCategoryMapping.Product.IsActive ||
				!productCategoryMapping.Product.IsStoreOutlet {
				continue
			}
			query := s.db.Model(&models.OutletProduct{}).
				Where("product_id = ?", productCategoryMapping.ProductID).
				Where("outlet_id = ?", outlet_id).
				//Where("is_active = ?", true). // if want hide products that are inactive in the outlet, then uncomment this line
				Preload("Product").
				First(&outletProduct)
			fmt.Println(query.Error)
			// filter products that are inactive
			if !outletProduct.Product.IsActive {
				continue
			}
			if query.Error == nil {
				isFav := favProductIDs[productCategoryMapping.ProductID]
				// Product exists in outlet and is active, so add it
				products = append(products, ProductWithFavourite{
					Product:            productCategoryMapping.Product,
					IsFavourite:        isFav,
					IsActiveInOutlet:   &outletProduct.IsActive,                  // means this product is active in the outlet
					IsActiveInBusiness: &productCategoryMapping.Product.IsActive, // means this product is active in the business
				})
			}
		}
		sort.Slice(products, func(i, j int) bool {
			return products[i].Product.SortOrder < products[j].Product.SortOrder
		})
		responseOfGetAllProductsFromOutlet = append(responseOfGetAllProductsFromOutlet, ProductCategory{
			Category: productCategory,
			Products: products,
		})
	}

	isActiveInOutlet := (*bool)(nil) // pass nil to get all products that are active and inactive in the outlet

	// get all outlet products
	var allOutletProducts []models.OutletProduct
	allOutletProducts, err = customer_common.GetAllProductsFromOutletProduct(
		s.db,
		outlet_id,
		isActiveInOutlet, // is active in outlet (means this product is active in the outlet)
		true,             // is active in business (means this product is active in the business)
		true,             // is store outlet (means this product is belongs to physical store)
		false,            // is grab food (means this product is belongs to GrabFood platform)
		false,            // is shopee food (means this product is belongs to ShopeeFood platform)
	)
	if err != nil {
		return nil, err
	}
	allCategoryID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	favouritesCategoryID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	var favouritesProductsOnly []ProductWithFavourite
	var products []ProductWithFavourite
	for _, allOutletProduct := range allOutletProducts {
		if favProductIDs[allOutletProduct.Product.ID] {
			favouritesProductsOnly = append(favouritesProductsOnly, ProductWithFavourite{
				Product:            allOutletProduct.Product,
				IsFavourite:        true,
				IsActiveInOutlet:   &allOutletProduct.IsActive,         // means this product is active in the outlet
				IsActiveInBusiness: &allOutletProduct.Product.IsActive, // means this product is active in the business
			})
		}
		// Skip products that are inactive
		if !allOutletProduct.Product.IsActive {
			continue
		}
		products = append(products, ProductWithFavourite{
			Product:            allOutletProduct.Product,
			IsFavourite:        favProductIDs[allOutletProduct.Product.ID],
			IsActiveInOutlet:   &allOutletProduct.IsActive,         // means this product is active in the outlet
			IsActiveInBusiness: &allOutletProduct.Product.IsActive, // means this product is active in the business
		})
	}

	// Create the "All" category and prepend it to the slice
	allCategory := ProductCategory{
		Category: models.ProductCategory{
			ID:   allCategoryID,
			Name: "All",
		},
		Products: products,
	}
	// customer own favourites products
	favouritesCategory := ProductCategory{
		Category: models.ProductCategory{
			ID:   favouritesCategoryID,
			Name: "Favourites",
		},
		Products: favouritesProductsOnly,
	}

	// Prepend the "All" category to the response slice
	responseOfGetAllProductsFromOutlet = append([]ProductCategory{favouritesCategory}, responseOfGetAllProductsFromOutlet...)
	responseOfGetAllProductsFromOutlet = append([]ProductCategory{allCategory}, responseOfGetAllProductsFromOutlet...)

	outlet, err := customer_common.GetOutletByID(s.db, outlet_id)
	if err != nil {
		return nil, err
	}
	return &GetAllProductsFromOutletResponse{
		ProductCategories:  responseOfGetAllProductsFromOutlet,
		OutletStatus:       &outlet.OutletStatus,
		OnlineOrderEnabled: &outlet.OnlineOrderEnabled,
	}, nil
}

// Get All products from business (Public API)
//
//encore:api public method=GET path=/api/customers/outlets/products/all/:business_id
func (s *Service) GetAllProductsFromBusiness(ctx context.Context, business_id uuid.UUID) (*GetAllProductsFromOutletResponse, error) {

	var responseOfGetAllProductsFromOutlet []ProductCategory
	var productCategories []models.ProductCategory
	query := s.db.Model(&models.ProductCategory{}).
		Where("business_id = ?", business_id).
		Find(&productCategories)
	if query.Error != nil {
		return nil, query.Error
	}

	// get all products from outlet and map with product category and favourite
	for _, productCategory := range productCategories {
		var productCategoryMappings []models.ProductCategoryMapping
		var products []ProductWithFavourite
		query := s.db.Model(&models.ProductCategoryMapping{}).
			Where("product_category_id = ?", productCategory.ID).
			Preload("Product").
			Find(&productCategoryMappings)
		if query.Error != nil {
			return nil, query.Error
		}
		for _, productCategoryMapping := range productCategoryMappings {
			// Skip products that are inactive, or not in store outlet
			if !productCategoryMapping.Product.IsActive ||
				!productCategoryMapping.Product.IsStoreOutlet {
				continue
			}

			isFav := false
			products = append(products, ProductWithFavourite{
				Product:            productCategoryMapping.Product,
				IsFavourite:        isFav,
				IsActiveInBusiness: &productCategoryMapping.Product.IsActive,
			})
		}
		// ascending sort by sort_order
		sort.Slice(products, func(i, j int) bool {
			return products[i].Product.SortOrder < products[j].Product.SortOrder
		})
		responseOfGetAllProductsFromOutlet = append(responseOfGetAllProductsFromOutlet, ProductCategory{
			Category: productCategory,
			Products: products,
		})
	}
	// get all outlet products
	allProducts, err := customer_common.GetAllProductsFromBusiness(
		s.db,
		business_id,
		true,
		true,
		false,
		false,
	)
	if err != nil {
		return nil, err
	}

	allCategoryID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	var products []ProductWithFavourite
	for _, allProduct := range allProducts {
		products = append(products, ProductWithFavourite{
			Product:            allProduct,
			IsFavourite:        false,
			IsActiveInBusiness: &allProduct.IsActive,
		})
	}
	// Create the "All" category and prepend it to the slice
	allCategory := ProductCategory{
		Category: models.ProductCategory{
			ID:   allCategoryID,
			Name: "All",
		},
		Products: products,
	}

	// Prepend the "All" category to the response slice
	responseOfGetAllProductsFromOutlet = append([]ProductCategory{allCategory}, responseOfGetAllProductsFromOutlet...)

	return &GetAllProductsFromOutletResponse{
		ProductCategories: responseOfGetAllProductsFromOutlet,
	}, nil
}

// Customer API to get all favourites products
//
//encore:api auth method=GET path=/api/customers/outlets/products/favourites/:customer_id
func (s *Service) GetAllFavouritesProducts(ctx context.Context, customer_id uuid.UUID) (*GetAllFavouritesProductsResponse, error) {
	favProducts, err := customer_common.GetAllFavouriteProductsByCustomerID(s.db, customer_id, false)
	if err != nil {
		return nil, err
	}
	var products []ProductWithFavourite
	for _, favProduct := range favProducts {
		products = append(products, ProductWithFavourite{
			Product:     favProduct.Product,
			IsFavourite: true,
		})
	}

	return &GetAllFavouritesProductsResponse{
		Products: products,
	}, nil
}

// Customer API to get modifiers based on product id
//
//encore:api auth method=GET path=/api/customers/outlets/products/modifiers/:outlet_id/:product_id
func (s *Service) GetModifiersBasedOnProductID(ctx context.Context, outlet_id uuid.UUID, product_id uuid.UUID) (*GetModifiersBasedOnProductIDResponse, error) {
	var ProductModifierMappings []models.ProductModifierMapping
	query := s.db.Model(&models.ProductModifierMapping{}).Where("product_id = ?", product_id).Preload("ModifierGroup").Find(&ProductModifierMappings)
	if query.Error != nil {
		return nil, query.Error
	}
	var modifiers []CustomModifierGroupWithModifiers
	for _, ProductModifierMapping := range ProductModifierMappings {
		var modifierOptions []models.ModifierOptions
		var err error
		query := s.db.Model(&models.ModifierOptions{}).Where("modifier_group_id = ?", ProductModifierMapping.ModifierGroupID).Find(&modifierOptions)
		if query.Error != nil {
			return nil, query.Error
		}
		modifierOptions, err = customer_common.CheckModifierOptionIsActiveAtOutletLevel(s.db, outlet_id, modifierOptions)
		if err != nil {
			return nil, err
		}

		modifiers = append(modifiers, CustomModifierGroupWithModifiers{
			ModifierGroup:   *ProductModifierMapping.ModifierGroup,
			ModifierOptions: modifierOptions,
			MaxSelection:    ProductModifierMapping.MaxSelection,
		})
	}
	return &GetModifiersBasedOnProductIDResponse{
		Modifiers: modifiers,
	}, nil
}

// Customer API to get TAX configuration
//
//encore:api auth method=GET path=/api/customers/business/tax-configuration/:business_id
func (s *Service) GetTaxConfiguration(ctx context.Context, business_id uuid.UUID) (*GetTaxConfigurationResponse, error) {
	businessConfiguration, err := customer_common.GetTaxConfigurationByBusinessID(s.db, business_id)
	if err != nil {
		return nil, err
	}
	return &GetTaxConfigurationResponse{
		ServiceChargePercentage: *businessConfiguration.ServiceChargePercentage,
		TaxPercentage:           *businessConfiguration.ServiceTaxPercentage,
		IsTaxIncludedInPrice:    businessConfiguration.IsTaxIncludedInPrice,
	}, nil
}

// Customer API to create order
//
//encore:api auth method=POST path=/api/customers/orders/create
func (s *Service) CustomerCreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	authData := auth.Data()
	customerData, ok := authData.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}
	customerID := customerData.ID
	req.CustomerID = &customerID
	req.UserID = uuid.Nil

	jsonDetails, _ := json.Marshal(req)

	logActivity := &models.ActivityLog{
		Activity: constants.LOG_ACTION_CREATE_ORDER,
		Status:   constants.LOG_STATUS_SUCCESS,
		Details:  fmt.Sprintf("Request: %+v", string(jsonDetails)),
	}
	s.db.Create(logActivity)

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	order, err := CreateOrder(ctx, trx, req)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// mark redeemed voucher as used (after may move to after payment success)
	if req.VoucherApplied != nil && req.VoucherApplied.VoucherID != nil && *req.VoucherApplied.VoucherID != uuid.Nil {
		err = MarkRedeemedVoucherAsUsed(trx, *req.VoucherApplied.CustomerVoucherID, *req.VoucherApplied.CustomerVoucherCode, customerID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	if req.VoucherApplied != nil && req.VoucherApplied.VoucherID != nil && *req.VoucherApplied.VoucherID != uuid.Nil {
		// apply changes on amount of voucher (after may move to after payment success)
		err = ApplyChangesOnAmountOfVoucher(trx, *req.VoucherApplied.VoucherID, order.GrossTotal)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
	}
	paymentReq := customer_payments.InitiatePaymentRequest{
		OrderID:       order.ID,
		PaymentMethod: string(constants.PaymentMethodFPX),
		Provider:      payment.Maybank, // Default for now, could be dynamic
	}

	// get the payment link using the existing transaction
	paymentLink, err := customer_payments.InitiatePaymentWithTx(ctx, &paymentReq, trx)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	err = trx.Commit().Error
	if err != nil {
		return nil, err
	}
	return &CreateOrderResponse{
		Message:           "Order created successfully",
		OrderID:           order.ID,
		PaymentLink:       paymentLink.PaymentURL,
		PaymentMethod:     string(paymentLink.PaymentMethod),
		PaymentURL:        paymentLink.PaymentURL,
		TransactionID:     paymentLink.TransactionID,
		TransactionNumber: paymentLink.TransactionNumber,
	}, nil
}

// create order for customer only this
func CreateOrder(ctx context.Context, trx *gorm.DB, req *CreateOrderRequest) (*models.Order, error) {
	var outlet models.Outlet
	result := trx.Where("id=?", req.OutletID).First(&outlet)
	if result.Error != nil {
		return nil, result.Error
	}

	// SCHOOL INTEGRATION: Trigger identity check if outlet is school-based
	if outlet.IsSchoolOutlet {
		if req.StudentID == "" {
			return nil, &errs.Error{Code: errs.InvalidArgument, Message: "StudentID is required for this school-based outlet"}
		}
		
		// Internal call to identity service proxy
		verifyResp, err := identity.VerifyStudent(ctx, &identity.VerifyRequest{
			StudentID:   req.StudentID,
			ParentPhone: req.ParentPhone,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to verify student identity: %v", err)
		}
		if !verifyResp.IsValid {
			return nil, &errs.Error{Code: errs.PermissionDenied, Message: "Identity verification failed for student " + req.StudentID}
		}
		// Verification successful, proceed to create order
	}

	// Generate order number
	todayFull := time.Now()
	orderDate := todayFull
	if req.OrderDate != nil {
		orderDate = *req.OrderDate
	}

	orderNumber, err := customer_common.GenerateOrderNumber(trx, req.OutletID)
	if err != nil {
		return nil, err
	}
	paymentChannel := models.PaymentChannel(req.PaymentChannel)

	orderType := DetermineOrderType(req.OrderType)

	// 1. Create order
	order := &models.Order{
		BusinessID:                outlet.BusinessID,
		OutletID:                  req.OutletID,
		OrderNumber:               orderNumber,
		OrderDate:                 orderDate,
		OrderType:                 orderType,
		InvoiceNumber:             req.InvoiceNumber,
		InvoiceDate:               todayFull,
		OrderStatus:               models.OrderStatusPending,
		GrossTotal:                req.GrossTotal,
		NetTotal:                  req.NetTotal,
		DeliveryFee:               &req.DeliveryFee,
		RoundedAmount:             req.RoundedAmount,
		RoundedNetTotal:           req.RoundedNetTotal,
		ServiceCharge:             req.ServiceCharge,
		ServiceChargePercentage:   req.ServiceChargePercentage,
		TaxCharge:                 req.TaxCharge,
		TaxPercentage:             req.TaxPercentage,
		DiscountType:              &req.DiscountType,
		DiscountAmount:            req.DiscountAmount,
		DiscountPercentage:        req.DiscountPercentage,
		VoucherDiscountAmount:     req.VoucherDiscountAmount,
		VoucherDiscountType:       req.VoucherDiscountType,
		VoucherDiscountPercentage: req.VoucherDiscountPercentage,
		PaymentMethod:             req.PaymentMethod,
		PaymentStatus:             models.PaymentStatusPending,
		PaymentChannel:            &paymentChannel,
		Notes:                     req.Notes,
		TableNumber:               req.TableNumber,
		Platform:                  req.Platform,
		PlatformOrderID:           req.PlatformOrderID,
		PlatformState:             req.PlatformState,
		PickupAt:                  req.PickupAt,
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

	// 2. create order details
	orderDetails := &models.OrderDetails{
		OrderID:                 order.ID,
		CustomerName:            req.CustomerName,
		CustomerPhone:           req.CustomerPhone,
		CustomerAddress:         req.CustomerAddress,
		CustomerLatitude:        req.CustomerLatitude,
		CustomerLongitude:       req.CustomerLongitude,
		EstimatedOrderReadyTime: req.EstimatedOrderReadyTime,
		MaxOrderReadyTime:       req.MaxOrderReadyTime,
		NewOrderReadyTime:       req.NewOrderReadyTime,
		GrabShortOrderNum:       req.GrabShortOrderNum,
		VoucherID:               req.VoucherApplied.VoucherID,
		CustomerVoucherID:       req.VoucherApplied.CustomerVoucherID,
		// Map Student Info
		StudentID:               &req.StudentID,
		ParentPhone:             &req.ParentPhone,
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
			ItemNotes: product.ItemNotes,
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

	var amount float32
	if order.PaymentMethod == string(constants.PaymentMethodCash) {
		amount = order.RoundedNetTotal
	} else {
		amount = order.NetTotal
	}

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

	return order, nil
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


// GetDeliveryQuote removed as GrabFood integration is disabled

// mark redeemed voucher as used
func MarkRedeemedVoucherAsUsed(trx *gorm.DB, voucherID uuid.UUID, voucherCode string, customerID uuid.UUID) error {
	var voucher models.CustomerVoucher
	query := trx.Model(&models.CustomerVoucher{}).
		Where("id = ?", voucherID).
		Where("customer_id = ?", customerID).
		Where("voucher_code = ?", voucherCode).
		First(&voucher)
	if query.Error != nil {
		return query.Error
	}

	voucher.Used = true
	timeNow := time.Now()
	voucher.UsedAt = &timeNow
	query = trx.Save(&voucher)
	if query.Error != nil {
		return query.Error
	}

	return nil
}

// apply changes on amount of voucher
func ApplyChangesOnAmountOfVoucher(trx *gorm.DB, voucherID uuid.UUID, grossTotal float32) error {
	var voucher models.Voucher
	result := trx.Model(&models.Voucher{}).
		Where("id = ?", voucherID).
		First(&voucher)
	if result.Error != nil {
		return result.Error
	}
	err := customer_common.VoucherRequirementsCheck(trx, voucher, grossTotal)
	if err != nil {
		return err
	}
	// apply changes on voucher usage
	voucher.CurrentUsage += 1
	result = trx.Save(&voucher)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// Customer API to get all orders (without product and selected modifier)
//
//encore:api auth method=GET path=/api/customers/orders/:order_status/:page/:page_size
func (s *Service) GetAllOrders(ctx context.Context, order_status string, page int, page_size int) (*GetAllOrdersResponse, error) {
	// prevent page from being less than 1 and 0 page size
	if page < 1 || page_size < 1 {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid page or page size",
		}
	}

	authData := auth.Data()
	customerData, ok := authData.(*models.User)
	if !ok {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid auth data type",
		}
	}

	var customer models.Customer
	query := s.db.Model(&models.Customer{}).Where("id= ?", customerData.ID).First(&customer)
	if query.Error != nil {
		return nil, query.Error
	}

	var orders []CustomerOrderResponse
	query = s.db.Model(&models.Order{}).
		Preload("Outlet").
		Where("customer_id = ?", customer.ID)

	/*
		Active Group
		OrderStatusPending            = "pending"
		OrderStatusPreparing          = "preparing"
		OrderStatusReady              = "ready"
		OrderStatusOnTheWay           = "on_the_way"

		Past Group
		OrderStatusCollected          = "collected"
		OrderStatusDelivered          = "delivered"
		OrderStatusCompleted          = "completed"
		OrderStatusCancelled          = "cancelled"
	*/
	// Only add order_status filter if it's not "all"
	if order_status != "" && order_status != "all" {
		if order_status == "active" {
			//query = query.Where("order_status = ? ", "pending")
			query = query.Where("order_status = ? OR order_status = ? OR order_status = ? OR order_status = ?", "pending", "preparing", "ready", "on_the_way")
		} else if order_status == "past" {
			query = query.Where("order_status = ? OR order_status = ? OR order_status = ? OR order_status = ?", "completed", "cancelled", "delivered", "collected")
		} else if order_status == "non_cancelled" {
			query = query.Where("order_status != ?", "cancelled")
		} else {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "invalid order status",
			}
		}
	}

	// count the max page
	var totalCount int64
	query.Count(&totalCount)
	maxPage := int(math.Ceil(float64(totalCount) / float64(page_size)))
	if page > maxPage {
		return &GetAllOrdersResponse{
			Orders:  []CustomerOrderResponse{},
			MaxPage: maxPage,
		}, nil
	}

	// Use Find instead of Scan to properly load preloaded associations
	var orderModels []models.Order
	query = query.Order("created_at DESC").
		Offset((page - 1) * page_size).
		Limit(page_size).
		Preload("Outlet").
		Preload("OrderDetails").
		Preload("OrderItems").
		Preload("OrderItems.Product").
		Preload("OrderItems.SelectedModifierGroups").
		Preload("OrderItems.SelectedModifierGroups.ModifierGroup").
		Preload("OrderItems.SelectedModifierGroups.ModifierOptions").
		Find(&orderModels)

	if query.Error != nil {
		return nil, query.Error
	}

	// Manually map the Order models to CustomerOrderResponse
	for _, order := range orderModels {
		customerOrder := CustomerOrderResponse{
			ID:                      order.ID,
			Outlet:                  *order.Outlet,
			OrderNumber:             order.OrderNumber,
			OrderDate:               order.OrderDate,
			OrderType:               order.OrderType,
			InvoiceNumber:           order.InvoiceNumber,
			InvoiceDate:             order.InvoiceDate,
			OrderStatus:             order.OrderStatus,
			Platform:                (*string)(order.Platform),
			PlatformOrderID:         order.PlatformOrderID,
			PlatformState:           order.PlatformState,
			GrossTotal:              order.GrossTotal,
			NetTotal:                order.NetTotal,
			RoundedAmount:           order.RoundedAmount,
			RoundedNetTotal:         order.RoundedNetTotal,
			AmountReceived:          order.AmountReceived,
			TaxCharge:               order.TaxCharge,
			TaxPercentage:           order.TaxPercentage,
			ServiceCharge:           order.ServiceCharge,
			ServiceChargePercentage: order.ServiceChargePercentage,
			DiscountType:            *order.DiscountType,
			DiscountAmount:          order.DiscountAmount,
			DiscountPercentage:      order.DiscountPercentage,
			PaymentMethod:           order.PaymentMethod,
			PaymentStatus:           order.PaymentStatus,
			Notes:                   order.Notes,
			TableNumber:             order.TableNumber,
			EInvoiceSubmissionID:    order.EInvoiceSubmissionID,
			EInvoiceStatus:          (*string)(order.EInvoiceStatus),
			EInvoiceURL:             order.EInvoiceURL,
			EInvoiceRejectedReason:  order.EInvoiceRejectedReason,
			OrderDetails:            *order.OrderDetails,
			OrderItems:              order.OrderItems,
			PickupAt:                order.PickupAt,
			CompletedAt:             order.CompletedAt,
			PointsRewarded:          order.PointsRewarded,
			PointsRewardedAt:        order.PointsRewardedAt,
			ExpRewarded:             order.ExpRewarded,
			ExpRewardedAt:           order.ExpRewardedAt,
			CreatedAt:               order.CreatedAt,
			UpdatedAt:               order.UpdatedAt,
		}
		orders = append(orders, customerOrder)
	}

	return &GetAllOrdersResponse{
		Orders:  orders,
		MaxPage: maxPage,
	}, nil
}

// Customer API to get a order with details
//
//encore:api auth method=GET path=/api/customers/orders/:order_id
func (s *Service) GetOrder(ctx context.Context, order_id uuid.UUID) (*GetOrderResponse, error) {
	order, err := customer_common.GetOrderByID(s.db, order_id, true)
	if err != nil {
		return nil, err
	}
	pointsRewarded, err := customer_common.GetPointsRewardedBasedOnOrderID(s.db, order_id)
	if err != nil {
		return nil, err
	}

	var totalPointsRewarded int
	for _, pointRewarded := range pointsRewarded {
		totalPointsRewarded += pointRewarded.PointsEarned
	}

	customerOrder := CustomerOrderResponse{
		ID:                      order.ID,
		Outlet:                  *order.Outlet,
		OrderNumber:             order.OrderNumber,
		OrderDate:               order.OrderDate,
		OrderType:               order.OrderType,
		InvoiceNumber:           order.InvoiceNumber,
		InvoiceDate:             order.InvoiceDate,
		OrderStatus:             order.OrderStatus,
		Platform:                (*string)(order.Platform),
		PlatformOrderID:         order.PlatformOrderID,
		PlatformState:           order.PlatformState,
		GrossTotal:              order.GrossTotal,
		NetTotal:                order.NetTotal,
		RoundedAmount:           order.RoundedAmount,
		RoundedNetTotal:         order.RoundedNetTotal,
		DeliveryFee:             *order.DeliveryFee,
		AmountReceived:          order.AmountReceived,
		TaxCharge:               order.TaxCharge,
		TaxPercentage:           order.TaxPercentage,
		ServiceCharge:           order.ServiceCharge,
		ServiceChargePercentage: order.ServiceChargePercentage,
		DiscountType:            *order.DiscountType,
		DiscountAmount:          order.DiscountAmount,
		DiscountPercentage:      order.DiscountPercentage,
		PaymentMethod:           order.PaymentMethod,
		PaymentStatus:           order.PaymentStatus,
		Notes:                   order.Notes,
		TableNumber:             order.TableNumber,
		EInvoiceSubmissionID:    order.EInvoiceSubmissionID,
		EInvoiceStatus:          (*string)(order.EInvoiceStatus),
		EInvoiceURL:             order.EInvoiceURL,
		EInvoiceRejectedReason:  order.EInvoiceRejectedReason,
		OrderDetails:            *order.OrderDetails,
		OrderItems:              order.OrderItems,
		Customer:                order.Customer,
		PointEarned:             &totalPointsRewarded,
		PickupAt:                order.PickupAt,
		CompletedAt:             order.CompletedAt,
		PointsRewarded:          order.PointsRewarded,
		PointsRewardedAt:        order.PointsRewardedAt,
		ExpRewarded:             order.ExpRewarded,
		ExpRewardedAt:           order.ExpRewardedAt,
		CreatedAt:               order.CreatedAt,
		UpdatedAt:               order.UpdatedAt,
	}

	return &GetOrderResponse{
		Order: customerOrder,
	}, nil
}

//encore:api public method=GET path=/api/customers/order/:order_id
func (s *Service) GetOrderDetails(ctx context.Context, order_id uuid.UUID) (*models.Order, error) {
	var order models.Order
	if err := s.db.Where("id = ?", order_id).Preload("Outlet").Preload("OrderItems.Product").Preload("OrderItems.SelectedModifierGroups.ModifierOptions").First(&order).Error; err != nil {
		return nil, err
	}

	return &order, nil
}

// Customer API to get product and selected modifier
//
//encore:api auth method=GET path=/api/customers/orders-with-product/:order_id
func (s *Service) GetProductAndSelectedModifier(ctx context.Context, order_id uuid.UUID) (*GetProductAndSelectedModifierResponse, error) {
	var order models.Order
	query := s.db.Model(&models.Order{}).Where("id = ?", order_id).First(&order)
	if query.Error != nil {
		return nil, query.Error
	}

	var orderItems []models.OrderItem
	query = s.db.Model(&models.OrderItem{}).Where("order_id = ?", order_id).Preload("Product").Find(&orderItems)
	if query.Error != nil {
		return nil, query.Error
	}

	// Extract order item IDs for the IN clause
	var orderItemIDs []uuid.UUID
	for _, item := range orderItems {
		orderItemIDs = append(orderItemIDs, item.ID)
	}

	var selectedModifierGroups []models.SelectedModifierGroup
	query = s.db.Model(&models.SelectedModifierGroup{}).Where("order_item_id IN (?)", orderItemIDs).Preload("ModifierOptions").Find(&selectedModifierGroups)
	if query.Error != nil {
		return nil, query.Error
	}

	var orderItemsAndSelectedModifier []OrderItemAndSelectedModifier
	for _, orderItem := range orderItems {
		orderItem := OrderItemAndSelectedModifier{
			OrderItem:              orderItem,
			SelectedModifierGroups: selectedModifierGroups,
		}
		orderItemsAndSelectedModifier = append(orderItemsAndSelectedModifier, orderItem)
	}

	return &GetProductAndSelectedModifierResponse{
		OrderItems: orderItemsAndSelectedModifier,
	}, nil
}

// API to hard delete order if the order is still in pending status
//
//encore:api auth method=DELETE path=/api/customers/order/delete/:outlet_id/:order_id
func (s *Service) DeleteOrder(ctx context.Context, outlet_id uuid.UUID, order_id uuid.UUID) (*common.BasicResponse, error) {
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var order models.Order
	result := trx.Where("id=? AND outlet_id=? AND order_status=?", order_id, outlet_id, models.OrderStatusPending).First(&order)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// delete selected modifier group
	result = trx.Unscoped().Delete(&models.SelectedModifierGroup{}, "order_item_id IN (SELECT id FROM order_items WHERE order_id = ?)", order_id)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// delete all order items
	result = trx.Unscoped().Delete(&models.OrderItem{}, "order_id = ? ", order_id)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		trx.Rollback()
		return nil, result.Error
	}

	// delete all transaction
	result = trx.Unscoped().Delete(&models.Transaction{}, "order_id = ?", order_id)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// delete the order details
	result = trx.Unscoped().Delete(&models.OrderDetails{}, "order_id = ?", order_id)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// delete the order
	result = trx.Unscoped().Delete(&order)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Order deleted successfully",
	}, nil
}

// API to return order details and check if the order is reorderable
//
//encore:api auth method=POST path=/api/customers/order/reorder
func (s *Service) ReorderOrderDetails(ctx context.Context, req *ReorderOrderDetailsRequest) (*ReorderOrderDetailsResponse, error) {
	isChange := false
	custMsg := ""
	order, err := customer_common.GetOrderByIDWithCustomerID(s.db, req.OrderID, true, req.CustomerID)
	if err != nil {
		return nil, errors.New("order not found")
	}
	// get product ids from outlet products
	productIDsToRemove, isProductChange, custProductMsg, err := CheckOutletProductAvailability(s.db, req.OutletID, req.OrderID, true)
	if err != nil {
		return nil, errors.New("error checking outlet product availability")
	}

	// get order items left based on product ids to remove
	orderItemsLeft, _ := GetOrderItemsLeft(order.OrderItems, productIDsToRemove)
	order.OrderItems = orderItemsLeft

	// remove unavailable modifier groups and modifier options for all order items
	orderItemsAfterRemoveUnavailableModifierGroups, isModifierChange, custModifierMsg, err := GetLatestOrderItemsAfterRemoveUnavailableModifierGroups(
		s.db,
		order.OrderItems,
		req.OutletID,
	)
	if err != nil {
		return nil, errors.New("error getting modifier group ids to remove")
	}
	order.OrderItems = orderItemsAfterRemoveUnavailableModifierGroups
	isChange = isProductChange || isModifierChange
	custMsg += custProductMsg
	custMsg += custModifierMsg
	if custMsg == "" {
		custMsg = "Order validated successfully"
	}

	// apply changes on gross total
	if isChange {
		order, err = RecalculatePaymentDetails(s.db, *order)
		if err != nil {
			return nil, errors.New("error recalculating payment details")
		}
	}

	// convert order to custom order
	var customOrder CustomOrder
	customer_common.AutoMapStructFields(order, &customOrder)
	for i := range customOrder.OrderItems {
		orderItem := &customOrder.OrderItems[i]
		customOrder.OrderItems[i].SelectedModifierGroups, err = AssignMaxSelectionToSelectedModifierGroups(
			s.db,
			orderItem.ProductID,
			orderItem.SelectedModifierGroups,
		)
		if err != nil {
			return nil, err
		}
	}

	return &ReorderOrderDetailsResponse{
		Message:  custMsg,
		IsValid:  len(order.OrderItems) > 0,
		IsChange: isChange,
		Order:    customOrder,
	}, nil
}

// func to get order items left
func GetOrderItemsLeft(orderItems []models.OrderItem, productIDsToRemove []uuid.UUID) ([]models.OrderItem, []uuid.UUID) {
	orderItemsLeft := []models.OrderItem{}
	productIDsLeft := []uuid.UUID{}
	for _, orderItem := range orderItems {
		if !slices.Contains(productIDsToRemove, orderItem.ProductID) {
			orderItemsLeft = append(orderItemsLeft, orderItem)
			productIDsLeft = append(productIDsLeft, orderItem.ProductID)
			continue
		}
	}
	return orderItemsLeft, productIDsLeft
}

// func to check outlet product availability
func CheckOutletProductAvailability(trx *gorm.DB, outlet_id uuid.UUID, order_id uuid.UUID, is_active bool) ([]uuid.UUID, bool, string, error) {
	orderItemToRemove := []uuid.UUID{}
	custMsg := ""
	isChange := false
	// get product ids from outlet products
	productIDsFromOutletProducts, err := customer_common.GetProductIDsFromOutletProducts(trx, outlet_id, is_active)
	if err != nil {
		return nil, isChange, custMsg, errors.New("error getting product ids from outlet products")
	}
	// get product ids from order items
	productIDsFromOrderItems, err := customer_common.GetProductIDsFromOrderItems(trx, order_id)
	if err != nil {
		return nil, isChange, custMsg, errors.New("error getting product ids from order items")
	}
	// compare product ids from outlet products
	for _, productID := range productIDsFromOrderItems {
		// enter this if mean product is not available in outlet products
		if !slices.Contains(productIDsFromOutletProducts, productID) {
			orderItemToRemove = append(orderItemToRemove, productID)
			// generate customer message
			var product models.Product
			err := trx.Model(&models.Product{}).Where("id = ?", productID).Select("name").First(&product).Error
			if err == nil {
				custMsg += product.Name + " is not available\n"
			}
			isChange = true
			continue
		}
	}

	return orderItemToRemove, isChange, custMsg, nil
}

// func to get latest order items configuration after remove unavailable modifier groups
func GetLatestOrderItemsAfterRemoveUnavailableModifierGroups(trx *gorm.DB, orderItems []models.OrderItem, outletID uuid.UUID) ([]models.OrderItem, bool, string, error) {
	isChange := false
	custMsg := ""
	productModifierGroupIDsCache := make(map[uuid.UUID][]uuid.UUID)

	for i, _ := range orderItems {
		orderItem := &orderItems[i]
		var validSelectedModifierGroups []models.SelectedModifierGroup
		// check if product id and modifier group ids are in cache then reuse it
		productModifierGroupIDs, ok := productModifierGroupIDsCache[orderItem.ProductID]
		err := error(nil)
		if !ok {
			// not in cache, then get product modifier ids
			// get product modifier ids
			productModifierGroupIDs, err = customer_common.GetModifierGroupIDsFromProductModifierMapping(trx, orderItem.ProductID)
			if err != nil {
				return nil, isChange, custMsg, errors.New("error getting product modifier ids")
			}
			// add to cache
			productModifierGroupIDsCache[orderItem.ProductID] = productModifierGroupIDs
		}
		// remove selected modifier groups that are not in product modifier ids
		for _, smg := range orderItem.SelectedModifierGroups {
			// if modifier group id is in product modifier group ids, then add to valid selected modifier groups
			if slices.Contains(productModifierGroupIDs, smg.ModifierGroupID) {
				// Check if modifier option is available
				// if error, then directly remove it, no need to return error
				validSmg, isChangeModifierOption, custModifierOptionMsg, removedAmount, _ := RemoveUnavailableModifierOptions(trx, smg, outletID)
				// update subtotal after remove unavailable modifier group and option
				// if is available, the removed amount is 0, thus minus also would not affect the subtotal
				orderItem.SubTotal -= removedAmount
				// Only add if the modifier option is still valid (not removed)
				if validSmg != nil {
					validSelectedModifierGroups = append(validSelectedModifierGroups, *validSmg)
				}
				isChange = isChange || isChangeModifierOption
				custMsg += custModifierOptionMsg
				continue
			}
			// if reach here
			// means modifier group is removed due to invalid/inactive/unavailable
			isChange = true
			custMsg += smg.ModifierGroup.Name + " is not available\n"
			// update subtotal after remove unavailable modifier group and option
			orderItem.SubTotal -= smg.ModifierOptions.PriceAdjustment * float32(smg.ModifierOptionQuantity)
		}
		orderItem.SelectedModifierGroups = validSelectedModifierGroups
	}
	return orderItems, isChange, custMsg, nil
}

// func to remove unavailable modifier options for the selected modifier group
func RemoveUnavailableModifierOptions(trx *gorm.DB, selectedModifierGroup models.SelectedModifierGroup, outletID uuid.UUID) (*models.SelectedModifierGroup, bool, string, float32, error) {
	isChange := false
	custMsg := ""
	removedAmount := float32(0)
	// Check if the modifier option is available at outlet level
	var outletModifierOption models.OutletModifierOption
	err := trx.Model(&models.OutletModifierOption{}).
		Where("outlet_id = ? AND modifier_options_id = ?", outletID, selectedModifierGroup.ModifierOptionsID).
		Preload("ModifierOptions").
		First(&outletModifierOption).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Modifier option not available at this outlet
			custMsg += selectedModifierGroup.ModifierOptions.Name + " is not available\n"
			removedAmount = selectedModifierGroup.ModifierOptions.PriceAdjustment * float32(selectedModifierGroup.ModifierOptionQuantity)
			return nil, isChange, custMsg, removedAmount, err
		}
		return nil, isChange, custMsg, removedAmount, err
	}

	// Check if the modifier option is active at outlet level
	if !outletModifierOption.IsActive {
		isChange = true
		custMsg += selectedModifierGroup.ModifierOptions.Name + " is not available\n"
		removedAmount = selectedModifierGroup.ModifierOptions.PriceAdjustment * float32(selectedModifierGroup.ModifierOptionQuantity)
		return nil, isChange, custMsg, removedAmount, err
	}

	// Check if the modifier option is active globally
	if !outletModifierOption.ModifierOptions.IsActive {
		isChange = true
		custMsg += selectedModifierGroup.ModifierOptions.Name + " is not available\n"
		removedAmount = selectedModifierGroup.ModifierOptions.PriceAdjustment * float32(selectedModifierGroup.ModifierOptionQuantity)
		return nil, isChange, custMsg, removedAmount, err
	}

	// If all checks pass, return the selected modifier group as valid
	return &selectedModifierGroup, isChange, custMsg, removedAmount, nil
}

// func to recalculate all the payment details
func RecalculatePaymentDetails(trx *gorm.DB, order models.Order) (*models.Order, error) {
	// recalculate gross total
	order.GrossTotal = 0
	for i := range order.OrderItems {
		orderItem := &order.OrderItems[i]
		order.GrossTotal += orderItem.SubTotal
	}
	// get tax congig
	taxConfig, err := customer_common.GetTaxConfigurationByBusinessID(trx, order.BusinessID)
	if err != nil {
		return nil, err
	}
	// calculate tax charge
	taxCharge := order.GrossTotal * (*taxConfig.ServiceTaxPercentage / 100)
	// calculate service charge
	serviceCharge := order.GrossTotal * (*taxConfig.ServiceChargePercentage / 100)

	// round to 2 decimal places
	order.GrossTotal = customer_common.RoundToTwoDecimals(order.GrossTotal)

	taxCharge = customer_common.RoundToTwoDecimals(taxCharge)
	serviceCharge = customer_common.RoundToTwoDecimals(serviceCharge)
	order.TaxCharge = taxCharge
	order.ServiceCharge = serviceCharge
	order.ServiceChargePercentage = *taxConfig.ServiceChargePercentage
	order.TaxPercentage = *taxConfig.ServiceTaxPercentage

	netTotal := order.GrossTotal + taxCharge + serviceCharge
	order.NetTotal = netTotal
	order.RoundedNetTotal = netTotal
	return &order, nil
}

// API to get all product modifier mappings based on product ids
//
//encore:api auth method=POST path=/api/customers/products/modifiers/all
func (s *Service) AllProductModifierMappings(ctx context.Context, req *AllProductModifierMappingsRequest) (*AllProductModifierMappingsResponse, error) {
	productModifierMappings, err := customer_common.GetProductModifierMappingsByProductIDs(
		s.db,
		req.ProductIDs,
		true,
		true,
	)
	if err != nil {
		return nil, err
	}
	// is valid mean has no modifier mapping, hence valid
	return &AllProductModifierMappingsResponse{
		Message:                 "Product modifier mappings fetched successfully",
		ProductModifierMappings: productModifierMappings,
	}, nil
}

// assign max selection to selected modifier groups
func AssignMaxSelectionToSelectedModifierGroups(trx *gorm.DB, productID uuid.UUID, selectedModifierGroups []CustomSelectedModifierGroup) ([]CustomSelectedModifierGroup, error) {
	if len(selectedModifierGroups) == 0 {
		return selectedModifierGroups, nil
	}

	groupIDs := make([]uuid.UUID, 0, len(selectedModifierGroups))
	for i := range selectedModifierGroups {
		groupIDs = append(groupIDs, selectedModifierGroups[i].ModifierGroupID)
	}

	maxSelByGroup, err := customer_common.GetMaxSelectionFromProductModifierMapping(trx, productID, groupIDs)
	if err != nil {
		return nil, err
	}

	for i := range selectedModifierGroups {
		gid := selectedModifierGroups[i].ModifierGroupID
		maxSel, ok := maxSelByGroup[gid]
		if !ok {
			// Preserve previous behavior of error if mapping is missing
			return nil, fmt.Errorf("max selection not found for product %s and modifier group %s", productID.String(), gid.String())
		}
		selectedModifierGroups[i].MaxSelection = maxSel
	}
	return selectedModifierGroups, nil
}

// API to get product with favourite
//
//encore:api auth method=GET path=/api/customers/outlets/products/favourite/:outlet_id/:product_id
func (s *Service) GetProductWithFavourite(ctx context.Context, outlet_id uuid.UUID, product_id uuid.UUID) (*GetProductWithFavouriteResponse, error) {
	customer, err := customer_common.GetCustomerDataFromAuthData(auth.Data)
	if err != nil {
		return nil, err
	}
	product, err := customer_common.GetOutletProductByProductID(s.db, product_id)
	if err != nil {
		return nil, err
	}
	isFav, _ := customer_common.GetProductIsFavouriteByProductID(s.db, customer.ID, product_id)

	return &GetProductWithFavouriteResponse{
		Message: "Product fetched successfully",
		Data: ProductWithFavourite{
			Product:     product.Product,
			IsFavourite: isFav,
		},
	}, nil

}

// API to cancel order
//
//encore:api auth method=POST path=/api/customers/order/cancel
func (s *Service) CancelOrder(ctx context.Context, req *CancelOrderRequest) (*common.BasicResponse, error) {

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	transaction, err := customer_common.GetTransactionByOrderID(trx, req.OrderID)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	customerID := transaction.Order.CustomerID
	if customerID != nil && *customerID != req.CustomerID {
		trx.Rollback()
		return nil, errors.New("failed to cancel order")
	}

	transaction.PaymentStatus = models.PaymentStatusVoided
	transaction.Order.OrderStatus = models.OrderStatusCancelled
	transaction.Order.PaymentStatus = models.PaymentStatusVoided
	result := trx.Save(&transaction)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	result = trx.Save(&transaction.Order)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	return &common.BasicResponse{
		Message: "Order cancelled successfully",
	}, nil
}

// API to complete order of customer
//
//encore:api auth method=POST path=/api/customers/order/complete
func (s *Service) CompleteOrder(ctx context.Context, req *CompleteOrderRequest) (*common.BasicResponse, error) {
	customer, err := customer_common.GetCustomerDataFromAuthData(auth.Data)
	if err != nil {
		return nil, err
	}
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	order, err := customer_common.GetOrderByID(trx, req.OrderID, true)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// ENSURE SAME CUSTOMER
	if customer.ID != *order.CustomerID {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "You are not allowed to complete this order",
		}
	}

	// ENSURE PICKUP ONLY
	if order.OrderType != models.OrderTypePickup && order.OrderType != models.OrderTypePickupLater {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "You are not allowed to complete this order",
		}
	}

	order.OrderStatus = models.OrderStatusCompleted
	result := trx.Save(&order)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// process sale
	err = customer_common.ProcessSale(trx, ctx, req.OrderID)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// send notification to the customer as if the customer id is not nil
	if order.CustomerID != nil && order.Platform != nil && *order.Platform == models.PlatformMembershipApp {
		// notification logic removed
	}

	// trigger digital signage removed

	return &common.BasicResponse{
		Message: "Order completed successfully",
	}, nil
}
