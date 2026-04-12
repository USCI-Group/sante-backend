package customer_users

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"math"

	"encore.app/auth_service"
	"encore.app/aws_ses"
	"encore.app/common"
	"encore.app/customers/customer_auth"
	"encore.app/customers/customer_common"
	"encore.app/database"
	"encore.app/database/models"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 50km is the largest allowed distance for a customer to be able to see an outlet
var largestAllowedDistance = 50.0

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
	return &Service{db: db}, nil
}

// CreateCustomerUserParams is the parameters for creating a customer user
type CreateCustomerAccountParams struct {
	FirstName              string    `json:"first_name" valid:"required~FirstName is required"`
	LastName               string    `json:"last_name" valid:"required~LastName is required"`
	Email                  string    `json:"email" valid:"required~Email is required,email~Invalid email format"`
	Password               string    `json:"password" valid:"required~Password is required"`
	PhoneNumber            string    `json:"phone_number" valid:"required~PhoneNumber is required"`
	DateOfBirth            string    `json:"date_of_birth" valid:"required~Date of Birth is required"`
	BusinessID             uuid.UUID `json:"business_id"`
	DeviceID               *string   `json:"device_id"`
	FCMToken               *string   `json:"fcm_token"`
	IsAgreeToTerms         bool      `json:"is_agree_to_terms" valid:"required~You must agree to the terms and conditions"`
	IsAgreeToPrivacyPolicy bool      `json:"is_agree_to_privacy_policy" valid:"required~You must agree to the privacy policy"`
	IsNewsletterSubscribed bool      `json:"is_newsletter_subscribed"`
}

type CreateCustomerAccountResponse struct {
	Customer     *models.Customer `json:"customer"`
	JWTToken     string           `json:"jwt_token"`
	RefreshToken string           `json:"refresh_token"`
}

type VerifyEmailResponse struct {
	Message string `json:"message"`
}

type TokenInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	TokenExpiry time.Time `json:"token_expiry"`
}

type LoginCustomerUserRequest struct {
	Email    string `json:"email" valid:"required~Email is required,email~Invalid email format"`
	Password string `json:"password" valid:"required~Password is required"`
}

type LoginCustomerUserResponse struct {
	//Customer *models.Customer `json:"customer"`
	Response int    `json:"response"`
	JWTToken string `json:"jwt_token"`
	Message  string `json:"message"`
}

type GetUserResponse struct {
	User *models.Customer `json:"user"`
}

type ToggleProductToFavouriteRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	ProductID  uuid.UUID `json:"product_id"`
}

type CreateCustomerDeliveryAddressRequest struct {
	CustomerID uuid.UUID      `json:"customer_id"`
	Address    common.Address `json:"address"`
	Name       string         `json:"name"`
	IsDefault  bool           `json:"is_default"`
	Latitude   *float64       `json:"latitude"`
	Longitude  *float64       `json:"longitude"`
}

type GetAllCustomerDeliveryAddressResponse struct {
	Message                   string                           `json:"message"`
	CustomerDeliveryAddresses []models.CustomerDeliveryAddress `json:"customer_delivery_addresses"`
}

type RemoveCustomerDeliveryAddressRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	AddressID  uuid.UUID `json:"address_id"`
}

type UpdateCustomerDeliveryAddressRequest struct {
	CustomerID uuid.UUID      `json:"customer_id"`
	AddressID  uuid.UUID      `json:"address_id"`
	Address    common.Address `json:"address"`
	Name       string         `json:"name"`
	IsDefault  bool           `json:"is_default"`
	Latitude   *float64       `json:"latitude"`
	Longitude  *float64       `json:"longitude"`
}

type GetNearestOutletRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	BusinessID uuid.UUID `json:"business_id"`
	Address    string    `json:"address" valid:"required~Address is required"`
	Radius     float64   `json:"radius,omitempty"` // in kilometers, default 10km
	Latitude   *float64  `json:"latitude"`
	Longitude  *float64  `json:"longitude"`
}

type GetNearestOutletResponse struct {
	Message  string         `json:"message"`
	Outlet   *models.Outlet `json:"outlet"`
	Distance Distance       `json:"distance"`
}

type Distance struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

type SearchOutletRequest struct {
	SearchKey string `json:"search_key"`
}

type SearchOutletResponse struct {
	Message string          `json:"message"`
	Outlets []models.Outlet `json:"outlets"`
}

type GetSurroundingOutletsRequest struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Radius     float64   `json:"radius"`
	BusinessID uuid.UUID `json:"business_id"`
}

type GetSurroundingOutletsResponse struct {
	Message string          `json:"message"`
	Outlets []models.Outlet `json:"outlets"`
}

type GetRecentVisitedOutletsResponse struct {
	Message string          `json:"message"`
	Outlets []models.Outlet `json:"outlets"`
}

type DeactivateAccountRequest struct {
	Comment string `json:"comment"`
}

var permissionDeniedError = &errs.Error{
	Code:    errs.PermissionDenied,
	Message: "Permission denied to geocode address",
}

// createCustomerUser creates a new customer user
//
//encore:api public method=POST path=/api/customers/create-account
func (s *Service) CreateCustomerAccount(ctx context.Context, req *CreateCustomerAccountParams) (*CreateCustomerAccountResponse, error) {

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//manually add format to RFC 3339
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		fmt.Println("error parsing date of birth", err)
		return nil, err
	}

	// Check if email already exists
	var updateExistingAcc bool = false
	var existingCustomer models.Customer
	if err := s.db.Where("email = ? AND business_id = ?", req.Email, req.BusinessID).First(&existingCustomer).Error; err == nil {
		// check email is verified and active or not
		// if email is verified and active, return that account already exists
		if existingCustomer.EmailVerified && existingCustomer.IsActive {
			return nil, &errs.Error{
				Code:    errs.AlreadyExists,
				Message: "Account already exists",
			}
		}
		updateExistingAcc = true // mean it will update existing account data instead of creating new account
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	possiblePhoneNumbers := common.GetPossiblePhoneNumbers(req.PhoneNumber)
	if err := s.db.Where("phone_number IN (?) AND business_id = ?", possiblePhoneNumbers, req.BusinessID).First(&existingCustomer).Error; err == nil {
		// check email is verified and active or not
		// if email is verified and active, return that account already exists
		if existingCustomer.EmailVerified && existingCustomer.IsActive {
			return nil, &errs.Error{
				Code:    errs.AlreadyExists,
				Message: "Phone number already registered with an account",
			}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Validate password
	if err := common.ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	//encode password
	encryptedPass, err := common.EncodePassword(req.Password)
	if err != nil {
		return nil, err
	}

	//create customer user
	customer := &models.Customer{
		FirstName:              req.FirstName,
		LastName:               req.LastName,
		Email:                  req.Email,
		Password:               encryptedPass,
		PhoneNumber:            common.SanitizePhoneNumber(req.PhoneNumber),
		DateOfBirth:            dob,
		CreatedAt:              time.Now(),
		UpdatedAt:              nil,
		EmailVerified:          false,
		BusinessID:             req.BusinessID,
		IsAgreeToTerms:         req.IsAgreeToTerms,
		IsAgreeToPrivacyPolicy: req.IsAgreeToPrivacyPolicy,
		IsNewsletterSubscribed: req.IsNewsletterSubscribed,
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

	// update existing account data instead of creating new account
	if updateExistingAcc {
		customer.ID = existingCustomer.ID
		customer.ProfilePicture = existingCustomer.ProfilePicture
		customer.IsActive = existingCustomer.IsActive
		if err := trx.Save(customer).Error; err != nil {
			trx.Rollback()
			return nil, err
		}
	} else {
		if err := trx.Create(customer).Error; err != nil {
			trx.Rollback()
			return nil, err
		}

	}

	//get membership id based on business id
	var membership models.Membership
	if err := trx.Where("business_id = ? AND tier_level = ?", req.BusinessID, 0).First(&membership).Error; err != nil {
		trx.Rollback()
		return nil, err
	}

	//create customer membership
	// if update existing account, check if customer membership already exists, if not create it
	// if create new account, create customer membership directly
	if updateExistingAcc {
		var existingCustomerMembership models.CustomerMembership
		if err := trx.Where("customer_id = ?", customer.ID).First(&existingCustomerMembership).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// Only create if it doesn't exist
				customerMembership := &models.CustomerMembership{
					CustomerID:   customer.ID,
					MembershipID: membership.ID,
					ExpiryDate:   nil,
					Points:       0,
					CreatedAt:    time.Now(),
					UpdatedAt:    nil,
					DeletedAt:    gorm.DeletedAt{},
				}
				if err := trx.Create(customerMembership).Error; err != nil {
					log.Printf("unable to create customer membership: %v", err)
					trx.Rollback()
					return nil, err
				}
			} else {
				trx.Rollback()
				return nil, err
			}
		}

	} else {
		customerMembership := &models.CustomerMembership{
			CustomerID:   customer.ID,
			MembershipID: membership.ID,
			ExpiryDate:   nil,
			Points:       0,
			CreatedAt:    time.Now(),
			UpdatedAt:    nil,
			DeletedAt:    gorm.DeletedAt{},
		}

		if err := trx.Create(customerMembership).Error; err != nil {
			log.Printf("unable to create customer membership: %v", err)
			trx.Rollback()
			return nil, err
		}

	}

	// Generate JWT token
	/* token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        customer.ID.String(),
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"login_type": "customer",
	})

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(secretsKeys.jwtSecretKey))
	if err != nil {
		log.Printf("unable to generate refresh token: %v", err)
		trx.Rollback()
		return nil, err
	}
	// Generate refresh token
	tokens, err := auth_service.GenerateRefreshTokenForUser()
	if err != nil {
		log.Printf("unable to generate refresh token: %v", err)
		trx.Rollback()
		return nil, err
	}

	customerToken := &models.CustomerToken{
		CustomerID:   customer.ID,
		RefreshToken: tokens.HashedToken,
		ExpiredAt:    time.Now().Add(30 * 24 * time.Hour),
		DeviceID:     req.DeviceID,
		FCMToken:     req.FCMToken,
	}

	if err := trx.Create(customerToken).Error; err != nil {
		log.Printf("unable to create customer token: %v", err)
		trx.Rollback()
		return nil, err
	} */

	err = trx.Commit().Error
	if err != nil {
		log.Printf("unable to commit transaction: %v", err)
		trx.Rollback()
		return nil, err
	}

	customer_auth.SendVerificationCode(ctx, &customer_auth.SendVerificationCodeRequest{
		CustomerID:  customer.ID.String(),
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Action:      models.EmailActionTypeVerifyEmail,
	})

	return &CreateCustomerAccountResponse{
		Customer:     customer,
		JWTToken:     "",
		RefreshToken: "",
		//JWTToken:     signedToken,
		//RefreshToken: tokens.PlainRefreshToken,
	}, nil

}

//encore:api public method=GET path=/customers/verify-email/:token
func (s *Service) VerifyEmail(ctx context.Context, token string) (*VerifyEmailResponse, error) {
	//verify email
	//decode token
	fmt.Println("token", token)
	tokenInfo, err := common.DecodeToken(token)
	if err != nil {
		return nil, err
	}
	//check if token is expired
	if tokenInfo.TokenExpiry.Before(time.Now()) {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Token expired",
		}
	}
	//update customer email verified to true
	if err := s.db.Model(&models.Customer{}).Where("id = ?", tokenInfo.UserID).Update("email_verified", true).Error; err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{
		Message: "Email verified successfully",
	}, nil

}

// get user
//
//encore:api auth method=GET path=/api/customers/get/user/:customer_id
func (s *Service) GetUser(ctx context.Context, customer_id string) (*GetUserResponse, error) {
	var customer models.Customer
	result := s.db.Where("id=?", customer_id).
		First(&customer)

	if result.Error != nil {
		return nil, result.Error
	}
	customer.Password = ""

	return &GetUserResponse{
		User: &customer,
	}, nil
}

// API to toggle product to favourite
//
//encore:api auth method=POST path=/api/customers/toggle/favourite/product
func (s *Service) ToggleProductToFavourite(ctx context.Context, req *ToggleProductToFavouriteRequest) (*common.BasicResponse, error) {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if req.CustomerID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Customer ID is required",
		}
	}

	if req.ProductID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Product ID is required",
		}
	}
	// CHECK if product is already in favourite
	var existingFav models.CustomerFavouriteProduct
	result := trx.Where("customer_id = ? AND product_id = ?", req.CustomerID, req.ProductID).First(&existingFav)
	fmt.Println("existingFav", existingFav)
	if result.Error == nil {
		// product exist then delete it
		result = trx.Unscoped().Delete(&existingFav)
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
			Message: "Product removed from favourite",
		}, nil
	}

	// add product to favourite
	customerFavProduct := &models.CustomerFavouriteProduct{
		CustomerID: req.CustomerID,
		ProductID:  req.ProductID,
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
		DeletedAt:  gorm.DeletedAt{},
	}

	result = trx.Create(customerFavProduct)
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
		Message: "Product added to favourite",
	}, nil
}

// API to create customer delivery address
//
//encore:api auth method=POST path=/api/customers/create/address
func (s *Service) CreateCustomerDeliveryAddress(ctx context.Context, req *CreateCustomerDeliveryAddressRequest) (*common.BasicResponse, error) {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	customerDeliveryAddress := &models.CustomerDeliveryAddress{
		CustomerID: req.CustomerID,
		Address:    req.Address,
		Name:       req.Name,
		IsDefault:  req.IsDefault,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
		DeletedAt:  gorm.DeletedAt{},
	}

	// if is default, set all other address to not default
	if req.IsDefault {
		customerDeliveryAddresses, err := customer_common.GetCustomerDeliveryAddressOnly(trx, req.CustomerID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		// set all other address to not default
		for _, address := range customerDeliveryAddresses {
			address.IsDefault = false
			trx.Save(&address)
		}
	}

	result := trx.Create(customerDeliveryAddress)
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
		Message: "Customer delivery address created successfully",
	}, nil
}

// API to get customer all delivery address
//
//encore:api auth method=GET path=/api/customers/get/all/address/:customer_id
func (s *Service) GetAllCustomerDeliveryAddress(ctx context.Context, customer_id uuid.UUID) (*GetAllCustomerDeliveryAddressResponse, error) {
	customerDeliveryAddresses, err := customer_common.GetCustomerDeliveryAddressOnly(s.db, customer_id)
	if err != nil {
		return nil, err
	}

	return &GetAllCustomerDeliveryAddressResponse{
		Message:                   "Customer delivery addresses fetched successfully",
		CustomerDeliveryAddresses: customerDeliveryAddresses,
	}, nil
}

// API to remove customer delivery address
//
//encore:api auth method=POST path=/api/customers/remove/address
func (s *Service) RemoveCustomerDeliveryAddress(ctx context.Context, req *RemoveCustomerDeliveryAddressRequest) (*common.BasicResponse, error) {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if req.CustomerID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Customer ID is required",
		}
	}
	if req.AddressID == uuid.Nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Address ID is required",
		}
	}

	result := trx.Model(&models.CustomerDeliveryAddress{}).
		Where("id = ? AND customer_id = ?", req.AddressID, req.CustomerID).
		Unscoped().
		Delete(&models.CustomerDeliveryAddress{})
	if result.RowsAffected == 0 {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Address not found",
		}
	}
	if result.RowsAffected > 1 {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Multiple addresses deleted at the same time, rollback transaction",
		}
	}
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
		Message: "Customer delivery address removed successfully",
	}, nil
}

// API to update the customer delivery address
//
//encore:api auth method=POST path=/api/customers/update/address
func (s *Service) UpdateCustomerDeliveryAddress(ctx context.Context, req *UpdateCustomerDeliveryAddressRequest) (*common.BasicResponse, error) {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	customerDeliveryAddress, err := customer_common.GetCustomerDeliveryAddressByID(trx, req.CustomerID, req.AddressID)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	customerDeliveryAddress.Address = req.Address
	customerDeliveryAddress.Name = req.Name
	customerDeliveryAddress.Latitude = req.Latitude
	customerDeliveryAddress.Longitude = req.Longitude
	if req.IsDefault {
		customerDeliveryAddresses, err := customer_common.GetCustomerDeliveryAddressOnly(trx, req.CustomerID)
		if err != nil {
			trx.Rollback()
			return nil, err
		}
		// set all other address to not default
		for _, address := range customerDeliveryAddresses {
			address.IsDefault = false
			trx.Save(&address)
		}
	}
	customerDeliveryAddress.IsDefault = req.IsDefault

	result := trx.Save(customerDeliveryAddress)
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
		Message: "Customer delivery address updated successfully",
	}, nil
}

// Geocode address to coordinates using OpenStreetMap Nominatim
func (s *Service) geocodeAddress(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"

	// Build query parameters
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	// Make request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
		case http.StatusForbidden:
			return 0, 0, permissionDeniedError
		default:
			return 0, 0, &errs.Error{
				Code:    errs.Internal,
				Message: "Failed to geocode address: " + resp.Status,
			}
		}
	}

	// Parse response
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no coordinates found for address: %s", address)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, err
	}

	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, lon, nil
}

// Calculate distance between two coordinates using Haversine formula
func (s *Service) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth's radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// API to get nearest outlet based on customer location
//
//encore:api auth method=POST path=/api/customers/outlets/nearest
func (s *Service) GetNearestOutlet(ctx context.Context, req *GetNearestOutletRequest) (*GetNearestOutletResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// if customer has latitude and longitude, use it
	var customerLat float64
	var customerLon float64
	var errInGeocode error
	// if latitude and longitude are provided, use them
	if req.Latitude != nil && req.Longitude != nil {
		customerLat = *req.Latitude
		customerLon = *req.Longitude
	} else {
		// Convert customer address to coordinates
		// Skip geocoding for now (2025-12-31) to be replaced with google maps api
		// customerLat, customerLon, errInGeocode = s.geocodeAddress(req.Address)
		if errInGeocode != nil {
			return nil, errInGeocode
		}
	}

	// Get all open outlets with their addresses
	outlets, err := customer_common.GetOutletAddress(s.db, req.BusinessID, false)
	if err != nil {
		return nil, err
	}

	var nearestOutlet *models.Outlet
	var shortestDistance float64 = math.MaxFloat64
	// Find nearest outlet
	// Skip geocoding for now (2025-12-31) to be replaced with google maps api
	stopGeocoding := true
	for _, outlet := range outlets {
		var outletLat float64
		var outletLon float64
		shouldGeocode := true
		var errInGeocode error

		// if outlet has latitude and longitude, use it
		if outlet.Latitude != nil && outlet.Longitude != nil {
			outletLat = *outlet.Latitude
			outletLon = *outlet.Longitude
			shouldGeocode = false
		}
		// enter this if the outlet has no coordinates in request body
		if shouldGeocode {
			if stopGeocoding {
				continue
			}

			combinedAddress := customer_common.CombineAddress(outlet.Address)
			outletLat, outletLon, errInGeocode = s.geocodeAddress(combinedAddress)
			if errInGeocode != nil {
				if errInGeocode.Error() == permissionDeniedError.Error() {
					stopGeocoding = true
				}
				continue // Skip outlets with invalid addresses
			}
			outlet.Latitude = &outletLat
			outlet.Longitude = &outletLon
			s.db.Save(&outlet)
		}
		distance := s.calculateDistance(customerLat, customerLon, outletLat, outletLon)
		if distance < shortestDistance && distance <= largestAllowedDistance {
			shortestDistance = distance
			nearestOutlet = &outlet
		}
	}

	if len(outlets) == 0 {
		return nil, fmt.Errorf("no outlets found")
	}
	if nearestOutlet == nil {
		nearestOutlet = closestOutletBasedOnAddressString(outlets, req.Address)
	}

	distance := Distance{
		Value: shortestDistance,
		Unit:  "km",
	}

	return &GetNearestOutletResponse{
		Message:  "Nearest outlet fetched successfully",
		Outlet:   nearestOutlet,
		Distance: distance,
	}, nil
}

// API to search and get outlet based on search key
//
//encore:api auth method=POST path=/api/customers/outlets/search
func (s *Service) SearchOutlet(ctx context.Context, req *SearchOutletRequest) (*SearchOutletResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	outlets, err := customer_common.GetOutletBySearchKey(s.db, req.SearchKey)
	if err != nil {
		return nil, err
	}

	return &SearchOutletResponse{
		Message: "Outlets fetched successfully",
		Outlets: outlets,
	}, nil
}

// API to get surrounding outlets based on user latitude and longitude
//
//encore:api auth method=POST path=/api/customers/outlets/surrounding
func (s *Service) GetSurroundingOutlets(ctx context.Context, req *GetSurroundingOutletsRequest) (*GetSurroundingOutletsResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	outlets, err := customer_common.GetAllOutletsBasedOnBusinessID(s.db, req.BusinessID, models.OutletStatusOpen)
	if err != nil {
		return nil, err
	}

	surroundingOutlets, err := s.checkSurroundingOutlets(outlets, req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		return nil, err
	}

	return &GetSurroundingOutletsResponse{
		Message: "Surrounding outlets fetched successfully",
		Outlets: surroundingOutlets,
	}, nil
}

// function to check surrounding outlets
func (s *Service) checkSurroundingOutlets(
	outlets []models.Outlet,
	userLatitude float64,
	userLongitude float64,
	radius float64,
) ([]models.Outlet, error) {
	var surroundingOutlets []models.Outlet
	// Skip geocoding for now (2025-12-31) to be replaced with google maps api
	stopGeocoding := true
	for _, outlet := range outlets {
		var outletLat float64
		var outletLon float64
		shouldGeocode := true
		if outlet.Latitude != nil {
			outletLat = *outlet.Latitude
			shouldGeocode = false
		}
		if outlet.Longitude != nil {
			outletLon = *outlet.Longitude
			shouldGeocode = false
		}
		// enter this if the outlet has no coordinates in db
		if shouldGeocode {
			if stopGeocoding {
				continue
			}
			combinedAddress := customer_common.CombineAddress(outlet.Address)
			outletLatitude, outletLongitude, err := s.geocodeAddress(combinedAddress)
			if err != nil {
				if err.Error() == permissionDeniedError.Error() {
					stopGeocoding = true
				}
				continue // Skip outlets with invalid addresses
			}
			outletLat = outletLatitude
			outletLon = outletLongitude
		}
		distance := s.calculateDistance(userLatitude, userLongitude, outletLat, outletLon)
		if distance <= radius {
			// overwrite the outlet coordinates with the coordinates from the geocode
			if shouldGeocode {
				outlet.Latitude = &outletLat
				outlet.Longitude = &outletLon
			}
			surroundingOutlets = append(surroundingOutlets, outlet)
		}
	}
	return surroundingOutlets, nil
}

// API to get recent visited outlets
//
//encore:api auth method=GET path=/api/customers/outlets/recent/visited
func (s *Service) GetRecentVisitedOutlets(ctx context.Context) (*GetRecentVisitedOutletsResponse, error) {
	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("user", user)

	var visitedOutlets []uuid.UUID
	result := s.db.Model(&models.Order{}).
		Where("customer_id = ?", user.ID).
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Pluck("outlet_id", &visitedOutlets)
	if result.Error != nil {
		return nil, result.Error
	}
	var outlets []models.Outlet
	result = s.db.Model(&models.Outlet{}).Where("id IN (?)", visitedOutlets).Find(&outlets)
	if result.Error != nil {
		return nil, result.Error
	}
	return &GetRecentVisitedOutletsResponse{
		Message: "Recent visited outlets fetched successfully",
		Outlets: outlets,
	}, nil
}

// API to deactivate customer account
//
//encore:api auth method=POST path=/api/customers/deactivate/account
func (s *Service) DeactivateAccount(ctx context.Context, req *DeactivateAccountRequest) error {

	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return err
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", user.ID).First(&customer).Error; err != nil {
		return err
	}

	if !customer.IsActive {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Your account has already been deactivated.",
		}
	}

	var business models.Business
	if err := s.db.Where("id = ?", customer.BusinessID).First(&business).Error; err != nil {
		return err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	customer.IsActive = false
	if err := trx.Save(&customer).Error; err != nil {
		trx.Rollback()
		return err
	}

	if err := aws_ses.SendDeactivateAccountEmail(customer, business.Name); err != nil {
		trx.Rollback()
		return err
	}

	if err := trx.Commit().Error; err != nil {
		trx.Rollback()
		return err
	}

	feedbackQuestion, err := s.getLeavingReasonFeedbackQuestion(business.ID)
	if err != nil {
		return nil
	}

	var customerFeedback models.Feedback

	customerFeedback.CustomerID = customer.ID
	customerFeedback.FeedbackQuestionID = feedbackQuestion.ID
	customerFeedback.Comment = req.Comment

	err = s.db.Create(&customerFeedback).Error
	if err != nil {
		return err
	}

	return nil
}

// API to reactivate customer account
//
//encore:api public method=GET path=/api/account/reactivate/:token
func (s *Service) ReactivateAccount(ctx context.Context, token string) (*common.BasicResponse, error) {
	tokenInfo, err := common.DecodeToken(token)
	if err != nil {
		return nil, err
	}

	if tokenInfo.TokenExpiry.Before(time.Now()) {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Token expired",
		}
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", tokenInfo.UserID).First(&customer).Error; err != nil {
		return nil, err
	}

	var business models.Business
	if err := s.db.Where("id = ?", customer.BusinessID).First(&business).Error; err != nil {
		return nil, err
	}

	customer.IsActive = true
	if err := s.db.Save(&customer).Error; err != nil {
		return nil, err
	}

	aws_ses.SendReactivateAccountEmail(customer, business.Name)

	return &common.BasicResponse{
		Message: "Your account has been reactivated",
	}, nil

}

func (s *Service) getLeavingReasonFeedbackQuestion(business_id uuid.UUID) (*models.FeedbackQuestion, error) {
	var feedbackQuestion models.FeedbackQuestion
	if err := s.db.Where("business_id = ? AND section = ? AND question = ?", business_id, "service", models.LeavingReasonFeedbackQuestion).First(&feedbackQuestion).Error; err != nil {

		var count int64
		s.db.Model(&models.FeedbackQuestion{}).Where("business_id = ?", business_id).Count(&count)

		createFeedbackQuestion := models.FeedbackQuestion{
			BusinessID: business_id,
			Section:    "service",
			Question:   models.LeavingReasonFeedbackQuestion,
			IsActive:   true,
			SortOrder:  int(count) + 1,
		}
		s.db.Create(&createFeedbackQuestion)

		return &createFeedbackQuestion, nil
	}

	return &feedbackQuestion, nil
}

func closestOutletBasedOnAddressString(outlets []models.Outlet, address string) *models.Outlet {

	// If no outlets or empty address, return nil
	if len(outlets) == 0 || address == "" {
		return nil
	}

	// We'll use simple string comparison (Levenshtein distance)
	// to compare the input address with each outlet's address and find the most similar one.
	var closest *models.Outlet
	var minDistance int = -1
	var bestScore = 0

	for i, outlet := range outlets {
		score := 0
		// Check common address fields as score
		if outlet.Address.Country != "" {
			if strings.Contains(strings.ToLower(address), strings.ToLower(outlet.Address.Country)) {
				score += 1
			}
		}
		if outlet.Address.State != "" {
			if strings.Contains(strings.ToLower(address), strings.ToLower(outlet.Address.State)) {
				score += 1
			}
		}
		if outlet.Address.City != "" {
			if strings.Contains(strings.ToLower(address), strings.ToLower(outlet.Address.City)) {
				score += 1
			}
		}
		if outlet.Address.PostalCode != "" {
			if strings.Contains(strings.ToLower(address), strings.ToLower(outlet.Address.PostalCode)) {
				score += 1
			}
		}

		if score > bestScore {
			bestScore = score
			closest = &outlets[i]
			log.Println("outlet score", score)
			continue
		}

		// Combine the outlet address fields into a single address string
		outletStr := customer_common.CombineAddress(outlet.Address)

		distance := common.LevenshteinDistance(address, outletStr)
		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			closest = &outlets[i]
		}
	}

	return closest
}
