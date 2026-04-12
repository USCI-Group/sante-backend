package finances

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"encore.app/auth_service"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/common_operations"
	"encore.app/database"
	"encore.app/database/models"
	"encore.dev/beta/auth"
	"encore.dev/types/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Service{db: db}, nil
}

type GetTransactionsRequest struct {
	Page           int                       `json:"page"`
	PageSize       int                       `json:"page_size"`
	ProductIDs     []uuid.UUID               `json:"product_ids"`
	PaymentMethods []constants.PaymentMethod `json:"payment_methods"`
	BusinessID     *uuid.UUID                `json:"business_id"`
	OutletID       *uuid.UUID                `json:"outlet_id"`
	OutletGroupID  *uuid.UUID                `json:"outlet_group_id"`
	StartDate      *time.Time                `json:"start_date"`
	EndDate        *time.Time                `json:"end_date"`
	Search         *string                   `json:"search"`
}

type GetPayoutReportRequest struct {
	Page          int        `json:"page"`
	PageSize      int        `json:"page_size"`
	OutletID      *uuid.UUID `json:"outlet_id"`
	OutletGroupID *uuid.UUID `json:"outlet_group_id"`
	Date          time.Time  `json:"date"`
}

type GetPayoutReportResponse struct {
	Meta common.Pagination       `json:"meta"`
	Data []models.ExpensesOutlet `json:"data"`
}

type ItemDetail struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	Price       float32   `json:"price"`
	ImageURL    string    `json:"image_url"`
}

type Transaction struct {
	OrderID         uuid.UUID    `json:"order_id"`
	OutletName      string       `json:"outlet_name"`
	ReceiptNo       string       `json:"receipt_no"`
	ProductNames    string       `json:"product_names"`
	TransactionDate time.Time    `json:"transaction_date"`
	PaymentMethod   string       `json:"payment_method"`
	Price           float32      `json:"price"`
	ItemDetails     []ItemDetail `json:"item_details"`
	Status          string       `json:"status"`
	Platform        string       `json:"platform"`
}

type GetTransactionsResponse struct {
	Meta common.Pagination `json:"meta"`
	Data []Transaction     `json:"data"`
}

type FinanceSummary struct {
	SalesInfo              SalesInfo                `json:"sales_info"`
	SalesByCategory        []SalesByCategory        `json:"sales_by_category"`
	SalesByProduct         []SalesByProduct         `json:"sales_by_product"`
	SalesByModifierOptions []SalesByModifierOptions `json:"sales_by_modifier_options"`
	SalesByEmployee        []SalesByEmployee        `json:"sales_by_employee"`
}

type SalesInfo struct {
	GrossSales                    float32 `json:"gross_sales"`
	TotalDiscount                 float32 `json:"total_discount"`
	TotalServiceCharge            float32 `json:"total_service_charge"`
	TotalTax                      float32 `json:"total_tax"`
	TotalNetSales                 float32 `json:"total_net_sales"`
	TotalRoundedNetSales          float32 `json:"total_rounded_net_sales"`
	TotalRoundedDownAmount        float32 `json:"total_rounded_down_amount"`
	TotalRoundedUpAmount          float32 `json:"total_rounded_up_amount"`
	TotalRedeemPoints             int     `json:"total_redeem_points"`
	TotalCashRounding             float32 `json:"total_cash_rounding"`
	TotalCost                     float32 `json:"total_cost"`
	GrossProfit                   float32 `json:"gross_profit"`
	NetProfit                     float32 `json:"net_profit"`
	NoOfSalesTrans                int     `json:"no_of_sales_trans"`
	AverageSalesPerTrans          float32 `json:"average_sales_per_trans"`
	NoOfVoidedTrans               int     `json:"no_of_voided_trans"`
	TotalVoidedAmount             float32 `json:"total_voided_amount"`
	TotalCustomerSignUpMembership int     `json:"total_customer_sign_up_membership"`
	MemberSales                   float32 `json:"member_sales"`
	NonMemberSales                float32 `json:"non_member_sales"`
	MemberSalesQuantity           int     `json:"member_sales_quantity"`
	NonMemberSalesQuantity        int     `json:"non_member_sales_quantity"`
	NoOfUnpaidOrders              int     `json:"no_of_unpaid_orders"`
	UnpaidOrdersAmount            float32 `json:"unpaid_orders_amount"`
	CashSales                     float32 `json:"cash_sales"`
	DuitnowQRSales                float32 `json:"duitnow_qr_sales"`
	CardSales                     float32 `json:"card_sales"`
	CashClosing                   float32 `json:"cash_closing"`
}

type SalesByCategory struct {
	ProductCategory models.ProductCategory `json:"product_category"`
	GrossSales      float32                `json:"gross_sales"`
	QuantitySold    int                    `json:"quantity_sold"`
	TotalDiscount   float32                `json:"total_discount"`
}

type SalesByProduct struct {
	Product             models.Product `json:"product"`
	GrossSales          float32        `json:"gross_sales"`
	QuantitySold        int            `json:"quantity_sold"`
	TotalCost           float32        `json:"total_cost"`
	TotalDiscount       float32        `json:"total_discount"`
	TotalProfit         float32        `json:"total_profit"`
	SoldByPaymentMethod map[string]int `json:"sold_by_payment_method"`
}

type SalesByModifierOptions struct {
	ModifierOptions models.ModifierOptions `json:"modifier_options"`
	GrossSales      float32                `json:"gross_sales"`
	QuantitySold    int                    `json:"quantity_sold"`
	TotalDiscount   float32                `json:"total_discount"`
}

type SalesByEmployee struct {
	Employee             models.User `json:"employee"`
	NumberOfSales        int         `json:"number_of_sales"`
	TotalNetSales        float32     `json:"total_net_sales"`
	TotalRoundedNetSales float32     `json:"total_rounded_net_sales"`
}

type GetFullReportRequest struct {
	BusinessID    *uuid.UUID `json:"business_id"`
	OutletID      *uuid.UUID `json:"outlet_id"`
	OutletGroupID *uuid.UUID `json:"outlet_group_id"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
}

type FullReport struct {
	OutletID       uuid.UUID `json:"outlet_id"`
	OutletName     string    `json:"outlet_name"`
	OutletImageURL string    `json:"outlet_image_url"`
	Cash           float32   `json:"cash"`
	QR             float32   `json:"qr"`
	Grab           float32   `json:"grab"`
	Shopee         float32   `json:"shopee"`
	TotalSales     float32   `json:"total_sales"`

	// Staff and Labor
	ExpenseCategoryStaffSalary    float32 `json:"expense_category_staff_salary"`
	ExpenseCategoryPartTimerDaily float32 `json:"expense_category_part_timer_daily"`

	// Utilities
	ExpenseCategoryGas         float32 `json:"expense_category_gas"`
	ExpenseCategoryWaterCharge float32 `json:"expense_category_water_charge"`

	// Kitchen and Cleaning Supplies
	ExpenseCategoryDishwashingLiquidPaste float32 `json:"expense_category_dishwashing_liquid_paste"`
	ExpenseCategoryCookingOil             float32 `json:"expense_category_cooking_oil"`
	ExpenseCategoryTissue                 float32 `json:"expense_category_tissue"`
	ExpenseCategoryGarbageBag             float32 `json:"expense_category_garbage_bag"`
	ExpenseCategoryGlove                  float32 `json:"expense_category_glove"`

	// Transportation
	ExpenseCategoryTransportationPetrol float32 `json:"expense_category_transportation_petrol"`
	ExpenseCategoryPetrolAllowanceCrew  float32 `json:"expense_category_petrol_allowance_crew"`
	ExpenseCategoryUpkeepOfVehicle      float32 `json:"expense_category_upkeep_of_vehicle"`

	// Administrative
	ExpenseCategoryLicenseFee            float32 `json:"expense_category_license_fee"`
	ExpenseCategoryPrintingAndStationery float32 `json:"expense_category_printing_and_stationery"`
	ExpenseCategoryOutletSupplies        float32 `json:"expense_category_outlet_supplies"`

	// Maintenance
	ExpenseCategoryUpkeepOfOutlet      float32 `json:"expense_category_upkeep_of_outlet"`
	ExpenseCategoryUpkeepOfCentreHouse float32 `json:"expense_category_upkeep_of_centre_house"`

	// Other
	ExpenseCategoryOther float32 `json:"expense_category_other"`

	TotalExpenses float32 `json:"total_expenses"`
}

type GetFullReportResponse struct {
	TotalSalesInfo     float32      `json:"total_sales_info"`
	TotalSalesProduct  float32      `json:"total_sales_product"`
	TotalSalesModifier float32      `json:"total_sales_modifier"`
	FullReports        []FullReport `json:"full_reports"`
}

type FinanceSummaryRequest struct {
	OutletID  uuid.UUID `json:"outlet_id"`
	StartDate time.Time `json:"start_date" validate:"required"` // Please pass in time in utc 0 such as "start_date": "2026-01-08T00:00:00Z",
	EndDate   time.Time `json:"end_date" validate:"required"`   // Please pass in time in utc 0 such as "end_date": "2026-01-08T00:00:00Z",
}

type FinanceOverviewPerformance struct {
	SalesByProductOutlet   []SalesByProduct         `json:"sales_by_product_outlet"` // store outlet
	SalesByProductGrab     []SalesByProduct         `json:"sales_by_product_grab"`   // grabfood
	SalesByProductShopee   []SalesByProduct         `json:"sales_by_product_shopee"` // shopeefood
	SalesByModifier        []SalesByModifierOptions `json:"sales_by_modifier"`
	SalesByPaymentMethod   []SalesByPaymentMethod   `json:"sales_by_payment_method"`
	HourlySalesPerformance HourlySalesPerformance   `json:"hourly_sales_performance"`
}

type SalesByPaymentMethod struct {
	PaymentMethod string  `json:"payment_method"`
	TotalSales    float32 `json:"total_sales"`
	TotalQuantity int     `json:"total_quantity"`
}

type HourlySalesPerformance struct {
	TotalSales  float32      `json:"total_sales"`
	HourlySales []HourlySale `json:"hourly_sales"`
}

type HourlySale struct {
	TotalSales float32 `json:"total_sales"`
	Time       string  `json:"time"`
}

/*
 /$$$$$$$$ /$$                                                           /$$$  /$$$$$$        /$$               /$$           /$$$
| $$_____/|__/                                                          /$$_/ /$$__  $$      | $$              |__/          |_  $$
| $$       /$$ /$$$$$$$   /$$$$$$  /$$$$$$$   /$$$$$$$  /$$$$$$        /$$/  | $$  \ $$  /$$$$$$$ /$$$$$$/$$$$  /$$ /$$$$$$$   \  $$
| $$$$$   | $$| $$__  $$ |____  $$| $$__  $$ /$$_____/ /$$__  $$      | $$   | $$$$$$$$ /$$__  $$| $$_  $$_  $$| $$| $$__  $$   | $$
| $$__/   | $$| $$  \ $$  /$$$$$$$| $$  \ $$| $$      | $$$$$$$$      | $$   | $$__  $$| $$  | $$| $$ \ $$ \ $$| $$| $$  \ $$   | $$
| $$      | $$| $$  | $$ /$$__  $$| $$  | $$| $$      | $$_____/      |  $$  | $$  | $$| $$  | $$| $$ | $$ | $$| $$| $$  | $$   /$$/
| $$      | $$| $$  | $$|  $$$$$$$| $$  | $$|  $$$$$$$|  $$$$$$$       \  $$$| $$  | $$|  $$$$$$$| $$ | $$ | $$| $$| $$  | $$ /$$$/
|__/      |__/|__/  |__/ \_______/|__/  |__/ \_______/ \_______/        \___/|__/  |__/ \_______/|__/ |__/ |__/|__/|__/  |__/|___/
*/

//encore:api auth method=POST path=/api/admin/finances/transactions
func (s *Service) GetTransactions(ctx context.Context, params *GetTransactionsRequest) (*GetTransactionsResponse, error) {
	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	var orders []models.Order

	query := s.db.Preload("Outlet").Order("order_date DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	queryCount := s.db.Model(&models.Order{})

	// If product IDs are provided, filter orders that contain these products
	if len(params.ProductIDs) > 0 {
		// Join with order_items and filter by product_id
		query = query.Joins("JOIN order_items ON orders.id = order_items.order_id").
			Where("order_items.product_id IN ?", params.ProductIDs).
			Group("orders.id") // Group to avoid duplicate orders

		queryCount = queryCount.Joins("JOIN order_items ON orders.id = order_items.order_id").
			Where("order_items.product_id IN ?", params.ProductIDs).
			Group("orders.id")
	}

	if len(params.PaymentMethods) > 0 {
		query = query.Where("payment_method IN ?", params.PaymentMethods)
		queryCount = queryCount.Where("payment_method IN ?", params.PaymentMethods)
	}

	if !auth_service.IsSanteAdmin() {
		params.BusinessID = auth_service.GetUserBusinessID()
	}

	if params.BusinessID != nil {
		query = query.Where("business_id = ?", params.BusinessID)
		queryCount = queryCount.Where("business_id = ?", params.BusinessID)
	}

	if params.OutletID != nil {
		query = query.Where("outlet_id = ?", params.OutletID)
		queryCount = queryCount.Where("outlet_id = ?", params.OutletID)
	}

	if params.OutletGroupID != nil {
		outletIDs, err := common_operations.GetOutletIDsByGroupID(s.db, *params.OutletGroupID)
		if err == nil {
			query = query.Where("outlet_id IN ?", outletIDs)
			queryCount = queryCount.Where("outlet_id IN ?", outletIDs)
		}
	}

	if params.StartDate != nil {
		startDate, _ := common.GetStartOfDay(*params.StartDate)
		query = query.Where("order_date >= ?", startDate.UTC())
		queryCount = queryCount.Where("order_date >= ?", startDate.UTC())
	}

	if params.EndDate != nil {
		endDate, _ := common.GetEndOfDay(*params.EndDate)
		query = query.Where("order_date < ?", endDate.UTC())
		queryCount = queryCount.Where("order_date < ?", endDate.UTC())
	}

	if params.Search != nil {
		query = query.Where("invoice_number LIKE ? OR order_number LIKE ?", "%"+*params.Search+"%", "%"+*params.Search+"%")
		queryCount = queryCount.Where("invoice_number LIKE ? OR order_number LIKE ?", "%"+*params.Search+"%", "%"+*params.Search+"%")
	}

	if err := query.Find(&orders).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := queryCount.Count(&total).Error; err != nil {
		return nil, err
	}

	var transactions []Transaction
	for _, order := range orders {
		var orderItems []models.OrderItem
		if err := s.db.Where("order_id = ?", order.ID).Preload("Product").Find(&orderItems).Error; err != nil {
			return nil, err
		}

		var productNames string

		for _, orderItem := range orderItems {
			if !strings.Contains(productNames, orderItem.Product.Name) {
				productNames += orderItem.Product.Name + " + "
			}

		}
		productNames = strings.TrimSuffix(productNames, " + ")

		platform := "store_outlet"
		if order.Platform != nil {
			platform = string(*order.Platform)
		}

		transaction := Transaction{
			OrderID:         order.ID,
			OutletName:      order.Outlet.Name,
			ReceiptNo:       order.InvoiceNumber,
			ProductNames:    productNames,
			TransactionDate: order.OrderDate,
			PaymentMethod:   order.PaymentMethod,
			Price:           float32(order.NetTotal),
			Status:          order.OrderStatus,
			Platform:        platform,
		}

		transactions = append(transactions, transaction)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &GetTransactionsResponse{
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: totalPages,
			Total:      total,
		},
		Data: transactions,
	}, nil
}

//encore:api auth method=GET path=/api/admin/finances/transaction-details/:orderID
func (s *Service) GetTransactionDetails(ctx context.Context, orderID uuid.UUID) (*Transaction, error) {
	var order models.Order
	if err := s.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		return nil, err
	}

	var orderItems []models.OrderItem
	if err := s.db.Where("order_id = ?", order.ID).Preload("Product").Find(&orderItems).Error; err != nil {
		return nil, err
	}

	itemDetailsMap := make(map[uuid.UUID]ItemDetail)
	for _, orderItem := range orderItems {
		itemDetailsMap = addItemDetails(itemDetailsMap, ItemDetail{
			ProductID:   orderItem.Product.ID,
			ProductName: orderItem.Product.Name,
			Quantity:    orderItem.Quantity,
			Price:       float32(orderItem.SubTotal),
			ImageURL:    orderItem.Product.ImageURL,
		})

		var selectedModifierGroups []models.SelectedModifierGroup
		err := s.db.Where("order_item_id = ?", orderItem.ID).Find(&selectedModifierGroups).Error
		if err != nil {
			return nil, err
		}

		for _, selectedModifierGroup := range selectedModifierGroups {
			modifierOptionID := selectedModifierGroup.ModifierOptionsID

			var modifierOptions []models.ModifierOptions
			err := s.db.Where("id = ?", modifierOptionID).Find(&modifierOptions).Error
			if err != nil {
				return nil, err
			}

			for _, modifierOption := range modifierOptions {
				itemDetailsMap = addItemDetails(itemDetailsMap, ItemDetail{
					ProductID:   modifierOption.ID,
					ProductName: modifierOption.Name,
					Quantity:    selectedModifierGroup.ModifierOptionQuantity,
					Price:       modifierOption.PriceAdjustment * float32(selectedModifierGroup.ModifierOptionQuantity),
				})
			}
		}
	}

	var itemDetails []ItemDetail
	for _, itemDetail := range itemDetailsMap {
		itemDetails = append(itemDetails, itemDetail)
	}

	transaction := Transaction{
		OrderID:         order.ID,
		ReceiptNo:       order.InvoiceNumber,
		TransactionDate: order.OrderDate,
		PaymentMethod:   order.PaymentMethod,
		Price:           float32(order.NetTotal),
		ItemDetails:     itemDetails,
		Status:          order.OrderStatus,
	}

	return &transaction, nil
}

//encore:api auth method=POST path=/api/admin/finances/payout-reports
func (s *Service) GetPayoutReport(ctx context.Context, params *GetPayoutReportRequest) (*GetPayoutReportResponse, error) {
	d := auth.Data()
	user := d.(*models.User)

	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	var expenses []models.ExpensesOutlet

	query := s.db.Preload("Outlet").Order("created_at DESC").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	queryCount := s.db.Model(&models.ExpensesOutlet{})

	// filter by outlet
	if params.OutletID != nil {
		query = query.Where("outlet_id = ?", params.OutletID)
		queryCount = queryCount.Where("outlet_id = ?", params.OutletID)
	}

	// filter by outlet group
	if params.OutletGroupID != nil {
		outletIDs, err := common_operations.GetOutletIDsByGroupID(s.db, *params.OutletGroupID)
		if err == nil {
			query = query.Where("outlet_id IN ?", outletIDs)
			queryCount = queryCount.Where("outlet_id IN ?", outletIDs)
		}
	}

	// filter by business
	if params.OutletID == nil && params.OutletGroupID == nil {
		outletsIDs := []uuid.UUID{}
		s.db.Model(&models.Outlet{}).Where("business_id = ?", user.BusinessID).Pluck("id", &outletsIDs)
		query = query.Where("outlet_id IN ?", outletsIDs)
		queryCount = queryCount.Where("outlet_id IN ?", outletsIDs)
	}

	if params.Date != (time.Time{}) {
		query = query.Where("DATE(expenses_date) = ?", params.Date.Format("2006-01-02"))
		queryCount = queryCount.Where("DATE(expenses_date) = ?", params.Date.Format("2006-01-02"))
	}

	if err := query.Find(&expenses).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := queryCount.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &GetPayoutReportResponse{
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: totalPages,
			Total:      total,
		},
		Data: expenses,
	}, nil
}

//encore:api auth method=GET path=/api/admin/finances/payout-details/:expensesOutletID
func (s *Service) GetPayoutDetails(ctx context.Context, expensesOutletID uuid.UUID) (*models.ExpensesOutlet, error) {
	var expenses models.ExpensesOutlet
	err := s.db.Model(&models.ExpensesOutlet{}).Where("id = ?", expensesOutletID).Find(&expenses).Error
	return &expenses, err
}

//encore:api auth method=POST path=/api/admin/finances/full-reports
func (s *Service) GetFullReport(ctx context.Context, params *GetFullReportRequest) (*GetFullReportResponse, error) {
	businessID := auth_service.GetUserBusinessID()

	startDate := time.Now()
	endDate := time.Now()
	if params.StartDate != nil && params.EndDate != nil {
		startDate = *params.StartDate
		endDate = *params.EndDate
	}
	startDate, _ = common.GetStartOfDay(startDate)
	endDate, _ = common.GetEndOfDay(endDate)

	var netTotal float32
	var salesProduct float32
	var salesModifier float32
	fullReports := []FullReport{}

	queryNetTotal := s.db.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as net_total").
		Where("orders.order_status = ?", models.OrderStatusCompleted)
	querySalesProduct := s.db.Model(&models.Order{}).
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Select("COALESCE(SUM(order_items.sub_total), 0) as sub_total").
		Where("orders.order_status = ?", models.OrderStatusCompleted)
	querySalesModifier := s.db.Model(&models.Order{}).
		Joins("JOIN order_items ON orders.id = order_items.order_id").
		Joins("JOIN selected_modifier_groups ON order_items.id = selected_modifier_groups.order_item_id").
		Joins("JOIN modifier_options ON selected_modifier_groups.modifier_options_id = modifier_options.id").
		Select("COALESCE(SUM(selected_modifier_groups.modifier_option_quantity * modifier_options.price_adjustment), 0) as modifier_option_quantity").
		Where("orders.order_status = ?", models.OrderStatusCompleted)
	queryFullReports := s.db.Model(&models.Outlet{}).
		Select("id as outlet_id, name as outlet_name, image_url as outlet_image_url")

	// filter by outlet
	if params.OutletID != nil {
		queryNetTotal = queryNetTotal.Where("outlet_id = ?", *params.OutletID)
		querySalesProduct = querySalesProduct.Where("orders.outlet_id = ?", *params.OutletID)
		querySalesModifier = querySalesModifier.Where("orders.outlet_id = ?", *params.OutletID)
		queryFullReports = queryFullReports.Where("id = ?", *params.OutletID)
	}

	// filter by outlet group
	if params.OutletGroupID != nil {
		outletIDs, err := common_operations.GetOutletIDsByGroupID(s.db, *params.OutletGroupID)
		if err == nil {
			queryNetTotal = queryNetTotal.Where("outlet_id IN (?)", outletIDs)
			querySalesProduct = querySalesProduct.Where("orders.outlet_id IN (?)", outletIDs)
			querySalesModifier = querySalesModifier.Where("orders.outlet_id IN (?)", outletIDs)
			queryFullReports = queryFullReports.Where("id IN (?)", outletIDs)
		}
	}

	// filter by business
	if params.OutletID == nil && params.OutletGroupID == nil {
		queryNetTotal = queryNetTotal.Where("business_id = ?", businessID)
		querySalesProduct = querySalesProduct.Where("orders.business_id = ?", businessID)
		querySalesModifier = querySalesModifier.Where("orders.business_id = ?", businessID)
		queryFullReports = queryFullReports.Where("business_id = ?", businessID)
	}

	// filter by date range
	queryNetTotal = queryNetTotal.Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC())
	querySalesProduct = querySalesProduct.Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC())
	querySalesModifier = querySalesModifier.Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC())
	if err := queryNetTotal.Scan(&netTotal).Error; err != nil {
		netTotal = 0
	}
	if err := querySalesProduct.Scan(&salesProduct).Error; err != nil {
		salesProduct = 0
	}
	if err := querySalesModifier.Scan(&salesModifier).Error; err != nil {
		salesModifier = 0
	}
	queryFullReports.Find(&fullReports)

	for i := 0; i < len(fullReports); i++ {
		var order models.Order
		if err := s.db.Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", fullReports[i].OutletID, startDate.UTC(), endDate.UTC()).First(&order).Error; err != nil {
			// Remove this outlet from fullReports if no orders found
			fullReports = append(fullReports[:i], fullReports[i+1:]...)
			i-- // Adjust index since we removed an element
			continue
		}
	}

	for i := range fullReports {
		fullReports[i] = s.getFullReportSaleDetails(fullReports[i].OutletID, fullReports[i], startDate, endDate)
		fullReports[i] = s.getOverallExpenses(fullReports[i].OutletID, fullReports[i], startDate, endDate)
	}

	fullReportResponse := GetFullReportResponse{
		TotalSalesInfo:     netTotal,
		TotalSalesProduct:  salesProduct,
		TotalSalesModifier: salesModifier,
		FullReports:        fullReports,
	}

	return &fullReportResponse, nil
}

func (s *Service) getFullReportSaleDetails(outletID uuid.UUID, fullReport FullReport, startDate time.Time, endDate time.Time) FullReport {
	var cashTotal float32
	var qrTotal float32
	var grabTotal float32
	var shopeeTotal float32

	s.db.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND payment_method = ?", outletID, startDate.UTC(), endDate.UTC(), models.CasePaymentCash).
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Select("COALESCE(SUM(net_total), 0)").Scan(&cashTotal)

	s.db.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND payment_method IN (?, ?)", outletID, startDate.UTC(), endDate.UTC(), models.CasePaymentStaticQR, models.CasePaymentEWallet).
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Select("COALESCE(SUM(net_total), 0)").Scan(&qrTotal)

	s.db.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND platform = ?", outletID, startDate.UTC(), endDate.UTC(), models.PlatformGrabFood).
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Select("COALESCE(SUM(net_total), 0)").Scan(&grabTotal)

	s.db.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND platform = ?", outletID, startDate.UTC(), endDate.UTC(), models.PlatformShopeeFood).
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Select("COALESCE(SUM(net_total), 0)").Scan(&shopeeTotal)

	fullReport.Cash = cashTotal
	fullReport.QR = qrTotal
	fullReport.Grab = grabTotal
	fullReport.Shopee = shopeeTotal
	fullReport.TotalSales = cashTotal + qrTotal + grabTotal + shopeeTotal
	return fullReport
}

func (s *Service) getOverallExpenses(outletID uuid.UUID, fullReport FullReport, startDate time.Time, endDate time.Time) FullReport {
	// Staff and Labor
	staffSalaryTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryStaffSalary, startDate, endDate)
	partTimerDailyTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryPartTimerDaily, startDate, endDate)

	// Utilities
	gasTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryGas, startDate, endDate)
	waterChargeTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryWaterCharge, startDate, endDate)

	// Kitchen and Cleaning Supplies
	dishwashingLiquidPasteTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryDishwashingLiquidPaste, startDate, endDate)
	cookingOilTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryCookingOil, startDate, endDate)
	tissueTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryTissue, startDate, endDate)
	garbageBagTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryGarbageBag, startDate, endDate)
	gloveTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryGlove, startDate, endDate)

	// Transportation
	transportationPetrolTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryTransportationPetrol, startDate, endDate)
	petrolAllowanceCrewTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryPetrolAllowanceCrew, startDate, endDate)
	upkeepOfVehicleTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryUpkeepOfVehicle, startDate, endDate)

	// Administrative
	licenseFeeTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryLicenseFee, startDate, endDate)
	printingAndStationeryTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryPrintingAndStationery, startDate, endDate)
	outletSuppliesTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryOutletSupplies, startDate, endDate)

	// Maintenance
	upkeepOfOutletTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryUpkeepOfOutlet, startDate, endDate)
	upkeepOfCentreHouseTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryUpkeepOfCentreHouse, startDate, endDate)

	// Other
	otherTotal := s.getExpensesByCategory(outletID, models.ExpenseCategoryOther, startDate, endDate)

	// Assign values to FullReport struct
	fullReport.ExpenseCategoryStaffSalary = staffSalaryTotal
	fullReport.ExpenseCategoryPartTimerDaily = partTimerDailyTotal
	fullReport.ExpenseCategoryGas = gasTotal
	fullReport.ExpenseCategoryWaterCharge = waterChargeTotal
	fullReport.ExpenseCategoryDishwashingLiquidPaste = dishwashingLiquidPasteTotal
	fullReport.ExpenseCategoryCookingOil = cookingOilTotal
	fullReport.ExpenseCategoryTissue = tissueTotal
	fullReport.ExpenseCategoryGarbageBag = garbageBagTotal
	fullReport.ExpenseCategoryGlove = gloveTotal
	fullReport.ExpenseCategoryTransportationPetrol = transportationPetrolTotal
	fullReport.ExpenseCategoryPetrolAllowanceCrew = petrolAllowanceCrewTotal
	fullReport.ExpenseCategoryUpkeepOfVehicle = upkeepOfVehicleTotal
	fullReport.ExpenseCategoryLicenseFee = licenseFeeTotal
	fullReport.ExpenseCategoryPrintingAndStationery = printingAndStationeryTotal
	fullReport.ExpenseCategoryOutletSupplies = outletSuppliesTotal
	fullReport.ExpenseCategoryUpkeepOfOutlet = upkeepOfOutletTotal
	fullReport.ExpenseCategoryUpkeepOfCentreHouse = upkeepOfCentreHouseTotal
	fullReport.ExpenseCategoryOther = otherTotal

	// Calculate total expenses
	fullReport.TotalExpenses = staffSalaryTotal + partTimerDailyTotal + gasTotal + waterChargeTotal +
		dishwashingLiquidPasteTotal + cookingOilTotal + tissueTotal + garbageBagTotal + gloveTotal +
		transportationPetrolTotal + petrolAllowanceCrewTotal + upkeepOfVehicleTotal + licenseFeeTotal +
		printingAndStationeryTotal + outletSuppliesTotal + upkeepOfOutletTotal + upkeepOfCentreHouseTotal + otherTotal

	return fullReport
}

func (s *Service) getExpensesByCategory(outletID uuid.UUID, category models.ExpenseCategory, startDate time.Time, endDate time.Time) float32 {
	var totalExpenses float32
	s.db.Model(&models.ExpensesOutlet{}).
		Where("outlet_id = ? AND expenses_date >= ? AND expenses_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Where("expenses_category = ?", category).
		Select("COALESCE(SUM(expenses_amount), 0)").Scan(&totalExpenses)

	return totalExpenses
}

//encore:api auth method=POST path=/api/admin/finances/overview-performance
func (s *Service) GetFinanceOverviewPerformance(ctx context.Context, params *GetFullReportRequest) (*FinanceOverviewPerformance, error) {
	trx := s.db

	startDate, _ := common.GetStartOfDay(*params.StartDate)
	endDate, _ := common.GetEndOfDay(*params.EndDate)

	// Sales By Product (Store Outlet, Grab, Shopee)
	platform := models.PlatformStoreOutlet
	salesByProductOutlet, err := s.GetSalesByProduct(params.BusinessID, params.OutletID, params.OutletGroupID, &platform, startDate, endDate, true, trx)
	if err != nil {
		salesByProductOutlet = []SalesByProduct{}
	}

	platform = models.PlatformGrabFood
	salesByProductGrab, err := s.GetSalesByProduct(params.BusinessID, params.OutletID, params.OutletGroupID, &platform, startDate, endDate, true, trx)
	if err != nil {
		salesByProductGrab = []SalesByProduct{}
	}

	platform = models.PlatformShopeeFood
	salesByProductShopee, err := s.GetSalesByProduct(params.BusinessID, params.OutletID, params.OutletGroupID, &platform, startDate, endDate, true, trx)
	if err != nil {
		salesByProductShopee = []SalesByProduct{}
	}

	// Sales By Modifier Options
	salesByModifierOptions, err := s.GetSalesByModifierOptions(params.BusinessID, params.OutletID, startDate, endDate, false, trx)
	if err != nil {
		salesByModifierOptions = []SalesByModifierOptions{}
	}

	// Sales By Payment Method
	salesByPaymentMethods, err := s.GetSalesByPaymentMethod(params.BusinessID, params.OutletID, startDate, endDate, trx)
	if err != nil {
		salesByPaymentMethods = []SalesByPaymentMethod{}
	}

	// Hourly Sales Performance
	hourlySalesPerformance, err := s.GetHourlySalesPerformance(params.BusinessID, params.OutletID, startDate, endDate, trx)
	if err != nil {
		hourlySalesPerformance = &HourlySalesPerformance{}
	}

	// Sort by number of products sold in descending order
	sort.Slice(salesByProductOutlet, func(i, j int) bool {
		return salesByProductOutlet[i].QuantitySold > salesByProductOutlet[j].QuantitySold
	})
	sort.Slice(salesByProductGrab, func(i, j int) bool {
		return salesByProductGrab[i].QuantitySold > salesByProductGrab[j].QuantitySold
	})
	sort.Slice(salesByProductShopee, func(i, j int) bool {
		return salesByProductShopee[i].QuantitySold > salesByProductShopee[j].QuantitySold
	})
	sort.Slice(salesByModifierOptions, func(i, j int) bool {
		return salesByModifierOptions[i].QuantitySold > salesByModifierOptions[j].QuantitySold
	})

	// Sort salesByPaymentMethod in descending order by TotalSales (total sales)
	sort.Slice(salesByPaymentMethods, func(i, j int) bool {
		return salesByPaymentMethods[i].TotalSales > salesByPaymentMethods[j].TotalSales
	})

	financeOverviewPerformance := FinanceOverviewPerformance{
		SalesByProductOutlet:   salesByProductOutlet,
		SalesByProductGrab:     salesByProductGrab,
		SalesByProductShopee:   salesByProductShopee,
		SalesByModifier:        salesByModifierOptions,
		SalesByPaymentMethod:   salesByPaymentMethods,
		HourlySalesPerformance: *hourlySalesPerformance,
	}

	return &financeOverviewPerformance, nil
}

func (s *Service) GetSalesByPaymentMethod(businessID *uuid.UUID, outletID *uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) ([]SalesByPaymentMethod, error) {

	var salesByPaymentMethods []SalesByPaymentMethod

	for _, paymentMethod := range models.PaymentMethods {
		var salesByPaymentMethod SalesByPaymentMethod
		query := trx.Model(&models.Order{}).
			Select("COALESCE(SUM(net_total), 0) as total_sales, COUNT(*) as total_quantity").
			Where("payment_method = ?", paymentMethod).
			Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC()).
			Where("orders.order_status = ?", models.OrderStatusCompleted)
		if outletID != nil {
			query = query.Where("outlet_id = ?", outletID)
		} else {
			query = query.Where("business_id = ?", businessID)
		}
		query.Scan(&salesByPaymentMethod)
		salesByPaymentMethod.PaymentMethod = paymentMethod

		salesByPaymentMethods = append(salesByPaymentMethods, salesByPaymentMethod)
	}

	salesByPaymentMethods = extractPaymentMethodToGrab(salesByPaymentMethods)

	return salesByPaymentMethods, nil
}

func extractPaymentMethodToGrab(salesByPaymentMethods []SalesByPaymentMethod) []SalesByPaymentMethod {
	// Find indices for "CASH" and "CASHLESS"
	var cashIndex, cashlessIndex int = -1, -1
	var cashTotalSales float32 = 0
	var cashlessTotalSales float32 = 0
	var cashTotalQuantity int = 0
	var cashlessTotalQuantity int = 0

	for i, pm := range salesByPaymentMethods {
		if pm.PaymentMethod == "CASH" {
			cashIndex = i
			cashTotalSales = pm.TotalSales
			cashTotalQuantity = pm.TotalQuantity
		}
		if pm.PaymentMethod == "CASHLESS" {
			cashlessIndex = i
			cashlessTotalSales = pm.TotalSales
			cashlessTotalQuantity = pm.TotalQuantity
		}
	}
	if cashIndex != -1 && cashlessIndex != -1 {
		// Remove CASHLESS first (higher index), then CASH
		if cashIndex > cashlessIndex {
			// Remove CASH
			salesByPaymentMethods = append(salesByPaymentMethods[:cashIndex], salesByPaymentMethods[cashIndex+1:]...)
			// Remove CASHLESS (its index is unchanged)
			salesByPaymentMethods = append(salesByPaymentMethods[:cashlessIndex], salesByPaymentMethods[cashlessIndex+1:]...)
		} else {
			// Remove CASHLESS
			salesByPaymentMethods = append(salesByPaymentMethods[:cashlessIndex], salesByPaymentMethods[cashlessIndex+1:]...)
			// Remove CASH (its index is unchanged)
			salesByPaymentMethods = append(salesByPaymentMethods[:cashIndex], salesByPaymentMethods[cashIndex+1:]...)
		}
		// Insert "grab" with summed values at the position of the first removed
		grab := SalesByPaymentMethod{
			PaymentMethod: "grab",
			TotalSales:    cashTotalSales + cashlessTotalSales,
			TotalQuantity: cashTotalQuantity + cashlessTotalQuantity,
		}
		// Insert at the lower of the two indices
		insertAt := cashIndex
		if cashlessIndex < cashIndex {
			insertAt = cashlessIndex
		}
		if insertAt > len(salesByPaymentMethods) {
			insertAt = len(salesByPaymentMethods)
		}
		salesByPaymentMethods = append(salesByPaymentMethods[:insertAt],
			append([]SalesByPaymentMethod{grab}, salesByPaymentMethods[insertAt:]...)...)
	}
	return salesByPaymentMethods
}

func (s *Service) GetHourlySalesPerformance(businessID *uuid.UUID, outletID *uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) (*HourlySalesPerformance, error) {

	var hourlySalesPerformance HourlySalesPerformance

	var totalSales float32
	query := trx.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as total_sales").
		Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC()).
		Where("orders.order_status = ?", models.OrderStatusCompleted)
	if outletID != nil {
		query = query.Where("outlet_id = ?", outletID)
	} else {
		query = query.Where("business_id = ?", businessID)
	}
	query.Scan(&totalSales)

	hourlySalesPerformance.TotalSales = totalSales
	// hourlySalesPerformance.HourlySales = []HourlySale{
	// 	{
	// 		TotalSales: 0,
	// 		Time:       "09:00",
	// 	},
	// }

	hour := 9
	startMinute := 0
	endMinute := 60

	for i := hour; i < 20; i++ {
		hourlySale := HourlySale{}
		query = trx.Model(&models.Order{}).
			Select("COALESCE(SUM(net_total), 0) as total_sales").
			Where("EXTRACT(HOUR FROM order_date AT TIME ZONE 'Asia/Kuala_Lumpur') = ? AND EXTRACT(MINUTE FROM order_date AT TIME ZONE 'Asia/Kuala_Lumpur') BETWEEN ? AND ? AND Date(order_date) BETWEEN ? AND ?",
				hour,
				startMinute,
				endMinute,
				startDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"))
		if outletID != nil {
			query = query.Where("outlet_id = ?", outletID)
		} else {
			query = query.Where("business_id = ?", businessID)
		}
		query.Scan(&hourlySale.TotalSales)
		hourlySale.Time = fmt.Sprintf("%d:%02d %s", hour%12, startMinute, map[bool]string{true: "PM", false: "AM"}[hour >= 12])
		hourlySalesPerformance.HourlySales = append(hourlySalesPerformance.HourlySales, hourlySale)
		hour++
	}

	return &hourlySalesPerformance, nil
}

func addItemDetails(itemDetailsMap map[uuid.UUID]ItemDetail, itemDetail ItemDetail) map[uuid.UUID]ItemDetail {
	productID := itemDetail.ProductID
	if _, ok := itemDetailsMap[productID]; !ok {
		itemDetailsMap[productID] = ItemDetail{
			ProductID:   productID,
			ProductName: itemDetail.ProductName,
			Quantity:    itemDetail.Quantity,
			Price:       itemDetail.Price,
			ImageURL:    itemDetail.ImageURL,
		}
	} else {
		existingItem := itemDetailsMap[productID]
		itemDetailsMap[productID] = ItemDetail{
			ProductID:   productID,
			ProductName: existingItem.ProductName,
			Quantity:    existingItem.Quantity + itemDetail.Quantity,
			Price:       existingItem.Price + itemDetail.Price,
			ImageURL:    existingItem.ImageURL,
		}
	}
	return itemDetailsMap
}

// encore:api auth method=GET path=/api/admin/finances/void-receipt/:order_id
func (s *Service) VoidReceipt(ctx context.Context, order_id uuid.UUID) error {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	var order models.Order
	result := trx.Where("id=?", order_id).First(&order)
	if result.Error != nil {
		trx.Rollback()
		return result.Error
	}
	order.OrderStatus = models.OrderStatusCancelled
	result = trx.Save(&order)
	if result.Error != nil {
		trx.Rollback()
		return result.Error
	}
	var transaction models.Transaction
	result = trx.Where("order_id = ?", order_id).First(&transaction)
	if result.Error != nil {
		trx.Rollback()
		return result.Error
	}
	transaction.PaymentStatus = models.PaymentStatusVoided
	result = trx.Save(&transaction)
	if result.Error != nil {
		trx.Rollback()
		return result.Error
	}

	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return result.Error
	}

	return nil
}

/*
 /$$$$$$$$ /$$                                                           /$$$ /$$$$$$$   /$$$$$$   /$$$$$$  /$$$
| $$_____/|__/                                                          /$$_/| $$__  $$ /$$__  $$ /$$__  $$|_  $$
| $$       /$$ /$$$$$$$   /$$$$$$  /$$$$$$$   /$$$$$$$  /$$$$$$        /$$/  | $$  \ $$| $$  \ $$| $$  \__/  \  $$
| $$$$$   | $$| $$__  $$ |____  $$| $$__  $$ /$$_____/ /$$__  $$      | $$   | $$$$$$$/| $$  | $$|  $$$$$$    | $$
| $$__/   | $$| $$  \ $$  /$$$$$$$| $$  \ $$| $$      | $$$$$$$$      | $$   | $$____/ | $$  | $$ \____  $$   | $$
| $$      | $$| $$  | $$ /$$__  $$| $$  | $$| $$      | $$_____/      |  $$  | $$      | $$  | $$ /$$  \ $$   /$$/
| $$      | $$| $$  | $$|  $$$$$$$| $$  | $$|  $$$$$$$|  $$$$$$$       \  $$$| $$      |  $$$$$$/|  $$$$$$/ /$$$/
|__/      |__/|__/  |__/ \_______/|__/  |__/ \_______/ \_______/        \___/|__/       \______/  \______/ |___/
*/

// api to get finance summary report for POS / Z-Report (for both admin and POS)
//
//encore:api auth method=POST path=/api/admin/finances/summary
func (s *Service) GetFinanceSummary(ctx context.Context, params *FinanceSummaryRequest) (*FinanceSummary, error) {

	trx := s.db

	startDate, _ := common.GetStartOfDay(params.StartDate)
	endDate, _ := common.GetEndOfDay(params.EndDate)

	// Sales Info
	salesInfo, err := s.GetSalesInfo(params.OutletID, startDate, endDate, trx)
	if err != nil {
		salesInfo = SalesInfo{}
	}
	// Sales By Category
	salesByCategory, err := s.GetSalesByCategory(params.OutletID, startDate, endDate, trx)
	if err != nil {
		salesByCategory = []SalesByCategory{}
	}
	// Sales By Product
	salesByProduct, err := s.GetSalesByProduct(nil, &params.OutletID, nil, nil, startDate, endDate, true, trx)
	if err != nil {
		salesByProduct = []SalesByProduct{}
	}
	// Sales By Modifier Group
	salesByModifierOptions, err := s.GetSalesByModifierOptions(nil, &params.OutletID, startDate, endDate, true, trx)
	if err != nil {
		salesByModifierOptions = []SalesByModifierOptions{}
	}
	// Sales By Employee
	salesByEmployee, err := s.GetSalesByEmployee(params.OutletID, startDate, endDate, trx)
	if err != nil {
		salesByEmployee = []SalesByEmployee{}
	}

	return &FinanceSummary{
		SalesInfo:              salesInfo,
		SalesByCategory:        salesByCategory,
		SalesByProduct:         salesByProduct,
		SalesByModifierOptions: salesByModifierOptions,
		SalesByEmployee:        salesByEmployee,
	}, nil
}

func (s *Service) GetSalesInfo(outletID uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) (SalesInfo, error) {

	var grossSales float32
	var totalDiscount float32
	var totalServiceCharge float32
	var totalTax float32
	var totalNetSales float32
	var totalRoundedNetSales float32
	var totalRoundedAmount float32
	var totalRoundedDownAmount float32
	var totalRoundedUpAmount float32
	var totalRedeemPoints int
	var totalCashRounding float32
	var totalCost float32
	var grossProfit float32
	var netProfit float32
	var noOfSalesTrans int
	var averageSalesPerTrans float32
	var noOfVoidedTrans int
	var totalVoidedAmount float32 //  it calculate from rounded net total
	var totalCustomerSignUpMembership int
	var memberSales float32    //  it calculate from rounded net total
	var nonMemberSales float32 //  it calculate from rounded net total
	var memberSalesQuantity int
	var nonMemberSalesQuantity int
	var noOfUnpaidOrders int
	var unpaidOrdersAmount float32 // it calculate from rounded net total
	var cashSales float32
	var duitnowQRSales float32
	var cardSales float32
	var cashClosing float32

	// count order table as this is base
	var orderCountAll int64

	// count order table with customer id (mean is member)
	var orderCountMembership int64
	result := trx.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.customer_id IS NOT NULL").
		Count(&orderCountMembership)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// count order table without customer id (mean is non member)
	var orderCountNonMembership int64
	result = trx.Model(&models.Order{}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("orders.customer_id IS NULL").
		Count(&orderCountNonMembership)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// total order count
	orderCountAll = orderCountMembership + orderCountNonMembership

	// gross sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(gross_total), 0) as gross_sales").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&grossSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// total discount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(discount_amount), 0) as total_discount").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalDiscount)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// total service charge
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(service_charge), 0) as total_service_charge").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalServiceCharge)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// total tax
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(tax_charge), 0) as total_tax").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalTax)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// net sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as total_net_sales").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalNetSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// rounded net sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_net_total), 0) as total_rounded_net_sales").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalRoundedNetSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// total rounded amount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_amount), 0) as total_rounded_amount").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalRoundedAmount)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// total voided number
	result = trx.Model(&models.Order{}).
		Select("COALESCE(COUNT(*), 0) as no_of_voided_trans").
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusVoided}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&noOfVoidedTrans)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// total voided amount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_net_total), 0) as total_voided_amount").
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusVoided}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&totalVoidedAmount)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// member sales amount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_net_total), 0) as member_sales").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("orders.customer_id IS NOT NULL").
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&memberSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// non member sales amount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_net_total), 0) as non_member_sales").
		Where("orders.order_status in (?)", []string{models.OrderStatusCompleted}).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Where("orders.customer_id IS NULL").
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&nonMemberSales)

	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	// unpaid orders amount
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(rounded_net_total), 0) as unpaid_orders").
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusPending}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&unpaidOrdersAmount)

	// unpaid orders number
	result = trx.Model(&models.Order{}).
		Select("COALESCE(COUNT(*), 0) as no_of_unpaid_orders").
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusPending}).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Scan(&noOfUnpaidOrders)

	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// total redeem points (future implementation)
	totalRedeemPoints = 0
	// total cash rounding (future implementation)
	totalCashRounding = 0

	// calculate total rounded down and up amount
	if totalRoundedAmount > 0 {
		totalRoundedUpAmount = totalRoundedAmount
	} else if totalRoundedAmount < 0 {
		totalRoundedDownAmount = totalRoundedAmount * -1
	} else {
		totalRoundedDownAmount = 0
		totalRoundedUpAmount = 0
	}
	// total cost
	totalCost, err := s.GetTotalCost(outletID, startDate, endDate, trx)
	if err != nil {
		return SalesInfo{}, err
	}
	// gross profit
	grossProfit = grossSales - totalCost
	// net profit
	netProfit = totalNetSales - totalCost
	// no of sales trans
	noOfSalesTrans = int(orderCountAll)
	// Calculate average sales per transaction
	if noOfSalesTrans == 0 {
		averageSalesPerTrans = 0
	} else {
		averageSalesPerTrans = totalNetSales / float32(noOfSalesTrans)
	}

	// total customer sign up membership (future implementation)
	totalCustomerSignUpMembership = 0

	// member sales quantity
	memberSalesQuantity = int(orderCountMembership)
	// non member sales quantity
	nonMemberSalesQuantity = int(orderCountNonMembership)

	// cash sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as cash_sales").
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND payment_method = ?", outletID, startDate.UTC(), endDate.UTC(), models.CasePaymentCash).
		Scan(&cashSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// duitnow qr sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as duitnow_qr_sales").
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND (payment_method = ? OR payment_method = ? OR payment_method = ?)", outletID, startDate.UTC(), endDate.UTC(), models.CasePaymentEWallet, models.CasePaymentStaticQR, models.CasePaymentStaticQR).
		Scan(&duitnowQRSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}
	// card sales
	result = trx.Model(&models.Order{}).
		Select("COALESCE(SUM(net_total), 0) as card_sales").
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Where("outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ? AND (payment_method = ? OR payment_method = ?)", outletID, startDate.UTC(), endDate.UTC(), models.CasePaymentCreditCard, models.CasePaymentDebitCard).
		Scan(&cardSales)
	if result.Error != nil {
		return SalesInfo{}, result.Error
	}

	var stockReports []models.StockReport
	result = trx.Model(&models.StockReport{}).
		Select("created_at, cash_closing").
		Where("outlet_id = ?", outletID).
		Where("created_at >= ? AND created_at < ?", startDate.UTC(), endDate.UTC()).
		Where("cash_closing > 0").
		Find(&stockReports)

	if result.Error != nil {
		cashClosing = 0
	} else {
		var dateList []string
		for _, stockReport := range stockReports {
			date := stockReport.CreatedAt.Format("2006-01-02")
			if !common.ArrayContains(dateList, date) {
				dateList = append(dateList, date)
				cashClosing += float32(stockReport.CashClosing.InexactFloat64())
			}
		}
	}

	return SalesInfo{
		GrossSales:                    grossSales,
		TotalDiscount:                 totalDiscount,
		TotalServiceCharge:            totalServiceCharge,
		TotalTax:                      totalTax,
		TotalNetSales:                 totalNetSales,
		TotalRoundedNetSales:          totalRoundedNetSales,
		TotalRoundedDownAmount:        totalRoundedDownAmount,
		TotalRoundedUpAmount:          totalRoundedUpAmount,
		TotalRedeemPoints:             totalRedeemPoints,
		TotalCashRounding:             totalCashRounding,
		TotalCost:                     totalCost,
		GrossProfit:                   grossProfit,
		NetProfit:                     netProfit,
		NoOfSalesTrans:                noOfSalesTrans,
		AverageSalesPerTrans:          averageSalesPerTrans,
		NoOfVoidedTrans:               noOfVoidedTrans,
		TotalVoidedAmount:             totalVoidedAmount,
		TotalCustomerSignUpMembership: totalCustomerSignUpMembership,
		MemberSales:                   memberSales,
		NonMemberSales:                nonMemberSales,
		MemberSalesQuantity:           memberSalesQuantity,
		NonMemberSalesQuantity:        nonMemberSalesQuantity,
		NoOfUnpaidOrders:              noOfUnpaidOrders,
		UnpaidOrdersAmount:            unpaidOrdersAmount,
		CashSales:                     cashSales,
		DuitnowQRSales:                duitnowQRSales,
		CardSales:                     cardSales,
		CashClosing:                   cashClosing,
	}, nil

}

func (s *Service) GetTotalCost(outletID uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) (float32, error) {
	// fetch all order items for outler and date in one query
	var orderItems []models.OrderItem
	result := trx.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Where("orders.outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Find(&orderItems)
	if result.Error != nil {
		return 0, result.Error
	}

	// gather product IDs
	productIDs := make([]uuid.UUID, 0, len(orderItems))
	for _, item := range orderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// fetch all product ingredient mappings for product IDs in one query
	var productIngredientMappings []models.ProductIngredientMapping
	result = trx.Where("product_id IN ?", productIDs).Find(&productIngredientMappings)
	if result.Error != nil {
		return 0, result.Error
	}

	// gather ingredient IDs
	ingredientIDs := make([]uuid.UUID, 0, len(productIngredientMappings))
	for _, item := range productIngredientMappings {
		ingredientIDs = append(ingredientIDs, item.IngredientID)
	}
	// fetch all ingredients for ingredient IDs in one query
	var ingredients []models.Ingredient
	result = trx.Where("id IN ?", ingredientIDs).Find(&ingredients)
	if result.Error != nil {
		return 0, result.Error
	}

	// create map for quick lookup
	ingredientMap := make(map[uuid.UUID]models.Ingredient)
	for _, ing := range ingredients {
		ingredientMap[ing.ID] = ing
	}

	// calculate total cost
	var totalCost float32
	for _, item := range orderItems {
		// loop through all product ingredient mappings for the product
		for _, piMapping := range productIngredientMappings {
			// check if the product ingredient mapping is for the product
			if piMapping.ProductID == item.ProductID {
				// get ingredient information
				ing := ingredientMap[piMapping.IngredientID]
				// convert quantity to standard unit
				quantity := constants.ConvertToTargetUnit(piMapping.Unit, float32(piMapping.Quantity), ing.Unit)
				cost := ing.PricePerUnit * (quantity / ing.Quantity) * float32(item.Quantity)
				totalCost += cost
			}
		}
	}
	return totalCost, nil
}

// total cost orders -> order_items -> product_id -> product_ingredient_mappings -> ingredient_id -> ingredient
func (s *Service) GetTotalCostLegacy(outletID uuid.UUID, date time.Time, trx *gorm.DB) (float32, error) {
	var orders []models.Order
	result := trx.Model(&models.Order{}).
		Where("outlet_id = ? AND Date(order_date)=?", outletID, date.Format("2006-01-02")).
		Find(&orders)
	if result.Error != nil {
		return 0, result.Error
	}

	var totalCost float32
	for _, order := range orders {
		var orderItems []models.OrderItem
		result = trx.Model(&models.OrderItem{}).
			Where("order_id = ?", order.ID).
			Find(&orderItems)
		if result.Error != nil {
			return 0, result.Error
		}
		for _, orderItem := range orderItems {
			cost, err := s.ProductCostCalculation(orderItem.ProductID, trx)
			if err != nil {
				return 0, err
			}
			totalCost += cost
		}
	}

	return totalCost, nil
}

// the cost of product is the sum of the cost of all ingredients and the cost of the product itself (one product)
func (s *Service) ProductCostCalculation(productID uuid.UUID, trx *gorm.DB) (float32, error) {

	// get product and related ingredients in product_ingredient_mappings table
	var productIngredients []models.ProductIngredientMapping
	result := trx.Where("product_id = ?", productID).Find(&productIngredients)
	if result.Error != nil {
		return 0, result.Error
	}

	var totalCost float32
	for _, prodproductIngredient := range productIngredients {
		// get ingredient information from ingredient table
		var ingredient models.Ingredient
		result = trx.Where("id = ?", prodproductIngredient.IngredientID).First(&ingredient)
		if result.Error != nil {
			return 0, result.Error
		}

		pricePerUnit := ingredient.PricePerUnit
		quantityOfProductIngredient := prodproductIngredient.Quantity
		standardUnit := ingredient.Unit
		unitInProductIngredient := prodproductIngredient.Unit

		// convert unit to standard unit
		quantityOfProductIngredient = constants.ConvertToTargetUnit(unitInProductIngredient, quantityOfProductIngredient, standardUnit)

		// calculate cost of product ingredient
		costOfProductIngredient := pricePerUnit * quantityOfProductIngredient

		// add to total cost
		totalCost += costOfProductIngredient
	}

	return totalCost, nil

}

// GetSalesByCategory calculates sales statistics grouped by product category for a specific outlet and date
//
// Process flow:
// 1. Fetches all order items for the given outlet and date
// 2. Collects all product IDs from those order items
// 3. Loads product category mappings (with preloaded categories) for those products
// 4. Aggregates sales data (gross sales, quantity sold, discounts) by category
//
// Returns:
// - []SalesByCategory: Slice containing sales data grouped by product category
// - error: Any database error that occurred during the process
//
// Data sources:
// - order_items: Contains product quantities and subtotals
// - orders: Filters by outlet and date
// - product_category_mappings: Links products to their categories
// - product_categories: Provides category details (preloaded)
func (s *Service) GetSalesByCategory(outletID uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) ([]SalesByCategory, error) {

	var outlet models.Outlet
	result := trx.Model(&models.Outlet{}).
		Where("id = ?", outletID).
		Find(&outlet)
	if result.Error != nil {
		return []SalesByCategory{}, result.Error
	}

	var productCategories []models.ProductCategory
	result = trx.Model(&models.ProductCategory{}).
		Where("business_id = ?", outlet.BusinessID).
		Find(&productCategories)
	if result.Error != nil {
		return []SalesByCategory{}, result.Error
	}

	var orderItems []models.OrderItem
	result = trx.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		//Joins("JOIN products ON products.id = order_items.product_id").
		Where("orders.outlet_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, startDate.UTC(), endDate.UTC()).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted}).
		Find(&orderItems)
	if result.Error != nil {
		return []SalesByCategory{}, result.Error
	}

	// gather product IDs
	productIDs := make([]uuid.UUID, 0, len(orderItems))
	for _, item := range orderItems {
		productIDs = append(productIDs, item.ProductID)
	}

	var productCategoryMappings []models.ProductCategoryMapping
	result = trx.Model(&models.ProductCategoryMapping{}).
		Preload("ProductCategory").
		Where("product_id IN ?", productIDs).
		//Group("product_category_id").
		Find(&productCategoryMappings)
	if result.Error != nil {
		return []SalesByCategory{}, result.Error
	}

	// map product id (key) to product category (value)
	productCategoryMap := make(map[uuid.UUID]models.ProductCategory)
	for _, mapping := range productCategoryMappings {
		productCategoryMap[mapping.ProductID] = mapping.ProductCategory
	}

	// calculate sales by category
	categorySales := make(map[uuid.UUID]SalesByCategory)
	for _, item := range orderItems {
		category, exists := productCategoryMap[item.ProductID]
		if !exists {
			continue
		}
		_, ok := categorySales[category.ID]
		if !ok {
			categorySales[category.ID] = SalesByCategory{
				ProductCategory: category,
				GrossSales:      item.SubTotal,
				QuantitySold:    item.Quantity,
				TotalDiscount:   item.Order.DiscountAmount,
			}
		} else {
			existing := categorySales[category.ID]
			existing.GrossSales += item.SubTotal
			existing.QuantitySold += item.Quantity
			existing.TotalDiscount += item.Order.DiscountAmount
			categorySales[category.ID] = existing
		}
	}
	// add other categories which are not in the categorySales map
	for _, category := range productCategories {
		_, exists := categorySales[category.ID]
		if !exists {
			categorySales[category.ID] = SalesByCategory{
				ProductCategory: category,
				GrossSales:      0,
				QuantitySold:    0,
				TotalDiscount:   0,
			}
		}
	}

	var salesByCategory []SalesByCategory
	for _, sales := range categorySales {
		salesByCategory = append(salesByCategory, sales)
	}

	return salesByCategory, nil

}

// GetSalesByProduct calculates sales statistics grouped by product for a specific outlet and date
func (s *Service) GetSalesByProduct(businessID *uuid.UUID, outletID *uuid.UUID, outletGroupID *uuid.UUID, platform *models.Platform, startDate time.Time, endDate time.Time, excludeProductThatIsModifier bool, trx *gorm.DB) ([]SalesByProduct, error) {
	// First get the business ID for this outlet
	if businessID == nil {
		var outlet models.Outlet
		if err := trx.Where("id = ?", outletID).First(&outlet).Error; err != nil {
			return []SalesByProduct{}, err
		}
		businessID = &outlet.BusinessID
	}

	// Fetch all products for this business
	// TODO: To add filter by order platform as well
	var products []models.Product
	queryProducts := trx.Model(&models.Product{}).
		Where("business_id = ?", *businessID)

	if platform != nil {
		switch *platform {
		case models.PlatformStoreOutlet:
			queryProducts = queryProducts.Where("is_store_outlet = ?", true)
		case models.PlatformGrabFood:
			queryProducts = queryProducts.Where("is_grab_food = ?", true)
		case models.PlatformShopeeFood:
			queryProducts = queryProducts.Where("is_shopee_food = ?", true)
		}
	}

	if excludeProductThatIsModifier {
		queryProducts = queryProducts.Where("modifier_options_id IS NULL")
	}

	result := queryProducts.Find(&products)

	if result.Error != nil {
		return []SalesByProduct{}, result.Error
	}

	// Fetch all order items for the outlet and date in a single query
	var orderItems []models.OrderItem
	utcStartDate := startDate.UTC()
	utcEndDate := endDate.UTC()
	query := trx.Model(&models.OrderItem{}).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Preload("Order", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "discount_amount", "payment_method")
		}).
		Preload("Product", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "modifier_options_id")
		}).
		Where("orders.order_date >= ? AND orders.order_date < ?", utcStartDate, utcEndDate).
		Where("orders.order_status = ?", models.OrderStatusCompleted).
		Where("orders.payment_status IN (?)", []string{models.PaymentStatusCompleted})

	query = query.Where("orders.business_id = ?", businessID)

	if outletID != nil {
		query = query.Where("orders.outlet_id = ?", outletID)
	}

	if outletGroupID != nil {
		outletIDs, err := common_operations.GetOutletIDsByGroupID(trx, *outletGroupID)
		if err == nil {
			query = query.Where("orders.outlet_id IN (?)", outletIDs)
		}
	}

	result = query.Find(&orderItems)

	if result.Error != nil {
		return []SalesByProduct{}, result.Error
	}

	// Group order items by product ID
	productSales := make(map[uuid.UUID]struct {
		GrossSales    float32
		QuantitySold  int
		TotalDiscount float32
	})
	productSoldByPaymentMethod := make(map[uuid.UUID]map[string]int)

	for _, item := range orderItems {
		if excludeProductThatIsModifier && item.Product.ModifierOptionsID != nil && *item.Product.ModifierOptionsID != uuid.Nil {
			continue
		}
		entry := productSales[item.ProductID]
		entry.GrossSales += item.SubTotal
		entry.QuantitySold += item.Quantity
		entry.TotalDiscount += item.Order.DiscountAmount
		productSales[item.ProductID] = entry

		// Count product sold by payment method
		if _, exists := productSoldByPaymentMethod[item.ProductID]; !exists {
			productSoldByPaymentMethod[item.ProductID] = make(map[string]int)
		}
		productSoldByPaymentMethod[item.ProductID][item.Order.PaymentMethod] += item.Quantity
	}

	// Calculate costs and profits in bulk
	productCosts := make(map[uuid.UUID]float32)
	for _, product := range products {
		cost, err := s.ProductCostCalculation(product.ID, trx)
		if err != nil {
			return []SalesByProduct{}, err
		}
		productCosts[product.ID] = cost
	}

	// Build the final result including products with no sales
	var salesByProduct []SalesByProduct
	for _, product := range products {
		sales, exists := productSales[product.ID]
		if !exists {
			sales = struct {
				GrossSales    float32
				QuantitySold  int
				TotalDiscount float32
			}{0, 0, 0}
		}

		soldByPaymentMethod, exists := productSoldByPaymentMethod[product.ID]
		if !exists {
			soldByPaymentMethod = nil
		}

		cost := productCosts[product.ID]
		// calculate total profit
		var totalProfit float32
		if sales.GrossSales == 0 && sales.QuantitySold == 0 {
			totalProfit = 0
		} else {
			totalProfit = sales.GrossSales - cost
		}
		salesByProduct = append(salesByProduct, SalesByProduct{
			Product:             product,
			GrossSales:          sales.GrossSales,
			QuantitySold:        sales.QuantitySold,
			TotalCost:           cost,
			TotalDiscount:       sales.TotalDiscount,
			TotalProfit:         totalProfit,
			SoldByPaymentMethod: soldByPaymentMethod,
		})
	}

	return salesByProduct, nil
}

// GetSalesByModifierOptions calculates sales statistics grouped by modifier options for a specific outlet and date
func (s *Service) GetSalesByModifierOptions(businessID *uuid.UUID, outletID *uuid.UUID, startDate time.Time, endDate time.Time, excludeProductThatIsModifier bool, trx *gorm.DB) ([]SalesByModifierOptions, error) {
	if businessID == nil {
		// First get the business ID for this outlet
		var outlet models.Outlet
		if err := trx.Where("id = ?", outletID).First(&outlet).Error; err != nil {
			return []SalesByModifierOptions{}, err
		}
		businessID = &outlet.BusinessID
	}

	// get modifier options only for this business
	var modifierOptions []models.ModifierOptions
	result := trx.Model(&models.ModifierOptions{}).
		Joins("JOIN modifier_groups ON modifier_groups.id = modifier_options.modifier_group_id").
		Preload("ModifierGroup").
		Where("modifier_groups.business_id = ?", *businessID).
		Find(&modifierOptions)
	if result.Error != nil {
		return []SalesByModifierOptions{}, result.Error
	}

	modifierOptionsMap := make(map[uuid.UUID]models.ModifierOptions)
	for _, mo := range modifierOptions {
		modifierOptionsMap[mo.ID] = mo
	}

	// map the modifier options as key
	salesMap := make(map[uuid.UUID]SalesByModifierOptions)

	var selectedModifierGroup []models.SelectedModifierGroup
	query := trx.Model(&models.SelectedModifierGroup{}).
		Joins("JOIN order_items ON order_items.id = selected_modifier_groups.order_item_id").
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Preload("OrderItem.Order", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "discount_amount") // only fetch these columns
		}).
		Where("orders.order_date >= ? AND orders.order_date < ?", startDate.UTC(), endDate.UTC()).
		Where("orders.order_status = ?", models.OrderStatusCompleted)

	if outletID != nil {
		query = query.Where("orders.outlet_id = ?", outletID)
	} else {
		query = query.Where("orders.business_id = ?", businessID)
	}

	result = query.Find(&selectedModifierGroup)
	if result.Error != nil {
		return []SalesByModifierOptions{}, result.Error
	}

	// count modifiers
	for _, smg := range selectedModifierGroup {
		mo, exists := modifierOptionsMap[smg.ModifierOptionsID]
		if !exists {
			continue
		}

		entry := salesMap[mo.ID]
		entry.ModifierOptions = mo
		entry.GrossSales += float32(smg.ModifierOptionQuantity) * mo.PriceAdjustment
		entry.QuantitySold += smg.ModifierOptionQuantity
		entry.TotalDiscount += smg.OrderItem.Order.DiscountAmount
		salesMap[mo.ID] = entry
	}

	// count product that is flag as modifier
	if !excludeProductThatIsModifier {
		// Query product in order item
		var orderItems []models.OrderItem
		query := trx.Model(&models.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Joins("JOIN products ON products.id = order_items.product_id").
			Preload("Product").
			Where("orders.order_date >= ? AND orders.order_date < ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
			Where("products.modifier_options_id IS NOT NULL").
			Where("orders.order_status = ?", models.OrderStatusCompleted)

		if outletID != nil {
			query = query.Where("orders.outlet_id = ?", outletID)
		} else {
			query = query.Where("orders.business_id = ?", businessID)
		}
		result = query.Find(&orderItems)
		if result.Error != nil {
			return []SalesByModifierOptions{}, result.Error
		}

		// count product that is flag as modifier
		for _, item := range orderItems {
			mo, exists := modifierOptionsMap[*item.Product.ModifierOptionsID]
			if !exists {
				continue
			}

			entry := salesMap[mo.ID]
			entry.ModifierOptions = mo
			entry.GrossSales += float32(item.Quantity) * item.Product.Price
			entry.QuantitySold += item.Quantity
			entry.TotalDiscount += item.Order.DiscountAmount
			salesMap[mo.ID] = entry
		}
	}

	// add other modifier options which are not in the salesMap
	for _, mo := range modifierOptions {
		_, exists := salesMap[mo.ID]
		if !exists {
			salesMap[mo.ID] = SalesByModifierOptions{
				ModifierOptions: mo,
				GrossSales:      0,
				QuantitySold:    0,
				TotalDiscount:   0,
			}
		}

	}

	var salesByModifierOptions []SalesByModifierOptions
	for _, sales := range salesMap {
		salesByModifierOptions = append(salesByModifierOptions, sales)
	}
	return salesByModifierOptions, nil

}

// GetSalesByEmployee calculates sales statistics grouped by employee for a specific outlet and date
func (s *Service) GetSalesByEmployee(outletID uuid.UUID, startDate time.Time, endDate time.Time, trx *gorm.DB) ([]SalesByEmployee, error) {
	/* Employee      models.User `json:"employee"`
	NumberOfSales int         `json:"number_of_sales"`
	TotalNetSales float32     `json:"total_net_sales"` */
	var outlet models.Outlet
	result := trx.Model(&models.Outlet{}).
		Where("id = ?", outletID).
		Find(&outlet)
	if result.Error != nil {
		return []SalesByEmployee{}, result.Error
	}

	var employees []models.User
	result = trx.Model(&models.User{}).
		Where("business_id = ? AND outlet_id = ?", outlet.BusinessID, outletID).
		Find(&employees)
	if result.Error != nil {
		return []SalesByEmployee{}, result.Error
	}

	var employeeSales []SalesByEmployee
	for _, employee := range employees {
		var numberOfSales int64
		result = trx.Model(&models.Order{}).
			Select("COUNT(*) as number_of_sales").
			Where("outlet_id = ? AND user_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, employee.ID, startDate.UTC(), endDate.UTC()).
			Scan(&numberOfSales)
		if result.Error != nil {
			return []SalesByEmployee{}, result.Error
		}

		var totalNetSales float32
		result = trx.Model(&models.Order{}).
			Select("COALESCE(SUM(net_total), 0) as total_net_sales").
			Where("outlet_id = ? AND user_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, employee.ID, startDate.UTC(), endDate.UTC()).
			Scan(&totalNetSales)
		if result.Error != nil {
			return []SalesByEmployee{}, result.Error
		}

		var totalRoundedNetSales float32
		result = trx.Model(&models.Order{}).
			Select("COALESCE(SUM(rounded_net_total), 0) as total_rounded_net_sales").
			Where("outlet_id = ? AND user_id = ? AND orders.order_date >= ? AND orders.order_date < ?", outletID, employee.ID, startDate.UTC(), endDate.UTC()).
			Scan(&totalRoundedNetSales)
		if result.Error != nil {
			return []SalesByEmployee{}, result.Error
		}

		employeeSales = append(employeeSales, SalesByEmployee{
			Employee:             employee,
			NumberOfSales:        int(numberOfSales),
			TotalNetSales:        totalNetSales,
			TotalRoundedNetSales: totalRoundedNetSales,
		})
	}

	return employeeSales, nil

}
