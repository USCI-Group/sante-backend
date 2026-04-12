package customer_common

import (
	"errors"
	"fmt"
	"time"

	"encore.app/common"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// POINT RULE MANAGEMENT & UTILITIES
// ========================================

// calculate points for each action type
func CalculatePointsBasedOnRule(trx *gorm.DB, pointRule models.PointRule, amount float32) (int, error) {
	// if min amount is not nil and amount is less than min amount, skip
	if pointRule.MinAmount != nil && amount < *pointRule.MinAmount {
		return 0, nil
	}
	// if points multiplier is not nil and greater than 0, use points multiplier to calculate points
	if pointRule.PointsMultiplier != nil && *pointRule.PointsMultiplier > 0 {
		return int(amount * float32(*pointRule.PointsMultiplier)), nil
	} else {
		// if points multiplier is nil or 0, use points earned to calculate points
		return int(*pointRule.PointsEarned), nil
	}
}

// function to add points transaction
func AddPointsTransaction(
	trx *gorm.DB,
	customer_id uuid.UUID,
	point_rule_id *uuid.UUID,
	points_earned int,
	order_id *uuid.UUID,
	details *string,
) error {
	fmt.Println("AddPointsAfterPayment", customer_id, point_rule_id, points_earned)
	pointTransaction := models.PointTransaction{
		PointRuleID:  point_rule_id,
		CustomerID:   customer_id,
		PointsEarned: points_earned,
		OrderID:      order_id,
		EarnedAt:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    nil,
		DeletedAt:    gorm.DeletedAt{},
		Details:      details,
	}
	result := trx.Create(&pointTransaction)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// function to add points to customer membership (table which has include record customer points)
func AddPointsToCustomerMembership(trx *gorm.DB, customer_id uuid.UUID, points_earned int) error {
	var customerMembership models.CustomerMembership
	result := trx.Model(&models.CustomerMembership{}).Where("customer_id = ?", customer_id).First(&customerMembership)
	if result.Error != nil {
		return result.Error
	}
	customerMembership.Points += points_earned
	result = trx.Save(&customerMembership)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// get point rule by business id and action type with or without membership id
func GetPointRuleByBusinessIDAndActionTypeAndMembershipID(trx *gorm.DB, business_id uuid.UUID, action_types []models.ActionType, membership_id uuid.UUID, additonalWhereSQL *string, additonalOrderSQL *string) ([]models.PointRule, error) {
	var pointRules []models.PointRule
	//either membership_id matches OR membership_id is NULL
	query := trx.Model(&models.PointRule{}).
		Where("business_id = ? AND action_type IN (?)", business_id, action_types).
		Where("membership_id = ? OR membership_id IS NULL", membership_id).
		Where("is_active = ?", true)
	if additonalWhereSQL != nil {
		query = query.Where(*additonalWhereSQL)
	}
	if additonalOrderSQL != nil {
		query = query.Order(*additonalOrderSQL)
	}

	result := query.Find(&pointRules)
	if result.Error != nil {
		return nil, result.Error
	}

	return pointRules, nil
}

// get all daily check in point rules
func GetAllDailyCheckInPointRules(trx *gorm.DB, business_id uuid.UUID) ([]models.PointRule, error) {
	var pointRules []models.PointRule
	query := trx.Model(&models.PointRule{}).
		Where("business_id = ? AND action_type = ?", business_id, models.ActionTypeDailyCheckin).
		Where("is_active = ?", true)

	result := query.Find(&pointRules)
	if result.Error != nil {
		return nil, result.Error
	}

	return pointRules, nil
}

// get customer daily check in record - returns the latest transaction from the provided point rule IDs
func GetCustomerDailyCheckInRecord(trx *gorm.DB, customer_id uuid.UUID, point_rule_ids []uuid.UUID) (*models.PointTransaction, error) {
	var pointTransaction models.PointTransaction
	query := trx.Model(&models.PointTransaction{}).
		Preload("PointRule").
		Where("customer_id = ?", customer_id).
		Where("point_rule_id IN (?)", point_rule_ids).
		Order("created_at DESC").
		Limit(1)
	result := query.First(&pointTransaction)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil if no records found
		}
		return nil, result.Error
	}

	return &pointTransaction, nil
}

// check customer is checked in today
func CheckCustomerIsCheckedInToday(trx *gorm.DB, business_id uuid.UUID, customer_id uuid.UUID) (bool, error) {
	// get all daily check in point rules
	allDailyCheckInPointRules, err := GetAllDailyCheckInPointRules(trx, business_id)
	if err != nil {
		return false, err
	}

	// get point rule ids
	pointRuleIds := []uuid.UUID{}
	for _, pointRule := range allDailyCheckInPointRules {
		pointRuleIds = append(pointRuleIds, pointRule.ID)
	}

	// get customer daily check in record
	customerDailyCheckInRecord, err := GetCustomerDailyCheckInRecord(trx, customer_id, pointRuleIds)
	if err != nil {
		return false, err
	}

	if customerDailyCheckInRecord != nil {
		startOfDay, _ := common.GetStartOfDay(time.Now())
		startOfDayCustomerDailyCheckInRecord, _ := common.GetStartOfDay(customerDailyCheckInRecord.CreatedAt)

		if startOfDay.Equal(startOfDayCustomerDailyCheckInRecord) {
			return true, nil
		}
	}

	return false, nil
}

// get point transaction by customer id
func GetPointTransactionByCutomerID(trx *gorm.DB, customer_id uuid.UUID) ([]models.PointTransaction, error) {
	var pointTransactions []models.PointTransaction
	query := trx.Model(&models.PointTransaction{}).
		Where("customer_id = ?", customer_id).
		Preload("PointRule").
		Order("earned_at DESC")
	result := query.Find(&pointTransactions)
	if result.Error != nil {
		return nil, result.Error
	}
	return pointTransactions, nil
}

// get points rewarded based on
func GetPointsRewardedBasedOnOrderID(trx *gorm.DB, order_id uuid.UUID) ([]models.PointTransaction, error) {
	var pointTransactions []models.PointTransaction
	query := trx.Model(&models.PointTransaction{}).
		Where("order_id = ?", order_id).
		Order("earned_at DESC")
	result := query.Find(&pointTransactions)
	if result.Error != nil {
		return nil, result.Error
	}
	return pointTransactions, nil
}

// func to calculate points based on rule
func CalculatePointsBasedOnRules(trx *gorm.DB, customerID uuid.UUID, businessID uuid.UUID, actionTypes []models.ActionType, amount float32) (int, error) {
	customerMembership, err := GetCustomerMembershipByCustomerID(trx, customerID)
	if err != nil {
		return 0, err
	}

	// get point rules based on action type and business id with membership id if provided
	pointRules, err := GetPointRuleByBusinessIDAndActionTypeAndMembershipID(trx, businessID, actionTypes, customerMembership.MembershipID, nil, nil)
	if err != nil {
		return 0, err
	}

	// calculate points for each action type
	var pointsAccumulated float32
	for _, pointRule := range pointRules {
		points, err := CalculatePointsBasedOnRule(trx, pointRule, amount)
		if err != nil {
			return 0, err
		}
		pointsAccumulated += float32(points)
	}

	return int(pointsAccumulated), nil
}
