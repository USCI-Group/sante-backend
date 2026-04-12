// Package outlets provides outlet management functionality
package outlets

import (
	"context"
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
	"encore.app/common_operations"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/beta/auth"
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

// initService initializes the outlet service.
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
type CreateOutletParams struct {
	BusinessID         uuid.UUID      `json:"business_id" gorm:"type:uuid"`
	Name               string         `json:"name" gorm:"type:varchar(255)" valid:"required~Name is required"`
	Email              string         `json:"email" gorm:"type:varchar(255)" valid:"required~Email is required,email~Invalid email format"`
	Phone              string         `json:"phone" gorm:"type:varchar(255)" valid:"required~Phone is required"`
	Address            common.Address `json:"address" gorm:"embedded"`
	Website            string         `json:"website,omitempty" gorm:"type:varchar(255)"`
	RegistrationNumber string         `json:"registration_number,omitempty" gorm:"type:varchar(255)" valid:"required~Registration number is required"`
	ImageURL           string         `json:"image_url,omitempty"`
	Latitude           *float64       `json:"latitude,omitempty"`
	Longitude          *float64       `json:"longitude,omitempty"`
}

type UpdateOutletParams struct {
	ID                 uuid.UUID      `json:"id"`
	Name               string         `json:"name"`
	Email              string         `json:"email"`
	Phone              string         `json:"phone"`
	Address            common.Address `json:"address" gorm:"embedded"`
	Website            string         `json:"website"`
	RegistrationNumber string         `json:"registration_number"`
	ImageURL           string         `json:"image_url"`
	Latitude           *float64       `json:"latitude,omitempty"`
	Longitude          *float64       `json:"longitude,omitempty"`
}

type GetOutletsParams struct {
	BusinessID uuid.UUID `json:"business_id"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Filter     struct {
		Search string `json:"search"`
	} `json:"filter"`
}

type DeleteOutletParams struct {
	ID           uuid.UUID `json:"id"`
	Confirmation string    `json:"confirmation" valid:"required~Confirmation is required"`
}

type UpdateMerchantSecretsRequest struct {
	OutletID                    uuid.UUID `json:"outlet_id" `
	FiuuVerifyKey               string    `json:"fiuu_verify_key"`
	FiuuApplicationCode         string    `json:"fiuu_application_code"`
	FiuuSecretKey               string    `json:"fiuu_secret_key"`
	FiuuOfflineSecretKey        string    `json:"fiuu_offline_secret_key"`
	FiuuCloudERCSecretKey       string    `json:"fiuu_cloud_erc_secret_key"`
	FiuuCloudERCAccountID       string    `json:"fiuu_cloud_erc_account_id"`
	FiuuMerchantID              string    `json:"fiuu_merchant_id"`
	FiuuCloudERCAccountPassword string    `json:"fiuu_cloud_erc_account_password"`
	GrabStoreID                 string    `json:"grab_store_id"`
}

type VTToOutletTerminalMap struct {
	VtID         string    `json:"vt_id" valid:"required~VT ID is required"`
	OutletID     uuid.UUID `json:"outlet_id"`
	Confirmation bool      `json:"confirmation"`
}

type VTToOutletTerminalMapResponse struct {
	OutletTerminalID uuid.UUID `json:"outlet_terminal_id"`
	VtID             string    `json:"vt_id" valid:"required~VT ID is required"`
}

// RESPONSE STRUCTURES
type OutletResponse struct {
	Meta common.Pagination `json:"meta"`
	Data []models.Outlet   `json:"data"`
}

type OutletOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OutletOptionsResponse struct {
	Data []OutletOption `json:"data"`
}

// request parameters to set outlet product active status
type SetOutletProductActiveStatusRequest struct {
	OutletID  uuid.UUID `json:"outlet_id" `
	ProductID uuid.UUID `json:"product_id" `
	IsActive  bool      `json:"is_active"`
}

type UpdateOutletStatusRequest struct {
	OutletID             uuid.UUID `json:"outlet_id"`
	Status               string    `json:"status" valid:"required~Status is required"`
	AutoCloseOnlineOrder bool      `json:"auto_close_online_order" default:"false"`
}

type GetOutletModeResponse struct {
	Message string `json:"message"`
	Mode    string `json:"mode"`
}

type CreateOutletGroupRequest struct {
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Description string    `json:"description"`
}

type OutletGroupResponse struct {
	OutletGroups  []models.OutletGroup `json:"outlet_groups"`
	OutletOptions []OutletOption       `json:"outlet_options"`
}

type AssignOutletToGroupRequest struct {
	OutletGroupID uuid.UUID `json:"outlet_group_id"`
	OutletID      uuid.UUID `json:"outlet_id"`
}

type AssignUserToGroupRequest struct {
	OutletGroupID uuid.UUID `json:"outlet_group_id"`
	UserID        uuid.UUID `json:"user_id"`
}

type BindPlatformIDRequest struct {
	OutletID      uuid.UUID `json:"outlet_id" valid:"required"`
	ShopeeStoreID *string   `json:"shopee_store_id,omitempty"`
	GrabStoreID   *string   `json:"grab_store_id,omitempty"`
}

type BindPlatformIDResponse struct {
	Message string `json:"message"`
}

func (s *Service) GetOutletByID(ctx context.Context, outlet_id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	result := s.db.Where("id = ?", outlet_id).First(&outlet)
	if result.Error != nil {
		return nil, result.Error
	}

	return &outlet, nil
}

//encore:api auth method=POST path=/api/admin/outlet/create
func (s *Service) CreateOutlet(ctx context.Context, params *CreateOutletParams) (*models.Outlet, error) {
	if err := middleware.CheckPermission(constants.CreateOutletAction, &params.BusinessID, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
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

	// Check if registration number already exists
	// Comment to allow outlets with the same ssm for now
	// var existingOutlet models.Outlet
	// if err := trx.Where("registration_number = ? AND business_id = ?", params.RegistrationNumber, params.BusinessID).First(&existingOutlet).Error; err == nil {
	// 	trx.Rollback()

	// 	return nil, &errs.Error{
	// 		Code:    errs.AlreadyExists,
	// 		Message: "A business with this registration number already exists",
	// 	}
	// }

	outlet := &models.Outlet{
		BusinessID:         params.BusinessID,
		Name:               params.Name,
		Email:              params.Email,
		Phone:              params.Phone,
		Address:            params.Address,
		Website:            params.Website,
		RegistrationNumber: params.RegistrationNumber,
		Latitude:           params.Latitude,
		Longitude:          params.Longitude,
	}

	if err := trx.Create(outlet).Error; err != nil {
		trx.Rollback()
		return nil, err
	}

	// create merchant secret if not exists
	var merchantSecret models.MerchantSecret
	err := s.db.Where("outlet_id = ?", outlet.ID).First(&merchantSecret).Error
	if err != nil {
		merchantSecret = models.MerchantSecret{
			OutletID: outlet.ID,
		}
		s.db.Create(&merchantSecret)
	}

	commit_err := trx.Commit().Error
	if commit_err != nil {
		return nil, commit_err
	}

	// e_invoice removed

	return outlet, nil
}

//encore:api auth method=GET path=/api/admin/outlet/get/:id
func (s *Service) GetOutlet(ctx context.Context, id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	if err := s.db.Preload("Business").First(&outlet, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Business or outlet not found",
			}
		}
		return nil, err
	}

	// Check if user has permission to access this outlet
	if err := middleware.CheckPermission(constants.ReadOutletAction, &outlet.BusinessID, &id); err != nil {
		return nil, err
	}

	return &outlet, nil
}

//encore:api auth method=POST path=/api/admin/outlet/update
func (s *Service) UpdateOutlet(ctx context.Context, params *UpdateOutletParams) (*models.Outlet, error) {
	// Get existing outlet first to check permissions
	outlet, err := s.GetOutlet(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	// Check if user has permission to access this outlet
	if err := middleware.CheckPermission(constants.UpdateOutletAction, &outlet.BusinessID, &params.ID); err != nil {
		return nil, err
	}

	if params.RegistrationNumber != outlet.RegistrationNumber || outlet.TIN == nil {
		outlet.RegistrationNumber = params.RegistrationNumber
		// e_invoice removed
	}

	if err := s.db.Model(&outlet).Omit("id, business_id, created_at, created_by").Updates(params).Error; err != nil {
		return nil, err
	}

	return outlet, nil
}

//encore:api auth method=POST path=/api/admin/outlet/get-all
func (s *Service) GetOutlets(ctx context.Context, params *GetOutletsParams) (*OutletResponse, error) {
	if err := middleware.CheckPermission(constants.ReadOutletAction, nil, nil); err != nil {
		return nil, err
	}

	d := auth.Data()
	user := d.(*models.User)

	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	var outlets []models.Outlet

	query := s.db.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	queryCount := s.db.Model(&models.Outlet{})

	if auth_service.IsSanteAdmin() {
		// Sante admin can access all outlets
		if params.BusinessID != uuid.Nil {
			query = query.Where("business_id = ?", params.BusinessID)
			queryCount = queryCount.Where("business_id = ?", params.BusinessID)
		}
	} else {
		query = query.Where("business_id = ?", user.BusinessID)
		queryCount = queryCount.Where("business_id = ?", user.BusinessID)
	}

	if params.Filter.Search != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?) OR LOWER(registration_number) LIKE LOWER(?)", "%"+params.Filter.Search+"%", "%"+params.Filter.Search+"%")
		queryCount = queryCount.Where("LOWER(name) LIKE LOWER(?) OR LOWER(registration_number) LIKE LOWER(?)", "%"+params.Filter.Search+"%", "%"+params.Filter.Search+"%")
	}

	if err := query.Find(&outlets).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := queryCount.Count(&total).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.PageSize)))

	return &OutletResponse{
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
		Data: outlets,
	}, nil
}

//encore:api auth method=GET path=/api/admin/outlet/get-all-without-pagination/:business_id
func (s *Service) GetOutletsWithoutPagination(ctx context.Context, business_id uuid.UUID) (*OutletResponse, error) {
	if err := middleware.CheckPermission(constants.ReadOutletAction, nil, nil); err != nil {
		return nil, err
	}

	d := auth.Data()
	user := d.(*models.User)

	var outlets []models.Outlet

	query := s.db.Model(&models.Outlet{})

	if auth_service.IsSanteAdmin() {
		// Sante admin can access all outlets
		if business_id != uuid.Nil {
			query = query.Where("business_id = ?", business_id)
		}
	} else {
		query = query.Where("business_id = ?", user.BusinessID)
	}

	if err := query.Find(&outlets).Error; err != nil {
		return nil, err
	}

	return &OutletResponse{
		Data: outlets,
	}, nil
}

// GetOutletOptions retrieves a list of outlets with minimal info for dropdown options
//
//encore:api auth method=GET path=/api/admin/outlet/options/:business_id
func (s *Service) GetOutletOptions(ctx context.Context, business_id uuid.UUID) (*OutletOptionsResponse, error) {
	var outlets []OutletOption
	if err := s.db.Model(&models.Outlet{}).Select("id", "name").Where("business_id = ?", business_id).Order("name ASC").Find(&outlets).Error; err != nil {
		return nil, err
	}

	return &OutletOptionsResponse{
		Data: outlets,
	}, nil
}

//encore:api auth method=DELETE path=/api/admin/outlet/delete/:id
func (s *Service) DeleteOutlet(ctx context.Context, id uuid.UUID) error {
	var outlet models.Outlet
	if err := s.db.First(&outlet, "id = ?", id).Error; err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.DeleteOutletAction, &outlet.BusinessID, &id); err != nil {
		return err
	}

	return s.db.Delete(&outlet).Error
}

//encore:api auth raw method=POST path=/api/admin/outlet/upload/image
func (s *Service) UploadOutletImage(w http.ResponseWriter, req *http.Request) {
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

	outlet, err := s.GetOutlet(context.TODO(), uuid.UUID(temp_uuid))
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
	document.DocPath = "business/" + outlet.Business.RegistrationNumber + "/outlet/" + outlet.Name + "/outlet_front" + file_extension

	// Replace spaces with underscores in document path
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	outlet.ImageURL = document_res.Url
	err = s.db.Save(&outlet).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth method=POST path=/api/admin/outlet/merchant-secrets/update
func (s *Service) UpdateMerchantSecrets(ctx context.Context, req *UpdateMerchantSecretsRequest) (*models.MerchantSecret, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	if !auth_service.IsSanteAdmin() {
		return nil, &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "Unauthorized access",
		}
	}

	outlet, err := s.GetOutlet(ctx, req.OutletID)
	if err != nil {
		return nil, err
	}

	encrypted_fiuu_secret_key, err := common.EncryptText(req.FiuuSecretKey)
	if err != nil {
		return nil, err
	}

	encrypted_fiuu_offline_secret_key, err := common.EncryptText(req.FiuuOfflineSecretKey)
	if err != nil {
		return nil, err
	}

	encrypted_fiuu_cloud_erc_secret_key, err := common.EncryptText(req.FiuuCloudERCSecretKey)
	if err != nil {
		return nil, err
	}

	encrypted_fiuu_cloud_erc_account_password, err := common.EncryptText(req.FiuuCloudERCAccountPassword)
	if err != nil {
		return nil, err
	}

	var merchantSecret models.MerchantSecret
	err = s.db.Where("outlet_id = ?", outlet.ID).First(&merchantSecret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			merchantSecret = models.MerchantSecret{
				OutletID:              req.OutletID,
				FiuuApplicationCode:   &req.FiuuApplicationCode,
				GrabStoreID:           &req.GrabStoreID,
				FiuuCloudERCAccountID: &req.FiuuCloudERCAccountID,
				FiuuMerchantID:        &req.FiuuMerchantID,
				FiuuVerifyKey:         &req.FiuuVerifyKey,
			}
			if req.FiuuSecretKey != "" {
				merchantSecret.FiuuSecretKey = &encrypted_fiuu_secret_key
			}
			if req.FiuuOfflineSecretKey != "" {
				merchantSecret.FiuuOfflineSecretKey = &encrypted_fiuu_offline_secret_key
			}
			if req.FiuuCloudERCSecretKey != "" {
				merchantSecret.FiuuCloudERCSecretKey = &encrypted_fiuu_cloud_erc_secret_key
			}
			if req.FiuuCloudERCAccountPassword != "" {
				merchantSecret.FiuuCloudERCAccountPassword = &encrypted_fiuu_cloud_erc_account_password
			}
			err = s.db.Create(&merchantSecret).Error
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		merchantSecret.FiuuApplicationCode = &req.FiuuApplicationCode
		merchantSecret.GrabStoreID = &req.GrabStoreID
		merchantSecret.FiuuCloudERCAccountID = &req.FiuuCloudERCAccountID
		merchantSecret.FiuuMerchantID = &req.FiuuMerchantID
		merchantSecret.FiuuVerifyKey = &req.FiuuVerifyKey
		if req.FiuuSecretKey != "" {
			merchantSecret.FiuuSecretKey = &encrypted_fiuu_secret_key
		}
		if req.FiuuOfflineSecretKey != "" {
			merchantSecret.FiuuOfflineSecretKey = &encrypted_fiuu_offline_secret_key
		}
		if req.FiuuCloudERCSecretKey != "" {
			merchantSecret.FiuuCloudERCSecretKey = &encrypted_fiuu_cloud_erc_secret_key
		}
		if req.FiuuCloudERCAccountPassword != "" {
			merchantSecret.FiuuCloudERCAccountPassword = &encrypted_fiuu_cloud_erc_account_password
		}
		if req.GrabStoreID == "" {
			merchantSecret.GrabStoreID = nil
		}
		err = s.db.Save(&merchantSecret).Error
		if err != nil {
			return nil, err
		}
	}
	return &merchantSecret, nil
}

// GetMerchantSecret retrieves the merchant secret for a specific outlet
//
//encore:api auth method=GET path=/api/admin/outlet/merchant-secrets/:outlet_id
func (s *Service) GetMerchantSecret(ctx context.Context, outlet_id uuid.UUID) (*models.MerchantSecret, error) {
	var merchantSecret models.MerchantSecret
	err := s.db.Where("outlet_id = ?", outlet_id).First(&merchantSecret).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Merchant secret not found for this outlet",
			}
		}
		return nil, err
	}

	return &merchantSecret, nil
}

// api to set outlet 's product active status
//
//encore:api auth method=PUT path=/api/admin/outlet/set-product-active-status
func (s *Service) SetOutletProductActiveStatus(ctx context.Context, req *SetOutletProductActiveStatusRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.UpdateProductAction, nil, &req.OutletID)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var outletProduct models.OutletProduct
	err = trx.Preload("Product").Where("outlet_id = ? AND product_id = ?", req.OutletID, req.ProductID).First(&outletProduct).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	outletProduct.IsActive = req.IsActive
	err = trx.Save(&outletProduct).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	commit_err := trx.Commit().Error
	if commit_err != nil {
		trx.Rollback()
		return nil, commit_err
	}

	// grabfood removed

	return &common.BasicResponse{
		Message: "Product active status set successfully",
	}, nil
}

// API to update outlet status
//
//encore:api auth method=PUT path=/api/admin/outlet/update-status
func (s *Service) UpdateOutletStatus(ctx context.Context, req *UpdateOutletStatusRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.UpdateOutletAction, nil, &req.OutletID)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var outlet models.Outlet
	err = trx.Where("id = ?", req.OutletID).First(&outlet).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	outlet.OutletStatus = models.OutletStatus(req.Status)
	err = trx.Save(&outlet).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// Auto closed/open online order if enabled
	if req.AutoCloseOnlineOrder {
		log.Println("Auto close online order enabled")
		err = common_operations.ToggleOnlineOrder(
			trx,
			req.OutletID,
			models.OutletStatus(req.Status),
		)
		if err != nil {
			log.Println("Error auto close online order:", err)
			trx.Rollback()
			return nil, err
		}
	}

	commit_err := trx.Commit().Error
	if commit_err != nil {
		trx.Rollback()
		return nil, commit_err
	}

	// grabfood removed

	return &common.BasicResponse{
		Message: "Outlet status updated successfully",
	}, nil
}

// API to get a outlet mode/status
//
//encore:api auth method=GET path=/api/admin/outlet/get-outlet-mode/:outlet_id
func (s *Service) GetOutletMode(ctx context.Context, outlet_id uuid.UUID) (*GetOutletModeResponse, error) {
	var outlet models.Outlet
	err := s.db.Where("id = ?", outlet_id).First(&outlet).Error
	if err != nil {
		return nil, err
	}

	if outlet.OutletStatus == "" {
		return &GetOutletModeResponse{
			Message: "Outlet mode retrieved successfully",
			Mode:    "closed",
		}, nil
	}

	return &GetOutletModeResponse{
		Message: "Outlet mode retrieved successfully",
		Mode:    string(outlet.OutletStatus),
	}, nil
}

// API to upload outlet static QR
//
//encore:api auth raw method=POST path=/api/admin/outlet/upload-static-qr
func (s *Service) UploadOutletStaticQR(w http.ResponseWriter, req *http.Request) {
	outlet_id := req.FormValue("outlet_id")
	if outlet_id == "" {
		http.Error(w, "Outlet ID is required", http.StatusBadRequest)
		return
	}

	outlet_uuid, err := googleUUID.Parse(outlet_id)
	if err != nil {
		http.Error(w, "Invalid Outlet ID format", http.StatusBadRequest)
		return
	}

	fmt.Println("outlet_uuid", outlet_uuid)

	outlet := models.Outlet{}
	err = s.db.Model(&models.Outlet{}).Where("id = ?", outlet_uuid).First(&outlet).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	err = s.db.Model(&models.Business{}).Where("id = ?", outlet.BusinessID).First(&business).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(2 << 20); err != nil { // Limit to 2 MB
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
	document.DocPath = fmt.Sprintf("business/%s/outlet/%s/static_qr/%s_%d%s",
		business.RegistrationNumber,
		outlet.ID,
		outlet.Name,
		time.Now().UnixMilli(),
		file_extension)
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	outlet.OutletStaticQR = &document_res.Url

	err = s.db.Save(&outlet).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

//encore:api auth method=PUT path=/api/admin/outlet/set-modifier-option-active-status
func (s *Service) SetOutletModifierOptionActiveStatus(ctx context.Context, req *common.OutletModifierOption) error {
	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	var outletModifierOption models.OutletModifierOption
	err := trx.Model(&models.OutletModifierOption{}).Where("id = ?", req.ID).Preload("ModifierOptions").First(&outletModifierOption).Error
	if err != nil {
		trx.Rollback()
		return err
	}

	outletModifierOption.IsActive = req.IsActive
	err = trx.Save(&outletModifierOption).Error
	if err != nil {
		trx.Rollback()
		return err
	}

	commit_err := trx.Commit().Error
	if commit_err != nil {
		trx.Rollback()
		return commit_err
	}

	// grabfood removed

	return nil
}

//encore:api auth method=POST path=/api/admin/outlet/outlet-group/create
func (s *Service) CreateOutletGroup(ctx context.Context, req *CreateOutletGroupRequest) error {
	if err := middleware.CheckPermission(constants.CreateOutletAction, &req.BusinessID, nil); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	outletGroup := &models.OutletGroup{
		BusinessID:  req.BusinessID,
		Name:        req.Name,
		Description: req.Description,
	}
	if err := trx.Create(&outletGroup).Error; err != nil {
		trx.Rollback()
		return err
	}

	if commitErr := trx.Commit().Error; commitErr != nil {
		trx.Rollback()
		return commitErr
	}

	return nil
}

//encore:api auth method=PUT path=/api/admin/outlet/outlet-group/update
func (s *Service) UpdateOutletGroup(ctx context.Context, req *models.OutletGroup) error {
	if err := middleware.CheckPermission(constants.CreateOutletAction, &req.BusinessID, nil); err != nil {
		return err
	}

	var outletGroup models.OutletGroup
	err := s.db.Model(&models.OutletGroup{}).Where("id = ?", req.ID).First(&outletGroup).Error
	if err != nil {
		return err
	}

	outletGroup.Name = req.Name
	outletGroup.Description = req.Description

	err = s.db.Save(&outletGroup).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=GET path=/api/admin/outlet/outlet-group/get/:business_id
func (s *Service) GetOutletGroups(ctx context.Context, business_id uuid.UUID) (*OutletGroupResponse, error) {
	if err := middleware.CheckPermission(constants.CreateOutletAction, &business_id, nil); err != nil {
		return nil, err
	}

	var outletGroups []models.OutletGroup
	err := s.db.Model(&models.OutletGroup{}).Preload("Outlets").Preload("Users.GroupRole.Role").Where("business_id = ?", business_id).Order("created_at asc").Find(&outletGroups).Error
	if err != nil {
		return nil, err
	}

	return &OutletGroupResponse{
		OutletGroups: outletGroups,
	}, nil
}

//encore:api auth method=GET path=/api/admin/outlet/outlet-group/get-by-user/:user_id
func (s *Service) GetOutletGroupsByUser(ctx context.Context, user_id uuid.UUID) (*OutletGroupResponse, error) {
	var outletGroups []models.OutletGroup
	err := s.db.Joins("JOIN outlet_groups_users ogu ON ogu.outlet_group_id = outlet_groups.id").
		Joins("JOIN users ON users.id = ogu.user_id").
		Where("users.id = ?", user_id).
		Preload("Outlets").
		Preload("Users").
		Find(&outletGroups).Error
	if err != nil {
		return nil, err
	}

	if len(outletGroups) == 0 {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "No outlet group found for this user",
		}
	}

	outletOptions := []OutletOption{}
	for _, outletGroup := range outletGroups {
		for _, outlet := range outletGroup.Outlets {
			outletOptions = append(outletOptions, OutletOption{
				ID:   outlet.ID,
				Name: outlet.Name,
			})
		}
	}

	return &OutletGroupResponse{
		OutletGroups:  outletGroups,
		OutletOptions: outletOptions,
	}, nil
}

//encore:api auth method=POST path=/api/admin/outlet/outlet-group/assign-outlet
func (s *Service) AssignOutletToGroup(ctx context.Context, req *AssignOutletToGroupRequest) error {
	var outletGroup models.OutletGroup
	err := s.db.Preload("Outlets").Where("id = ?", req.OutletGroupID).First(&outletGroup).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.CreateOutletAction, &outletGroup.BusinessID, nil); err != nil {
		return err
	}

	// Check if the OutletID is already in the Outlets list of the OutletGroup
	for _, outlet := range outletGroup.Outlets {
		if outlet.ID == req.OutletID {
			return &errs.Error{
				Code:    errs.AlreadyExists,
				Message: "This outlet is already assigned to this group",
			}
		}
	}

	// Add the OutletID to the Outlets list of the OutletGroup
	outletGroup.Outlets = append(outletGroup.Outlets, &models.Outlet{
		ID: req.OutletID,
	})

	err = s.db.Save(&outletGroup).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=POST path=/api/admin/outlet/outlet-group/unassign-outlet
func (s *Service) UnassignOutletFromGroup(ctx context.Context, req *AssignOutletToGroupRequest) error {
	var outletGroup models.OutletGroup
	err := s.db.Preload("Outlets").Where("id = ?", req.OutletGroupID).First(&outletGroup).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.CreateOutletAction, &outletGroup.BusinessID, nil); err != nil {
		return err
	}

	// Remove the outlet from the outlet_groups_outlets join table
	err = s.db.Model(&outletGroup).Association("Outlets").Delete(&models.Outlet{ID: req.OutletID})
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=POST path=/api/admin/outlet/outlet-group/assign-user
func (s *Service) AssignUserToGroup(ctx context.Context, req *AssignUserToGroupRequest) error {
	var outletGroup models.OutletGroup
	err := s.db.Preload("Users").Where("id = ?", req.OutletGroupID).First(&outletGroup).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.CreateOutletAction, &outletGroup.BusinessID, nil); err != nil {
		return err
	}

	// Check if the OutletID is already in the Outlets list of the OutletGroup
	for _, user := range outletGroup.Users {
		if user.ID == req.UserID {
			return &errs.Error{
				Code:    errs.AlreadyExists,
				Message: "This user is already assigned to this group",
			}
		}
	}

	// Add the UserID to the Users list of the OutletGroup
	outletGroup.Users = append(outletGroup.Users, &models.User{
		ID: req.UserID,
	})

	err = s.db.Save(&outletGroup).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=POST path=/api/admin/outlet/outlet-group/unassign-user
func (s *Service) UnassignUserFromGroup(ctx context.Context, req *AssignUserToGroupRequest) error {
	var outletGroup models.OutletGroup
	err := s.db.Preload("Users").Where("id = ?", req.OutletGroupID).First(&outletGroup).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.CreateOutletAction, &outletGroup.BusinessID, nil); err != nil {
		return err
	}

	// Remove the user from the outlet_groups_users join table
	err = s.db.Model(&outletGroup).Association("Users").Delete(&models.User{ID: req.UserID})
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=DELETE path=/api/admin/outlet/outlet-group/delete/:outlet_group_id
func (s *Service) DeleteOutletGroup(ctx context.Context, outlet_group_id uuid.UUID) error {
	var outletGroup models.OutletGroup
	err := s.db.Where("id = ?", outlet_group_id).First(&outletGroup).Error
	if err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.CreateOutletAction, &outletGroup.BusinessID, nil); err != nil {
		return err
	}

	// First, delete associations in join tables (foreign keys)
	if err := s.db.Model(&outletGroup).Association("Outlets").Clear(); err != nil {
		return err
	}
	if err := s.db.Model(&outletGroup).Association("Users").Clear(); err != nil {
		return err
	}
	// Now, delete the outlet group itself
	err = s.db.Unscoped().Delete(&outletGroup).Error
	if err != nil {
		return err
	}

	return nil
}

// encore:api auth method=POST path=/api/admin/outlet/outlet-terminal/map-vt
func (s *Service) MapVTIDToOutletTerminal(ctx context.Context, req *VTToOutletTerminalMap) (*VTToOutletTerminalMapResponse, error) {
	var outletTerminal models.OutletTerminal
	err := s.db.Where("vt_id = ?", req.VtID).First(&outletTerminal).Error

	if err == nil && !req.Confirmation {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "Fiuu VT is already mapped to this outlet terminal. Please confirm to remap.",
		}
	} else if err == nil && req.Confirmation {
		err = s.db.Unscoped().Delete(&outletTerminal).Error
		if err != nil {
			return nil, err
		}
	}

	outletTerminal = models.OutletTerminal{
		OutletID: req.OutletID,
		VtID:     req.VtID,
	}

	err = s.db.Create(&outletTerminal).Error
	if err != nil {
		return nil, err
	}

	return &VTToOutletTerminalMapResponse{
		OutletTerminalID: outletTerminal.ID,
		VtID:             outletTerminal.VtID,
	}, nil
}

// APi to create outlet operation time

//encore:api auth method=PUT path=/api/admin/outlet/merchant-secrets/bind-platform-id
func (s *Service) BindPlatformIDToMerchantSecret(ctx context.Context, req *BindPlatformIDRequest) (*BindPlatformIDResponse, error) {
	if req.OutletID == uuid.Nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "Outlet ID is required"}
	}
	if req.ShopeeStoreID == nil && req.GrabStoreID == nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: "At least ShopeeStoreID or GrabStoreID must be provided"}
	}

	var merchantSecret models.MerchantSecret
	err := s.db.Where("outlet_id = ?", req.OutletID).First(&merchantSecret).Error
	if err != nil {
		return nil, err
	}

	if req.ShopeeStoreID != nil {
		merchantSecret.ShopeeStoreID = req.ShopeeStoreID
	}
	if req.GrabStoreID != nil {
		merchantSecret.GrabStoreID = req.GrabStoreID
	}

	if err := s.db.Save(&merchantSecret).Error; err != nil {
		return nil, err
	}

	return &BindPlatformIDResponse{Message: "Platform ID(s) bound successfully"}, nil
}
