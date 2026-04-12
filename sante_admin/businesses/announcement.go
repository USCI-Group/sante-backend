package businesses

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/types/uuid"
)

type GetAnnouncementListResponse struct {
	AnnouncementList []models.Announcement `json:"announcement_list"`
}

type CreateAnnouncementRequest struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	ImageURL   string    `json:"image_url"`
	IsActive   bool      `json:"is_active"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
}

type GetAnnouncementListRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	Search     *string   `json:"search"`
	Filter     *struct {
		Status *string `json:"status"`
	} `json:"filter"`
}

type UploadAnnouncementImageRequest struct {
	AnnouncementID uuid.UUID             `json:"announcement_id"`
	File           multipart.File        `json:"file"`
	Header         *multipart.FileHeader `json:"header"`
}

// API to get announcement list
//
//encore:api public method=GET path=/api/admin/business/announcement/get-all/:business_id
func (s *Service) GetActiveAnnouncementList(ctx context.Context, business_id uuid.UUID) (*GetAnnouncementListResponse, error) {
	now := time.Now()

	var announcementList []models.Announcement
	result := s.db.Model(&models.Announcement{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Where("start_date <= ? AND end_date >= ?", now, now).
		Order("is_active DESC, sort_order ASC").
		Find(&announcementList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAnnouncementListResponse{
		AnnouncementList: announcementList,
	}, nil
}

//encore:api public method=POST path=/api/admin/business/announcement/get-with-filters
func (s *Service) GetAnnouncementList(ctx context.Context, req *GetAnnouncementListRequest) (*GetAnnouncementListResponse, error) {
	var announcementList []models.Announcement
	query := s.db.Model(&models.Announcement{}).
		Where("business_id = ?", req.BusinessID).
		Order("is_active DESC, start_date ASC, end_date ASC")

	if req.Search != nil {
		query = query.Where("LOWER(title) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", "%"+*req.Search+"%", "%"+*req.Search+"%")
	}

	if req.Filter != nil && req.Filter.Status != nil {
		switch *req.Filter.Status {
		case "active":
			query = query.Where("is_active = ?", true)
			query = query.Where("end_date >= ?", time.Now())
		case "inactive":
			query = query.Where("is_active = ?", false)
			query = query.Where("end_date >= ?", time.Now())
		case "expired":
			query = query.Where("end_date < ?", time.Now())
		}
	}

	result := query.Find(&announcementList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAnnouncementListResponse{
		AnnouncementList: announcementList,
	}, nil
}

//encore:api auth raw method=POST path=/api/admin/business/announcement/create
func (s *Service) CreateAnnouncement(w http.ResponseWriter, req *http.Request) {
	business_id := req.FormValue("business_id")
	is_active := req.FormValue("is_active")
	title := req.FormValue("title")
	description := req.FormValue("description")
	start_date := req.FormValue("start_date")
	end_date := req.FormValue("end_date")

	if business_id == "" {
		http.Error(w, "business_id is required", http.StatusBadRequest)
		return
	}

	if is_active == "" {
		http.Error(w, "is_active is required", http.StatusBadRequest)
		return
	}
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	if description == "" {
		http.Error(w, "description is required", http.StatusBadRequest)
		return
	}
	if start_date == "" {
		http.Error(w, "start_date is required", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(time.RFC3339, start_date); err != nil {
		http.Error(w, "start_date format is invalid, must be Date format or RFC3339", http.StatusBadRequest)
		return
	}
	if end_date == "" {
		http.Error(w, "end_date is required", http.StatusBadRequest)
		return
	}
	if _, err := time.Parse(time.RFC3339, end_date); err != nil {
		http.Error(w, "end_date format is invalid, must be Date format or RFC3339", http.StatusBadRequest)
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

	if err := middleware.CheckPermission(constants.ManageAnnouncementAction, &business.ID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	var total_announcements int64
	s.db.Model(&models.Announcement{}).Where("business_id = ?", business_id).Count(&total_announcements)

	startDate, _ := time.Parse(time.RFC3339, start_date)
	endDate, _ := time.Parse(time.RFC3339, end_date)

	startDate, _ = common.GetStartOfDay(startDate)
	endDate, _ = common.GetEndOfDay(endDate)

	announcement := models.Announcement{
		BusinessID:  business.ID,
		Title:       title,
		Description: description,
		IsActive:    is_active == "true",
		StartDate:   startDate,
		EndDate:     endDate,
		SortOrder:   int(total_announcements) + 1,
	}

	err = s.db.Create(&announcement).Error
	if err != nil {
		http.Error(w, "Failed to create announcement", http.StatusInternalServerError)
		return
	}

	err = s.uploadAnnouncementImage(&UploadAnnouncementImageRequest{
		AnnouncementID: announcement.ID,
		File:           file,
		Header:         header,
	})
	if err != nil {
		http.Error(w, "Failed to upload announcement image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth raw method=POST path=/api/admin/business/announcement/update
func (s *Service) UpdateAnnouncement(w http.ResponseWriter, req *http.Request) {
	announcement_id := req.FormValue("id")
	is_active := req.FormValue("is_active")
	title := req.FormValue("title")
	description := req.FormValue("description")
	start_date := req.FormValue("start_date")
	end_date := req.FormValue("end_date")

	if announcement_id == "" {
		http.Error(w, "id(announcement_id) is required", http.StatusBadRequest)
		return
	}

	var announcement models.Announcement
	err := s.db.Where("id = ?", announcement_id).First(&announcement).Error
	if err != nil {
		http.Error(w, "Announcement not found", http.StatusNotFound)
		return
	}

	if err := middleware.CheckPermission(constants.ManageAnnouncementAction, &announcement.BusinessID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	announcement.IsActive = is_active == "true"

	if title != "" {
		announcement.Title = title
	}
	if description != "" {
		announcement.Description = description
	}
	if start_date != "" {
		if _, err := time.Parse(time.RFC3339, start_date); err != nil {
			http.Error(w, "start_date format is invalid, must be RFC3339", http.StatusBadRequest)
			return
		}
		startDate, _ := time.Parse(time.RFC3339, start_date)
		startDate, _ = common.GetStartOfDay(startDate)
		announcement.StartDate = startDate
	}
	if end_date != "" {
		if _, err := time.Parse(time.RFC3339, end_date); err != nil {
			http.Error(w, "end_date format is invalid, must be RFC3339", http.StatusBadRequest)
			return
		}
		endDate, _ := time.Parse(time.RFC3339, end_date)
		endDate, _ = common.GetEndOfDay(endDate)
		announcement.EndDate = endDate
	}

	err = s.db.Save(&announcement).Error
	if err != nil {
		http.Error(w, "Failed to update announcement", http.StatusInternalServerError)
		return
	}

	file, header, _ := req.FormFile("file")

	if file != nil && header != nil && header.Filename != "" {
		// delete old image from s3
		if announcement.ImageURL != "" {
			aws_s3.DeleteDocument(announcement.ImageURL)
		}

		err = s.uploadAnnouncementImage(&UploadAnnouncementImageRequest{
			AnnouncementID: announcement.ID,
			File:           file,
			Header:         header,
		})
		defer file.Close()
		if err != nil {
			http.Error(w, "Failed to upload announcement image", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth method=DELETE path=/api/admin/business/announcement/delete/:announcement_id
func (s *Service) DeleteAnnouncement(ctx context.Context, announcement_id uuid.UUID) error {
	var announcement models.Announcement
	err := s.db.Where("id = ?", announcement_id).First(&announcement).Error
	if err != nil {
		return err
	}
	if err := middleware.CheckPermission(constants.ManageAnnouncementAction, &announcement.BusinessID, nil); err != nil {
		return err
	}

	// remove old image from s3
	if announcement.ImageURL != "" {
		aws_s3.DeleteDocument(announcement.ImageURL)
	}

	err = s.db.Delete(&announcement).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) uploadAnnouncementImage(req *UploadAnnouncementImageRequest) error {
	var announcement models.Announcement
	err := s.db.Where("id = ?", req.AnnouncementID).First(&announcement).Error
	if err != nil {
		return err
	}

	business := models.Business{}
	err = s.db.Where("id = ?", announcement.BusinessID).First(&business).Error
	if err != nil {
		return err
	}

	var document aws_s3.Document
	document.File = req.File
	file_extension := filepath.Ext(req.Header.Filename)
	currentTime := time.Now().Unix()
	file_name := fmt.Sprintf("%s_%d%s", announcement.ID, currentTime, file_extension)
	document.DocPath = "business/" + business.RegistrationNumber + "/announcement/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		return err
	}

	// remove old image from s3
	if announcement.ImageURL != "" {
		aws_s3.DeleteDocument(announcement.ImageURL)
	}

	announcement.ImageURL = document_res.Url

	err = s.db.Save(&announcement).Error
	if err != nil {
		return err
	}

	return nil
}
