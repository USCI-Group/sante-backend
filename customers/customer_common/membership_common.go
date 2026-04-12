package customer_common

import (
	"errors"
	"math"

	"encore.app/customers/customer_type"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// CUSTOMER MEMBERSHIP STATS MANAGEMENT & UTILITIES
// ========================================

// get customer membership stats by customer id
func GetCustomerMembershipStatsByCustomerID(trx *gorm.DB, customer_id uuid.UUID, business_id uuid.UUID) (*models.CustomerMembershipStats, error) {
	var customerMembershipStats models.CustomerMembershipStats
	result := trx.Model(&models.CustomerMembershipStats{}).
		Where("customer_id = ? AND business_id = ?", customer_id, business_id).
		Preload("ProductPurchaseStats").
		First(&customerMembershipStats)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerMembershipStats, nil
}

// ========================================
// CUSTOMER MEMBERSHIP MANAGEMENT & UTILITIES
// ========================================

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

// get customer membership by customer id
func GetMembershipOfNextTier(trx *gorm.DB, customerMembership models.CustomerMembership) (*models.Membership, error) {
	currentTierLvl := customerMembership.Membership.TierLevel
	nextTierLvl := currentTierLvl + 1
	var membership models.Membership
	result := trx.Model(&models.Membership{}).
		Where("tier_level = ?", nextTierLvl).
		Where("business_id = ?", customerMembership.Membership.BusinessID).
		Preload("Benefits").
		Preload("UpgradeRules.Product").
		First(&membership)
	if result.Error != nil {
		// mean next tier is the highest tier
		return nil, errors.New("next tier is the highest tier")
	}

	return &membership, nil

}

// get all products required by upgrade rule
func GetAllProductsRequiredByUpgradeRule(trx *gorm.DB, upgradeRules []models.MembershipUpgradeRule) ([]models.MembershipUpgradeRule, error) {
	var productsRequired []models.MembershipUpgradeRule
	for _, rule := range upgradeRules {
		if rule.ProductID == nil {
			continue
		}
		productsRequired = append(productsRequired, rule)
	}
	return productsRequired, nil
}

// get membership by id
func GetMembershipByID(trx *gorm.DB, membership_id uuid.UUID) (*models.Membership, error) {
	var membership models.Membership
	result := trx.Model(&models.Membership{}).Where("id = ?", membership_id).First(&membership)
	if result.Error != nil {
		return nil, result.Error
	}
	return &membership, nil
}

// function to get experience points rate by business id
func GetExperiencePointsByBusinessID(
	trx *gorm.DB,
	business_id uuid.UUID,
	totalSpendingAmount float32,
	products []customer_type.ProductWithQuantity,
) (int, error) {
	var businessConfiguration models.BusinessConfiguration
	query := trx.Select("experience_points_per_currency", "membership_upgrade_method").
		Model(&models.BusinessConfiguration{}).
		Where("business_id = ?", business_id).
		First(&businessConfiguration)
	if query.Error != nil {
		return 0, nil
	}

	membershipUpgradeMethod := *businessConfiguration.MembershipUpgradeMethod
	experiencePointsRate := *businessConfiguration.ExperiencePointsPerCurrency

	var totalExperiencePoints float32

	// flatten the products to get the product ids
	productsIDs := make([]uuid.UUID, len(products))
	for i, product := range products {
		productsIDs[i] = product.ProductID
	}

	// map of product id : quantity (refer to quantity of product in cart)
	productQuantities := make(map[uuid.UUID]int)
	for i, product := range products {
		productQuantities[productsIDs[i]] = product.Quantity
	}

	switch membershipUpgradeMethod {
	case models.MembershipUpgradeMethodPriceToExperience:
		totalExperiencePoints = totalSpendingAmount * float32(experiencePointsRate)
	case models.MembershipUpgradeMethodProductToExperience:
		products, err := GetProductsByIDs(trx, productsIDs)
		if err != nil {
			return 0, err
		}
		for _, product := range products {
			productExp := float32(product.ExperiencePoints)
			if productExp == 0 {
				continue
			}
			quantity := productQuantities[product.ID]
			if quantity == 0 {
				continue
			}
			totalExperiencePoints += productExp * float32(quantity)
		}
	}

	return int(math.Floor(float64(totalExperiencePoints))), nil
}
