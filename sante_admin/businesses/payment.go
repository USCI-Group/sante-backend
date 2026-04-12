package businesses

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/database/models"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	"gorm.io/gorm"

	_ "embed"
)

//go:embed payment_methods_membership.json
var paymentMethodsMembershipJSON []byte

//go:embed payment_methods_pos.json
var paymentMethodsPOSJSON []byte

type CreatePaymentMethodConfigurationRequest struct {
	BusinessID         *uuid.UUID             `json:"business_id"`
	PaymentMethod      models.PaymentMethod   `json:"payment_method"`
	PaymentChannel     models.PaymentChannel  `json:"payment_channel"`
	PaymentPlatform    models.PaymentPlatform `json:"payment_platform"`
	PaymentChannelCode *string                `json:"payment_channel_code"` // some payment channel like FPX has multiple codes, so we need to store the code here
	IsActive           bool                   `json:"is_active"`            // the hierachy level isActive -> isMaintenance -> isVisible
	IsMaintenance      bool                   `json:"is_maintenance"`       // the hierachy level isActive -> isMaintenance -> isVisible
	IsVisible          bool                   `json:"is_visible"`           // the hierachy level isActive -> isMaintenance -> isVisible

	// Add validation fields
	MinAmount         *float32 `json:"min_amount"`          // Minimum transaction amount
	MaxAmount         *float32 `json:"max_amount"`          // Maximum transaction amount
	ProcessingFee     *float32 `json:"processing_fee"`      // Fixed processing fee
	ProcessingFeeRate *float32 `json:"processing_fee_rate"` // Percentage processing fee (0.025 = 2.5%)
	Currency          string   `json:"currency"`            // Currency code
	SettlementDays    *int     `json:"settlement_days"`     // Days to settlement
	Priority          int      `json:"priority"`            // Display priority (higher = shown first)

	// Configuration metadata
	DisplayName string  `json:"display_name"` // Custom display name
	Description *string `json:"description"`  // Description for users
	IconURL     *string `json:"icon_url"`     // Payment method icon
	ColorCode   *string `json:"color_code"`   // Hex color for UI

	// Compliance fields
	ComplianceStatus models.ComplianceStatus `json:"compliance_status"` // active, suspended, under_review
	ComplianceNotes  *string                 `json:"compliance_notes"`
	ValidFrom        *time.Time              `json:"valid_from"`
	ValidUntil       *time.Time              `json:"valid_until"`
}

type CreatePaymentMethodConfigurationResponse struct {
	Message string `json:"message"`
}

type CreateDefaultPaymentMethodConfigurationRequest struct {
	BusinessID *uuid.UUID `json:"business_id"`
}

type CreateDefaultPaymentMethodConfigurationResponse struct {
	Message string `json:"message"`
}

// function to create payment method configuration
func CreatePaymentMethod(trx *gorm.DB, businessID uuid.UUID, req *CreatePaymentMethodConfigurationRequest) (*models.PaymentMethodConfiguration, error) {
	paymentMethodConfiguration := models.PaymentMethodConfiguration{
		BusinessID:         businessID,
		PaymentMethod:      req.PaymentMethod,
		PaymentChannel:     req.PaymentChannel,
		PaymentPlatform:    req.PaymentPlatform,
		PaymentChannelCode: req.PaymentChannelCode,
		IsActive:           req.IsActive,
		IsMaintenance:      req.IsMaintenance,
		IsVisible:          req.IsVisible,
		MinAmount:          req.MinAmount,
		MaxAmount:          req.MaxAmount,
		ProcessingFee:      req.ProcessingFee,
		ProcessingFeeRate:  req.ProcessingFeeRate,
		Currency:           req.Currency,
		SettlementDays:     req.SettlementDays,
		Priority:           req.Priority,
		DisplayName:        req.DisplayName,
		Description:        req.Description,
		IconURL:            req.IconURL,
		ColorCode:          req.ColorCode,
		ComplianceStatus:   req.ComplianceStatus,
		ComplianceNotes:    req.ComplianceNotes,
		ValidFrom:          req.ValidFrom,
		ValidUntil:         req.ValidUntil,
	}

	result := trx.Create(&paymentMethodConfiguration)
	if result.Error != nil {
		return nil, result.Error
	}

	return &paymentMethodConfiguration, nil
}

// API to create payment method configuration
//
//encore:api auth method=POST path=/api/admin/business/payment/method/create
func (s *Service) CreatePaymentMethodConfiguration(ctx context.Context, req *CreatePaymentMethodConfigurationRequest) (*CreatePaymentMethodConfigurationResponse, error) {

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var user *models.User
	var ok bool
	var businessID uuid.UUID
	if req.BusinessID == nil || *req.BusinessID == uuid.Nil {
		user, ok = auth.Data().(*models.User)
		if !ok {
			return nil, fmt.Errorf("invalid auth data type")
		} else {
			businessID = *user.BusinessID
		}
	} else {
		businessID = *req.BusinessID
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if trx.Error != nil {
		return nil, trx.Error
	}

	_, err := CreatePaymentMethod(trx, businessID, req)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	result := trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	return &CreatePaymentMethodConfigurationResponse{
		Message: "Payment method configuration created successfully",
	}, nil
}

// load both default payment methods for POS and MEMBERSHIP
func SeedDefaultPaymentMethods(trx *gorm.DB, businessID uuid.UUID) error {
	err := SeedDefaultMembershipPaymentMethods(trx, businessID)
	if err != nil {
		return err
	}
	err = SeedDefaultPOSPaymentMethods(trx, businessID)
	if err != nil {
		return err
	}
	return nil
}

// load the payment_methods_pos.json
func SeedDefaultPOSPaymentMethods(trx *gorm.DB, businessID uuid.UUID) error {
	paymentMethodsPOS := []models.PaymentMethodConfiguration{}
	json.Unmarshal(paymentMethodsPOSJSON, &paymentMethodsPOS)
	for _, paymentMethod := range paymentMethodsPOS {
		paymentMethod.BusinessID = businessID
		trx.Create(&paymentMethod)
	}
	return nil
}

// load the payment_methods_membership.json
func SeedDefaultMembershipPaymentMethods(trx *gorm.DB, businessID uuid.UUID) error {
	paymentMethodsMembership := []models.PaymentMethodConfiguration{}
	json.Unmarshal(paymentMethodsMembershipJSON, &paymentMethodsMembership)
	for _, paymentMethod := range paymentMethodsMembership {
		paymentMethod.BusinessID = businessID
		trx.Create(&paymentMethod)
	}
	return nil
}

// API to create default payment method configuration
//
//encore:api auth method=POST path=/api/admin/business/payment/method/create/default
func (s *Service) CreateDefaultPaymentMethodConfiguration(ctx context.Context, req *CreateDefaultPaymentMethodConfigurationRequest) (*CreateDefaultPaymentMethodConfigurationResponse, error) {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	var user *models.User
	var ok bool
	var businessID uuid.UUID
	if req.BusinessID == nil || *req.BusinessID == uuid.Nil {
		user, ok = auth.Data().(*models.User)
		if !ok {
			return nil, fmt.Errorf("invalid auth data type")
		} else {
			businessID = *user.BusinessID
		}
	} else {
		businessID = *req.BusinessID
	}

	result := SeedDefaultPaymentMethods(trx, businessID)
	if result != nil {
		trx.Rollback()
		return nil, result
	}

	err := trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &CreateDefaultPaymentMethodConfigurationResponse{
		Message: "Default payment method configurations created successfully",
	}, nil
}

// API to upload payment method icon
//
//encore:api auth raw method=POST path=/api/admin/business/payment/method/upload/icon
func (s *Service) UploadPaymentMethodIcon(w http.ResponseWriter, req *http.Request) {

	payment_method_configuration_id := req.FormValue("payment_method_configuration_id")
	if payment_method_configuration_id == "" {
		http.Error(w, "Payment method configuration ID is required", http.StatusBadRequest)
		return
	}

	// Validate file size early
	if err := req.ParseMultipartForm(10 << 20); err != nil { // Limit to 10 MB
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

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if trx.Error != nil {
		errs.HTTPError(w, trx.Error)
		return
	}

	var payment_method_configuration models.PaymentMethodConfiguration
	err = trx.Model(&models.PaymentMethodConfiguration{}).
		Where("id = ?", payment_method_configuration_id).
		First(&payment_method_configuration).Error
	if err != nil {
		trx.Rollback()
		errs.HTTPError(w, err)
		return
	}

	user, ok := auth.Data().(*models.User)
	if !ok {
		trx.Rollback()
		http.Error(w, "Invalid auth data type", http.StatusUnauthorized)
		return
	}

	businessID := *user.BusinessID
	var business models.Business
	err = trx.Model(&models.Business{}).
		Where("id = ?", businessID).
		First(&business).Error
	if err != nil {
		trx.Rollback()
		errs.HTTPError(w, err)
		return
	}

	var document aws_s3.Document
	document.File = file
	file_extension := filepath.Ext(header.Filename)
	document.DocPath = "business/" + business.RegistrationNumber + "/payment_methods/" + string(payment_method_configuration.PaymentMethod) + "/" + payment_method_configuration_id + file_extension

	// Replace spaces with underscores in document path
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		trx.Rollback()
		errs.HTTPError(w, err)
		return
	}

	payment_method_configuration.IconURL = &document_res.Url
	err = trx.Save(&payment_method_configuration).Error
	if err != nil {
		trx.Rollback()
		errs.HTTPError(w, err)
		return
	}

	// Commit the transaction
	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
