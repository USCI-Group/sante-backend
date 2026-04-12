package businesses

import (
	"context"
	"fmt"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/auth_service"
	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
)

type GetFeedbackQuestionListResponse struct {
	Meta                 common.Pagination         `json:"meta"`
	FeedbackQuestionList []models.FeedbackQuestion `json:"feedback_question_list"`
}

type GetFeedbackQuestionListRequest struct {
	BusinessID uuid.UUID  `json:"business_id"`
	Search     *string    `json:"search"`
	DateFilter *time.Time `json:"date_filter"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

type UploadFeedbackQuestionImageRequest struct {
	FeedbackQuestionID uuid.UUID             `json:"feedback_question_id"`
	File               multipart.File        `json:"file"`
	Header             *multipart.FileHeader `json:"header"`
}

type AddCustomerFeedbackRequest struct {
	FeedbackQuestionID uuid.UUID `json:"feedback_question_id"`
	Rating             int       `json:"rating" validate:"required,min=1,max=5"`
	Comment            string    `json:"comment"`
}

type GetCustomerFeedbackResponse struct {
	FeedbackList []models.Feedback `json:"feedback_list"`
}

// API to get feedback question list
//
//encore:api public method=GET path=/api/admin/business/feedback/question/get-all/:business_id
func (s *Service) GetActiveFeedbackQuestionList(ctx context.Context, business_id uuid.UUID) (*GetFeedbackQuestionListResponse, error) {

	var feedbackQuestionList []models.FeedbackQuestion
	result := s.db.Model(&models.FeedbackQuestion{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Where("question != ?", models.LeavingReasonFeedbackQuestion).
		Order("is_active DESC, sort_order ASC").
		Find(&feedbackQuestionList)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetFeedbackQuestionListResponse{
		FeedbackQuestionList: feedbackQuestionList,
	}, nil
}

//encore:api auth method=POST path=/api/admin/business/feedback/question/get-with-filters
func (s *Service) GetFeedbackQuestionList(ctx context.Context, req *GetFeedbackQuestionListRequest) (*GetFeedbackQuestionListResponse, error) {

	if err := middleware.CheckPermission(constants.ManageFeedbackAction, &req.BusinessID, nil); err != nil {
		return nil, err
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	var feedbackQuestionList []models.FeedbackQuestion

	query := s.db.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize)
	queryCount := s.db.Model(&models.FeedbackQuestion{})

	if req.DateFilter != nil {
		startOfDay, err := common.GetStartOfDay(*req.DateFilter)
		if err != nil {
			return nil, err
		}
		endOfDay, err := common.GetEndOfDay(*req.DateFilter)
		if err != nil {
			return nil, err
		}
		query = query.Where("created_at >= ?", startOfDay)
		queryCount = queryCount.Where("created_at >= ?", startOfDay)
		query = query.Where("created_at <= ?", endOfDay)
		queryCount = queryCount.Where("created_at <= ?", endOfDay)
	}

	query = query.
		Where("business_id = ?", req.BusinessID).
		Where("question != ?", models.LeavingReasonFeedbackQuestion).
		Order("is_active DESC, sort_order ASC")
	queryCount = queryCount.
		Where("business_id = ?", req.BusinessID).
		Where("question != ?", models.LeavingReasonFeedbackQuestion)

	if req.Search != nil {
		query = query.Where("question ILIKE ? OR section ILIKE ?", "%"+*req.Search+"%", "%"+*req.Search+"%")
		queryCount = queryCount.Where("question ILIKE ? OR section ILIKE ?", "%"+*req.Search+"%", "%"+*req.Search+"%")
	}

	result := query.Find(&feedbackQuestionList)
	if result.Error != nil {
		return nil, result.Error
	}

	var total int64
	queryCount.Count(&total)

	return &GetFeedbackQuestionListResponse{
		Meta: common.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
			Total:      total,
		},
		FeedbackQuestionList: feedbackQuestionList,
	}, nil
}

//encore:api auth raw method=POST path=/api/admin/business/feedback/question/create
func (s *Service) CreateFeedbackQuestion(w http.ResponseWriter, req *http.Request) {
	business_id := req.FormValue("business_id")
	question := req.FormValue("question")
	section := req.FormValue("section")
	is_active := req.FormValue("is_active")

	if question == "" {
		http.Error(w, "question is required", http.StatusBadRequest)
		return
	}
	if section == "" {
		http.Error(w, "section is required", http.StatusBadRequest)
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

	if err := middleware.CheckPermission(constants.ManageFeedbackAction, &business.ID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	var total_feedback_questions int64
	s.db.Model(&models.FeedbackQuestion{}).Where("business_id = ?", business_id).Count(&total_feedback_questions)

	log.Println("header", header)

	feedbackQuestion := models.FeedbackQuestion{
		BusinessID: business.ID,
		Question:   question,
		Section:    strings.ToLower(section),
		IsActive:   is_active == "true",
		SortOrder:  int(total_feedback_questions) + 1,
	}

	err = s.db.Create(&feedbackQuestion).Error
	if err != nil {
		http.Error(w, "Failed to create feedback question", http.StatusInternalServerError)
		return
	}

	err = s.uploadFeedbackQuestionImage(&UploadFeedbackQuestionImageRequest{
		FeedbackQuestionID: feedbackQuestion.ID,
		File:               file,
		Header:             header,
	})
	if err != nil {
		http.Error(w, "Failed to upload feedback question image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth raw method=POST path=/api/admin/business/feedback/question/update
func (s *Service) UpdateFeedbackQuestion(w http.ResponseWriter, req *http.Request) {
	feedback_question_id := req.FormValue("id")
	is_active := req.FormValue("is_active")
	question := req.FormValue("question")
	section := req.FormValue("section")

	if feedback_question_id == "" {
		http.Error(w, "id(feedback_question_id) is required", http.StatusBadRequest)
		return
	}
	if question == "" {
		http.Error(w, "question is required", http.StatusBadRequest)
		return
	}
	if section == "" {
		http.Error(w, "section is required", http.StatusBadRequest)
		return
	}

	var feedbackQuestion models.FeedbackQuestion
	err := s.db.Where("id = ?", feedback_question_id).First(&feedbackQuestion).Error
	if err != nil {
		http.Error(w, "Feedback question not found", http.StatusNotFound)
		return
	}

	if err := middleware.CheckPermission(constants.ManageFeedbackAction, &feedbackQuestion.BusinessID, nil); err != nil {
		http.Error(w, "Permission denied", http.StatusForbidden)
		return
	}

	feedbackQuestion.IsActive = is_active == "true"
	feedbackQuestion.Question = question
	feedbackQuestion.Section = strings.ToLower(section)

	err = s.db.Save(&feedbackQuestion).Error
	if err != nil {
		http.Error(w, "Failed to update feedback question", http.StatusInternalServerError)
		return
	}

	file, header, _ := req.FormFile("file")

	if file != nil && header != nil && header.Filename != "" {
		// delete old image from s3
		if feedbackQuestion.ImageURL != "" {
			aws_s3.DeleteDocument(feedbackQuestion.ImageURL)
		}

		err = s.uploadFeedbackQuestionImage(&UploadFeedbackQuestionImageRequest{
			FeedbackQuestionID: feedbackQuestion.ID,
			File:               file,
			Header:             header,
		})
		defer file.Close()
		if err != nil {
			http.Error(w, "Failed to upload feedback question image", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

//encore:api auth method=DELETE path=/api/admin/business/feedback/question/delete/:feedback_question_id
func (s *Service) DeleteFeedbackQuestion(ctx context.Context, feedback_question_id uuid.UUID) error {
	var feedbackQuestion models.FeedbackQuestion
	err := s.db.Where("id = ?", feedback_question_id).First(&feedbackQuestion).Error
	if err != nil {
		return err
	}
	if err := middleware.CheckPermission(constants.ManageFeedbackAction, &feedbackQuestion.BusinessID, nil); err != nil {
		return err
	}

	// remove old image from s3
	if feedbackQuestion.ImageURL != "" {
		aws_s3.DeleteDocument(feedbackQuestion.ImageURL)
	}

	err = s.db.Delete(&feedbackQuestion).Error
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) uploadFeedbackQuestionImage(req *UploadFeedbackQuestionImageRequest) error {
	var feedbackQuestion models.FeedbackQuestion
	err := s.db.Where("id = ?", req.FeedbackQuestionID).First(&feedbackQuestion).Error
	if err != nil {
		return err
	}

	business := models.Business{}
	err = s.db.Where("id = ?", feedbackQuestion.BusinessID).First(&business).Error
	if err != nil {
		return err
	}

	var document aws_s3.Document
	document.File = req.File
	file_extension := filepath.Ext(req.Header.Filename)
	currentTime := time.Now().Unix()
	file_name := fmt.Sprintf("%s_%d%s", feedbackQuestion.ID, currentTime, file_extension)
	document.DocPath = "business/" + business.RegistrationNumber + "/feedback_question/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		return err
	}

	// remove old image from s3
	if feedbackQuestion.ImageURL != "" {
		aws_s3.DeleteDocument(feedbackQuestion.ImageURL)
	}

	feedbackQuestion.ImageURL = document_res.Url

	err = s.db.Save(&feedbackQuestion).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=POST path=/api/customer/feedback/add
func (s *Service) AddCustomerFeedback(ctx context.Context, req *AddCustomerFeedbackRequest) error {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return err
	}

	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return err
	}

	var customerFeedback models.Feedback

	customerFeedback.CustomerID = user.ID
	customerFeedback.FeedbackQuestionID = req.FeedbackQuestionID
	customerFeedback.Rating = &req.Rating
	customerFeedback.Comment = req.Comment

	err = s.db.Create(&customerFeedback).Error
	if err != nil {
		return err
	}

	return nil
}

//encore:api auth method=GET path=/api/customer/feedback/get/:customer_id
func (s *Service) GetCustomerFeedback(ctx context.Context, customer_id uuid.UUID) (*GetCustomerFeedbackResponse, error) {
	var feedbackList []models.Feedback
	err := s.db.Where("customer_id = ?", customer_id).Preload("FeedbackQuestion").Find(&feedbackList).Error
	if err != nil {
		return nil, err
	}

	return &GetCustomerFeedbackResponse{
		FeedbackList: feedbackList,
	}, nil
}
