package customer_auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"encore.app/auth_service"
	"encore.app/aws_ses"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/customers/customer_common"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/messaging"
	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db          *gorm.DB
	redisClient *redis.Client
}

var _ = cron.NewJob("delete-expired-refresh-token", cron.JobConfig{
	Every:    1 * cron.Hour,
	Endpoint: PrivateDeleteExpiredRefreshTokenCustomer,
})

type TokenInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	TokenExpiry time.Time `json:"token_expiry"`
}

var secretsKeys struct {
	jwtSecretKey string
}

type CheckTokenRequest struct {
	Token string `json:"token" valid:"required~Token is required"`
}

type CheckTokenResponse struct {
	Response int       `json:"response"`
	Message  string    `json:"message"`
	UserID   uuid.UUID `json:"user_id"`
}

type RenewJWTTokenParams struct {
	UserID       string  `json:"user_id"`
	RefreshToken string  `json:"refresh_token"`
	DeviceID     *string `json:"device_id"`
	FCMToken     *string `json:"fcm_token"`
}

type RenewJWTTokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	Token        string `json:"token"`
	UserID       string `json:"user_id"`
}

type ForgetPasswordValidationRequest struct {
	Email string `json:"email" valid:"required~Email is required,email~Invalid email format"`
}

type ForgetPasswordValidationResponse struct {
	Message string `json:"message"`
}

type PasswordResetRequest struct {
	Email              string    `json:"email" valid:"required~Email is required,email~Invalid email format"`
	BusinessID         uuid.UUID `json:"business_id" valid:"required~Business ID is required"`
	ForgetPassAuthCode string    `json:"forget_pass_auth_code" valid:"required~Forget pass auth code is required"`
	NewPassword        string    `json:"new_password" valid:"required~New password is required"`
	ConfirmNewPassword string    `json:"confirm_new_password" valid:"required~Confirm new password is required"`
}

type PasswordResetResponse struct {
	Message string `json:"message"`
}

type DeleteExpiredRefreshTokenResponse struct {
	Message string `json:"message"`
}

type LoginParams struct {
	Email           string  `json:"email"`
	Password        string  `json:"password"`
	GoogleAuthToken *string `json:"google_auth_token"`
	AppleAuthToken  *string `json:"apple_auth_token"`
	PhoneNumber     *string `json:"phone_number"`
	OTP             *string `json:"otp"` // OTP is a 6 digit code sent to the customer's phone number
	// SmsToken        *string   `json:"sms_token"`
	DeviceID   string    `json:"device_id" valid:"required~Device ID is required"`
	FCMToken   *string   `json:"fcm_token"`
	BusinessID uuid.UUID `json:"business_id"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
}

type CheckTokenTestResponse struct {
	Valid bool `json:"valid"`
}

type SendVerificationCodeRequest struct {
	CustomerID  string                 `json:"customer_id"`
	Email       string                 `json:"email" valid:"required~Email is required,email~Invalid email format"`
	PhoneNumber string                 `json:"phone_number" valid:"required~Phone number is required"`
	Action      models.EmailActionType `json:"action" valid:"required~Action is required"`
	BusinessID  uuid.UUID              `json:"business_id"`
}

type PhoneOTPRequest struct {
	PhoneNumber string    `json:"phone_number"`
	BusinessID  uuid.UUID `json:"business_id"`
	OTP         string    `json:"otp"`
}

type SendVerificationCodeResponse struct {
	Message string `json:"message"`
}

type VerificationCodeInfo struct {
	Code        string    `redis:"code"`
	ExpiredAt   time.Time `redis:"expired_at"`
	UserID      uuid.UUID `redis:"user_id"`
	Email       string    `redis:"email"`
	PhoneNumber string    `redis:"phone_number"`
}

type VerifyVerificationCodeRequest struct {
	CustomerID  string    `json:"customer_id"`
	BusinessID  uuid.UUID `json:"business_id"`
	InputCode   string    `json:"input_code"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phone_number"`
	Action      string    `json:"action"`
}

type VerifyVerificationCodeResponse struct {
	Message            string `json:"message"`
	IsVerified         bool   `json:"is_verified"`
	ForgetPassAuthCode string `json:"forget_pass_auth_code"`
}

type VerifyForgetPasswordCodeRequest struct {
	Email      string    `json:"email" valid:"required~Email is required,email~Invalid email format"`
	BusinessID uuid.UUID `json:"business_id"`
	InputCode  string    `json:"input_code"`
}

var VerificationCodeExpiredTime = 10 * time.Minute

// initService initializes the auth service.
func initService() (*Service, error) {
	common.LoadEnv()

	secretsKeys.jwtSecretKey = os.Getenv("JWT_SECRET_KEY")

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// init redis client
	address := "localhost:6379"
	if !common.IsLocal() {
		address = "redis:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: address,
		DB:   0,
	})

	// test redis connection
	_, err = redisClient.Ping().Result()
	if err != nil {
		log.Println("**")
		log.Println("**")
		log.Println("  Redis is not running!")
		log.Println("  Please follow the steps in 'sante-backend/README_CONFIG.md' and run redis.")
		log.Println("**")
		log.Println("**")
	}

	return &Service{db: db, redisClient: redisClient}, nil
}

// Login authenticates a customer user and returns a JWT token
//
//encore:api public method=POST path=/api/customers/auth/login
func (s *Service) CustomerLogin(ctx context.Context, params *LoginParams) (*LoginResponse, error) {
	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: err.Error(),
		}
	}

	disableEmailCheck := false
	disablePasswordCheck := false

	if params.GoogleAuthToken != nil && *params.GoogleAuthToken != "" {
		email, err := s.ValidateGoogleAuthToken(ctx, *params.GoogleAuthToken, params.BusinessID)
		if err != nil {
			return nil, err
		}
		params.Email = strings.ToLower(email)
		disablePasswordCheck = true
	}

	if params.AppleAuthToken != nil && *params.AppleAuthToken != "" {
		email, err := ValidateAppleAuthToken(*params.AppleAuthToken)
		if err != nil {
			return nil, err
		}
		params.Email = strings.ToLower(email)
		disablePasswordCheck = true
	}

	if params.OTP != nil && *params.OTP != "" {
		if err := s.VerifyPhoneOTP(ctx, &PhoneOTPRequest{PhoneNumber: *params.PhoneNumber, BusinessID: params.BusinessID, OTP: *params.OTP}); err != nil {
			return nil, err
		}

		disableEmailCheck = true
		disablePasswordCheck = true
	}

	// if params.SmsToken != nil && *params.SmsToken != "" {
	// 	phoneNumber, err := firebase.VerifyFirebaseToken(ctx, *params.SmsToken)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	params.PhoneNumber = phoneNumber
	// 	disableEmailCheck = true
	// 	disablePasswordCheck = true
	// }

	var customer models.Customer

	if !disableEmailCheck {
		if params.Email == "" || !govalidator.IsEmail(params.Email) {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "Invalid email format",
			}
		}

		if err := s.db.Model(&models.Customer{}).Where("email = ? AND business_id = ?", params.Email, params.BusinessID).First(&customer).Error; err != nil {
			return nil, errors.New("This email has not yet been registered.")
		}
	} else {
		possiblePhoneNumbers := common.GetPossiblePhoneNumbers(*params.PhoneNumber)

		if err := s.db.Model(&models.Customer{}).Where("phone_number IN (?) AND business_id = ?", possiblePhoneNumbers, params.BusinessID).First(&customer).Error; err != nil {
			return nil, errors.New("This phone number has not yet been registered.")
		}
	}

	/* UAT BYPASS: Allow login without verification for now
	if !customer.EmailVerified {
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Email is not verified",
		}
	}
	*/

	if !customer.IsActive {
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Your account has already been deactivated.",
		}
	}

	/* activityLog := &models.ActivityLog{
		Activity:       constants.LOG_ACTION_LOGIN,
		Status:         constants.LOG_STATUS_SUCCESS,
		ActionBy:       user.FirstName + " " + user.Surname,
		ActionByUserID: user.ID,
	} */

	// Compare passwords
	if !disablePasswordCheck {
		if params.Password == "" {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "Password is required",
			}
		}

		err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(params.Password))
		if err != nil {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "invalid credentials",
			}
		}
	}

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": customer.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		//"exp":        time.Now().Add(30 * time.Second).Unix(),
		"login_type": "customer",
	})

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(secretsKeys.jwtSecretKey))
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate token",
		}
	}

	// Generate refresh token
	tokens, err := auth_service.GenerateRefreshTokenForUser()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate refresh token",
		}
	}

	userToken := &models.CustomerToken{
		CustomerID:   customer.ID,
		RefreshToken: tokens.HashedToken,
		ExpiredAt:    time.Now().Add(30 * 24 * time.Hour),
		DeviceID:     &params.DeviceID,
		FCMToken:     params.FCMToken,
	}

	// save refresh token to database
	var existingUserToken models.CustomerToken
	var result *gorm.DB
	if params.DeviceID != "" {
		result = s.db.Where("customer_id = ? AND device_id = ?", customer.ID, params.DeviceID).First(&existingUserToken)
	} else {
		result = s.db.Where("customer_id = ? AND device_id IS NULL", customer.ID).First(&existingUserToken)
		if result.Error != nil {
			result = s.db.Where("customer_id = ?", customer.ID).First(&existingUserToken)
		}
	}
	if result.Error != nil {
		fmt.Println("no existing refresh token")
		result = s.db.Create(&userToken)
		if result.Error != nil {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to save new refresh token",
			}
		}
	} else {
		// override existing refresh token
		existingUserToken.RefreshToken = userToken.RefreshToken
		existingUserToken.ExpiredAt = userToken.ExpiredAt
		if params.DeviceID != "" {
			existingUserToken.DeviceID = userToken.DeviceID
		}
		if params.FCMToken != nil && *params.FCMToken != "" {
			existingUserToken.FCMToken = params.FCMToken
		}
		result = s.db.Save(&existingUserToken)
		if result.Error != nil {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to override refresh token",
			}
		}
	}

	//s.db.Create(activityLog)

	return &LoginResponse{
		Token:        signedToken,
		RefreshToken: tokens.PlainRefreshToken,
		UserID:       customer.ID.String(),
	}, nil
}

// generate new refresh token
//
//encore:api public method=POST path=/api/customers/auth/token/renew
func (s *Service) RenewJWTTokenCustomer(ctx context.Context, req *RenewJWTTokenParams) (*RenewJWTTokenResponse, error) {
	trx := s.db.Begin()
	defer trx.Rollback()
	if trx.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to start transaction",
		}
	}
	var existingCustomerToken []models.CustomerToken
	var result *gorm.DB

	if req.DeviceID != nil && *req.DeviceID != "" {
		// New logic: Look for device-specific token
		result = trx.Where("customer_id = ? AND device_id = ?", req.UserID, *req.DeviceID).
			Where("expired_at > ?", time.Now()).
			Order("expired_at DESC").
			Find(&existingCustomerToken)
	} else {
		// Old logic: Look for legacy token (device_id IS NULL) OR try to find any token for this user
		result = trx.Where("customer_id = ? AND device_id IS NULL", req.UserID).
			Where("expired_at > ?", time.Now()).
			Order("expired_at DESC").
			Find(&existingCustomerToken)

		// If no legacy token found, try to find any token for this user (for existing data)
		if result.Error != nil {
			result = trx.Where("customer_id = ?", req.UserID).
				Where("expired_at > ?", time.Now()).
				Order("expired_at DESC").
				Find(&existingCustomerToken)
		}
	}
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Refresh token not found",
		}
	}

	// loop through existingCustomerToken and find the one that has the same refresh token
	var isFoundSameRefreshToken bool = false
	var existingEquivalentToken models.CustomerToken
	for _, existingToken := range existingCustomerToken {
		err := bcrypt.CompareHashAndPassword([]byte(existingToken.RefreshToken), []byte(req.RefreshToken))
		// this mean not the same refresh token
		if err != nil {
			continue
		} else {
			// this mean the same refresh token
			isFoundSameRefreshToken = true
			existingEquivalentToken = existingToken
			break
		}
	}

	if !isFoundSameRefreshToken {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "Invalid refresh token",
		}
	}

	var customer models.Customer
	result = trx.First(&customer, "id = ?", req.UserID)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Unauthenticated,
			Message: "invalid user id",
		}
	}

	customerTokens, err := auth_service.GenerateRefreshTokenForUser()
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate new refresh token",
		}
	}

	//existingEquivalentToken.RefreshToken = customerTokens.HashedToken
	// this is use to change existing equivalent token to expire in 10 minutes so that if any user edge issue such as network issue cause latest refresh token is not captured, it will still be valid for 10 minutes
	existingEquivalentToken.ExpiredAt = time.Now().Add(10 * time.Minute)

	result = trx.Save(&existingEquivalentToken)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to save equivalent refresh token",
		}
	}

	customerTokenLatest := &models.CustomerToken{
		CustomerID:   customer.ID,
		RefreshToken: customerTokens.HashedToken,
		ExpiredAt:    time.Now().Add(30 * 24 * time.Hour),
		DeviceID:     req.DeviceID,
		FCMToken:     req.FCMToken,
	}

	result = trx.Create(&customerTokenLatest)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to save latest refresh token",
		}
	}
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": customer.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		//"exp":        time.Now().Add(30 * time.Second).Unix(),
		"login_type": "customer",
	})

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(secretsKeys.jwtSecretKey))
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate token",
		}
	}
	fmt.Println("signedToken", signedToken)

	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to commit transaction",
		}
	}

	return &RenewJWTTokenResponse{
		RefreshToken: customerTokens.PlainRefreshToken,
		Token:        signedToken,
		UserID:       customer.ID.String(),
	}, nil
}

// Forget password API (before validation)
//
//encore:api public method=POST path=/api/customers/auth/password/forget/validation/send-code
func (s *Service) ForgetPasswordSendCodeValidation(ctx context.Context, req *ForgetPasswordValidationRequest) (*ForgetPasswordValidationResponse, error) {
	if req.Email == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Email is required",
		}
	}
	var customer models.Customer
	result := s.db.First(&customer, "email = ?", req.Email)
	if result.Error != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Email is not registered",
		}
	}
	_, err := SendVerificationCode(ctx, &SendVerificationCodeRequest{
		CustomerID:  customer.ID.String(),
		Email:       customer.Email,
		PhoneNumber: customer.PhoneNumber,
		Action:      models.EmailActionTypeForgetPassword,
	})
	if err != nil {
		return nil, err
	}

	return &ForgetPasswordValidationResponse{
		Message: "Email is valid",
	}, nil
}

// Reset password API (after validation)
//
//encore:api public method=POST path=/api/customers/auth/password/reset
func (s *Service) PasswordReset(ctx context.Context, req *PasswordResetRequest) (*PasswordResetResponse, error) {
	if req.NewPassword == "" || req.ConfirmNewPassword == "" || req.ForgetPassAuthCode == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "New password, confirm new password, and forget pass auth code are required",
		}
	}
	if req.NewPassword != req.ConfirmNewPassword {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "New password and confirm new password do not match",
		}
	}

	trx := s.db.Begin()
	defer trx.Rollback()
	if trx.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to start transaction",
		}
	}

	// get customer data
	customer, err := customer_common.GetCustomerDataFromEmailAndBusinessID(s.db, req.Email, req.BusinessID)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Email is not registered",
		}
	}

	// verification of verification code
	isVerified, err := s.VerificationOfVerificationCode(customer.ID.String(), models.EmailActionTypeForgetPassword, req.ForgetPassAuthCode)
	if err != nil {
		return nil, err
	}
	if !isVerified {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Verification code is invalid",
		}
	}

	encryptedPass, err := common.EncodePassword(req.NewPassword)
	if err != nil {
		trx.Rollback()
		return nil, err
	}
	customer.Password = encryptedPass
	result := trx.Save(&customer)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to save new password",
		}
	}

	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to commit transaction",
		}
	}
	// delete verification code from redis (even failed also let it continue)
	s.DeleteVerificationCodeFromRedis(customer.ID.String(), models.EmailActionTypeForgetPassword)

	return &PasswordResetResponse{
		Message: "Password changed successfully",
	}, nil
}

// delete all expired refresh token for customer
//
//encore:api public method=POST path=/api/customers/auth/refreshtoken/expired/delete/all
func (s *Service) DeleteExpiredRefreshTokenCustomer(ctx context.Context) (*DeleteExpiredRefreshTokenResponse, error) {
	trx := s.db.Begin()
	defer trx.Rollback()
	if trx.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to start transaction",
		}
	}

	timeNow := time.Now().Local()
	fmt.Println("timeNow", timeNow)

	//result := trx.Model(&models.CustomerToken{}).Where("expired_at < ?", timeNow).Unscoped().Delete(&models.CustomerToken{})
	result := trx.Model(&models.CustomerToken{}).Where("expired_at < ?", timeNow).Delete(&models.CustomerToken{})
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to delete expired refresh token",
		}
	}

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to commit transaction",
		}
	}

	return &DeleteExpiredRefreshTokenResponse{
		Message: "Expired refresh token deleted successfully",
	}, nil
}

// delete all expired refresh token for customer
//
//encore:api private
func (s *Service) PrivateDeleteExpiredRefreshTokenCustomer(ctx context.Context) (*DeleteExpiredRefreshTokenResponse, error) {
	trx := s.db.Begin()
	defer trx.Rollback()
	if trx.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to start transaction",
		}
	}

	timeNow := time.Now().Local()
	fmt.Println("timeNow", timeNow)

	//result := trx.Model(&models.CustomerToken{}).Where("expired_at < ?", timeNow).Unscoped().Delete(&models.CustomerToken{})
	result := trx.Model(&models.CustomerToken{}).Where("expired_at < ?", timeNow).Delete(&models.CustomerToken{})
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to delete expired refresh token",
		}
	}

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to commit transaction",
		}
	}

	return &DeleteExpiredRefreshTokenResponse{
		Message: "Expired refresh token deleted successfully",
	}, nil
}

// CheckToken validates if the current session is authenticated
//
//encore:api public method=POST path=/api/customers/auth/check-token
func (s *Service) CheckToken(ctx context.Context, req *CheckTokenRequest) (*CheckTokenResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &CheckTokenResponse{
			Response: http.StatusBadRequest,
			Message:  err.Error(),
			UserID:   uuid.UUID{},
		}, nil
	}

	fmt.Println("check token")
	token, err := common.DecodeToken(req.Token)
	if err != nil {
		return &CheckTokenResponse{
			Response: http.StatusUnauthorized,
			Message:  "Token is invalid",
			UserID:   uuid.UUID{},
		}, nil
	}
	fmt.Println("token", token.UserID)
	fmt.Println("token expiry", token.TokenExpiry)

	// check if token is expired
	if token.TokenExpiry.Before(time.Now()) {
		return &CheckTokenResponse{
			Response: http.StatusUnauthorized,
			Message:  "Token is expired",
			UserID:   uuid.UUID{},
		}, nil
	}

	return &CheckTokenResponse{
		Response: http.StatusOK,
		Message:  "Token is valid",
		UserID:   token.UserID,
	}, nil
}

// another check token test using auth service
//
//encore:api auth method=GET path=/api/customers/auth/check-customer-token
func (s *Service) CheckCustomerToken(ctx context.Context) (*CheckTokenTestResponse, error) {
	val := CheckTokenTestResponse{
		Valid: true,
	}
	return &val, nil
}

// API to send 6 digit code to customer email
//
//encore:api public method=POST path=/api/customers/send-verification-code
func (s *Service) SendVerificationCode(ctx context.Context, req *SendVerificationCodeRequest) (*SendVerificationCodeResponse, error) {
	// normally is scenario forget password to enter this if statement.
	// as forget password assume that customer ID is not provided from client side.
	if req.CustomerID == "" {
		if req.Action == models.EmailActionTypeForgetPassword {
			//get customer id from email
			customer, err := customer_common.GetCustomerDataFromEmailAndBusinessID(s.db, req.Email, req.BusinessID)
			if err != nil {
				return nil, &errs.Error{
					Code:    errs.InvalidArgument,
					Message: "failed to get customer",
				}
			}
			req.CustomerID = customer.ID.String()
			req.PhoneNumber = customer.PhoneNumber
			// otherwise is scenario register to enter this if customer ID is not provided.
		} else {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "CustomerID cannot be empty",
			}
		}
	}
	var customer models.Customer
	query := s.db.Model(&models.Customer{}).
		Where("id = ?", req.CustomerID).First(&customer)
	if query.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get customer",
		}
	}
	var business models.Business
	query = s.db.Model(&models.Business{}).Where("id = ?", customer.BusinessID).First(&business)
	if query.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get business",
		}
	}

	// get key by action
	key, err := customer_common.GetKeyByAction(req.CustomerID, req.Action)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid action",
		}
	}

	linkOrCode, err := VerificationMethod(s.db, req.CustomerID, req.Action)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid action",
		}
	}
	// save code to redis
	fields := map[string]interface{}{
		"code":         linkOrCode,
		"expired_at":   strconv.FormatInt(time.Now().Add(VerificationCodeExpiredTime).Unix(), 10),
		"email":        req.Email,
		"phone_number": req.PhoneNumber,
		"user_id":      req.CustomerID,
	}

	result := s.redisClient.HMSet(key, fields)
	if result.Err() != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to save verification code to redis when generate verification code",
		}
	}

	if req.Action == models.EmailActionTypeForgetPassword {
		err = aws_ses.SendPasswordResetEmail(
			req.Email,
			customer.FirstName+" "+customer.LastName,
			linkOrCode,
			business.Name,
		)
		if err != nil {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to send password reset email",
			}
		}
	} else {
		err = aws_ses.SendEmailVerificationCode(
			req.Email,
			customer.FirstName+" "+customer.LastName,
			linkOrCode,
			business.Name,
		)
		if err != nil {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to send email verification code",
			}
		}
	}

	printVerificationCodes(s.redisClient)

	return &SendVerificationCodeResponse{
		Message: "Verification code sent to email",
	}, nil
}

// API to verify 6 digit code (for register)
//
//encore:api public method=POST path=/api/customers/verify-verification-code
func (s *Service) VerifyVerificationCode(ctx context.Context, req *VerifyVerificationCodeRequest) (*VerifyVerificationCodeResponse, error) {
	if req.CustomerID == "" {
		return nil, errors.New("CustomerID cannot be empty")
	}
	if req.BusinessID == uuid.Nil {
		return nil, errors.New("BusinessID cannot be empty")
	}

	key, err := customer_common.GetKeyByAction(req.CustomerID, models.EmailActionType(req.Action))
	if err != nil {
		return nil, err
	}
	fields, err := s.redisClient.HGetAll(key).Result()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get verification code from redis",
		}
	}
	// check code input is same as the code in redis
	if fields["code"] != req.InputCode {
		printVerificationCodes(s.redisClient)
		return &VerifyVerificationCodeResponse{
			Message:    "Verification code is invalid",
			IsVerified: false,
		}, nil
	}
	// check code is expired
	expiredAtUnix, err := strconv.ParseInt(fields["expired_at"], 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to parse expired at",
		}
	}
	expiredAt := time.Unix(expiredAtUnix, 0)
	if time.Now().After(expiredAt) {
		printVerificationCodes(s.redisClient)
		return &VerifyVerificationCodeResponse{
			Message:    "Verification code is expired",
			IsVerified: false,
		}, nil
	}

	trx := s.db.Begin()
	defer trx.Rollback()
	if trx.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to start transaction",
		}
	}

	var customer models.Customer
	result := trx.Model(&models.Customer{}).Where("id = ? AND business_id = ?", req.CustomerID, req.BusinessID).First(&customer)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get customer",
		}
	}
	customer.EmailVerified = true
	result = trx.Save(&customer)
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to update customer email verified",
		}
	}
	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to commit transaction",
		}
	}

	// delete verification code from redis
	err = s.redisClient.Del(key).Err()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to delete verification code from redis",
		}
	}

	printVerificationCodes(s.redisClient)

	return &VerifyVerificationCodeResponse{
		Message:    "Verification code verified",
		IsVerified: true,
	}, nil
}

// API to verify 6 digit code (for forget password)
//
//encore:api public method=POST path=/api/customers/auth/password/forget/validation/code
func (s *Service) VerifyForgetPasswordCode(ctx context.Context, req *VerifyForgetPasswordCodeRequest) (*VerifyVerificationCodeResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if req.BusinessID == uuid.Nil {
		return nil, errors.New("businessID cannot be empty")
	}
	if req.InputCode == "" {
		return nil, errors.New("input code cannot be empty")
	}

	action := models.EmailActionTypeForgetPassword
	var customer models.Customer
	result := s.db.Model(&models.Customer{}).Where("email = ? AND business_id = ?", req.Email, req.BusinessID).First(&customer)
	if result.Error != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get customer",
		}
	}
	key, err := customer_common.GetKeyByAction(customer.ID.String(), action)
	if err != nil {
		return nil, err
	}
	fields, err := s.redisClient.HGetAll(key).Result()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get verification code from redis",
		}
	}
	// check code input is same as the code in redis
	if fields["code"] != req.InputCode {
		printVerificationCodes(s.redisClient)
		return &VerifyVerificationCodeResponse{
			Message:    "Verification code is invalid",
			IsVerified: false,
		}, nil
	}
	// check code is expired
	expiredAtUnix, err := strconv.ParseInt(fields["expired_at"], 10, 64)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to parse expired at",
		}
	}
	expiredAt := time.Unix(expiredAtUnix, 0)
	if time.Now().After(expiredAt) {
		printVerificationCodes(s.redisClient)
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Verification code is expired",
		}
	}

	// delete verification code from redis
	err = s.redisClient.Del(key).Err()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to delete verification code from redis",
		}
	}

	// generate new random 6 digit code
	code, err := GenerateRandom6DigitCode()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate new random 6 digit code",
		}
	}
	// save new code to redis
	newFields := map[string]interface{}{
		"code": code,
		//"expired_at":   time.Now().Add(30 * time.Minute).Unix(),
		"expired_at":   strconv.FormatInt(time.Now().Add(30*time.Minute).Unix(), 10),
		"email":        req.Email,
		"phone_number": customer.PhoneNumber,
		"user_id":      customer.ID.String(),
	}

	err = s.redisClient.HMSet(key, newFields).Err()
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to save verification code to redis",
		}
	}

	printVerificationCodes(s.redisClient)

	return &VerifyVerificationCodeResponse{
		Message:            "Verification code verified",
		IsVerified:         true,
		ForgetPassAuthCode: code,
	}, nil
}

//encore:api public method=POST path=/api/customers/send-phone-otp
func (s *Service) SendPhoneOTP(ctx context.Context, req *PhoneOTPRequest) error {
	if req.PhoneNumber == "" {
		return errors.New("phone number cannot be empty")
	}
	if req.BusinessID == uuid.Nil {
		return errors.New("businessID cannot be empty")
	}

	business := models.Business{}
	if err := s.db.Where("id = ?", req.BusinessID).First(&business).Error; err != nil {
		return errors.New("business not found")
	}

	possiblePhoneNumbers := common.GetPossiblePhoneNumbers(req.PhoneNumber)

	customer := models.Customer{}
	if err := s.db.Model(&models.Customer{}).Where("phone_number IN (?) AND business_id = ?", possiblePhoneNumbers, req.BusinessID).First(&customer).Error; err != nil {
		return errors.New("This phone number has not yet been registered.")
	}

	phoneNumber := common.SanitizePhoneNumber(req.PhoneNumber)
	phoneNumber = common.ClearCountryCodeFromPhoneNumber(phoneNumber)

	code, err := GenerateRandom6DigitCode()
	if err != nil {
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate new random 6 digit code",
		}
	}

	key := "phone_otp:" + business.ID.String() + ":" + phoneNumber
	minutes := 1 // 1 minute
	expiry := time.Duration(minutes) * time.Minute
	s.redisClient.Set(key, code, expiry)

	errorMessage := "SMS Anchor Credentials not found"
	var userData models.SystemData
	if err := s.db.Where("info_type = ?", string(constants.SystemDataInfoTypeAnchorSMSUser)).First(&userData).Error; err != nil {
		return fmt.Errorf(errorMessage)
	}
	var passData models.SystemData
	if err := s.db.Where("info_type = ?", string(constants.SystemDataInfoTypeAnchorSMSPassword)).First(&passData).Error; err != nil {
		return fmt.Errorf(errorMessage)
	}

	ref := "UqqU/KwtfnD"
	return messaging.SendSMS(phoneNumber, "Pretzley: Your code is "+code+".\nThis OTP is valid for "+strconv.Itoa(minutes)+" min.\nDO NOT SHARE THIS OTP. Ref: "+ref, userData.InfoValue, passData.InfoValue)
}

func (s *Service) VerifyPhoneOTP(ctx context.Context, req *PhoneOTPRequest) error {
	if req.PhoneNumber == "" {
		return errors.New("phone number cannot be empty")
	}
	if req.BusinessID == uuid.Nil {
		return errors.New("businessID cannot be empty")
	}
	if req.OTP == "" {
		return errors.New("OTP cannot be empty")
	}

	phoneNumber := common.SanitizePhoneNumber(req.PhoneNumber)
	phoneNumber = common.ClearCountryCodeFromPhoneNumber(phoneNumber)
	key := "phone_otp:" + req.BusinessID.String() + ":" + phoneNumber
	val, err := s.redisClient.Get(key).Result()
	if err != nil {
		return errors.New("OTP is expired")
	}
	if val != req.OTP {
		return errors.New("OTP code is invalid")
	}

	s.redisClient.Del(key)
	return nil
}

// test print verification codes
func printVerificationCodes(redisClient *redis.Client) {
	// Get all keys that match verification code pattern
	keys, err := redisClient.Keys("verification_code*").Result()
	if err != nil {
		log.Println("failed to get verification code keys from redis", err)
		return
	}

	forgetPassKeys, err := redisClient.Keys("forget_password_code*").Result()
	if err != nil {
		log.Println("failed to get forget pass code keys from redis", err)
		return
	}

	log.Println("=== All Verification Codes ===")
	for _, key := range keys {
		log.Printf("Key: %s\n", key)
		fields, err := redisClient.HGetAll(key).Result()
		if err != nil {
			log.Println("failed to get fields from redis for key:", key, err)
			continue
		}

		// Print fields nicely
		for field, value := range fields {
			log.Printf("  %s: %s\n", field, value)
		}
		log.Println("---")
	}
	for _, key := range forgetPassKeys {
		log.Printf("Key: %s\n", key)
		fields, err := redisClient.HGetAll(key).Result()
		if err != nil {
			log.Println("failed to get fields from redis for key:", key, err)
			continue
		}
		for field, value := range fields {
			log.Printf("  %s: %s\n", field, value)
		}
		log.Println("---")
	}
	log.Println("=== End Verification Codes ===")
}

// determine send 6 digit code or link
func VerificationMethod(trx *gorm.DB, customer_id string, action models.EmailActionType) (string, error) {
	if action == models.EmailActionTypeForgetPassword {
		// not sure use which approach link or code
		//link, err := GenerateUniqueForgetPasswordLink(trx, customer_id)
		code, err := GenerateRandom6DigitCode()
		if err != nil {
			return "", err
		}
		return code, nil
	} else if action == models.EmailActionTypeVerifyEmail {
		code, err := GenerateRandom6DigitCode()
		if err != nil {
			return "", err
		}
		return code, nil
	} else {
		return "", errors.New("invalid action")
	}
}

// generate unique forget password link
func GenerateUniqueForgetPasswordLink(trx *gorm.DB, customer_id string) (string, error) {
	// generate JWT token
	expiredTime := time.Now().Add(30 * time.Minute).Unix()
	token, err := GenerateJWTToken(trx, customer_id, expiredTime, "customer_forget_password")
	if err != nil {
		return "", err
	}
	if common.IsProduction() {
		link := "https://app.sante.my/api/customers/auth/password/forget/change/" + token
		return link, nil
	} else if common.IsStaging() {
		link := "https://sante-staging.afeddigital.dev/api/customers/auth/password/forget/change/" + token
		return link, nil
	} else if common.IsLocal() {
		link := "http://localhost:3000/api/customers/auth/password/forget/change/" + token
		return link, nil
	} else {
		return "", errors.New("invalid environment")
	}
}

// generate JWT token
func GenerateJWTToken(trx *gorm.DB, customer_id string, expiredTime int64, loginType string) (string, error) {
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        customer_id,
		"exp":        expiredTime,
		"login_type": loginType,
	})

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(secretsKeys.jwtSecretKey))
	if err != nil {
		return "", &errs.Error{
			Code:    errs.Internal,
			Message: "failed to generate token",
		}
	}
	return signedToken, nil
}

// generate random 6 digit code
func GenerateRandom6DigitCode() (string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", r.Intn(1000000))
	return code, nil
}

// verification of verification code
func (s *Service) VerificationOfVerificationCode(customer_id string, action models.EmailActionType, inputCode string) (bool, error) {
	key, err := customer_common.GetKeyByAction(customer_id, action)
	if err != nil {
		return false, err
	}
	fields, err := s.redisClient.HGetAll(key).Result()
	if err != nil {
		return false, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to get verification code from redis",
		}
	}
	// check code input is same as the code in redis
	if fields["code"] != inputCode {
		printVerificationCodes(s.redisClient)
		return false, nil
	}
	// check code is expired
	expiredAtUnix, err := strconv.ParseInt(fields["expired_at"], 10, 64)
	if err != nil {
		return false, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to parse expired at",
		}
	}
	expiredAt := time.Unix(expiredAtUnix, 0)
	if time.Now().After(expiredAt) {
		printVerificationCodes(s.redisClient)
		return false, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Verification code is expired",
		}
	}
	return true, nil
}

// delete verification code from redis
func (s *Service) DeleteVerificationCodeFromRedis(customer_id string, action models.EmailActionType) error {
	key, err := customer_common.GetKeyByAction(customer_id, action)
	if err != nil {
		fmt.Println("failed to get key from redis", err)
		return err
	}
	err = s.redisClient.Del(key).Err()
	if err != nil {
		fmt.Println("failed to delete verification code from redis", err)
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to delete verification code from redis",
		}
	}
	fmt.Println("deleted verification code from redis", key)
	return nil
}

func (s *Service) ValidateGoogleAuthToken(ctx context.Context, token string, businessID uuid.UUID) (string, error) {
	var businessConfiguration models.BusinessConfiguration
	if err := s.db.First(&businessConfiguration, "business_id = ?", businessID).Error; err != nil {
		return "", errors.New("failed to get business configuration")
	}
	if businessConfiguration.GoogleAuthClientID == nil {
		return "", errors.New("google auth client id is not set for this business")
	}

	payload, err := idtoken.Validate(ctx, token, *businessConfiguration.GoogleAuthClientID)
	if err != nil {
		return "", err
	}

	email, _ := payload.Claims["email"].(string)

	return email, nil
}

func ValidateAppleAuthToken(identityToken string) (string, error) {
	// Parse token WITHOUT validating first (to extract kid)
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(identityToken, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return "", errors.New("invalid token header: missing kid")
	}

	// Fetch Apple public keys
	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var keys appleKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return "", err
	}

	var publicKey *rsa.PublicKey

	for _, key := range keys.Keys {
		if key.Kid == kid {
			nBytes, _ := base64.RawURLEncoding.DecodeString(key.N)
			eBytes, _ := base64.RawURLEncoding.DecodeString(key.E)

			n := new(big.Int).SetBytes(nBytes)
			e := new(big.Int).SetBytes(eBytes)

			publicKey = &rsa.PublicKey{
				N: n,
				E: int(e.Int64()),
			}
			break
		}
	}

	if publicKey == nil {
		return "", errors.New("unable to find matching Apple public key")
	}

	// Now validate token properly
	verifiedToken, err := jwt.Parse(identityToken, func(token *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok || !verifiedToken.Valid {
		return "", errors.New("invalid token")
	}

	// Validate issuer
	if claims["iss"] != appleIssuer {
		return "", errors.New("invalid issuer")
	}

	// Validate audience
	if claims["aud"] != common.GetBundleID() {
		return "", errors.New("invalid audience")
	}

	// Validate expiration
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return "", errors.New("invalid exp claim")
	}
	if time.Unix(int64(expFloat), 0).Before(time.Now()) {
		return "", errors.New("token expired")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", errors.New("email not present in token claims")
	}

	return email, nil
}
