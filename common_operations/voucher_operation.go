package common_operations

import (
	"strings"
	"time"

	"encore.app/database/models"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// get voucher by voucher id or voucher code in customer voucher table
func GetVoucherByVoucherCodeInVoucherTable(
	trx *gorm.DB,
	voucherID uuid.UUID,
	voucherCode string,
) (*models.Voucher, error) {
	var voucher models.Voucher

	db := trx.Model(&models.Voucher{})

	if voucherID != uuid.Nil {
		db = db.Where("id = ?", voucherID)
	}
	if voucherCode != "" {
		db = db.Where("voucher_code = ?", voucherCode)
	}

	db = db.Where("is_active = ?", true)

	if err := db.First(&voucher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "voucher not found",
			}
		}
		return nil, err
	}
	return &voucher, nil
}

// get voucher by voucher id or voucher code in customer voucher table
func GetVoucherByVoucherCodeInCustomerVoucherTable(
	trx *gorm.DB,
	voucherID uuid.UUID,
	voucherCode string,
) (*models.CustomerVoucher, error) {
	var customerVoucher models.CustomerVoucher

	db := trx.Model(&models.CustomerVoucher{}).
		Where("used = ?", false)

	if voucherID != uuid.Nil {
		db = db.Where("voucher_id = ?", voucherID)
	}
	if voucherCode != "" {
		db = db.Where("voucher_code = ?", voucherCode)
	}

	if err := db.Preload("Voucher").First(&customerVoucher).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "voucher not found",
			}
		}
		return nil, err
	}
	return &customerVoucher, nil
}

func VoucherRequirementsFullCheck(
	trx *gorm.DB,
	voucher models.Voucher,
	grossTotal float32,
	OrderMethod models.VoucherEligibleOrderMethod,
	Platform models.VoucherPlatform,
	outletID uuid.UUID,
	productIDs []uuid.UUID,
	productCategoryIDs []uuid.UUID,
) error {
	// check valid from
	validFrom := voucher.ValidFrom
	validTo := voucher.ValidTo
	// Check MinPurchase only if it's not nil
	if voucher.MinPurchase != nil && *voucher.MinPurchase > grossTotal {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "voucher is not meet the minimum spend",
		}
	}
	// max purchase not null, not zero, and less than gross total
	// fulfill this condition mean is invalid, else valid
	if voucher.MaxPurchase != nil && *voucher.MaxPurchase > 0 && *voucher.MaxPurchase < grossTotal {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this spending amount",
		}
	}
	timeNow := time.Now()
	if timeNow.Before(validFrom) || timeNow.After(validTo) {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "voucher is invalid",
		}
	}
	// current usage + current redemption must not exceed/equal max redemption
	currentRedemption := voucher.CurrentRedemptions
	currentUsage := voucher.CurrentUsage
	totalUsage := currentRedemption + currentUsage
	// CHECK is it reach max redemption
	// max redemption not null, not zero, and less than total usage
	// fulfill this condition mean is invalid, else valid
	maxRedemption := voucher.MaxRedemption
	if maxRedemption != nil && *maxRedemption > 0 && totalUsage >= *maxRedemption {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher has reached the maximum usage",
		}
	}

	// check voucher eligible order method
	if !CheckVoucherEligibleOrderMethod(voucher, OrderMethod) {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this order method",
		}
	}
	// check voucher eligible platform
	if !CheckVoucherEligiblePlatform(voucher, Platform) {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this platform",
		}
	}
	// get voucher eligibility rules by voucher id
	voucherEligibilityRules, err := GetVoucherEligibilityRulesByVoucherID(trx, voucher.ID)
	if err != nil {
		return nil
	}
	var outletCheck bool = false
	var productCheck = make(map[uuid.UUID]bool, 0) // product id (key) and bool (value)
	var productCategoryCheck bool = true           // need to change to false after implement product category check
	var userCheck bool = true                      // need to change to false after implement user check

	// initial product check with false
	for _, productID := range productIDs {
		productCheck[productID] = false
	}
	var eligibilityOutletExist bool = false // this is to check if the voucher has outlet eligibility rule
	var eligibiltyProductExist bool = false // this is to check if the voucher has product eligibility rule
	if len(voucherEligibilityRules) > 0 {
		for _, egRule := range voucherEligibilityRules {
			ruleType := egRule.EligibilityRuleType
			if ruleType == models.VoucherEligibilityRuleOutlet {
				// check outlet id is not nil
				if outletID == uuid.Nil {
					return &errs.Error{
						Code:    errs.InvalidArgument,
						Message: "Outlet id is required",
					}
				}
				eligibilityOutletExist = true
				if outletID == *egRule.OutletID {
					outletCheck = true
				}
				continue
			}

			if ruleType == models.VoucherEligibilityRuleProduct {
				if len(productIDs) == 0 {
					return &errs.Error{
						Code:    errs.InvalidArgument,
						Message: "Product ids are required",
					}
				}
				eligibiltyProductExist = true
				productCheck[*egRule.ProductID] = true
				continue
			}

			if ruleType == models.VoucherEligibilityRuleProductCategory {
				// future implementation
				continue
			}

			if ruleType == models.VoucherEligibilityRuleUser {
				// future implementation
				continue
			}
		}
	}

	if eligibilityOutletExist && !outletCheck {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this outlet",
		}
	}

	// only check product check if eligibilty product exist
	if eligibiltyProductExist {
		for _, check := range productCheck {
			if !check {
				return &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "This voucher is not applicable for this product",
				}
			}
		}
	}

	if !productCategoryCheck {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this product category",
		}
	}

	if !userCheck {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "This voucher is not applicable for this user",
		}
	}

	return nil
}

// check voucher eligible order method
func CheckVoucherEligibleOrderMethod(voucher models.Voucher, orderMethod models.VoucherEligibleOrderMethod) bool {
	eligibleOrderMethod := voucher.EligibleOrderMethod
	pickupAndDelivery := models.VoucherEligibleOrderMethodPickupAndDelivery
	// ensure dont check if is support both pickup and delivery
	compare := strings.EqualFold(string(eligibleOrderMethod), string(pickupAndDelivery))
	if !compare {
		// compare with order method
		compare = strings.EqualFold(string(eligibleOrderMethod), string(orderMethod))
	}
	return compare
}

// check voucher eligible platform
func CheckVoucherEligiblePlatform(voucher models.Voucher, platform models.VoucherPlatform) bool {
	eligiblePlatform := voucher.EligiblePlatform
	// ensure dont check if is support all platforms
	compare := strings.EqualFold(string(eligiblePlatform), string(models.VoucherPlatformAll))
	if !compare {
		// compare with platform
		compare = strings.EqualFold(string(eligiblePlatform), string(platform))
	}
	return compare
}

// get voucher eligibility rules by voucher id
func GetVoucherEligibilityRulesByVoucherID(trx *gorm.DB, voucherID uuid.UUID) ([]models.VoucherEligibilityRule, error) {
	var voucherEligibilityRules []models.VoucherEligibilityRule
	query := trx.
		Model(&models.VoucherEligibilityRule{}).
		Where("voucher_id = ?", voucherID).
		Order("created_at ASC")
	result := query.Find(&voucherEligibilityRules)
	if result.Error != nil {
		return nil, result.Error
	}
	return voucherEligibilityRules, nil
}

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
func ApplyChangesOnAmountOfVoucherWithoutCheck(trx *gorm.DB, voucherID uuid.UUID, grossTotal float32) error {
	var voucher models.Voucher
	result := trx.Model(&models.Voucher{}).
		Where("id = ?", voucherID).
		First(&voucher)
	if result.Error != nil {
		return result.Error
	}
	// apply changes on voucher usage
	voucher.CurrentUsage += 1
	result = trx.Save(&voucher)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
