package businesses

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/auth_service"
	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	googleUUID "github.com/google/uuid" // For generating new UUIDs
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

// initService initializes the business service.
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

// REQUEST PARAMETERS STRUCTURES
type CreateBusinessParams struct {
	Name               string         `json:"name" valid:"required~Name is required"`
	Email              string         `json:"email" valid:"required~Email is required"`
	Phone              string         `json:"phone" valid:"required~Phone is required"`
	Address            common.Address `json:"address" valid:"required~Address is required"`
	Website            string         `json:"website" valid:"required~Website is required"`
	RegistrationNumber string         `json:"registration_number" valid:"required~Registration Number is required"`
}

type GetAllBusinessesParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type BusinessesResponse struct {
	Meta common.Pagination `json:"meta"`
	Data []models.Business `json:"data"`
}

type BusinessOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type BusinessOptionsResponse struct {
	Data []BusinessOption `json:"data"`
}

type UpdateBusinessParams struct {
	ID      uuid.UUID      `json:"id"`
	Name    string         `json:"name"`
	Email   string         `json:"email"`
	Phone   string         `json:"phone"`
	Address common.Address `json:"address" gorm:"embedded"`
	Website string         `json:"website"`
	LogoURL string         `json:"logo_url"`
}

type DeleteBusinessParams struct {
	ID           uuid.UUID `json:"id"`
	Confirmation string    `json:"confirmation" valid:"required~Confirmation is required"`
}

// request for migrate membership (involve auto creation of new membership, and auto migration of customer_memberships, point_rules, redemption_rules, vouchers)
type MigrateMembershipRequest struct {
	BusinessID   uuid.UUID `json:"business_id"`
	OldTierID    uuid.UUID `json:"old_tier_id"`
	NewTierName  string    `json:"new_tier_name"`
	NewTierLevel int       `json:"new_tier_level"`
	NewTierPrice float32   `json:"new_tier_price"`
}

// request for get all points rules for business
type GetPointsRulesRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type MembershipPointRules struct {
	MembershipID uuid.UUID          `json:"membership_id"`
	PointRules   []models.PointRule `json:"point_rules"`
}

// response for get all points rules for business
type GetPointsRulesResponse struct {
	Message string                 `json:"message"`
	Data    []MembershipPointRules `json:"data"`
}

// request for update points rules for business
type UpdatePointsRulesRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	//MembershipID                 uuid.UUID `json:"membership_id"`
	PointRuleID                  uuid.UUID `json:"point_rule_id" default:""`
	ActionType                   string    `json:"action_type" default:""`
	PointsEarned                 int       `json:"points_earned" default:"0"`
	PointsMultiplier             int       `json:"points_multiplier" default:"0"`
	MinAmount                    float32   `json:"min_amount" default:"0"`
	MaxPointsPerAction           int       `json:"max_points_per_action" default:"0"`
	MaxPointsPerDay              int       `json:"max_points_per_day" default:"0"`
	MaxPointsPerMonth            int       `json:"max_points_per_month" default:"0"`
	MaxPointsPerYear             int       `json:"max_points_per_year" default:"0"`
	IsActive                     bool      `json:"is_active" default:"false"`
	IsGeneralRule                bool      `json:"is_general_rule" default:"false"`
	MaxPointsPerCustomer         int       `json:"max_points_per_customer" default:"0"`
	MaxPointsPerCustomerPerDay   int       `json:"max_points_per_customer_per_day" default:"0"`
	MaxPointsPerCustomerPerMonth int       `json:"max_points_per_customer_per_month" default:"0"`
	MaxPointsPerCustomerPerYear  int       `json:"max_points_per_customer_per_year" default:"0"`
}

// request for delete points rules for business
type DeletePointsRulesRequest struct {
	BusinessID  uuid.UUID `json:"business_id"`
	PointRuleID uuid.UUID `json:"point_rule_id"`
}

type AddRedemptionRulesRequest struct {
	BusinessID                  uuid.UUID `json:"business_id"`
	RuleName                    string    `json:"rule_name" valid:"required~Rule Name is required"`
	Description                 string    `json:"description" valid:"required~Description is required"`
	TermsAndConditions          string    `json:"terms_and_conditions" valid:"required~Terms and Conditions is required"`
	MinAmount                   float32   `json:"min_amount" `
	MaxAmount                   float32   `json:"max_amount" `
	MaxRedemption               int       `json:"max_redemption" `
	MaxRedemptionPerCustomer    int       `json:"max_redemption_per_customer" `
	MaxRedemptionPerDay         int       `json:"max_redemption_per_day" `
	MaxRedemptionPerMonth       int       `json:"max_redemption_per_month"`
	MaxRedemptionPerYear        int       `json:"max_redemption_per_year"`
	MaxRedemptionPerTransaction int       `json:"max_redemption_per_transaction"`
	MaxRedemptionPerOrder       int       `json:"max_redemption_per_order"`
	ApplicableTierLevel         int       `json:"applicable_tier_level"`
	IsActive                    bool      `json:"is_active"`
	ValidFrom                   time.Time `json:"valid_from"`
	ValidTo                     time.Time `json:"valid_to"`
	IsAllowAllPaymentType       bool      `json:"is_allow_all_payment_type"`
	DiscountPercentage          float32   `json:"discount_percentage"`
	ExchangeRate                float32   `json:"exchange_rate"`
}

// request for get all redemption rules for business
type GetRedemptionRulesRequest struct {
	//BusinessID uuid.UUID `json:"business_id" valid:"required~Business ID is required"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type RedemptionRules struct {
	BusinessID      uuid.UUID                    `json:"business_id"`
	RedemptionRules []models.PointRedemptionRule `json:"redemption_rules"`
}

// response for get all redemption rules for business
type GetRedemptionRulesResponse struct {
	Message string            `json:"message"`
	Data    []RedemptionRules `json:"data"`
}

// request for update redemption rules for business
type UpdateRedemptionRulesRequest struct {
	BusinessID                  uuid.UUID  `json:"business_id"`
	RuleName                    *string    `json:"rule_name" valid:"required~Rule Name is required"`
	Description                 *string    `json:"description" valid:"required~Description is required"`
	TermsAndConditions          *string    `json:"terms_and_conditions" valid:"required~Terms and Conditions is required"`
	MinAmount                   *float32   `json:"min_amount" `
	MaxAmount                   *float32   `json:"max_amount" `
	MaxRedemption               *int       `json:"max_redemption" `
	MaxRedemptionPerCustomer    *int       `json:"max_redemption_per_customer" `
	MaxRedemptionPerDay         *int       `json:"max_redemption_per_day" `
	MaxRedemptionPerMonth       *int       `json:"max_redemption_per_month"`
	MaxRedemptionPerYear        *int       `json:"max_redemption_per_year"`
	MaxRedemptionPerTransaction *int       `json:"max_redemption_per_transaction"`
	MaxRedemptionPerOrder       *int       `json:"max_redemption_per_order"`
	ApplicableTierLevel         *int       `json:"applicable_tier_level"`
	IsActive                    *bool      `json:"is_active"`
	ValidFrom                   *time.Time `json:"valid_from"`
	ValidTo                     *time.Time `json:"valid_to"`
	IsAllowAllPaymentType       *bool      `json:"is_allow_all_payment_type"`
	DiscountPercentage          *float32   `json:"discount_percentage"`
	ExchangeRate                *float32   `json:"exchange_rate"`
}

type DeleteRedemptionRulesRequest struct {
	BusinessID       uuid.UUID `json:"business_id"`
	RedemptionRuleID uuid.UUID `json:"redemption_rule_id"`
}

type CreateProductWastageTypeRequest struct {
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Description string    `json:"description" valid:"required~Description is required"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active" valid:"required~Is Active is required"`
}

type GetAllProductWastageTypeResponse struct {
	Message string                      `json:"message"`
	Data    []models.ProductWastageType `json:"data"`
}

// CreateBusiness creates a new business
//
//encore:api auth method=POST path=/api/admin/business/create
func (s *Service) CreateBusiness(ctx context.Context, params *CreateBusinessParams) (*models.Business, error) {
	if err := middleware.CheckPermission(constants.CreateBusinessAction, nil, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	// Check if registration number already exists
	var existingBusiness models.Business
	if err := s.db.Where("registration_number = ?", params.RegistrationNumber).First(&existingBusiness).Error; err == nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "A business with this registration number already exists",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	business := &models.Business{
		Name:               params.Name,
		Email:              params.Email,
		Phone:              params.Phone,
		Address:            params.Address,
		Website:            params.Website,
		RegistrationNumber: params.RegistrationNumber,
	}

	if err := s.db.Create(business).Error; err != nil {
		return nil, err
	}
	//create membership default table for business
	membership := &models.Membership{
		BusinessID: business.ID,
		TierName:   "Basic",
		TierLevel:  0,
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
		DeletedAt:  gorm.DeletedAt{},
	}

	if err := s.db.Create(membership).Error; err != nil {
		log.Printf("unable to create membership: %v", err)
		return nil, err
	}

	_, err := s.SaveBusinessConfiguration(ctx, &SaveBusinessConfigurationParams{
		BusinessID: business.ID,
	})
	if err != nil {
		return nil, err
	}

	return business, nil
}

// GetBusiness retrieves a business by ID
//
//encore:api auth method=GET path=/api/admin/business/get/:id
func (s *Service) GetBusiness(ctx context.Context, id uuid.UUID) (*models.Business, error) {
	if err := middleware.CheckPermission(constants.ReadBusinessAction, &id, nil); err != nil {
		return nil, err
	}

	var business models.Business
	if err := s.db.First(&business, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &business, nil
}

// GetAllBusinesses retrieves all businesses
//
//encore:api auth method=POST path=/api/admin/business/get-all
func (s *Service) GetBusinesses(ctx context.Context, params *GetAllBusinessesParams) (*BusinessesResponse, error) {
	if err := middleware.CheckPermission(constants.ReadBusinessAction, nil, nil); err != nil {
		return nil, err
	}

	var businesses []models.Business

	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	offset := (params.Page - 1) * params.PageSize
	if err := s.db.Offset(offset).Limit(params.PageSize).Find(&businesses).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := s.db.Model(&models.Business{}).Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &BusinessesResponse{
		Data: businesses,
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			TotalPages: totalPages,
			Total:      int64(total),
		},
	}, nil
}

// GetBusinessOptions retrieves a list of businesses with minimal info for dropdown options
//
//encore:api auth method=GET path=/api/admin/business/options
func (s *Service) GetBusinessOptions(ctx context.Context) (*BusinessOptionsResponse, error) {
	var businesses []BusinessOption
	query := s.db.Model(&models.Business{}).Select("id", "name")
	if !auth_service.IsSanteAdmin() {
		business_id := auth_service.GetUserBusinessID()
		query = query.Where("id = ?", business_id)
	}

	if err := query.Find(&businesses).Error; err != nil {
		return nil, err
	}

	return &BusinessOptionsResponse{
		Data: businesses,
	}, nil
}

// UpdateBusiness updates an existing business
//
//encore:api auth method=POST path=/api/admin/business/update
func (s *Service) UpdateBusiness(ctx context.Context, params *UpdateBusinessParams) (*models.Business, error) {
	if err := middleware.CheckPermission(constants.UpdateBusinessAction, &params.ID, nil); err != nil {
		return nil, err
	}

	business, err := s.GetBusiness(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	if err := s.db.Model(&business).Omit("id, created_at, created_by").Updates(params).Error; err != nil {
		return nil, err
	}
	return business, nil
}

// UploadBusinessLogo uploads a logo for a business
//
//encore:api auth raw method=POST path=/api/admin/business/upload/logo
func (s *Service) UploadBusinessLogo(w http.ResponseWriter, req *http.Request) {
	// log.Println("Received file:", req.FormFile("file"))
	log.Println("Received id:", req.FormValue("id"))

	business_id := req.FormValue("id")
	if business_id == "" {
		http.Error(w, "Employee ID is required", http.StatusBadRequest)
		return
	}

	temp_uuid, err := googleUUID.Parse(business_id)
	if err != nil {
		http.Error(w, "Invalid Business ID format", http.StatusBadRequest)
		return
	}

	business, err := s.GetBusiness(context.TODO(), uuid.UUID(temp_uuid))
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
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

	var document aws_s3.Document
	document.File = file
	file_extension := filepath.Ext(header.Filename)
	document.DocPath = "business/" + business.RegistrationNumber + "/logo" + file_extension
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	var business_params = UpdateBusinessParams{
		ID:      business.ID,
		LogoURL: document_res.Url,
	}
	s.UpdateBusiness(context.TODO(), &business_params)

	w.WriteHeader(http.StatusOK)
}

// DeleteBusiness soft deletes a business
//
//encore:api auth method=DELETE path=/api/admin/business/delete/:id
func (s *Service) DeleteBusiness(ctx context.Context, id uuid.UUID) error {
	if err := middleware.CheckPermission(constants.DeleteBusinessAction, &id, nil); err != nil {
		return err
	}

	return s.db.Delete(&models.Business{}, "id = ?", id).Error
}

// This API serve as a migration API before deleting a membership
// migration of customer_memberships, point_rules, redemption_rules, vouchers to new membership
//
//encore:api auth method=POST path=/api/business/membership/migrate-membership
func (s *Service) MigrateMembership(ctx context.Context, req *MigrateMembershipRequest) (*common.BasicResponse, error) {
	//validate request
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware (for now use UpdateMembershipAction permission)
	err := middleware.CheckPermission(constants.UpdateMembershipAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//check if business exists
	var business models.Business
	if err = s.db.First(&business, "id = ?", req.BusinessID).Error; err != nil {
		return nil, err
	}

	//migrate membership
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//create new membership
	newMembership := models.Membership{
		BusinessID: req.BusinessID,
		TierName:   req.NewTierName,
		TierLevel:  req.NewTierLevel,
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
		DeletedAt:  gorm.DeletedAt{},
	}
	err = tx.Create(&newMembership).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	//migrate customer_memberships
	err = tx.Model(&models.CustomerMembership{}).Where("membership_id = ?", req.OldTierID).Update("membership_id", newMembership.ID).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	//migrate point_rules
	err = tx.Model(&models.PointRule{}).Where("membership_id = ?", req.OldTierID).Updates(map[string]interface{}{
		"membership_id": newMembership.ID,
	}).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// TODO: migrate redemption_rules
	// TODO: migrate vouchers

	//commit transaction
	tx.Commit()
	return &common.BasicResponse{
		Message: "Membership migrated successfully",
	}, nil
}

// API to get all points rules for business ID==business id
//
//encore:api auth method=GET path=/api/business/points/get-points-rules/:id
func (s *Service) GetPointsRules(ctx context.Context, id uuid.UUID, req *GetPointsRulesRequest) (*GetPointsRulesResponse, error) {
	fmt.Println("business id ---->", id)
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.ReadPointsRulesAction, &id, nil)
	if err != nil {
		return nil, err
	}

	//check if business exists
	var business models.Business
	err = s.db.First(&business, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	//get memberships
	var memberships []models.Membership
	err = s.db.Where("business_id = ?", business.ID).Find(&memberships).Error
	if err != nil {
		return nil, err
	}
	var membershipPointRulesList []MembershipPointRules
	//var pointRules []models.PointRule
	for _, membership := range memberships {
		fmt.Println("Processing membership:", membership.ID)

		// Get all point rules for this membership ID
		var point_rules []models.PointRule
		err = s.db.Where("membership_id = ?", membership.ID).Find(&point_rules).Error
		if err != nil {
			return nil, err
		}

		// Append each valid point rule to the master list
		membershipPointRulesList = append(membershipPointRulesList, MembershipPointRules{
			MembershipID: membership.ID,
			PointRules:   point_rules,
		})
	}

	//testing purpose
	for _, membershipPointRules := range membershipPointRulesList {
		fmt.Println("membership point rules id ---->", membershipPointRules.MembershipID)
		for _, pointRule := range membershipPointRules.PointRules {
			fmt.Println("point rule id ---->", pointRule.ID)
		}
	}

	return &GetPointsRulesResponse{
		Message: "Points rules fetched successfully",
		Data:    membershipPointRulesList,
	}, nil
}

// API to update points rules for business
// suppose to update the point rule and customer limit rule at the same time, but not implemented yet.
//
//encore:api auth method=POST path=/api/business/points/update-points-rules
func (s *Service) UpdatePointsRules(ctx context.Context, req *UpdatePointsRulesRequest) (*common.BasicResponse, error) {
	//validate request struct
	fmt.Println("business id ---->", req.BusinessID)
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.UpdatePointsRulesAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//check if membership exists
	var memberships []models.Membership
	result := s.db.Where("business_id = ?", req.BusinessID).Find(&memberships)
	if result.Error != nil {
		return nil, result.Error
	}
	//if business has no memberships or business not found
	if result.RowsAffected == 0 {
		return nil, result.Error
	}
	fmt.Println("memberships ---->", memberships)

	//update point rule
	//membership id and point rule id must tally only will update the point rule
	var pointRules []models.PointRule
	updated := false
	for _, membership := range memberships {
		fmt.Println("Current membership id in loop---->", membership.ID)
		fmt.Println("Current point rule id in loop---->", req.PointRuleID)
		result := s.db.Model(&pointRules).Where("membership_id = ? AND id = ?", membership.ID, req.PointRuleID).Updates(map[string]interface{}{
			"action_type":           req.ActionType,
			"points_earned":         req.PointsEarned,
			"points_multiplier":     req.PointsMultiplier,
			"min_amount":            req.MinAmount,
			"max_points_per_action": req.MaxPointsPerAction,
			"max_points_per_day":    req.MaxPointsPerDay,
			"max_points_per_month":  req.MaxPointsPerMonth,
			"max_points_per_year":   req.MaxPointsPerYear,
			"is_active":             req.IsActive,
			"is_general_rule":       req.IsGeneralRule,
		})
		if result.Error != nil {
			return nil, result.Error
		}

		if result.RowsAffected > 0 {
			updated = true
			break
		}
	}

	if !updated {
		return nil, result.Error
	}

	// TODO update customer limit rules (not implemented yet)

	return &common.BasicResponse{
		Message: "Points rules updated successfully",
	}, nil
}

// API to delete points rules for business (soft delete)
//
//encore:api auth method=POST path=/api/business/points/delete-points-rules
func (s *Service) DeletePointsRules(ctx context.Context, req *DeletePointsRulesRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.DeletePointsRulesAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//check if business exists
	var memberships []models.Membership
	err = s.db.Where("business_id = ?", req.BusinessID).Find(&memberships).Error
	if err != nil {
		return nil, err
	}

	//begin transaction (use this to rollback if any error occurs)
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	//delete point rule
	var pointRules []models.PointRule
	deleted := false
	for _, membership := range memberships {
		fmt.Println("Current membership id in loop---->", membership.ID)
		fmt.Println("Current point rule id in loop---->", req.PointRuleID)
		result := tx.Where("membership_id = ? AND id = ?", membership.ID, req.PointRuleID).Delete(&pointRules)

		if result.Error != nil {
			tx.Rollback()
			return nil, result.Error
		}
		if result.RowsAffected > 0 {
			deleted = true
			break
		}
	}

	if !deleted {
		tx.Rollback()
		return nil, err
	}

	//commit transaction
	tx.Commit()

	return &common.BasicResponse{
		Message: "Points rules deleted successfully",
	}, nil
}

// API to add redemption rules for business
//
//encore:api auth method=POST path=/api/business/points/add-redemption-rules
func (s *Service) AddRedemptionRules(ctx context.Context, req *AddRedemptionRulesRequest) (*common.BasicResponse, error) {
	fmt.Println("Business ID ----->", req.BusinessID)
	fmt.Println("Rule Name ----->", req.RuleName)
	fmt.Println("Min Amount ----->", req.MinAmount)
	fmt.Println("Max Amount ----->", req.MaxAmount)
	fmt.Println("Max Redemption ----->", req.MaxRedemption)
	fmt.Println("Max Redemption Per Customer ----->", req.MaxRedemptionPerCustomer)
	fmt.Println("Max Redemption Per Day ----->", req.MaxRedemptionPerDay)
	fmt.Println("Max Redemption Per Month ----->", req.MaxRedemptionPerMonth)
	fmt.Println("Max Redemption Per Year ----->", req.MaxRedemptionPerYear)
	fmt.Println("Is Active ----->", req.IsActive)

	//validate request struct
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.CRURedemptionRulesAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//check the rules is already exists in the point redemption rules table for each business
	// one business can only have one point redemption rule
	var pointRedemptionRule models.PointRedemptionRule
	result := s.db.Where("business_id = ? AND is_active = ?", req.BusinessID, true).First(&pointRedemptionRule)
	fmt.Println("result ---->", result.Error)
	if result.Error == nil {
		return nil, err
	}
	//check if the rule already exists
	if result.RowsAffected > 0 {
		return nil, err
	}

	//add
	pointRedemptionRule = models.PointRedemptionRule{
		BusinessID:                  req.BusinessID,
		RuleName:                    req.RuleName,
		Description:                 req.Description,
		TermsAndConditions:          req.TermsAndConditions,
		MinAmount:                   req.MinAmount,
		MaxAmount:                   req.MaxAmount,
		MaxRedemption:               req.MaxRedemption,
		MaxRedemptionPerCustomer:    req.MaxRedemptionPerCustomer,
		MaxRedemptionPerDay:         req.MaxRedemptionPerDay,
		MaxRedemptionPerMonth:       req.MaxRedemptionPerMonth,
		MaxRedemptionPerYear:        req.MaxRedemptionPerYear,
		MaxRedemptionPerTransaction: req.MaxRedemptionPerTransaction,
		MaxRedemptionPerOrder:       req.MaxRedemptionPerOrder,
		ApplicableTierLevel:         req.ApplicableTierLevel,
		IsActive:                    req.IsActive,
		ValidFrom:                   req.ValidFrom,
		ValidTo:                     req.ValidTo,
		IsAllowAllPaymentType:       req.IsAllowAllPaymentType,
		DiscountPercentage:          req.DiscountPercentage,
		ExchangeRate:                req.ExchangeRate,
		CreatedAt:                   time.Now(),
		UpdatedAt:                   nil,
		DeletedAt:                   gorm.DeletedAt{},
	}

	err = s.db.Create(&pointRedemptionRule).Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Point redemption rules added successfully",
	}, nil
}

// API to get all redemption rules for business (this will get all redemption rules for all memberships of a business entity) id == business id
//
//encore:api auth method=GET path=/api/business/points/get-redemption-rules/:id
func (s *Service) GetRedemptionRules(ctx context.Context, id uuid.UUID, req *GetRedemptionRulesRequest) (*GetRedemptionRulesResponse, error) {
	//validate request struct
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.CRURedemptionRulesAction, &id, nil)
	if err != nil {
		return nil, err
	}

	//check if the business exists
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ?", id).First(&business).Error
	if err != nil {
		return nil, err
	}

	//get all redemption rules for the memberships
	var redemptionRulesList []RedemptionRules
	var rules []models.PointRedemptionRule
	err = s.db.Model(&models.PointRedemptionRule{}).Where("business_id = ?", id).Find(&rules).Error
	if err != nil {
		return nil, err
	}

	redemptionRulesList = append(redemptionRulesList, RedemptionRules{
		BusinessID:      business.ID,
		RedemptionRules: rules,
	})

	return &GetRedemptionRulesResponse{
		Message: "Redemption rules fetched successfully",
		Data:    redemptionRulesList,
	}, nil
}

// API to update redemption rules for business
//
//encore:api auth method=POST path=/api/business/points/update-redemption-rules
func (s *Service) UpdateRedemptionRules(ctx context.Context, req *UpdateRedemptionRulesRequest) (*common.BasicResponse, error) {
	//validate request struct
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.CRURedemptionRulesAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//business exists
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business).Error
	if err != nil {
		return nil, err
	}

	//existing point redemption rule
	var oldRedemptionRule models.PointRedemptionRule
	err = s.db.Model(&models.PointRedemptionRule{}).Where("business_id = ?", req.BusinessID).First(&oldRedemptionRule).Error
	if err != nil {
		return nil, err
	}

	//update data and check if the data is valid
	updateData := make(map[string]interface{})
	common.HandleFieldUpdate(&oldRedemptionRule.RuleName, req.RuleName, "rule_name", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.Description, req.Description, "description", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.TermsAndConditions, req.TermsAndConditions, "terms_and_conditions", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.MinAmount, req.MinAmount, "min_amount", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxAmount, req.MaxAmount, "max_amount", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemption, req.MaxRedemption, "max_redemption", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerCustomer, req.MaxRedemptionPerCustomer, "max_redemption_per_customer", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerDay, req.MaxRedemptionPerDay, "max_redemption_per_day", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerMonth, req.MaxRedemptionPerMonth, "max_redemption_per_month", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerYear, req.MaxRedemptionPerYear, "max_redemption_per_year", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerTransaction, req.MaxRedemptionPerTransaction, "max_redemption_per_transaction", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.MaxRedemptionPerOrder, req.MaxRedemptionPerOrder, "max_redemption_per_order", updateData, true)
	common.HandleFieldUpdate(&oldRedemptionRule.ApplicableTierLevel, req.ApplicableTierLevel, "applicable_tier_level", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.IsActive, req.IsActive, "is_active", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.ValidFrom, req.ValidFrom, "valid_from", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.ValidTo, req.ValidTo, "valid_to", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.IsAllowAllPaymentType, req.IsAllowAllPaymentType, "is_allow_all_payment_type", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.DiscountPercentage, req.DiscountPercentage, "discount_percentage", updateData, false)
	common.HandleFieldUpdate(&oldRedemptionRule.ExchangeRate, req.ExchangeRate, "exchange_rate", updateData, false)

	//update redemption rules
	result := s.db.Model(&models.PointRedemptionRule{}).Where("business_id = ?", req.BusinessID).Updates(updateData)
	//check if there was an error in the update
	if result.Error != nil {
		return nil, result.Error
	}
	//check if any row was affected (affected means updated)
	if result.RowsAffected == 0 {
		return nil, result.Error
	}

	//success
	return &common.BasicResponse{
		Message: "Redemption rules updated successfully",
	}, nil
}

// API to delete redemption rules for business
//
//encore:api auth method=POST path=/api/business/points/delete-redemption-rules
func (s *Service) DeleteRedemptionRules(ctx context.Context, req *DeleteRedemptionRulesRequest) (*common.BasicResponse, error) {
	//validate request struct
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	//check permission from middleware
	err := middleware.CheckPermission(constants.DRedemptionRulesAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	//check if business exists
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business).Error
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var redemptionRules []models.PointRedemptionRule
	deleted := false
	result := tx.Where("business_id = ? AND id = ?", business.ID, req.RedemptionRuleID).Delete(&redemptionRules)

	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}
	if result.RowsAffected > 0 {
		deleted = true
	}

	if !deleted {
		tx.Rollback()
		return nil, result.Error
	}

	tx.Commit()

	return &common.BasicResponse{
		Message: "Redemption rules deleted successfully",
	}, nil
}

// create product wastage type
//
//encore:api auth method=POST path=/api/business/reports/product-wastage-type/create
func (s *Service) CreateProductWastageType(ctx context.Context, req *CreateProductWastageTypeRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// check business existence
	var business models.Business
	query := s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business)
	if query.Error != nil {
		return nil, errors.New("business not found")
	}

	// check the request sort order is unique
	var existingProductWastageType models.ProductWastageType
	query = s.db.Model(&models.ProductWastageType{}).Where("business_id = ? AND sort_order = ?", req.BusinessID, req.SortOrder).First(&existingProductWastageType)
	if query.Error == nil {
		return nil, errors.New("sort order already exists")
	}

	productWastageType := models.ProductWastageType{
		BusinessID:  req.BusinessID,
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   nil,
		DeletedAt:   gorm.DeletedAt{},
	}

	err := s.db.Create(&productWastageType).Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Product wastage type created successfully",
	}, nil
}

// API to get all product wastage type
//
//encore:api auth method=GET path=/api/business/reports/product-wastage-type/get-all/:business_id
func (s *Service) GetAllProductWastageType(ctx context.Context, business_id uuid.UUID) (*GetAllProductWastageTypeResponse, error) {
	var productWastageTypes []models.ProductWastageType
	result := s.db.Model(&models.ProductWastageType{}).Where("business_id = ?", business_id).Find(&productWastageTypes)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAllProductWastageTypeResponse{
		Message: "Product wastage type retrieved successfully",
		Data:    productWastageTypes,
	}, nil
}
