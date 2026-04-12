package stocks

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"encore.app/auth_service"
	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/common_operations"
	"encore.app/sante_admin/outlets"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/firebase"
	"encore.app/middleware"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator" // For generating new UUIDs
	googleUUID "github.com/google/uuid" // For generating new UUIDs
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Service{db: db}, nil
}

type TransferInStockRequest struct {
	OutletID               uuid.UUID `json:"outlet_id"`
	IngredientID           uuid.UUID `json:"ingredient_id"`
	AmountOfIngredientItem int       `json:"amount_of_ingredient_item" valid:"required~Amount of ingredient item is required"`
	TransferType           string    `json:"transfer_type" valid:"required~Transfer type is required,in(in|out)~Transfer type must be either 'in' or 'out'"`
}

type UpdateStockRequest struct {
	OutletID           uuid.UUID `json:"outlet_id"`
	StockID            uuid.UUID `json:"stock_id"`
	SmallScaleUnit     string    `json:"small_scale_unit"`
	LargeScaleUnit     string    `json:"large_scale_unit"`
	SmallScaleQuantity float32   `json:"small_scale_quantity"`
	LargeScaleQuantity float32   `json:"large_scale_quantity"`
}

type GetAllStockOutletResponse struct {
	//Data []models.Stock `json:"data"`
	Data []StockWithSelectedUnit `json:"data"`
}

type StockWithSelectedUnit struct {
	models.Stock
	SelectedUnit string `json:"selected_unit"`
}

type OpenStockReportRequest struct {
	OutletID uuid.UUID `json:"outlet_id"`
	Stocks   []struct {
		IngredientID uuid.UUID `json:"ingredient_id"`
		Opening      float32   `json:"opening" valid:"required~Opening is required"`
		OpeningUnit  string    `json:"opening_unit" valid:"required~Opening unit is required"`
	} `json:"stocks" valid:"required~Stocks is required"`
	CashOpening float32 `json:"cash_opening" valid:"required~Cash opening is required"`
}

type CloseStockReportRequest struct {
	OutletID          uuid.UUID                               `json:"outlet_id"`
	ProductWastage    []CloseStockReportRequestProductWastage `json:"product_wastage"`
	IngredientWastage []struct {
		IngredientID          uuid.UUID `json:"ingredient_id"`
		NoOfIngredientWastage float32   `json:"no_of_ingredient_wastage"`
		UnitSelected          string    `json:"unit_selected"`
	} `json:"ingredient_wastage"`
	CashClosing float32 `json:"cash_closing"`
	Expenses    []struct {
		ExpensesID          uuid.UUID              `json:"expenses_id"`
		ExpensesCategory    models.ExpenseCategory `json:"expenses_category"`
		ExpensesDescription string                 `json:"expenses_description"`
		ExpensesAmount      float32                `json:"expenses_amount"`
		ExpensesDate        time.Time              `json:"expenses_date"`
	} `json:"expenses"`
	IngredientStocksClosing []struct {
		IngredientID uuid.UUID `json:"ingredient_id"`
		Closing      float32   `json:"closing"`
		UnitSelected string    `json:"unit_selected"`
	} `json:"ingredient_stocks_closing"`
}

type CloseStockReportRequestProductWastage struct {
	ProductID     uuid.UUID `json:"product_id"`
	WastageTypeID uuid.UUID `json:"wastage_type_id"`
	WastageAmount float32   `json:"wastage_amount"`
	ReportDate    time.Time `json:"report_date"`
	Notes         string    `json:"notes"`
}

type CloseStockReportResponse struct {
	Message string            `json:"message"`
	Data    []AddExpensesData `json:"data"`
}

type RecordProductWastageRequest struct {
	BusinessID     uuid.UUID                               `json:"business_id"`
	OutletID       uuid.UUID                               `json:"outlet_id"`
	ProductWastage []CloseStockReportRequestProductWastage `json:"product_wastage"`
}

type RecordIngredientWastageRequest struct {
	OutletID          uuid.UUID `json:"outlet_id"`
	IngredientWastage []struct {
		IngredientID          uuid.UUID `json:"ingredient_id"`
		NoOfIngredientWastage float32   `json:"no_of_ingredient_wastage"`
		UnitSelected          string    `json:"unit_selected"`
	} `json:"ingredient_wastage"`
}
type GetStockReportsRequest struct {
	BusinessID    *uuid.UUID `json:"business_id"`
	OutletID      *uuid.UUID `json:"outlet_id"`
	OutletGroupID *uuid.UUID `json:"outlet_group_id"`
	FromDate      *time.Time `json:"from_date"`
	ToDate        *time.Time `json:"to_date"`
}

type GetStockReportsResponse struct {
	Data []models.StockReport `json:"data"`
}

type CheckOutletStockOpeningResponse struct {
	Message            string `json:"message"`
	IsStockOpeningDone bool   `json:"is_stock_opening_done"`
}

type GetAllStockOpeningReportByDaysRequest struct {
	OutletID   uuid.UUID `json:"outlet_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SearchKey  string    `json:"search_key"`
}

type GetAllStockOpeningReportByDaysResponse struct {
	Message string `json:"message"`
	Data    []struct {
		Date             time.Time            `json:"date"`
		Count            int                  `json:"count"`
		StockOpeningData []models.StockReport `json:"stock_opening_data"`
	} `json:"data"`
}

type GetCashOpeningResponse struct {
	Message string `json:"message"`
	Data    struct {
		Date        time.Time `json:"date"`
		CashOpening float32   `json:"cash_opening"`
	} `json:"data"`
}

type AddExpensesRequest struct {
	OutletID uuid.UUID `json:"outlet_id"`
	Expenses []struct {
		ExpensesCategory    models.ExpenseCategory `json:"expenses_category" valid:"required~Expenses category is required"`
		ExpensesDescription string                 `json:"expenses_description" valid:"required~Expenses description is required"`
		ExpensesAmount      float32                `json:"expenses_amount"`
		ExpensesDate        time.Time              `json:"expenses_date" valid:"required~Expenses date is required"`
	} `json:"expenses"`
}

type ModifyExpensesRequest struct {
	OutletID uuid.UUID `json:"outlet_id"`
	Expenses []struct {
		ExpensesID          uuid.UUID              `json:"expenses_id"`
		ExpensesCategory    models.ExpenseCategory `json:"expenses_category" valid:"required~Expenses category is required"`
		ExpensesDescription string                 `json:"expenses_description" valid:"required~Expenses description is required"`
		ExpensesAmount      float32                `json:"expenses_amount"`
		ExpensesDate        time.Time              `json:"expenses_date" valid:"required~Expenses date is required"`
	} `json:"expenses"`
	Timezone *string `json:"timezone"`
}

type AddExpensesResponse struct {
	Message string            `json:"message"`
	Data    []AddExpensesData `json:"data"`
}

type AddExpensesData struct {
	ExpensesOutletID uuid.UUID `json:"expenses_outlet_id"`
}

type GetAllExpensesOfOutletByDaysRequest struct {
	OutletID   uuid.UUID `json:"outlet_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SearchKey  string    `json:"search_key"`
}

type GetAllExpensesOfOutletByDaysResponse struct {
	Message string                             `json:"message"`
	Data    []GetAllExpensesOfOutletByDaysData `json:"data"`
}

type GetAllExpensesOfOutletByDaysData struct {
	Date               time.Time               `json:"date"`
	TotalExpensesOfDay float32                 `json:"total_expenses_of_day"`
	Expenses           []models.ExpensesOutlet `json:"expenses"`
}

type CreateStockRequestRequest struct {
	RequesterOutletID uuid.UUID `json:"requester_outlet_id"`
	ResponderOutletID uuid.UUID `json:"responder_outlet_id"`
	RequestDate       time.Time `json:"request_date" valid:"required~Request date is required"`
	Remarks           string    `json:"remarks" valid:"required~Remarks is required"`
	//StockRequestedItems []models.StockRequestedItem `json:"stock_requested_items" valid:"required~Stock requested items is required"`
	StockRequestedItems []struct {
		IngredientID uuid.UUID `json:"ingredient_id"`
		UnitSelected string    `json:"unit_selected" `
		Quantity     float32   `json:"quantity" valid:"required~Quantity is required"`
	} `json:"stock_request_items" valid:"required~Stock request items is required"`
}

type GetAllStockRequestAndApprovalResponse struct {
	Message string                            `json:"message"`
	Data    GetAllStockRequestAndApprovalData `json:"data"`
}

type GetAllStockRequestAndApprovalData struct {
	StockRequest []struct {
		Date             time.Time             `json:"date"`
		StockRequestList []models.StockRequest `json:"stock_request_list"`
	} `json:"stock_request"`
	StockRequestApproval []struct {
		Date             time.Time             `json:"date"`
		StockRequestList []models.StockRequest `json:"stock_request_approval_list"`
	} `json:"stock_request_approval"`

	//StockRequestList         []models.StockRequest `json:"stock_request_list"`
	//StockRequestApprovalList []models.StockRequest `json:"stock_request_approval_list"`
}

type AmendStockRequestRequest struct {
	OutletID       uuid.UUID `json:"outlet_id"`
	StockRequestID uuid.UUID `json:"stock_request_id"`
	Remarks        string    `json:"remarks"`
	AmendedItems   []struct {
		IngredientID uuid.UUID `json:"ingredient_id" `
		UnitSelected string    `json:"unit_selected" valid:"required~Unit selected is required"`
		Quantity     float32   `json:"quantity" valid:"required~Quantity is required"`
	} `json:"amended_items" valid:"required~Amended items is required"`
}

type UpdateTransferInOutRequest struct {
	OutletID       uuid.UUID `json:"outlet_id" `
	StockRequestID uuid.UUID `json:"stock_request_id" `
}

type ReplyAmendStockRequestRequest struct {
	OutletID       uuid.UUID `json:"outlet_id" `
	StockRequestID uuid.UUID `json:"stock_request_id"`
	IsAccepted     bool      `json:"is_accepted"`
}

type ReplyAcceptRejectStockRequestRequest struct {
	OutletID       uuid.UUID `json:"outlet_id" `
	StockRequestID uuid.UUID `json:"stock_request_id" `
	IsAccepted     bool      `json:"is_accepted"`
}

type MarkStockRequestAsReceivedRequest struct {
	OutletID       uuid.UUID `json:"outlet_id" `
	StockRequestID uuid.UUID `json:"stock_request_id" `
}

type GetAllTransferInOfOutletRequest struct {
	OutletID   uuid.UUID `json:"outlet_id" `
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SearchKey  string    `json:"search_key"`
}

type GetAllTransferInOfOutletResponse struct {
	Message string `json:"message"`
	Data    []struct {
		StockRequestID         uuid.UUID       `json:"stock_request_id"`
		RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
		ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
		RequesterOutletName    string          `json:"requester_outlet_name"`
		ResponderOutletName    string          `json:"responder_outlet_name"`
		RequesterOutletAddress string          `json:"requester_outlet_address"`
		ResponderOutletAddress string          `json:"responder_outlet_address"`
		Status                 string          `json:"status"`
		Date                   time.Time       `json:"date"`
		RequestedItems         []RequestedItem `json:"requested_items"`
	} `json:"data"`
	TotalCount int `json:"total_count"` // total number of transfer in without pagination
}

type GetAllTransferOutOfOutletRequest struct {
	OutletID   uuid.UUID `json:"outlet_id" `
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SearchKey  string    `json:"search_key"`
}

type GetAllTransferOutOfOutletResponse struct {
	Message string `json:"message"`
	Data    []struct {
		StockRequestID         uuid.UUID       `json:"stock_request_id"`
		RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
		ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
		RequesterOutletName    string          `json:"requester_outlet_name"`
		ResponderOutletName    string          `json:"responder_outlet_name"`
		RequesterOutletAddress string          `json:"requester_outlet_address"`
		ResponderOutletAddress string          `json:"responder_outlet_address"`
		Status                 string          `json:"status"`
		Date                   time.Time       `json:"date"`
		RequestedItems         []RequestedItem `json:"requested_items"`
	} `json:"data"`
	TotalCount int `json:"total_count"` // total number of transfer out without pagination
}

type RequestedItem struct {
	IngredientID   uuid.UUID `json:"ingredient_id"`
	IngredientName string    `json:"ingredient_name"`
	UnitSelected   string    `json:"unit_selected"`
	Quantity       float32   `json:"quantity"`
}

type AdditionStockOpeningRequest struct {
	OutletID     uuid.UUID `json:"outlet_id"`
	IngredientID uuid.UUID `json:"ingredient_id"`
	Opening      float32   `json:"opening"`
	OpeningUnit  string    `json:"opening_unit"`
}

type GetAllProductWastageDailyByTypeStatRequest struct {
	OutletID   uuid.UUID `json:"outlet_id"`
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
	PageSize   int       `json:"page_size"`
	PageNumber int       `json:"page_number"`
	SearchKey  string    `json:"search_key"`
}
type GetAllProductWastageDailyByTypeStatResponse struct {
	Message string                    `json:"message"`
	Meta    common.Pagination         `json:"meta"`
	Data    []ProductWastageDailyStat `json:"data"`
}
type ProductWastageDailyStat struct {
	Date           time.Time                `json:"date"`
	Header         []string                 `json:"headers"`
	ProductWastage []ProductWastageItemStat `json:"product_wastage_items"`
}
type ProductWastageItemStat struct {
	ProductID    uuid.UUID          `json:"product_id"`
	ProductName  string             `json:"product_name"`
	WastageStats map[string]float32 `json:"wastage_stats"` // Changed from int to float32
	TotalWastage float32            `json:"total_wastage"` // Changed from int to float32
	SortOrder    int                `json:"sort_order"`
}

// API to add stock to outlet
//
//encore:api auth method=POST path=/api/admin/outlet/stock/transfer
func (s *Service) TransferStock(ctx context.Context, req *TransferInStockRequest) (*common.BasicResponse, error) {

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

	// Get ingredient details
	var ingredient models.Ingredient
	if err := trx.Where("id = ?", req.IngredientID).First(&ingredient).Error; err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Ingredient not found",
		}
	}

	ingredientQuantity := float32(req.AmountOfIngredientItem) * ingredient.Quantity

	stock := &models.Stock{
		OutletID:           req.OutletID,
		IngredientID:       req.IngredientID,
		Name:               ingredient.Name,
		SmallScaleUnit:     constants.GetSmallUnitFromLarge(ingredient.Unit),
		LargeScaleUnit:     constants.GetLargeUnitFromSmall(ingredient.Unit),
		SmallScaleQuantity: constants.ConvertToSmallUnit(ingredient.Unit, ingredientQuantity),
		LargeScaleQuantity: constants.ConvertToLargeUnit(ingredient.Unit, ingredientQuantity),
		CreatedAt:          time.Now(),
		UpdatedAt:          nil,
		DeletedAt:          gorm.DeletedAt{},
	}

	// if got existing stock, update the stock
	var existingStock models.Stock
	if err := trx.Where("outlet_id = ? AND ingredient_id = ?", req.OutletID, req.IngredientID).First(&existingStock).Error; err == nil {
		if req.TransferType == "in" {
			existingStock.SmallScaleQuantity += stock.SmallScaleQuantity
			existingStock.LargeScaleQuantity += stock.LargeScaleQuantity
		} else {
			existingStock.SmallScaleQuantity -= stock.SmallScaleQuantity
			existingStock.LargeScaleQuantity -= stock.LargeScaleQuantity
		}
		trx.Save(&existingStock)
	} else {
		// Add the stock
		result := trx.Create(stock)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	// Update stock report if it exists for today
	stockReport, err := common_operations.GetStockReportByIngredientForToday(trx, ctx, req.IngredientID, req.OutletID)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	if req.TransferType == "in" {
		//stockReport.TransferIn += ingredientQuantity
		*stockReport.TransferIn = stockReport.TransferIn.Add(common.ConvertFloat32ToDecimal(ingredientQuantity))
	} else {
		//stockReport.TransferOut += ingredientQuantity
		*stockReport.TransferOut = stockReport.TransferOut.Add(common.ConvertFloat32ToDecimal(ingredientQuantity))
	}
	if err := trx.Save(&stockReport).Error; err != nil {
		trx.Rollback()
		return nil, err
	}

	// Commit the transaction
	err = trx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Stock added successfully",
	}, nil

}

//
//encore:api auth method=GET path=/api/admin/outlet/stock/get-all/:outlet_id
func (s *Service) GetAllStockOutlet(ctx context.Context, outlet_id uuid.UUID) (*GetAllStockOutletResponse, error) {
	var outlet models.Outlet
	result := s.db.Where("id = ?", outlet_id).First(&outlet)
	if result.Error != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Outlet not found",
		}
	}

	// Get all ingredients for this outlet
	var ingredients []models.Ingredient
	if err := s.db.Where("business_id = ?", outlet.BusinessID).Find(&ingredients).Error; err != nil {
		return nil, err
	}

	var stocks []models.Stock
	result = s.db.Preload("Ingredient").Where("outlet_id = ?", outlet_id).Find(&stocks)
	if result.Error != nil {
		return nil, result.Error
	}

	stockMap := make(map[uuid.UUID]models.Stock)
	for _, stock := range stocks {
		stockMap[stock.IngredientID] = stock
	}

	var data []StockWithSelectedUnit

	for _, ingredient := range ingredients {
		stock, ok := stockMap[ingredient.ID]
		if !ok {
			stock = models.Stock{
				OutletID:           outlet_id,
				IngredientID:       ingredient.ID,
				Name:               ingredient.Name,
				SmallScaleUnit:     constants.GetSmallUnitFromLarge(ingredient.Unit),
				LargeScaleUnit:     constants.GetLargeUnitFromSmall(ingredient.Unit),
				SmallScaleQuantity: 0,
				LargeScaleQuantity: 0,
			}

			// Create new stock record for this ingredient
			if err := s.db.Create(&stock).Error; err != nil {
				return nil, &errs.Error{
					Code:    errs.Internal,
					Message: "Failed to create stock record",
				}
			}
		}
		// this is to remove the ingredient from the stock
		stockWithSelectedUnit := StockWithSelectedUnit{
			Stock: models.Stock{
				ID:                 stock.ID,
				OutletID:           outlet_id,
				IngredientID:       ingredient.ID,
				Ingredient:         nil,
				Name:               stock.Name,
				Description:        stock.Description,
				SmallScaleUnit:     constants.GetSmallUnitFromLarge(ingredient.Unit),
				LargeScaleUnit:     constants.GetLargeUnitFromSmall(ingredient.Unit),
				SmallScaleQuantity: constants.ConvertToSmallUnit(ingredient.Unit, stock.SmallScaleQuantity),
				LargeScaleQuantity: constants.ConvertToLargeUnit(ingredient.Unit, stock.LargeScaleQuantity),
				CreatedAt:          stock.CreatedAt,
				UpdatedAt:          stock.UpdatedAt,
				DeletedAt:          stock.DeletedAt,
			},
			SelectedUnit: string(ingredient.Unit),
		}
		data = append(data, stockWithSelectedUnit)
	}

	return &GetAllStockOutletResponse{
		Data: data,
	}, nil
}

// API to Initialize stock report for the day for all stocks of ingredients
//
//encore:api auth method=POST path=/api/admin/outlet/stock/opening-report
func (s *Service) OpenStockReportOfTheDay(ctx context.Context, req *OpenStockReportRequest) error {
	var outlet models.Outlet
	result := s.db.Where("id = ?", req.OutletID).First(&outlet)
	if result.Error != nil {
		return &errs.Error{
			Code:    errs.NotFound,
			Message: "Outlet not found",
		}
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return trx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	for _, stock := range req.Stocks {
		var currentStock models.Stock
		result := s.db.Preload("Ingredient").Where("ingredient_id = ? AND outlet_id = ?",
			stock.IngredientID, req.OutletID).
			First(&currentStock)

		if result.Error != nil {
			trx.Rollback()
			return &errs.Error{
				Code:    errs.NotFound,
				Message: "Stock not found",
			}
		}

		// save to stock
		currentStock.SmallScaleQuantity = constants.ConvertToSmallUnit(currentStock.Ingredient.Unit, stock.Opening)
		currentStock.LargeScaleQuantity = constants.ConvertToLargeUnit(currentStock.Ingredient.Unit, stock.Opening)
		if err := trx.Save(&currentStock).Error; err != nil {
			trx.Rollback()
			return err
		}

		// Check if a stock report already exists for this ingredient today
		existingReport, err := common_operations.GetStockReportByIngredientForToday(trx, ctx, stock.IngredientID, req.OutletID)
		if err != nil {
			trx.Rollback()
			return err
		}

		// convert opening to the same unit as the ingredient unit
		var ingredient models.Ingredient
		resultIngredient := s.db.Where("id = ?", stock.IngredientID).First(&ingredient)
		if resultIngredient.Error != nil {
			trx.Rollback()
			return resultIngredient.Error
		}
		openingConverted := stock.Opening
		normalizedOpeningUnit := strings.TrimPrefix(stock.OpeningUnit, "StockOpeningDialogItemUnit.")
		if normalizedOpeningUnit != string(ingredient.Unit) {
			openingConverted = constants.ConvertToTargetUnit(constants.UnitMeasurement(normalizedOpeningUnit), stock.Opening, ingredient.Unit)
		}

		// if report exists, update the opening
		openingConvertedWithDecimal := common.ConvertFloat32ToDecimal(openingConverted)
		existingReport.Opening = &openingConvertedWithDecimal
		existingReport.OpeningBySystem = &openingConvertedWithDecimal
		cashOpening := common.ConvertFloat32ToDecimal(req.CashOpening)
		existingReport.CashOpening = &cashOpening
		existingReport.CashClosing = nil
		if err := trx.Save(&existingReport).Error; err != nil {
			trx.Rollback()
			return err
		}
	}

	err := trx.Commit().Error
	if err != nil {
		return err
	}

	return nil
}

// CloseStockReportOfTheDay updates the stock reports with closing values for the day
//
//encore:api auth method=POST path=/api/admin/outlet/stock/close-report
func (s *Service) CloseStockReportOfTheDay(ctx context.Context, req *CloseStockReportRequest) (*CloseStockReportResponse, error) {
	d := auth.Data()
	user := d.(*models.User)
	if user == nil {
		return nil, errors.New("user data not found")
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

	// convert product wastage to ingredient
	if req.ProductWastage != nil {
		_, err := s.RecordProductWastage(ctx, &RecordProductWastageRequest{
			BusinessID:     *user.BusinessID,
			OutletID:       req.OutletID,
			ProductWastage: req.ProductWastage,
		})
		if err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	// apply ingredient wastage to the stock report
	if req.IngredientWastage != nil {
		_, err := s.RecordIngredientWastage(ctx, &RecordIngredientWastageRequest{
			OutletID:          req.OutletID,
			IngredientWastage: req.IngredientWastage,
		})
		if err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	expensesResponse := &AddExpensesResponse{
		Data: make([]AddExpensesData, 0),
	}

	expensesConverted := make([]struct {
		ExpensesCategory    models.ExpenseCategory `json:"expenses_category" valid:"required~Expenses category is required"`
		ExpensesDescription string                 `json:"expenses_description" valid:"required~Expenses description is required"`
		ExpensesAmount      float32                `json:"expenses_amount"`
		ExpensesDate        time.Time              `json:"expenses_date" valid:"required~Expenses date is required"`
	}, 0)
	for _, expense := range req.Expenses {
		if expense.ExpensesID != uuid.Nil {
			expensesResponse.Data = append(expensesResponse.Data, AddExpensesData{
				ExpensesOutletID: uuid.Nil,
			})
			continue
		}
		expensesConverted = append(expensesConverted, struct {
			ExpensesCategory    models.ExpenseCategory `json:"expenses_category" valid:"required~Expenses category is required"`
			ExpensesDescription string                 `json:"expenses_description" valid:"required~Expenses description is required"`
			ExpensesAmount      float32                `json:"expenses_amount"`
			ExpensesDate        time.Time              `json:"expenses_date" valid:"required~Expenses date is required"`
		}{
			ExpensesCategory:    expense.ExpensesCategory,
			ExpensesDescription: expense.ExpensesDescription,
			ExpensesAmount:      expense.ExpensesAmount,
			ExpensesDate:        expense.ExpensesDate,
		})
	}
	// add expenses (add new expenses only)
	newExpensesResponse, err := s.AddExpensesInBulk(ctx, &AddExpensesRequest{
		OutletID: req.OutletID,
		Expenses: expensesConverted,
	})
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	// merge the new expenses with the existing expenses
	expensesResponse.Data = append(expensesResponse.Data, newExpensesResponse.Data...)

	// Update Closing Stock Reports
	for _, ingredientStock := range req.IngredientStocksClosing {
		// Check if a stock report exists for this ingredient today
		existingReport, err := common_operations.GetStockReportByIngredientForToday(trx, ctx, ingredientStock.IngredientID, req.OutletID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}

		// check the selected unit is the same as the ingredient unit in database
		var ingredient models.Ingredient
		resultIngredient := s.db.Where("id = ?", ingredientStock.IngredientID).First(&ingredient)
		if resultIngredient.Error != nil {
			trx.Rollback()
			return nil, resultIngredient.Error
		}
		// convert the closing to the same unit as the ingredient unit
		if ingredientStock.UnitSelected != string(ingredient.Unit) {
			ingredientStock.Closing = constants.ConvertToTargetUnit(constants.UnitMeasurement(ingredientStock.UnitSelected), ingredientStock.Closing, ingredient.Unit)
		}

		// Update existing report with closing values
		openingBySystem := common.ConvertDecimalToFloat32(*existingReport.OpeningBySystem)
		closingBySystem := openingBySystem - existingReport.Sales + existingReport.Purchases + common.ConvertDecimalToFloat32(*existingReport.TransferIn) - common.ConvertDecimalToFloat32(*existingReport.TransferOut) - existingReport.Wastage
		cashClosing := common.ConvertFloat32ToDecimal(req.CashClosing)
		existingReport.CashClosing = &cashClosing
		closingWithDecimal := common.ConvertFloat32ToDecimal(ingredientStock.Closing)
		existingReport.Closing = &closingWithDecimal
		closingBySystemWithDecimal := common.ConvertFloat32ToDecimal(closingBySystem)
		existingReport.ClosingBySystem = &closingBySystemWithDecimal
		closingVariance := common.ConvertFloat32ToDecimal(ingredientStock.Closing - closingBySystem)
		existingReport.Variance = &closingVariance
		if err := trx.Save(&existingReport).Error; err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &CloseStockReportResponse{
		Message: "Stock report closed successfully",
		Data:    expensesResponse.Data,
	}, nil
}

func (s *Service) RecordProductWastage(ctx context.Context, req *RecordProductWastageRequest) (*common.BasicResponse, error) {
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

	// Update Stock Wastage
	for _, productWastage := range req.ProductWastage {
		// check if the product wastage amount is valid
		if productWastage.WastageAmount <= 0 {
			continue
		}
		// create product wastage report
		productWastageReport := models.ProductWastageReport{
			BusinessID:    req.BusinessID,
			OutletID:      req.OutletID,
			ProductID:     productWastage.ProductID,
			WastageTypeID: &productWastage.WastageTypeID,
			WastageAmount: productWastage.WastageAmount,
			ReportDate:    productWastage.ReportDate,
			Notes:         productWastage.Notes,
			CreatedAt:     time.Now(),
			UpdatedAt:     nil,
			DeletedAt:     gorm.DeletedAt{},
		}
		err := trx.Create(&productWastageReport).Error
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		// Get current stock for this product
		var productIngredientMappings []models.ProductIngredientMapping
		s.db.Where("product_id = ?", productWastage.ProductID).Preload("Ingredient").Find(&productIngredientMappings)

		for _, productIngredient := range productIngredientMappings {
			wastageIngredient := float32(productWastage.WastageAmount) * float32(productIngredient.Quantity)

			stockReport, err := common_operations.GetStockReportByIngredientForToday(trx, ctx, productIngredient.IngredientID, req.OutletID)
			if err != nil {
				trx.Rollback()
				return nil, err
			}

			stockReport.Wastage = constants.ConvertToTargetUnit(productIngredient.Unit, wastageIngredient, productIngredient.Ingredient.Unit)
			if err := trx.Save(&stockReport).Error; err != nil {
				trx.Rollback()
				return nil, err
			}
		}
	}

	err := trx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Product wastage recorded successfully",
	}, nil
}

func (s *Service) RecordIngredientWastage(ctx context.Context, req *RecordIngredientWastageRequest) (*common.BasicResponse, error) {
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

	for _, ingredientWastage := range req.IngredientWastage {
		stockReport, err := common_operations.GetStockReportByIngredientForToday(trx, ctx, ingredientWastage.IngredientID, req.OutletID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}

		ingredient, err := s.GetIngredientByID(ctx, ingredientWastage.IngredientID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}

		currentWastage := constants.ConvertToTargetUnit(constants.UnitMeasurement(ingredientWastage.UnitSelected), ingredientWastage.NoOfIngredientWastage, ingredient.Unit)

		stockReport.Wastage += currentWastage
		if err := trx.Save(&stockReport).Error; err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	err := trx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Ingredient wastage recorded successfully",
	}, nil
}

func (s *Service) GetIngredientByID(ctx context.Context, ingredient_id uuid.UUID) (*models.Ingredient, error) {
	if ingredient_id == uuid.Nil {
		return nil, &errs.Error{
			Message: "Ingredient ID is required",
		}
	}

	var ingredient models.Ingredient
	result := s.db.Where("id = ?", ingredient_id).First(&ingredient)
	if result.Error != nil {
		return nil, result.Error
	}

	return &ingredient, nil
}

// API to get stock reports for a specific outlet and date
//
//encore:api auth method=POST path=/api/admin/outlet/stock/reports
func (s *Service) GetStockReports(ctx context.Context, req *GetStockReportsRequest) (*GetStockReportsResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var stockReports []models.StockReport
	query := s.db.Where("business_id = ?", req.BusinessID)

	if req.OutletID != nil {
		query = query.Where("outlet_id = ?", req.OutletID)
	}

	if req.OutletGroupID != nil {
		outletIDs, err := common_operations.GetOutletIDsByGroupID(s.db, *req.OutletGroupID)
		if err != nil {
			return nil, err
		}
		query = query.Where("outlet_id IN (?)", outletIDs)
	}

	fromDate := time.Now()
	toDate := time.Now()

	if req.FromDate != nil {
		fromDate, _ = common.GetStartOfDay(*req.FromDate)
	}
	if req.ToDate != nil {
		toDate, _ = common.GetEndOfDay(*req.ToDate)
	}

	query = query.Where("created_at BETWEEN ? AND ?", fromDate.UTC(), toDate.UTC())

	// Preload related data
	query = query.Preload("Ingredient")

	// Execute the query
	if err := query.Find(&stockReports).Error; err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "Failed to retrieve stock reports",
		}
	}

	// if len(stockReports) == 0 {
	// 	return nil, &errs.Error{
	// 		Code:    errs.NotFound,
	// 		Message: "No stock reports found",
	// 	}
	// }

	// Group stock reports by ingredient name and sum the columns
	groupedReports := make(map[string]models.StockReport)
	for _, report := range stockReports {
		if report.Ingredient != nil {
			ingredientName := report.Ingredient.Name
			// Add the quantity to the existing total or initialize if first occurrence
			if _, exists := groupedReports[ingredientName]; !exists {
				groupedReports[ingredientName] = models.StockReport{
					Ingredient:      report.Ingredient,
					Opening:         decimalPtr(decimal.Zero),
					OpeningBySystem: decimalPtr(decimal.Zero),
					Sales:           0,
					Purchases:       0,
					TransferIn:      decimalPtr(decimal.Zero),
					TransferOut:     decimalPtr(decimal.Zero),
					Wastage:         0,
					Closing:         decimalPtr(decimal.Zero),
					ClosingBySystem: decimalPtr(decimal.Zero),
					Variance:        decimalPtr(decimal.Zero),
				}
			}
			currentReport := groupedReports[ingredientName]
			//reportCopy.Opening += report.Opening
			if currentReport.Opening != nil && report.Opening != nil {
				*currentReport.Opening = currentReport.Opening.Add(*report.Opening)
			}
			if currentReport.OpeningBySystem != nil && report.OpeningBySystem != nil {
				*currentReport.OpeningBySystem = currentReport.OpeningBySystem.Add(*report.OpeningBySystem)
			}
			currentReport.Sales += report.Sales
			currentReport.Purchases += report.Purchases
			//reportCopy.TransferIn += report.TransferIn
			//reportCopy.TransferOut += report.TransferOut
			if currentReport.TransferIn != nil && report.TransferIn != nil {
				*currentReport.TransferIn = currentReport.TransferIn.Add(*report.TransferIn)
			}
			if currentReport.TransferOut != nil && report.TransferOut != nil {
				*currentReport.TransferOut = currentReport.TransferOut.Add(*report.TransferOut)
			}
			currentReport.Wastage += report.Wastage
			if currentReport.Closing != nil && report.Closing != nil {
				*currentReport.Closing = currentReport.Closing.Add(*report.Closing)
			}
			// Recalculate: We will recalculate these in the final step
			// if currentReport.ClosingBySystem != nil && report.ClosingBySystem != nil {
			// 	*currentReport.ClosingBySystem = currentReport.ClosingBySystem.Add(*report.ClosingBySystem)
			// }
			// if currentReport.Variance != nil && report.Variance != nil {
			// 	*currentReport.Variance = currentReport.Variance.Add(*report.Variance)
			// }
			groupedReports[ingredientName] = currentReport
		}
	}

	var aggregatedReports []models.StockReport
	for _, report := range groupedReports {
		// Calculate ClosingBySystem & Variance
		purchasesDecimal := common.ConvertFloat32ToDecimal(report.Purchases)
		salesDecimal := common.ConvertFloat32ToDecimal(report.Sales)
		wastageDecimal := common.ConvertFloat32ToDecimal(report.Wastage)
		closingBySystem := report.Opening.Add(purchasesDecimal).Sub(salesDecimal).Add(*report.TransferIn).Sub(*report.TransferOut).Sub(wastageDecimal)

		// Set report for ClosingBySystem and Variance
		report.ClosingBySystem = decimalPtr(closingBySystem)
		report.Variance = decimalPtr(closingBySystem.Sub(*report.Closing))

		// Append report to aggregatedReports
		aggregatedReports = append(aggregatedReports, report)
	}

	ingredients := []models.Ingredient{}
	err := s.db.Model(&models.Ingredient{}).Where("business_id = ?", req.BusinessID).Find(&ingredients).Error
	if err != nil {
		return nil, err
	}

	for _, ingredient := range ingredients {
		found := false
		for _, report := range aggregatedReports {
			if report.Ingredient.ID == ingredient.ID {
				found = true
				break
			}
		}
		if !found {
			// Create a zeroed StockReport for this ingredient
			zero := decimal.Zero
			aggregatedReports = append(aggregatedReports, models.StockReport{
				IngredientID:    ingredient.ID,
				Ingredient:      &ingredient,
				Opening:         &zero,
				OpeningBySystem: &zero,
				Sales:           0,
				Purchases:       0,
				TransferIn:      &zero,
				TransferOut:     &zero,
				Wastage:         0,
				Closing:         &zero,
				ClosingBySystem: &zero,
				Variance:        &zero,
			})
		}
	}

	// Sort aggregatedReports by ingredient sort order in ascending order
	sort.Slice(aggregatedReports, func(i, j int) bool {
		return aggregatedReports[i].Ingredient.SortOrder < aggregatedReports[j].Ingredient.SortOrder
	})

	return &GetStockReportsResponse{
		Data: aggregatedReports,
	}, nil
}
func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

// API to check outlet stock opening is done at that day
//
//encore:api auth method=GET path=/api/admin/outlet/stock/opening-check/:outlet_id
func (s *Service) CheckOutletStockOpening(ctx context.Context, outlet_id uuid.UUID) (*CheckOutletStockOpeningResponse, error) {

	var stockReports []models.StockReport
	result := s.db.Where("outlet_id = ?", outlet_id).Where("DATE(created_at) = CURRENT_DATE").Find(&stockReports)
	if result.Error != nil {
		return nil, result.Error
	}

	if len(stockReports) == 0 {
		return &CheckOutletStockOpeningResponse{
			Message:            "Stock opening is not done at that day",
			IsStockOpeningDone: false,
		}, nil
	}

	return &CheckOutletStockOpeningResponse{
		Message:            "Stock opening is done at that day",
		IsStockOpeningDone: true,
	}, nil
}

// API to query stock report by outlet id and show by days (group by days)
//
//encore:api auth method=POST path=/api/admin/outlet/stock/opening-report-by-days
func (s *Service) GetAllStockReportByDays(ctx context.Context, req *GetAllStockOpeningReportByDaysRequest) (*GetAllStockOpeningReportByDaysResponse, error) {
	startDate, err := common.GetDateInFormatYYYYMMDDHHMMSSStartOfDay(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetDateInFormatYYYYMMDDHHMMSSEndOfDay(req.EndDate)

	fmt.Println("startDate", startDate)
	fmt.Println("endDate", endDate)
	if err != nil {
		return nil, err
	}
	if req.PageSize == 0 {
		req.PageSize = 30
	}
	if req.PageNumber < 0 {
		req.PageNumber = 1
	}

	offset := (req.PageNumber - 1) * req.PageSize

	// First get total count of reports in date range
	var distinctDayCount int64
	err = s.db.Model(&models.StockReport{}).
		Where("outlet_id = ?", req.OutletID).
		Where("created_at BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05")).
		//Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Select("COUNT(DISTINCT DATE(created_at))").
		Count(&distinctDayCount).Error
	if err != nil {
		return nil, err
	}

	//totalPages := (distinctDayCount / int64(page_size)) + 1
	//debug
	fmt.Println("distinctDayCount", distinctDayCount)
	fmt.Println("page_size", req.PageSize)

	// Calculate max pages
	maxPage := common.CalculateMaxPageNumber(distinctDayCount, req.PageSize)
	fmt.Println("maxPage", maxPage)
	if req.PageNumber > maxPage {
		fmt.Println("reached max page")
		return &GetAllStockOpeningReportByDaysResponse{
			Message: "No more stock reports found",
			Data: []struct {
				Date             time.Time            `json:"date"`
				Count            int                  `json:"count"`
				StockOpeningData []models.StockReport `json:"stock_opening_data"`
			}{},
		}, nil
	}

	var dates []time.Time
	result := s.db.Model(&models.StockReport{}).
		Select("DATE(created_at) as date").
		Where("outlet_id = ?", req.OutletID).
		Where("DATE(created_at) BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("DATE(created_at) DESC").
		Group("DATE(created_at)").
		Offset(offset).
		Limit(req.PageSize).
		Find(&dates)

	//debug
	for _, date := range dates {
		fmt.Println("date", date)
	}
	if result.Error != nil {
		return nil, result.Error
	}

	// Then get all reports for those dates
	var stockReports []models.StockReport
	result = s.db.Where("stock_reports.outlet_id = ?", req.OutletID).
		Where("DATE(stock_reports.created_at) IN (?)", dates).
		Scopes(common.GeneralSearch(req.SearchKey, "ingredients", "name")).
		Joins("FULL OUTER JOIN ingredients ON stock_reports.ingredient_id = ingredients.id").
		//Where("ingredient.name LIKE ?", "%"+req.SearchKey+"%").
		Preload("Ingredient").
		Order("stock_reports.created_at DESC").
		Find(&stockReports)

	fmt.Println("result", result)
	if result.RowsAffected == 0 {
		return &GetAllStockOpeningReportByDaysResponse{
			Message: "No more stock reports found",
			Data: []struct {
				Date             time.Time            `json:"date"`
				Count            int                  `json:"count"`
				StockOpeningData []models.StockReport `json:"stock_opening_data"`
			}{},
		}, nil
	}

	//debug
	// for _, report := range stockReports {
	// 	fmt.Println("report", report.Ingredient.Name, report.CreatedAt)
	// }

	if len(stockReports) == 0 {
		return &GetAllStockOpeningReportByDaysResponse{
			Message: "No stock reports found",
			Data: []struct {
				Date             time.Time            `json:"date"`
				Count            int                  `json:"count"`
				StockOpeningData []models.StockReport `json:"stock_opening_data"`
			}{},
		}, nil
	}

	// Group reports by day
	groupedReports := make(map[time.Time][]models.StockReport)
	for _, report := range stockReports {
		// Normalize the date by truncating time components
		date := report.CreatedAt.Truncate(24 * time.Hour)
		groupedReports[date] = append(groupedReports[date], report)
	}

	// Convert the map to the response format
	var responseData []struct {
		Date             time.Time            `json:"date"`
		Count            int                  `json:"count"`
		StockOpeningData []models.StockReport `json:"stock_opening_data"`
	}
	for date, reports := range groupedReports {
		responseData = append(responseData, struct {
			Date             time.Time            `json:"date"`
			Count            int                  `json:"count"`
			StockOpeningData []models.StockReport `json:"stock_opening_data"`
		}{
			Date:             date,
			Count:            len(reports),
			StockOpeningData: reports,
		})
	}

	// sort by date
	sort.Slice(responseData, func(i, j int) bool {
		// descending order (date from latest to oldest)
		return responseData[i].Date.After(responseData[j].Date)
	})

	return &GetAllStockOpeningReportByDaysResponse{
		Message: fmt.Sprintf("Stock reports retrieved successfully (page %d of %d)", req.PageNumber, maxPage),
		Data:    responseData,
	}, nil
}

// API to get cash opening of a day
//
//encore:api auth method=GET path=/api/admin/outlet/cash/opening/:outlet_id/:date
func (s *Service) GetCashOpening(ctx context.Context, outlet_id uuid.UUID, date string) (*GetCashOpeningResponse, error) {

	var stockReports models.StockReport
	result := s.db.Where("outlet_id = ? AND DATE(created_at) = ?", outlet_id, date).First(&stockReports)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetCashOpeningResponse{
		Message: "Cash opening retrieved successfully",
		Data: struct {
			Date        time.Time `json:"date"`
			CashOpening float32   `json:"cash_opening"`
		}{
			Date:        stockReports.CreatedAt,
			CashOpening: common.ConvertDecimalToFloat32(*stockReports.CashOpening),
		},
	}, nil
}

// API to create expenses
//
//encore:api auth method=POST path=/api/admin/outlet/add-expenses-in-bulk
func (s *Service) AddExpensesInBulk(ctx context.Context, req *AddExpensesRequest) (*AddExpensesResponse, error) {
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

	ids := make([]uuid.UUID, 0, len(req.Expenses))
	for _, expense := range req.Expenses {
		// Check if the expense category is valid
		if !models.IsValidExpenseCategory(string(expense.ExpensesCategory)) {
			trx.Rollback()
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "Invalid expense category",
			}
		}

		expense := models.ExpensesOutlet{
			OutletID:              req.OutletID,
			ExpensesDate:          expense.ExpensesDate,
			ExpensesAmount:        common.ConvertFloat32ToDecimal(expense.ExpensesAmount),
			ExpensesCategory:      expense.ExpensesCategory,
			ExpensesDescription:   expense.ExpensesDescription,
			ExpensesAttachmentUrl: "",
			CreatedAt:             time.Now(),
			UpdatedAt:             nil,
			DeletedAt:             gorm.DeletedAt{},
		}

		result := trx.Create(&expense)
		ids = append(ids, expense.ID)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	data := make([]AddExpensesData, len(ids))
	for i, id := range ids {
		data[i] = AddExpensesData{
			ExpensesOutletID: id,
		}
	}

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &AddExpensesResponse{
		Message: "Expenses added successfully",
		Data:    data,
	}, nil
}

// API to modify expenses
//
//encore:api auth method=POST path=/api/admin/outlet/modify/expenses
func (s *Service) ModifyExpensesInBulk(ctx context.Context, req *ModifyExpensesRequest) (*common.BasicResponse, error) {
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

	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	if req.Timezone != nil {
		loc, err = time.LoadLocation(*req.Timezone)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
	}

	for _, expense := range req.Expenses {
		expenses, err := common_operations.GetExpensesByID(trx, expense.ExpensesID, false)
		if err != nil {
			continue
		}
		// Only expenses from today can be modified (today in loc: user timezone or Malaysia)
		nowInTz := time.Now().In(loc)
		expenseInTz := expenses.ExpensesDate.In(loc)
		if expenseInTz.Year() != nowInTz.Year() || expenseInTz.Month() != nowInTz.Month() || expenseInTz.Day() != nowInTz.Day() {
			trx.Rollback()
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: fmt.Sprintf("Cannot modify expense: only expenses from today can be modified (expense date: %s)", expenseInTz.Format("2006-01-02")),
			}
		}
		expenses.ExpensesAmount = common.ConvertFloat32ToDecimal(expense.ExpensesAmount)
		expenses.ExpensesCategory = expense.ExpensesCategory
		expenses.ExpensesDescription = expense.ExpensesDescription
		expenses.ExpensesDate = expense.ExpensesDate
		result := trx.Save(expenses)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Expenses modified successfully",
	}, nil
}

// API to upload file for expenses
//
//encore:api auth raw method=POST path=/api/expenses/upload/file
func (s *Service) UploadExpensesFile(w http.ResponseWriter, req *http.Request) {
	expenses_outlet_id := req.FormValue("expenses_outlet_id")
	if expenses_outlet_id == "" {
		http.Error(w, "Expense ID is required", http.StatusBadRequest)
		return
	}

	temp_uuid, err := googleUUID.Parse(expenses_outlet_id)
	if err != nil {
		http.Error(w, "Invalid Expense ID format", http.StatusBadRequest)
		return
	}

	expense := models.ExpensesOutlet{}
	err = s.db.Model(&models.ExpensesOutlet{}).Where("id = ?", temp_uuid).First(&expense).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	outlet := models.Outlet{}
	err = s.db.Model(&models.Outlet{}).Where("id = ?", expense.OutletID).First(&outlet).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	err = s.db.Model(&models.Business{}).Where("id = ?", outlet.BusinessID).First(&business).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(2 << 20); err != nil { // Limit to 2 MB
		errs.HTTPError(w, err)
		return
	}

	// Get the file
	file, header, err := req.FormFile("file")
	if err != nil {
		errs.HTTPError(w, err)
		return
	}
	defer file.Close()

	var document aws_s3.Document
	document.File = file
	file_extension := filepath.Ext(header.Filename)
	document.DocPath = fmt.Sprintf("business/%s/outlet/%s/expenses/%s_%d%s",
		business.RegistrationNumber,
		outlet.Name,
		expense.ExpensesCategory,
		time.Now().UnixMilli(),
		file_extension)
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	expense.ExpensesAttachmentUrl = document_res.Url

	err = s.db.Save(&expense).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

// API to get all expenses of outlet and group by day
//
//encore:api auth method=POST path=/api/admin/outlet/expenses/get-expenses-by-day
func (s *Service) GetAllExpensesOfOutletByDays(ctx context.Context, req *GetAllExpensesOfOutletByDaysRequest) (*GetAllExpensesOfOutletByDaysResponse, error) {

	startDate, err := common.GetDateInFormatYYYYMMDDHHMMSSStartOfDay(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetDateInFormatYYYYMMDDHHMMSSEndOfDay(req.EndDate)
	if err != nil {
		return nil, err
	}
	if req.PageSize == 0 {
		req.PageSize = 30
	}
	if req.PageNumber < 0 {
		req.PageNumber = 1
	}

	offset := (req.PageNumber - 1) * req.PageSize

	// First, get distinct dates for pagination
	var distinctDates []time.Time
	result := s.db.Model(&models.ExpensesOutlet{}).
		Select("DISTINCT DATE(expenses_date) as date").
		Where("outlet_id = ?", req.OutletID).
		Where("expenses_date BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Scopes(common.GeneralSearch(req.SearchKey, "expenses_outlets", "expenses_category")).
		Order("date DESC").
		Offset(offset).
		Limit(req.PageSize).
		Pluck("date", &distinctDates)

	if result.Error != nil {
		return nil, result.Error
	}

	// If no dates found, return early
	if len(distinctDates) == 0 {
		return &GetAllExpensesOfOutletByDaysResponse{
			Message: "No expenses found",
			Data:    []GetAllExpensesOfOutletByDaysData{},
		}, nil
	}

	// Format dates for the IN clause
	var dateStrings []string
	for _, date := range distinctDates {
		dateStrings = append(dateStrings, date.Format("2006-01-02"))
	}

	// Now get all expenses for these specific dates
	var expenses []models.ExpensesOutlet
	result = s.db.Where("outlet_id = ?", req.OutletID).
		Where("DATE(expenses_date) IN ?", dateStrings).
		Scopes(common.GeneralSearch(req.SearchKey, "expenses_outlets", "expenses_category")).
		Order("expenses_date DESC, id DESC").
		Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}

	// debug
	for _, expense := range expenses {
		fmt.Println("expense", expense.ExpensesDate, "expenses category", expense.ExpensesCategory)
	}

	// group them by day
	grouped := make(map[time.Time][]models.ExpensesOutlet)
	for _, expense := range expenses {
		date := expense.ExpensesDate.Truncate(24 * time.Hour)
		grouped[date] = append(grouped[date], expense)
	}

	// convert to response format
	var responseData []GetAllExpensesOfOutletByDaysData
	for date, expenses := range grouped {
		// calculate total expenses of day
		totalExpenses := float32(0)
		for _, expense := range expenses {
			totalExpenses += common.ConvertDecimalToFloat32(expense.ExpensesAmount)
		}
		responseData = append(responseData, GetAllExpensesOfOutletByDaysData{
			Date:               date,
			TotalExpensesOfDay: totalExpenses,
			Expenses:           expenses,
		})
	}
	// sort by date

	sort.Slice(responseData, func(i, j int) bool {
		return responseData[i].Date.After(responseData[j].Date)
	})

	if len(responseData) == 0 {
		return &GetAllExpensesOfOutletByDaysResponse{
			Message: "No expenses found",
			Data:    []GetAllExpensesOfOutletByDaysData{},
		}, nil
	}

	return &GetAllExpensesOfOutletByDaysResponse{
		Message: "Expenses retrieved successfully",
		Data:    responseData,
	}, nil
}

// API to get all expenses of outlet of a specific date
//
//encore:api auth method=GET path=/api/admin/outlet/expenses/get-expenses-by-date/:outlet_id/:date
func (s *Service) GetAllExpensesOfOutletByDate(ctx context.Context, outlet_id uuid.UUID, date string) (*GetAllExpensesOfOutletByDaysResponse, error) {

	var expenses []models.ExpensesOutlet
	result := s.db.Where("outlet_id = ? AND DATE(expenses_date) = ?", outlet_id, date).Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}

	// convert to response format
	var responseData []GetAllExpensesOfOutletByDaysData
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, err
	}

	// calculate total expenses of day
	totalExpenses := float32(0)
	for _, expense := range expenses {
		totalExpenses += common.ConvertDecimalToFloat32(expense.ExpensesAmount)
	}
	responseData = append(responseData, GetAllExpensesOfOutletByDaysData{
		Date:               parsedDate,
		TotalExpensesOfDay: totalExpenses,
		Expenses:           expenses,
	})

	sort.Slice(responseData, func(i, j int) bool {
		return responseData[i].Date.After(responseData[j].Date)
	})

	return &GetAllExpensesOfOutletByDaysResponse{
		Message: "Expenses retrieved successfully",
		Data:    responseData,
	}, nil

}

// API to create stock request
//
//encore:api auth method=POST path=/api/admin/outlet/stock/request
func (s *Service) CreateStockRequest(ctx context.Context, req *CreateStockRequestRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}
	// Validate outlets exist
	var responderOutlet, requesterOutlet models.Outlet
	if err := s.db.Where("id = ?", req.ResponderOutletID).First(&responderOutlet).Error; err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Responder outlet not found",
		}
	}
	if err := s.db.Where("id = ?", req.RequesterOutletID).First(&requesterOutlet).Error; err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Requester outlet not found",
		}
	}
	// validate the stock reports of the requester and responder outlets for the request day
	requesterStockReport := models.StockReport{}
	responderStockReport := models.StockReport{}
	if err := s.db.Where("outlet_id = ?", req.RequesterOutletID).Where("DATE(created_at) = ?", req.RequestDate.Format("2006-01-02")).First(&requesterStockReport).Error; err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Requester Outlet is not online for the request day",
		}
	}
	if err := s.db.Where("outlet_id = ?", req.ResponderOutletID).Where("DATE(created_at) = ?", req.RequestDate.Format("2006-01-02")).First(&responderStockReport).Error; err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Responder Outlet is not online for the request day",
		}
	}
	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}
	// get user data from auth
	authData := auth.Data()
	userData, ok := authData.(*models.User)
	if !ok {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid auth data type",
		}
	}

	// Improved transaction handling
	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	// Validate all ingredients exist and are available at responder outlet
	for _, item := range req.StockRequestedItems {
		var ingredient models.Ingredient
		if err := trx.Where("id = ?", item.IngredientID).First(&ingredient).Error; err != nil {
			trx.Rollback()
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("Ingredient %s not found", item.IngredientID),
			}
		}

		// Check if ingredient is available at responder outlet
		var stock models.Stock
		if err := trx.Where("ingredient_id = ? AND outlet_id = ?",
			item.IngredientID, req.ResponderOutletID).First(&stock).Error; err != nil {
			// create stock if not exists
			stock = models.Stock{
				OutletID:           req.ResponderOutletID,
				IngredientID:       ingredient.ID,
				Name:               ingredient.Name,
				SmallScaleUnit:     constants.GetSmallUnitFromLarge(ingredient.Unit),
				LargeScaleUnit:     constants.GetLargeUnitFromSmall(ingredient.Unit),
				SmallScaleQuantity: 0,
				LargeScaleQuantity: 0,
			}

			// Create new stock record for this ingredient
			if err := trx.Create(&stock).Error; err != nil {
				trx.Rollback()
				return nil, &errs.Error{
					Code:    errs.Internal,
					Message: "Failed to create stock record for responder outlet",
				}
			}

			trx.Rollback()
			return nil, &errs.Error{
				Code:    errs.FailedPrecondition,
				Message: fmt.Sprintf("Ingredient %s not available at responder outlet", ingredient.Name),
			}
		}
	}

	stockRequest := models.StockRequest{
		RequesterOutletID: req.RequesterOutletID,
		RequestDate:       req.RequestDate,
		Remarks:           req.Remarks,
		RequesterID:       userData.ID,
		RequestStatus:     models.StockRequestStatusPending,
		ResponderOutletID: &req.ResponderOutletID,
		ResponderID:       nil,
		ResponseDate:      nil,
		ResponseStatus:    nil,
		CreatedAt:         time.Now(),
		UpdatedAt:         nil,
		DeletedAt:         gorm.DeletedAt{},
	}

	result := trx.Create(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, item := range req.StockRequestedItems {
		stockRequestedItem := models.StockRequestedItem{
			StockRequestID: stockRequest.ID,
			IngredientID:   item.IngredientID,
			UnitSelected:   constants.UnitMeasurement(item.UnitSelected),
			Quantity:       common.ConvertFloat32ToDecimal(item.Quantity),
			CreatedAt:      time.Now(),
			UpdatedAt:      nil,
			DeletedAt:      gorm.DeletedAt{},
		}

		result := trx.Create(&stockRequestedItem)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	title := "Stock Request from " + requesterOutlet.Name
	body := "Stock Request from " + requesterOutlet.Name + " received. View details in the stock history."
	SendStockRequestNotification(
		ctx,
		title,
		body,
		stockRequest.RequesterOutletID,
		*stockRequest.ResponderOutletID,
		true,
		trx,
	)
	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	return &common.BasicResponse{
		Message: "Stock request created successfully",
	}, nil
}

// API to get all stock request and Approval
//
//encore:api auth method=GET path=/api/admin/outlet/stock/request-history/:outlet_id
func (s *Service) GetAllStockRequestAndApproval(ctx context.Context, outlet_id uuid.UUID) (*GetAllStockRequestAndApprovalResponse, error) {

	// Initialize empty slices to ensure we always return lists (not nil)
	stockRequest := make([]struct {
		Date             time.Time             `json:"date"`
		StockRequestList []models.StockRequest `json:"stock_request_list"`
	}, 0)

	stockRequestApproval := make([]struct {
		Date             time.Time             `json:"date"`
		StockRequestList []models.StockRequest `json:"stock_request_approval_list"`
	}, 0)

	// Query for stock requests
	var stockRequestList []models.StockRequest
	result := s.db.Model(&models.StockRequest{}).
		Preload("RequestedItems").
		Preload("RequestedItems.Ingredient").
		// stock request therefore need responder outlet information
		Preload("ResponderOutlet").
		Where("requester_outlet_id = ?", outlet_id).
		Order("created_at DESC").
		Find(&stockRequestList)
	if result.Error != nil {
		return nil, result.Error
	}

	// Query for stock request approvals
	var stockRequestApprovalList []models.StockRequest
	result = s.db.Model(&models.StockRequest{}).
		Preload("RequestedItems").
		Preload("RequestedItems.Ingredient").
		// stock request approval therefore need requester outlet information
		Preload("RequesterOutlet").
		Where("responder_outlet_id = ?", outlet_id).
		Order("created_at DESC").
		Find(&stockRequestApprovalList)
	if result.Error != nil {
		return nil, result.Error
	}

	// Group stock requests by day
	if len(stockRequestList) > 0 {
		groupedStockRequest := make(map[time.Time][]models.StockRequest)
		for _, stockRequest := range stockRequestList {
			date := stockRequest.CreatedAt.Truncate(24 * time.Hour)
			groupedStockRequest[date] = append(groupedStockRequest[date], stockRequest)
		}

		for date, stockRequests := range groupedStockRequest {
			stockRequest = append(stockRequest, struct {
				Date             time.Time             `json:"date"`
				StockRequestList []models.StockRequest `json:"stock_request_list"`
			}{
				Date:             date,
				StockRequestList: stockRequests,
			})
		}
	}

	// Group stock request approvals by day
	if len(stockRequestApprovalList) > 0 {
		groupedStockRequestApproval := make(map[time.Time][]models.StockRequest)
		for _, stockRequest := range stockRequestApprovalList {
			date := stockRequest.CreatedAt.Truncate(24 * time.Hour)
			groupedStockRequestApproval[date] = append(groupedStockRequestApproval[date], stockRequest)
		}

		for date, stockRequests := range groupedStockRequestApproval {
			stockRequestApproval = append(stockRequestApproval, struct {
				Date             time.Time             `json:"date"`
				StockRequestList []models.StockRequest `json:"stock_request_approval_list"`
			}{
				Date:             date,
				StockRequestList: stockRequests,
			})
		}
	}

	return &GetAllStockRequestAndApprovalResponse{
		Message: "Stock request and approval retrieved successfully",
		Data: GetAllStockRequestAndApprovalData{
			StockRequest:         stockRequest,
			StockRequestApproval: stockRequestApproval,
		},
	}, nil
}

// API to amend stock request from reviewer/approver
//
//encore:api auth method=POST path=/api/admin/outlet/stock/request-amendment
func (s *Service) AmendStockRequest(ctx context.Context, req *AmendStockRequestRequest) (*common.BasicResponse, error) {

	fmt.Println("amendment request", req)

	if req.StockRequestID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Stock request ID is required",
		}
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// get user data from auth
	authData := auth.Data()
	userData, ok := authData.(*models.User)
	if !ok {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid auth data type",
		}
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p)
		}
	}()

	// update the stock request table
	var stockRequest models.StockRequest
	result := trx.Where("id=?", req.StockRequestID).Preload("ResponderOutlet").First(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	// update the stock request responser remarks, response status, response date and responder id
	stockRequest.ResponderRemarks = &req.Remarks
	responseStatus := models.StockRequestResponseStatusAmended
	stockRequest.ResponseStatus = &responseStatus
	now := time.Now()
	stockRequest.ResponseDate = &now
	stockRequest.ResponderID = &userData.ID

	result = trx.Save(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// loop through the amended items and update the stock requested item
	for _, amendedItem := range req.AmendedItems {
		var stockRequestedItem models.StockRequestedItem
		result := trx.Where("stock_request_id = ? AND ingredient_id = ?", req.StockRequestID, amendedItem.IngredientID).First(&stockRequestedItem)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		stockRequestedItem.UnitSelected = constants.UnitMeasurement(amendedItem.UnitSelected)
		stockRequestedItem.Quantity = common.ConvertFloat32ToDecimal(amendedItem.Quantity)

		result = trx.Save(&stockRequestedItem)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	title := "Stock Amendment"
	body := "A stock amendment request from " + stockRequest.ResponderOutlet.Name + " has been received. Please review the details in the stock history section."
	SendStockRequestNotification(
		ctx,
		title,
		body,
		stockRequest.RequesterOutletID,
		*stockRequest.ResponderOutletID,
		false,
		trx,
	)

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Stock request amended successfully",
	}, nil

}

// function to update transfer in out
func (s *Service) UpdateTransferInOut(ctx context.Context, req *UpdateTransferInOutRequest) (*common.BasicResponse, error) {
	if req.StockRequestID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Stock request ID is required",
		}
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p)
		}
	}()

	// update the stock request table
	var stockRequest models.StockRequest
	result := trx.Where("id=?", req.StockRequestID).Preload("RequestedItems").First(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, item := range stockRequest.RequestedItems {
		// get existing stock report
		var stockReportRequester models.StockReport
		result = trx.Where("ingredient_id = ? AND outlet_id = ?", item.IngredientID, stockRequest.RequesterOutletID).
			Where("DATE(created_at) = ?", stockRequest.CreatedAt.Format("2006-01-02")).
			First(&stockReportRequester)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		// update the stock report (requester add trasnfer in , responder add transfer out)
		//stockReportRequester.TransferIn = stockReportRequester.TransferIn + common.ConvertDecimalToFloat32(item.Quantity)
		if stockReportRequester.TransferIn != nil {
			*stockReportRequester.TransferIn = stockReportRequester.TransferIn.Add(item.Quantity)
		}

		result = trx.Save(&stockReportRequester)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
		// update the stock report (requester add trasnfer in , responder add transfer out)
		var stockReportResponder models.StockReport
		result = trx.Where("ingredient_id = ? AND outlet_id = ?", item.IngredientID, stockRequest.ResponderOutletID).
			Where("DATE(created_at) = ?", stockRequest.CreatedAt.Format("2006-01-02")).
			First(&stockReportResponder)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		if stockReportResponder.TransferOut != nil {
			*stockReportResponder.TransferOut = stockReportResponder.TransferOut.Add(item.Quantity)
		}

		result = trx.Save(&stockReportResponder)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		// update the stock table (requester outlet) (requester outlet stock increase)
		// if got existing stock, update the stock
		var existingStockOfRequester models.Stock
		result = trx.Where("outlet_id = ? AND ingredient_id = ?", stockRequest.RequesterOutletID, item.IngredientID).First(&existingStockOfRequester)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
		// convert the quantity to correct based on the small scale and large scale
		smallScaleQuantity := constants.ConvertToTargetUnit(item.UnitSelected, common.ConvertDecimalToFloat32(item.Quantity), existingStockOfRequester.SmallScaleUnit)
		largeScaleQuantity := constants.ConvertToTargetUnit(item.UnitSelected, common.ConvertDecimalToFloat32(item.Quantity), existingStockOfRequester.LargeScaleUnit)
		// add the quantity to the existing stock
		existingStockOfRequester.SmallScaleQuantity += smallScaleQuantity
		existingStockOfRequester.LargeScaleQuantity += largeScaleQuantity
		result = trx.Save(&existingStockOfRequester)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		// update the stock table (responder outlet) (responder outlet stock decrease)
		var existingStockOfResponder models.Stock
		result = trx.Where("outlet_id = ? AND ingredient_id = ?", stockRequest.ResponderOutletID, item.IngredientID).First(&existingStockOfResponder)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
		// convert the quantity to correct based on the small scale and large scale
		smallScaleQuantity = constants.ConvertToTargetUnit(item.UnitSelected, common.ConvertDecimalToFloat32(item.Quantity), existingStockOfResponder.SmallScaleUnit)
		largeScaleQuantity = constants.ConvertToTargetUnit(item.UnitSelected, common.ConvertDecimalToFloat32(item.Quantity), existingStockOfResponder.LargeScaleUnit)
		// subtract the quantity from the existing stock
		existingStockOfResponder.SmallScaleQuantity -= smallScaleQuantity
		existingStockOfResponder.LargeScaleQuantity -= largeScaleQuantity
		result = trx.Save(&existingStockOfResponder)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Transfer in out updated successfully",
	}, nil

}

// API to reply the amendment of stock request
//
//encore:api auth method=POST path=/api/admin/outlet/stock/request-amendment-reply
func (s *Service) ReplyAmendStockRequest(ctx context.Context, req *ReplyAmendStockRequestRequest) (*common.BasicResponse, error) {
	if req.StockRequestID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Stock request ID is required",
		}
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p)
		}
	}()

	// update the stock request table
	var stockRequest models.StockRequest
	result := trx.Where("id=?", req.StockRequestID).Preload("ResponderOutlet").First(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	if req.IsAccepted {
		stockRequest.RequestStatus = models.StockRequestStatusApproved
		responseStatus := models.StockRequestResponseStatusApproved
		stockRequest.ResponseStatus = &responseStatus
	} else {
		stockRequest.RequestStatus = models.StockRequestStatusCancelled
		responseStatus := models.StockRequestResponseStatusCancelled
		stockRequest.ResponseStatus = &responseStatus
	}

	result = trx.Save(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	title := "Stock Amendment Reply"
	body := "A stock amendment reply from " + stockRequest.ResponderOutlet.Name + " has been received. Please review the details in the stock history section."
	SendStockRequestNotification(
		ctx,
		title,
		body,
		stockRequest.RequesterOutletID,
		*stockRequest.ResponderOutletID,
		true,
		trx,
	)

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// update the stock report
	if req.IsAccepted {
		s.UpdateTransferInOut(ctx, &UpdateTransferInOutRequest{
			OutletID:       req.OutletID,
			StockRequestID: req.StockRequestID,
		})
	}

	return &common.BasicResponse{
		Message: "Stock request amended successfully",
	}, nil
}

// API to reply accept/reject stock request
//
//encore:api auth method=POST path=/api/admin/outlet/stock/request-accept-reject
func (s *Service) ReplyAcceptRejectStockRequest(ctx context.Context, req *ReplyAcceptRejectStockRequestRequest) (*common.BasicResponse, error) {
	if req.StockRequestID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Stock request ID is required",
		}
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p)
		}
	}()

	userData := auth.Data()
	user, ok := userData.(*models.User)
	if !ok {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid auth data type",
		}
	}

	// update the stock request table
	var stockRequest models.StockRequest
	result := trx.Where("id=?", req.StockRequestID).First(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	stockRequest.ResponderID = &user.ID
	now := time.Now()
	stockRequest.ResponseDate = &now

	var title string
	var body string

	if req.IsAccepted {
		stockRequest.RequestStatus = models.StockRequestStatusApproved
		responseStatus := models.StockRequestResponseStatusApproved
		stockRequest.ResponseStatus = &responseStatus
		title = "Stock Request"
		body = "Your Stock Request was accepted. View details in the stock history."
	} else {
		stockRequest.RequestStatus = models.StockRequestStatusRejected
		responseStatus := models.StockRequestResponseStatusRejected
		stockRequest.ResponseStatus = &responseStatus
		title = "Stock Request"
		body = "Your Stock Request was rejected. View details in the stock history."
	}

	result = trx.Save(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	SendStockRequestNotification(
		ctx,
		title,
		body,
		stockRequest.RequesterOutletID,
		*stockRequest.ResponderOutletID,
		false,
		trx,
	)

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Stock request amended successfully",
	}, nil
}

// API to mark stock request as received
//
//encore:api auth method=POST path=/api/admin/outlet/stock/request-received
func (s *Service) MarkStockRequestAsReceived(ctx context.Context, req *MarkStockRequestAsReceivedRequest) (*common.BasicResponse, error) {
	if req.StockRequestID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Stock request ID is required",
		}
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	defer func() {
		if p := recover(); p != nil {
			trx.Rollback()
			panic(p)
		}
	}()

	// update the stock request table
	var stockRequest models.StockRequest
	result := trx.Where("id=?", req.StockRequestID).Preload("RequesterOutlet").First(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	stockRequest.IsReceived = true
	stockRequest.RequestStatus = models.StockRequestStatusCompleted

	result = trx.Save(&stockRequest)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// update the stock report
	s.UpdateTransferInOut(ctx, &UpdateTransferInOutRequest{
		OutletID:       req.OutletID,
		StockRequestID: req.StockRequestID,
	})

	title := "Stock Transfer"
	body := "A stock transfer request from " + stockRequest.RequesterOutlet.Name + " has been received. Please review the details in the stock history section."
	SendStockRequestNotification(
		ctx,
		title,
		body,
		stockRequest.RequesterOutletID,
		*stockRequest.ResponderOutletID,
		true,
		trx,
	)

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Stock request marked as received successfully",
	}, nil
}

// API to get all Transfer In of a outlet (transfer in can equivalent to requester)
// RequestFrom is the outlet that transfer the stock into the outlet of (requester outlet)
//
//encore:api auth method=POST path=/api/admin/outlets/transfers-in
func (s *Service) GetAllTransferInOfOutlet(ctx context.Context, req *GetAllTransferInOfOutletRequest) (*GetAllTransferInOfOutletResponse, error) {
	startDate, err := common.GetDateInFormatYYYYMMDDHHMMSSStartOfDay(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetDateInFormatYYYYMMDDHHMMSSEndOfDay(req.EndDate)
	if err != nil {
		return nil, err
	}
	if req.PageSize == 0 {
		req.PageSize = 30
	}
	if req.PageNumber < 0 {
		req.PageNumber = 1
	}

	offset := (req.PageNumber - 1) * req.PageSize

	// calculate day count
	var dayCount int64
	result := s.db.Model(&models.StockRequest{}).
		Select("DATE(request_date) as date").
		Where("requester_outlet_id = ?", req.OutletID).
		Where("request_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05")).
		Count(&dayCount)
	if result.Error != nil {
		return nil, result.Error
	}

	// calculate max page number
	maxPage := common.CalculateMaxPageNumber(dayCount, req.PageSize)
	if req.PageNumber > maxPage {
		return &GetAllTransferInOfOutletResponse{
			Message: "No more transfer in found",
			Data: []struct {
				StockRequestID         uuid.UUID       `json:"stock_request_id"`
				RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
				ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
				RequesterOutletName    string          `json:"requester_outlet_name"`
				ResponderOutletName    string          `json:"responder_outlet_name"`
				RequesterOutletAddress string          `json:"requester_outlet_address"`
				ResponderOutletAddress string          `json:"responder_outlet_address"`
				Status                 string          `json:"status"`
				Date                   time.Time       `json:"date"`
				RequestedItems         []RequestedItem `json:"requested_items"`
			}{},
			TotalCount: 0,
		}, nil
	}

	var stockRequests []models.StockRequest
	result = s.db.Where("requester_outlet_id = ?", req.OutletID).
		Preload("RequestedItems").
		Preload("RequestedItems.Ingredient").
		Preload("RequesterOutlet").
		Preload("ResponderOutlet").
		Where("request_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05")).
		Scopes(common.GeneralSearch(req.SearchKey, "stock_requests", "stock_request_id")).
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Find(&stockRequests)
	if result.Error != nil {
		return nil, result.Error
	}

	data := make([]struct {
		StockRequestID         uuid.UUID       `json:"stock_request_id"`
		RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
		ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
		RequesterOutletName    string          `json:"requester_outlet_name"`
		ResponderOutletName    string          `json:"responder_outlet_name"`
		RequesterOutletAddress string          `json:"requester_outlet_address"`
		ResponderOutletAddress string          `json:"responder_outlet_address"`
		Status                 string          `json:"status"`
		Date                   time.Time       `json:"date"`
		RequestedItems         []RequestedItem `json:"requested_items"`
	}, 0, len(stockRequests))

	for _, stockRequest := range stockRequests {
		requestedItems := make([]RequestedItem, 0, len(stockRequest.RequestedItems))

		for _, requestedItem := range stockRequest.RequestedItems {
			requestedItems = append(requestedItems, RequestedItem{
				IngredientID:   requestedItem.IngredientID,
				IngredientName: requestedItem.Ingredient.Name,
				UnitSelected:   string(requestedItem.UnitSelected),
				Quantity:       common.ConvertDecimalToFloat32(requestedItem.Quantity),
			})
		}

		// combine the address of the requester and responder
		requesterAddress := stockRequest.RequesterOutlet.Address
		responderAddress := stockRequest.ResponderOutlet.Address
		combinedRequesterAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s", requesterAddress.StreetLine1, requesterAddress.StreetLine2, requesterAddress.StreetLine3, requesterAddress.PostalCode, requesterAddress.City, requesterAddress.Country)
		combinedResponderAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s", responderAddress.StreetLine1, responderAddress.StreetLine2, responderAddress.StreetLine3, responderAddress.PostalCode, responderAddress.City, responderAddress.Country)

		data = append(data, struct {
			StockRequestID         uuid.UUID       `json:"stock_request_id"`
			RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
			ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
			RequesterOutletName    string          `json:"requester_outlet_name"`
			ResponderOutletName    string          `json:"responder_outlet_name"`
			RequesterOutletAddress string          `json:"requester_outlet_address"`
			ResponderOutletAddress string          `json:"responder_outlet_address"`
			Status                 string          `json:"status"`
			Date                   time.Time       `json:"date"`
			RequestedItems         []RequestedItem `json:"requested_items"`
		}{
			StockRequestID:         stockRequest.ID,
			RequesterOutletID:      stockRequest.RequesterOutlet.ID,
			ResponderOutletID:      stockRequest.ResponderOutlet.ID,
			RequesterOutletName:    stockRequest.RequesterOutlet.Name,
			ResponderOutletName:    stockRequest.ResponderOutlet.Name,
			RequesterOutletAddress: combinedRequesterAddress,
			ResponderOutletAddress: combinedResponderAddress,
			Status:                 string(stockRequest.RequestStatus),
			Date:                   stockRequest.RequestDate,
			RequestedItems:         requestedItems,
		})
	}

	// Build count query with same filters as main query (without pagination)
	var totalCount int64
	countQuery := s.db.Model(&models.StockRequest{}).
		Where("requester_outlet_id = ?", req.OutletID).
		Where("request_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))
	// Apply search filter if provided
	if req.SearchKey != "" {
		countQuery = countQuery.Scopes(common.GeneralSearch(req.SearchKey, "stock_requests", "stock_request_id"))
	}
	result = countQuery.Count(&totalCount)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAllTransferInOfOutletResponse{
		Message:    "Transfer in retrieved successfully",
		Data:       data,
		TotalCount: int(totalCount),
	}, nil

}

// API to get all Transfer Out of a outlet (transfer out can equivalent to responder/reviewer/approver)
// RequestTo is the outlet that is requesting the transfer out
//
//encore:api auth method=POST path=/api/admin/outlets/transfers-out
func (s *Service) GetAllTransferOutOfOutlet(ctx context.Context, req *GetAllTransferOutOfOutletRequest) (*GetAllTransferOutOfOutletResponse, error) {
	startDate, err := common.GetDateInFormatYYYYMMDDHHMMSSStartOfDay(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetDateInFormatYYYYMMDDHHMMSSEndOfDay(req.EndDate)
	if err != nil {
		return nil, err
	}
	if req.PageSize == 0 {
		req.PageSize = 30
	}
	if req.PageNumber < 0 {
		req.PageNumber = 1
	}

	offset := (req.PageNumber - 1) * req.PageSize
	// calculate day count
	var dayCount int64
	result := s.db.Model(&models.StockRequest{}).
		Select("DATE(response_date) as date").
		Where("responder_outlet_id = ?", req.OutletID).
		Where("response_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05")).
		Count(&dayCount)
	if result.Error != nil {
		return nil, result.Error
	}
	// calculate max page number
	maxPage := common.CalculateMaxPageNumber(dayCount, req.PageSize)
	if req.PageNumber > maxPage {
		return &GetAllTransferOutOfOutletResponse{
			Message: "No more transfer out found",
			Data: []struct {
				StockRequestID         uuid.UUID       `json:"stock_request_id"`
				RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
				ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
				RequesterOutletName    string          `json:"requester_outlet_name"`
				ResponderOutletName    string          `json:"responder_outlet_name"`
				RequesterOutletAddress string          `json:"requester_outlet_address"`
				ResponderOutletAddress string          `json:"responder_outlet_address"`
				Status                 string          `json:"status"`
				Date                   time.Time       `json:"date"`
				RequestedItems         []RequestedItem `json:"requested_items"`
			}{},
		}, nil
	}

	var stockRequests []models.StockRequest
	result = s.db.Where("responder_outlet_id = ?", req.OutletID).
		Preload("RequestedItems").
		Preload("RequestedItems.Ingredient").
		Preload("RequesterOutlet").
		Preload("ResponderOutlet").
		Where("response_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05")).
		Scopes(common.GeneralSearch(req.SearchKey, "stock_requests", "stock_request_id")).
		Offset(offset).
		Limit(req.PageSize).
		Order("created_at DESC").
		Find(&stockRequests)
	if result.Error != nil {
		return nil, result.Error
	}

	data := make([]struct {
		StockRequestID         uuid.UUID       `json:"stock_request_id"`
		RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
		ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
		RequesterOutletName    string          `json:"requester_outlet_name"`
		ResponderOutletName    string          `json:"responder_outlet_name"`
		RequesterOutletAddress string          `json:"requester_outlet_address"`
		ResponderOutletAddress string          `json:"responder_outlet_address"`
		Status                 string          `json:"status"`
		Date                   time.Time       `json:"date"`
		RequestedItems         []RequestedItem `json:"requested_items"`
	}, 0, len(stockRequests))

	for _, stockRequest := range stockRequests {
		requestedItems := make([]RequestedItem, 0, len(stockRequest.RequestedItems))

		for _, requestedItem := range stockRequest.RequestedItems {
			requestedItems = append(requestedItems, RequestedItem{
				IngredientID:   requestedItem.IngredientID,
				IngredientName: requestedItem.Ingredient.Name,
				UnitSelected:   string(requestedItem.UnitSelected),
				Quantity:       common.ConvertDecimalToFloat32(requestedItem.Quantity),
			})
		}

		// combine the address of the requester and responder
		requesterAddress := stockRequest.RequesterOutlet.Address
		responderAddress := stockRequest.ResponderOutlet.Address
		combinedRequesterAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s", requesterAddress.StreetLine1, requesterAddress.StreetLine2, requesterAddress.StreetLine3, requesterAddress.PostalCode, requesterAddress.City, requesterAddress.Country)
		combinedResponderAddress := fmt.Sprintf("%s, %s, %s, %s, %s, %s", responderAddress.StreetLine1, responderAddress.StreetLine2, responderAddress.StreetLine3, responderAddress.PostalCode, responderAddress.City, responderAddress.Country)

		data = append(data, struct {
			StockRequestID         uuid.UUID       `json:"stock_request_id"`
			RequesterOutletID      uuid.UUID       `json:"requester_outlet_id"`
			ResponderOutletID      uuid.UUID       `json:"responder_outlet_id"`
			RequesterOutletName    string          `json:"requester_outlet_name"`
			ResponderOutletName    string          `json:"responder_outlet_name"`
			RequesterOutletAddress string          `json:"requester_outlet_address"`
			ResponderOutletAddress string          `json:"responder_outlet_address"`
			Status                 string          `json:"status"`
			Date                   time.Time       `json:"date"`
			RequestedItems         []RequestedItem `json:"requested_items"`
		}{
			StockRequestID:         stockRequest.ID,
			RequesterOutletID:      stockRequest.RequesterOutlet.ID,
			ResponderOutletID:      stockRequest.ResponderOutlet.ID,
			RequesterOutletName:    stockRequest.RequesterOutlet.Name,
			ResponderOutletName:    stockRequest.ResponderOutlet.Name,
			RequesterOutletAddress: combinedRequesterAddress,
			ResponderOutletAddress: combinedResponderAddress,
			Status:                 string(stockRequest.RequestStatus),
			Date:                   stockRequest.RequestDate,
			RequestedItems:         requestedItems,
		})
	}

	// Build count query with same filters as main query (without pagination)
	var totalCount int64
	countQuery := s.db.Model(&models.StockRequest{}).
		Where("responder_outlet_id = ?", req.OutletID).
		Where("response_date BETWEEN ? AND ?", startDate.Format("2006-01-02 15:04:05"), endDate.Format("2006-01-02 15:04:05"))
	// Apply search filter if provided
	if req.SearchKey != "" {
		countQuery = countQuery.Scopes(common.GeneralSearch(req.SearchKey, "stock_requests", "stock_request_id"))
	}
	result = countQuery.Count(&totalCount)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAllTransferOutOfOutletResponse{
		Message:    "Transfer out retrieved successfully",
		Data:       data,
		TotalCount: int(totalCount),
	}, nil

}

// API to add addition stock opening
//
//encore:api auth method=POST path=/api/admin/outlets/addition-stock-opening
func (s *Service) AdditionStockOpening(ctx context.Context, req *AdditionStockOpeningRequest) (*common.BasicResponse, error) {
	// check if the outlet is exist
	outlet, err := outlets.GetOutlet(ctx, req.OutletID)
	if err != nil {
		return nil, err
	}
	if outlet == nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "Outlet not found",
		}
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var stockReport models.StockReport
	result := trx.Model(&models.StockReport{}).
		Where("outlet_id = ?", req.OutletID).
		Where("ingredient_id = ?", req.IngredientID).
		Where("Date(created_at) = ?", time.Now().Format("2006-01-02")).
		First(&stockReport)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "Stock report not found",
		}
	}
	// modify the stock report and ensure only addition
	stockReport.Opening = decimalPtr(stockReport.Opening.Add(common.ConvertFloat32ToDecimal(req.Opening)))
	stockReport.OpeningBySystem = decimalPtr(stockReport.OpeningBySystem.Add(common.ConvertFloat32ToDecimal(req.Opening)))

	result = trx.Save(stockReport)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "Failed to modify stock report",
		}
	}

	if result.RowsAffected > 1 {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: "There is something wrong with the stock report",
		}
	}

	trx.Commit()
	return &common.BasicResponse{
		Message: "Addition stock opening added successfully",
	}, nil
}

// Don't rollback the transaction here, it will be handled by the caller if needed
func SendStockRequestNotification(
	ctx context.Context,
	title string,
	body string,
	requesterOutletID uuid.UUID,
	responderOutletID uuid.UUID,
	isResponderGetNotification bool,
	trx *gorm.DB) (*common.BasicResponse, error) {
	var users []models.User
	var outletIDToSendNotification uuid.UUID
	if isResponderGetNotification {
		outletIDToSendNotification = responderOutletID
	} else {
		outletIDToSendNotification = requesterOutletID
	}
	result := trx.Model(&models.User{}).Where("outlet_id = ?", outletIDToSendNotification).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	// insert notification into database (for all user which under a same outlet)
	var notificationIDs []uuid.UUID
	var deviceTokens []string

	for _, user := range users {
		notification, err := common_operations.InsertNotification(trx, ctx, &models.Notification{
			OutletID:         &outletIDToSendNotification,
			UserID:           &user.ID,
			FCMDeviceToken:   user.FCMDeviceToken,
			Title:            &title,
			Body:             &body,
			NotificationType: models.StockRequestNotification,
			IsRead:           false,
			ActionURL:        nil,
			ImageURL:         nil,
			ExpiredAt:        nil,
		})
		notificationIDs = append(notificationIDs, notification.ID)
		if user.FCMDeviceToken != nil {
			deviceTokens = append(deviceTokens, *user.FCMDeviceToken)
		} else {
			deviceTokens = append(deviceTokens, "")
		}
		if err != nil {
			return nil, err
		}
	}
	var actionURL string
	if isResponderGetNotification {
		actionURL = "/streetfood/request_history?initial_tab=1"
	} else {
		actionURL = "/streetfood/request_history?initial_tab=0"
	}
	firebase.SendNotificationToMultipleDevices(
		ctx,
		deviceTokens,
		title,
		body,
		notificationIDs,
		nil,
		&actionURL,
		models.StockRequestNotification,
		firebase.FirebaseAppTypePOS,
		nil,
	)
	return &common.BasicResponse{
		Message: "Stock request notification sent successfully",
	}, nil
}

// API to get stat of product wastage daily by type
//
//encore:api auth method=POST path=/api/admin/outlets/daily/type/product-wastage/stats
func (s *Service) GetAllProductWastageDailyByTypeStat(ctx context.Context, req *GetAllProductWastageDailyByTypeStatRequest) (*GetAllProductWastageDailyByTypeStatResponse, error) {
	if err := middleware.CheckPermission(constants.ReadStockAction, nil, nil); err != nil {
		return nil, err
	}

	startDate, err := common.GetDateInFormatYYYYMMDDHHMMSSStartOfDay(req.StartDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetDateInFormatYYYYMMDDHHMMSSEndOfDay(req.EndDate)
	if err != nil {
		return nil, err
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.PageNumber < 1 {
		req.PageNumber = 1
	}

	offset := (req.PageNumber - 1) * req.PageSize

	// Get total count of distinct days for pagination metadata
	var totalDistinctDays int64
	err = s.db.Model(&models.ProductWastageReport{}).
		Where("outlet_id = ?", req.OutletID).
		Where("DATE(product_wastage_reports.created_at) BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Joins("JOIN products ON product_wastage_reports.product_id = products.id").
		Joins("JOIN product_wastage_types ON product_wastage_reports.wastage_type_id = product_wastage_types.id").
		Scopes(common.GeneralSearch(req.SearchKey, "products", "name")).
		Select("COUNT(DISTINCT DATE(product_wastage_reports.created_at))").
		Count(&totalDistinctDays).Error
	if err != nil {
		return nil, err
	}

	// get outlet
	var outlet models.Outlet
	result := s.db.Model(&models.Outlet{}).
		Where("id = ?", req.OutletID).
		First(&outlet)
	if result.Error != nil {
		return nil, result.Error
	}

	// get full list of dynamic product wastage type
	var productWastageTypes []models.ProductWastageType
	result = s.db.Model(&models.ProductWastageType{}).
		Where("business_id = ?", outlet.BusinessID).
		Find(&productWastageTypes)
	if result.Error != nil {
		return nil, result.Error
	}

	var rawResults []struct {
		Date            time.Time `json:"date"`
		ProductID       uuid.UUID `json:"product_id"`
		ProductName     string    `json:"product_name"`
		WastageTypeID   uuid.UUID `json:"wastage_type_id"`
		WastageTypeName string    `json:"wastage_type_name"`
		TotalWastage    float32   `json:"total_wastage"` // Changed from int to float32
	}

	// First, get the distinct dates for pagination
	var distinctDates []time.Time
	dateQuery := s.db.Model(&models.ProductWastageReport{}).
		Where("outlet_id = ?", req.OutletID).
		Where("DATE(product_wastage_reports.created_at) BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Joins("JOIN products ON product_wastage_reports.product_id = products.id").
		Scopes(common.GeneralSearch(req.SearchKey, "products", "name")).
		Select("DISTINCT DATE(product_wastage_reports.created_at) as date").
		Order("DATE(product_wastage_reports.created_at) ASC").
		Offset(offset).
		Limit(req.PageSize).
		Pluck("date", &distinctDates)
	if dateQuery.Error != nil {
		return nil, dateQuery.Error
	}

	if len(distinctDates) == 0 {
		return &GetAllProductWastageDailyByTypeStatResponse{
			Message: "No product wastage daily found",
			Data:    []ProductWastageDailyStat{},
		}, nil
	}

	// Convert dates to string format for the IN clause
	var dateStrings []string
	for _, date := range distinctDates {
		dateStrings = append(dateStrings, date.Format("2006-01-02"))
	}

	result = s.db.Model(&models.ProductWastageReport{}).
		Where("outlet_id = ?", req.OutletID).
		Where("DATE(product_wastage_reports.created_at) IN ?", dateStrings).
		Joins("JOIN products ON product_wastage_reports.product_id = products.id").
		Joins("JOIN product_wastage_types ON product_wastage_reports.wastage_type_id = product_wastage_types.id").
		Scopes(common.GeneralSearch(req.SearchKey, "products", "name")).
		Select("DATE(product_wastage_reports.created_at) as date, product_wastage_reports.product_id as product_id, products.name as product_name, product_wastage_reports.wastage_type_id as wastage_type_id, product_wastage_types.name as wastage_type_name, SUM(product_wastage_reports.wastage_amount) as total_wastage").
		Group("DATE(product_wastage_reports.created_at), product_wastage_reports.product_id, products.name, product_wastage_reports.wastage_type_id, product_wastage_types.name").
		Find(&rawResults)
	if result.Error != nil {
		return nil, result.Error
	}

	productWastageDailyStat := make([]ProductWastageDailyStat, 0)
	dateProductMap := make(map[time.Time]map[uuid.UUID]*ProductWastageItemStat)
	for _, raw := range rawResults {
		date := raw.Date
		if dateProductMap[date] == nil {
			dateProductMap[date] = make(map[uuid.UUID]*ProductWastageItemStat)
		}
		productID := raw.ProductID
		if dateProductMap[date][productID] == nil {
			wastageStatsMap := make(map[string]float32) // Changed from int to float32
			for _, productWastageType := range productWastageTypes {
				wastageStatsMap[productWastageType.Name] = 0
			}
			dateProductMap[date][productID] = &ProductWastageItemStat{
				ProductID:    productID,
				ProductName:  raw.ProductName,
				WastageStats: wastageStatsMap,
				TotalWastage: 0,
			}
		}
		item := dateProductMap[date][productID]
		item.WastageStats[raw.WastageTypeName] += raw.TotalWastage
		item.TotalWastage += raw.TotalWastage
	}
	// get and create header list
	header := make([]string, 0, len(productWastageTypes))
	for _, productWastageType := range productWastageTypes {
		header = append(header, productWastageType.Name)
	}
	// Convert map to slice
	for date, productMap := range dateProductMap {
		products := make([]ProductWastageItemStat, 0, len(productMap))
		for _, product := range productMap {
			products = append(products, *product)
		}
		productWastageDailyStat = append(productWastageDailyStat, ProductWastageDailyStat{
			Date:           date,
			Header:         header,
			ProductWastage: products,
		})
	}

	// Sort by date
	sort.Slice(productWastageDailyStat, func(i, j int) bool {
		return productWastageDailyStat[i].Date.After(productWastageDailyStat[j].Date)
	})
	return &GetAllProductWastageDailyByTypeStatResponse{
		Message: "Product wastage daily stat found",
		Meta: common.Pagination{
			Page:       req.PageNumber,
			PageSize:   req.PageSize,
			Total:      totalDistinctDays,
			TotalPages: common.CalculateMaxPageNumber(totalDistinctDays, req.PageSize),
		},
		Data: productWastageDailyStat,
	}, nil
}

type EditStockReportRequest struct {
	OutletID     uuid.UUID                   `json:"outlet_id"`
	IngredientID uuid.UUID                   `json:"ingredient_id"`
	Date         time.Time                   `json:"date"`
	Column       constants.StockReportColumn `json:"column"`
	Value        float32                     `json:"value"`
}

// API to edit stock report
//
//encore:api auth method=POST path=/api/admin/outlets/stock-report/edit
func (s *Service) EditStockReport(ctx context.Context, req *EditStockReportRequest) error {
	outlet := models.Outlet{}
	err := s.db.First(&outlet, "id = ?", req.OutletID).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.UpdateStockAction, &outlet.BusinessID, nil); err != nil {
		return err
	}

	date, err := common.GetStartDateOfMonth(req.Date)
	if err != nil {
		return err
	}

	// Get Stock Report by ingredient,outlet,date
	// Create if not found
	stockReport, err := s.getStockReport(req, date, outlet.BusinessID)
	if err != nil {
		return err
	}

	ingredient := models.Ingredient{}
	err = s.db.First(&ingredient, "id = ?", req.IngredientID).Error
	if err != nil {
		return err
	}

	// Update the stock report based on column and value
	decimalValue := common.ConvertFloat32ToDecimal(req.Value)
	details := ""
	switch req.Column {
	case constants.StockReportColumnOpening:
		details = fmt.Sprintf("Opening Stock:-\nOutlet: %s\nIngredient: %s\nMonth: %s\nChanged from %s%s to %s%s", outlet.Name, ingredient.Name, date.Format("January 2006"), stockReport.Opening.String(), ingredient.Unit, decimalValue.String(), ingredient.Unit)
		stockReport.Opening = &decimalValue
	case constants.StockReportColumnClosing:
		details = fmt.Sprintf("Closing Stock:-\nOutlet: %s\nIngredient: %s\nMonth: %s\nChanged from %s%s to %s%s", outlet.Name, ingredient.Name, date.Format("January 2006"), stockReport.Closing.String(), ingredient.Unit, decimalValue.String(), ingredient.Unit)
		stockReport.Closing = &decimalValue
		s.updateOpeningStockReportForNextMonth(req, date, outlet.BusinessID)
	case constants.StockReportColumnPurchases:
		details = fmt.Sprintf("Purchases Stock:-\nOutlet: %s\nIngredient: %s\nMonth: %s\nChanged from %v%s to %v%s", outlet.Name, ingredient.Name, date.Format("January 2006"), stockReport.Purchases, ingredient.Unit, req.Value, ingredient.Unit)
		stockReport.Purchases = req.Value
	}

	err = s.db.Save(&stockReport).Error
	if err != nil {
		return err
	}

	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return err
	}

	activityLog := &models.ActivityLog{
		Activity:       constants.LOG_ACTION_UPDATE_STOCK_REPORT,
		Status:         constants.LOG_STATUS_SUCCESS,
		ActionByUserID: user.ID,
		Details:        details,
	}
	s.db.Create(activityLog)

	return nil
}

func (s *Service) getStockReport(req *EditStockReportRequest, date time.Time, businessID uuid.UUID) (*models.StockReport, error) {
	// Get Stock Report by ingredient,outlet,date
	// Create if not found
	startDate, err := common.GetStartOfDay(date)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetEndOfDay(date)
	if err != nil {
		return nil, err
	}

	stockReport := models.StockReport{}
	result := s.db.Model(&models.StockReport{}).
		Where("ingredient_id = ?", req.IngredientID).
		Where("outlet_id = ?", req.OutletID).
		Where("created_at BETWEEN ? AND ?", startDate.UTC(), endDate.UTC()).
		First(&stockReport)

	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}
		stockReport = models.StockReport{
			IngredientID:    req.IngredientID,
			OutletID:        req.OutletID,
			BusinessID:      businessID,
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
			CreatedAt:       startDate,
		}
		result = s.db.Create(&stockReport)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &stockReport, nil
}

func (s *Service) updateOpeningStockReportForNextMonth(req *EditStockReportRequest, date time.Time, businessID uuid.UUID) error {
	// Load Malaysia location
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		log.Fatal("Failed to load location:", err)
	}

	// Convert UTC date to Malaysia time
	dateInMY := date.In(loc)

	// Get first day of next month in Malaysia time
	firstOfThisMonth := time.Date(dateInMY.Year(), dateInMY.Month(), 1, 0, 0, 0, 0, loc)
	firstOfNextMonth := firstOfThisMonth.AddDate(0, 1, 0)

	stockReport, err := s.getStockReport(req, firstOfNextMonth, businessID)
	if err != nil {
		return err
	}

	decimalValue := common.ConvertFloat32ToDecimal(req.Value)
	stockReport.Opening = &decimalValue
	return s.db.Save(&stockReport).Error
}

// API to get product wastage reports
//
//encore:api auth method=POST path=/api/admin/product-wastage-reports/get-all
func (s *Service) GetProductWastageReports(ctx context.Context, req *GetStockReportsRequest) (*ProductWastageDailyStat, error) {
	outlet := models.Outlet{}
	s.db.First(&outlet, "id = ?", req.OutletID)

	if req.BusinessID == nil {
		req.BusinessID = &outlet.BusinessID
	}

	startDate, err := common.GetStartOfDay(*req.FromDate)
	if err != nil {
		return nil, err
	}
	endDate, err := common.GetEndOfDay(*req.ToDate)
	if err != nil {
		return nil, err
	}

	productWastageTypes := []models.ProductWastageType{}
	err = s.db.Model(&models.ProductWastageType{}).
		Where("business_id = ?", *req.BusinessID).
		Find(&productWastageTypes).Error
	if err != nil {
		return nil, err
	}

	var header = []string{}
	for _, productWastageType := range productWastageTypes {
		header = append(header, productWastageType.Name)
	}

	products := []models.Product{}
	err = s.db.Model(&models.Product{}).
		Where("business_id = ?", *req.BusinessID).
		Find(&products).Error
	if err != nil {
		return nil, err
	}

	ProductWastageItemStatList := []ProductWastageItemStat{}
	for _, product := range products {
		productWastageItemStat := ProductWastageItemStat{
			ProductID:    product.ID,
			ProductName:  product.Name,
			TotalWastage: 0,
			WastageStats: make(map[string]float32),
			SortOrder:    product.SortOrder,
		}
		wastageStatsMap := make(map[string]float32)

		for _, productWastageType := range productWastageTypes {
			var wastageAmount float32
			query := s.db.Model(&models.ProductWastageReport{}).
				Where("wastage_type_id = ?", productWastageType.ID).
				Where("product_id = ?", product.ID).
				Where("DATE(created_at) BETWEEN ? AND ?", startDate, endDate)

			if req.OutletID != nil {
				query = query.Where("outlet_id = ?", *req.OutletID)
			}

			if req.OutletGroupID != nil {
				outletIDs, err := common_operations.GetOutletIDsByGroupID(s.db, *req.OutletGroupID)
				if err != nil {
					return nil, err
				}
				query = query.Where("outlet_id IN (?)", outletIDs)
			}

			err = query.Select("SUM(wastage_amount) as total_wastage_amount").Scan(&wastageAmount).Error
			if err != nil {
				wastageAmount = 0
			}

			wastageStatsMap[productWastageType.Name] = wastageAmount
			productWastageItemStat.TotalWastage += wastageAmount
		}
		productWastageItemStat.WastageStats = wastageStatsMap
		ProductWastageItemStatList = append(ProductWastageItemStatList, productWastageItemStat)
	}

	sort.Slice(ProductWastageItemStatList, func(i, j int) bool {
		return ProductWastageItemStatList[i].SortOrder < ProductWastageItemStatList[j].SortOrder
	})

	return &ProductWastageDailyStat{
		Header:         header,
		ProductWastage: ProductWastageItemStatList,
	}, nil
}
