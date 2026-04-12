package businesses

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type GetDeliveryListResponse struct {
	DeliveryList []models.Delivery `json:"delivery_list"`
}

type CreateDeliveryRequest struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	ImageURL   string    `json:"image_url"`
	IsActive   bool      `json:"is_active"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
}

type GetDeliveryListRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	Search     *string   `json:"search"`
	Filter     *struct {
		DeliveryType *models.DeliveryType `json:"delivery_type"`
	} `json:"filter"`
}

type UploadDeliveryImageRequest struct {
	DeliveryID uuid.UUID             `json:"delivery_id"`
	File       multipart.File        `json:"file"`
	Header     *multipart.FileHeader `json:"header"`
}

// API to get delivery list
//
//encore:api public method=GET path=/api/admin/business/delivery/get-all/:business_id
func (s *Service) GetActiveDeliveryList(ctx context.Context, business_id uuid.UUID) (*GetDeliveryListResponse, error) {
	var deliveryList []models.Delivery
	result := s.db.Model(&models.Delivery{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Find(&deliveryList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetDeliveryListResponse{
		DeliveryList: deliveryList,
	}, nil
}

//encore:api public method=POST path=/api/admin/business/delivery/get-with-filters
func (s *Service) GetDeliveryList(ctx context.Context, req *GetDeliveryListRequest) (*GetDeliveryListResponse, error) {
	var deliveryList []models.Delivery
	query := s.db.Model(&models.Delivery{}).
		Where("business_id = ?", req.BusinessID).
		Order("created_at ASC")

	if req.Filter != nil && req.Filter.DeliveryType != nil {
		query = query.Where("delivery_type = ?", req.Filter.DeliveryType)
	}

	result := query.Find(&deliveryList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetDeliveryListResponse{
		DeliveryList: deliveryList,
	}, nil
}

//encore:api auth raw method=POST path=/api/admin/business/delivery/create
func (s *Service) CreateDelivery(w http.ResponseWriter, req *http.Request) {
	business_id := req.FormValue("business_id")
	is_active := req.FormValue("is_active")
	delivery_type := req.FormValue("delivery_type")

	if business_id == "" {
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	if is_active == "" {
		http.Error(w, "is_active is required", http.StatusBadRequest)
		return
	}
	if delivery_type == "" {
		http.Error(w, "delivery_type is required", http.StatusBadRequest)
		return
	}

	// Get the file
	file, header, err := req.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	const maxFileSize = 3 * 1024 * 1024 // 3 MB

	if req.ContentLength > 0 && req.ContentLength > maxFileSize {
		http.Error(w, "file size exceeds 3 MB limit", http.StatusBadRequest)
		return
	}

	business := models.Business{}
	if err = s.db.Where("id = ?", business_id).First(&business).Error; err != nil {
		http.Error(w, "Business not found", http.StatusNotFound)
		return
	}

	if err := middleware.CheckPermission(constants.ManageOrderMethodAction, &business.ID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	delivery := models.Delivery{
		BusinessID:   business.ID,
		IsActive:     is_active == "true",
		DeliveryType: models.DeliveryType(delivery_type),
	}

	existingActive := models.Delivery{}
	err = s.db.Where("business_id = ? AND is_active = ? AND delivery_type = ?", business_id, true, delivery_type).First(&existingActive).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		delivery.IsActive = true
	}

	err = s.db.Create(&delivery).Error
	if err != nil {
		http.Error(w, "Failed to create delivery", http.StatusInternalServerError)
		return
	}

	err = s.uploadDeliveryImage(&UploadDeliveryImageRequest{
		DeliveryID: delivery.ID,
		File:       file,
		Header:     header,
	})
	if err != nil {
		http.Error(w, "Failed to upload delivery image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(delivery)
}

//encore:api auth raw method=POST path=/api/admin/business/delivery/update
func (s *Service) UpdateDelivery(w http.ResponseWriter, req *http.Request) {
	delivery_id := req.FormValue("id")
	is_active := req.FormValue("is_active")
	delivery_type := req.FormValue("delivery_type")

	if delivery_id == "" {
		http.Error(w, "id(delivery_id) is required", http.StatusBadRequest)
		return
	}

	var delivery models.Delivery
	err := s.db.Where("id = ?", delivery_id).First(&delivery).Error
	if err != nil {
		http.Error(w, "Delivery not found", http.StatusNotFound)
		return
	}

	if err := middleware.CheckPermission(constants.ManageOrderMethodAction, &delivery.BusinessID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	delivery.IsActive = is_active == "true"

	delivery.DeliveryType = models.DeliveryType(delivery_type)

	err = s.db.Save(&delivery).Error
	if delivery.IsActive {
		// update all other delivery to inactive
		err = s.db.Model(&models.Delivery{}).
			Where("business_id = ?", delivery.BusinessID).
			Where("id != ?", delivery.ID).
			Where("delivery_type = ?", delivery.DeliveryType).
			Where("is_active = ?", true).
			Update("is_active", false).Error

	}
	if err != nil {
		http.Error(w, "Failed to update delivery", http.StatusInternalServerError)
		return
	}

	file, header, _ := req.FormFile("file")

	if file != nil && header != nil && header.Filename != "" {
		// delete old image from s3
		if delivery.ImageURL != "" {
			aws_s3.DeleteDocument(delivery.ImageURL)
		}

		err = s.uploadDeliveryImage(&UploadDeliveryImageRequest{
			DeliveryID: delivery.ID,
			File:       file,
			Header:     header,
		})
		defer file.Close()
		if err != nil {
			http.Error(w, "Failed to upload delivery image", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth method=DELETE path=/api/admin/business/delivery/delete/:delivery_id
func (s *Service) DeleteDelivery(ctx context.Context, delivery_id uuid.UUID) error {
	var delivery models.Delivery
	err := s.db.Where("id = ?", delivery_id).First(&delivery).Error
	if err != nil {
		return err
	}
	if err := middleware.CheckPermission(constants.ManageOrderMethodAction, &delivery.BusinessID, nil); err != nil {
		return err
	}

	// remove old image from s3
	if delivery.ImageURL != "" {
		aws_s3.DeleteDocument(delivery.ImageURL)
	}

	err = s.db.Delete(&delivery).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) uploadDeliveryImage(req *UploadDeliveryImageRequest) error {
	var delivery models.Delivery
	err := s.db.Where("id = ?", req.DeliveryID).First(&delivery).Error
	if err != nil {
		return err
	}

	business := models.Business{}
	err = s.db.Where("id = ?", delivery.BusinessID).First(&business).Error
	if err != nil {
		return err
	}

	var document aws_s3.Document
	document.File = req.File
	file_extension := filepath.Ext(req.Header.Filename)
	currentTime := time.Now().Unix()
	file_name := fmt.Sprintf("%s_%d%s", delivery.ID, currentTime, file_extension)
	document.DocPath = "business/" + business.RegistrationNumber + "/delivery/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		return err
	}

	// remove old image from s3
	if delivery.ImageURL != "" {
		aws_s3.DeleteDocument(delivery.ImageURL)
	}

	delivery.ImageURL = document_res.Url

	err = s.db.Save(&delivery).Error
	if err != nil {
		return err
	}

	return nil
}
