package businesses

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
)

type GetOnboardingListResponse struct {
	OnboardingList []models.Onboarding `json:"onboarding_list"`
}

type CreateOnboardingRequest struct {
	ID          uuid.UUID `json:"id"`
	BusinessID  uuid.UUID `json:"business_id"`
	Title       string    `json:"title" valid:"required~Title is required"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
}

type GetOnboardingListRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	Search     *string   `json:"search"`
}

type UpdateOnboardingMultipleRequest struct {
	OnboardingList []models.Onboarding `json:"onboarding_list"`
}

// API to get onboarding list
//
//encore:api public method=GET path=/api/admin/business/onboarding/get-all/:business_id
func (s *Service) GetActiveOnboardingList(ctx context.Context, business_id uuid.UUID) (*GetOnboardingListResponse, error) {
	var onboardingList []models.Onboarding
	result := s.db.Model(&models.Onboarding{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Order("is_active DESC, sort_order ASC").
		Find(&onboardingList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetOnboardingListResponse{
		OnboardingList: onboardingList,
	}, nil
}

//encore:api public method=POST path=/api/admin/business/onboarding/get-with-filters
func (s *Service) GetOnboardingList(ctx context.Context, req *GetOnboardingListRequest) (*GetOnboardingListResponse, error) {
	var onboardingList []models.Onboarding
	query := s.db.Model(&models.Onboarding{}).
		Where("business_id = ?", req.BusinessID).
		Order("is_active DESC, sort_order ASC")

	if req.Search != nil {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+*req.Search+"%", "%"+*req.Search+"%")
	}

	result := query.Find(&onboardingList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetOnboardingListResponse{
		OnboardingList: onboardingList,
	}, nil
}

//encore:api auth method=POST path=/api/admin/business/onboarding/create
func (s *Service) CreateOnboarding(ctx context.Context, req *CreateOnboardingRequest) (*models.Onboarding, error) {
	if err := middleware.CheckPermission(constants.ManageOnboardingAction, &req.BusinessID, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var total_onboardings int64
	s.db.Model(&models.Onboarding{}).Where("business_id = ?", req.BusinessID).Count(&total_onboardings)

	onboarding := models.Onboarding{
		BusinessID:  req.BusinessID,
		Title:       req.Title,
		Description: req.Description,
		IsActive:    req.IsActive,
		SortOrder:   int(total_onboardings) + 1,
	}

	err := s.db.Create(&onboarding).Error
	if err != nil {
		return nil, err
	}

	return &onboarding, nil
}

//encore:api auth method=POST path=/api/admin/business/onboarding/update
func (s *Service) UpdateOnboarding(ctx context.Context, req *CreateOnboardingRequest) error {
	if err := middleware.CheckPermission(constants.ManageOnboardingAction, &req.BusinessID, nil); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return err
	}

	var onboarding models.Onboarding
	err := s.db.Where("id = ?", req.ID).First(&onboarding).Error
	if err != nil {
		return err
	}

	onboarding.Title = req.Title
	onboarding.Description = req.Description
	onboarding.IsActive = req.IsActive

	err = s.db.Save(&onboarding).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=POST path=/api/admin/business/onboarding/update-multiple
func (s *Service) UpdateOnboardingMultiple(ctx context.Context, req *UpdateOnboardingMultipleRequest) error {
	if err := middleware.CheckPermission(constants.ManageOnboardingAction, &req.OnboardingList[0].BusinessID, nil); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return err
	}

	for _, onboarding := range req.OnboardingList {
		err := s.db.Save(&onboarding).Error
		if err != nil {
			return err
		}
	}

	return nil
}

//encore:api auth method=DELETE path=/api/admin/business/onboarding/delete/:onboarding_id
func (s *Service) DeleteOnboarding(ctx context.Context, onboarding_id uuid.UUID) error {
	var onboarding models.Onboarding
	err := s.db.Where("id = ?", onboarding_id).First(&onboarding).Error
	if err != nil {
		return err
	}
	if err := middleware.CheckPermission(constants.ManageOnboardingAction, &onboarding.BusinessID, nil); err != nil {
		return err
	}

	// remove old image from s3
	if onboarding.ImageURL != "" {
		aws_s3.DeleteDocument(onboarding.ImageURL)
	}

	err = s.db.Delete(&onboarding).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth raw method=POST path=/api/admin/business/onboarding/upload-image
func (s *Service) UploadOnboardingImage(w http.ResponseWriter, req *http.Request) {
	business_id := req.FormValue("business_id")
	onboarding_id := req.FormValue("onboarding_id")

	if business_id == "" {
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}
	if onboarding_id == "" {
		http.Error(w, "Onboarding ID is required", http.StatusBadRequest)
		return
	}

	business := models.Business{}
	err := s.db.Where("id = ?", business_id).First(&business).Error
	if err != nil {
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	if err := middleware.CheckPermission(constants.ManageOnboardingAction, &business.ID, nil); err != nil {
		errs.HTTPError(w, err)
		return
	}

	var onboarding models.Onboarding
	err = s.db.Where("id = ?", onboarding_id).First(&onboarding).Error
	if err != nil {
		http.Error(w, "Onboarding not found", http.StatusNotFound)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(3 << 20); err != nil { // Limit to 3 MB
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
	currentTime := time.Now().Unix()
	file_name := fmt.Sprintf("%s_%d%s", onboarding.Title, currentTime, file_extension)
	document.DocPath = "business/" + business.RegistrationNumber + "/onboarding/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// remove old image from s3
	if onboarding.ImageURL != "" {
		aws_s3.DeleteDocument(onboarding.ImageURL)
	}

	onboarding.ImageURL = document_res.Url

	err = s.db.Save(&onboarding).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
