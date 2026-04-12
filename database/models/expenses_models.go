package models

import (
	"time"

	"encore.dev/types/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type ExpenseCategory string

const (
	// Staff and Labor
	ExpenseCategoryStaffSalary    ExpenseCategory = "Staff Salary"
	ExpenseCategoryPartTimerDaily ExpenseCategory = "Part timer (Daily)"

	// Utilities
	ExpenseCategoryGas         ExpenseCategory = "Gas"
	ExpenseCategoryWaterCharge ExpenseCategory = "Water Charge"

	// Kitchen and Cleaning Supplies
	ExpenseCategoryDishwashingLiquidPaste ExpenseCategory = "Dishwashing liquid/paste"
	ExpenseCategoryCookingOil             ExpenseCategory = "Cooking Oil"
	ExpenseCategoryTissue                 ExpenseCategory = "Tissue"
	ExpenseCategoryGarbageBag             ExpenseCategory = "Garbage Bag"
	ExpenseCategoryGlove                  ExpenseCategory = "Glove"

	// Transportation
	ExpenseCategoryTransportationPetrol ExpenseCategory = "Transportation (petrol)"
	ExpenseCategoryPetrolAllowanceCrew  ExpenseCategory = "Petrol Allowance (crew)"
	ExpenseCategoryUpkeepOfVehicle      ExpenseCategory = "Upkeep of vehicle"

	// Administrative
	ExpenseCategoryLicenseFee            ExpenseCategory = "License fee"
	ExpenseCategoryPrintingAndStationery ExpenseCategory = "Printing & Stationery"
	ExpenseCategoryOutletSupplies        ExpenseCategory = "Outlet Supplies"

	// Maintenance
	ExpenseCategoryUpkeepOfOutlet      ExpenseCategory = "Upkeep of outlet"
	ExpenseCategoryUpkeepOfCentreHouse ExpenseCategory = "Upkeep of centre house"
	ExpenseCategoryLalamoveOrGrab      ExpenseCategory = "Lalamove/Grab"

	// Other
	ExpenseCategoryOther ExpenseCategory = "Other"
)

// Add this helper function
func IsValidExpenseCategory(category string) bool {
	switch ExpenseCategory(category) {
	case ExpenseCategoryStaffSalary,
		ExpenseCategoryPartTimerDaily,
		ExpenseCategoryGas,
		ExpenseCategoryWaterCharge,
		ExpenseCategoryDishwashingLiquidPaste,
		ExpenseCategoryCookingOil,
		ExpenseCategoryTissue,
		ExpenseCategoryGarbageBag,
		ExpenseCategoryGlove,
		ExpenseCategoryTransportationPetrol,
		ExpenseCategoryPetrolAllowanceCrew,
		ExpenseCategoryUpkeepOfVehicle,
		ExpenseCategoryLicenseFee,
		ExpenseCategoryPrintingAndStationery,
		ExpenseCategoryOutletSupplies,
		ExpenseCategoryUpkeepOfOutlet,
		ExpenseCategoryUpkeepOfCentreHouse,
		ExpenseCategoryLalamoveOrGrab,
		ExpenseCategoryOther:
		return true
	default:
		return false
	}
}

type ExpensesOutlet struct {
	ID                    uuid.UUID       `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	OutletID              uuid.UUID       `json:"outlet_id" gorm:"type:uuid"`
	Outlet                *Outlet         `json:"outlet" gorm:"foreignKey:OutletID"`
	ExpensesCategory      ExpenseCategory `json:"expenses_category" gorm:"type:varchar(255)"`
	ExpensesDate          time.Time       `json:"expenses_date" gorm:"type:date"`
	ExpensesAmount        decimal.Decimal `json:"expenses_amount" gorm:"type:decimal(20,2)"`
	ExpensesDescription   string          `json:"expenses_description" gorm:"type:text"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             *time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt  `json:"deleted_at,omitempty"`
	ExpensesAttachmentUrl string          `json:"expenses_attachment_url" gorm:"type:text"`
}
