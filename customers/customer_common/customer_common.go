package customer_common

import (
	"fmt"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// CUSTOMERS
// ========================================

// func to get customer data from auth data
func GetCustomerDataFromAuthData(authData func() any) (*models.User, error) {
	customerData, ok := authData().(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}
	return customerData, nil
}

func GetCustomerDataFromEmailAndBusinessID(trx *gorm.DB, email string, business_id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	result := trx.Model(&models.Customer{}).
		Where("email = ? AND business_id = ?", email, business_id).
		First(&customer)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customer, nil
}

// get all favourite products by customer id
func GetAllFavouriteProductsByCustomerID(trx *gorm.DB, customer_id uuid.UUID, isAllowInactive bool) ([]models.CustomerFavouriteProduct, error) {
	var favouriteProducts []models.CustomerFavouriteProduct
	var filteredFavouriteProducts []models.CustomerFavouriteProduct
	result := trx.Model(&models.CustomerFavouriteProduct{}).
		Where("customer_id = ?", customer_id).
		Preload("Product").
		Find(&favouriteProducts)
	if result.Error != nil {
		return []models.CustomerFavouriteProduct{}, result.Error
	}
	if !isAllowInactive {
		for _, favouriteProduct := range favouriteProducts {
			if favouriteProduct.Product.IsActive {
				filteredFavouriteProducts = append(filteredFavouriteProducts, favouriteProduct)
			}
		}
	}
	return filteredFavouriteProducts, nil
}

// get customer information with favourite products (preload)
func GetAllCustomerFavouriteProducts(trx *gorm.DB, customer_id uuid.UUID) (*models.Customer, error) {
	var customer models.Customer
	result := trx.Model(&models.Customer{}).
		Where("customer_id = ?", customer_id).
		Preload("CustomerFavouriteProducts").
		Find(&customer)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customer, nil
}

// get customer favourite product by product id
func GetProductIsFavouriteByProductID(trx *gorm.DB, customer_id uuid.UUID, product_id uuid.UUID) (bool, error) {
	var customerFavouriteProduct models.CustomerFavouriteProduct
	result := trx.Model(&models.CustomerFavouriteProduct{}).
		Where("customer_id = ? AND product_id = ?", customer_id, product_id).
		First(&customerFavouriteProduct)
	if result.Error != nil {
		return false, result.Error
	}
	return true, nil
}

// get customer delivery address only
func GetCustomerDeliveryAddressOnly(trx *gorm.DB, customer_id uuid.UUID) ([]models.CustomerDeliveryAddress, error) {
	var customerDeliveryAddresses []models.CustomerDeliveryAddress
	result := trx.Model(&models.CustomerDeliveryAddress{}).
		Where("customer_id = ?", customer_id).
		Find(&customerDeliveryAddresses)
	if result.Error != nil {
		return []models.CustomerDeliveryAddress{}, result.Error
	}
	return customerDeliveryAddresses, nil
}

// get customer delivery address by id
func GetCustomerDeliveryAddressByID(trx *gorm.DB, customer_id uuid.UUID, address_id uuid.UUID) (*models.CustomerDeliveryAddress, error) {
	var customerDeliveryAddress models.CustomerDeliveryAddress
	result := trx.Model(&models.CustomerDeliveryAddress{}).
		Where("customer_id = ? AND id = ?", customer_id, address_id).
		First(&customerDeliveryAddress)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerDeliveryAddress, nil
}
