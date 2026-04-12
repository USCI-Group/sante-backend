package common_operations

import (
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// get customer by phone number
func GetCustomerByPhoneNumber(trx *gorm.DB, phone_number string, business_id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	result := trx.Model(&models.Customer{}).
		Where("phone_number = ?", phone_number).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Where("email_verified = ?", true).
		First(&customer)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customer, nil
}

// get customer membership by customer id
func GetCustomerMembershipByCustomerID(trx *gorm.DB, customer_id uuid.UUID) (*models.CustomerMembership, error) {
	var customerMembership models.CustomerMembership
	result := trx.Model(&models.CustomerMembership{}).
		Where("customer_id = ?", customer_id).
		Preload("Membership").
		Preload("Membership.Benefits").
		Preload("Membership.UpgradeRules.Product").
		First(&customerMembership)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerMembership, nil
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
	point_rule_id uuid.UUID,
	points_earned int,
	order_id *uuid.UUID,
) error {
	pointTransaction := models.PointTransaction{
		PointRuleID:  &point_rule_id,
		CustomerID:   customer_id,
		PointsEarned: points_earned,
		OrderID:      order_id,
		EarnedAt:     time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    nil,
		DeletedAt:    gorm.DeletedAt{},
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
