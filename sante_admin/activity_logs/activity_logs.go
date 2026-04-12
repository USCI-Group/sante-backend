package activity_logs

import (
	"context"
	"math"
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/middleware"
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

type GetActivityLogsParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Filter   struct {
		FromDate *time.Time                   `json:"from_date"`
		ToDate   *time.Time                   `json:"to_date"`
		ActionBy *string                      `json:"action_by"`
		Status   *constants.ActivityLogStatus `json:"status"`
	} `json:"filter"`
}

type ActivityLogResponse struct {
	Meta common.Pagination    `json:"meta"`
	Data []models.ActivityLog `json:"data"`
}

//encore:api auth method=POST path=/api/admin/activity-log/get-all
func (s *Service) GetActivityLogs(ctx context.Context, params *GetActivityLogsParams) (*ActivityLogResponse, error) {
	if err := middleware.CheckPermission(constants.ActivityLogAction, nil, nil); err != nil {
		return nil, err
	}

	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	var activityLogs []models.ActivityLog

	query := s.db.Preload("User").Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	queryCount := s.db.Model(&models.ActivityLog{})

	if params.Filter.FromDate != nil {
		query = query.Where("created_at >= ?", params.Filter.FromDate)
		queryCount = queryCount.Where("created_at >= ?", params.Filter.FromDate)
	}

	if params.Filter.ToDate != nil {
		query = query.Where("created_at <= ?", params.Filter.ToDate)
		queryCount = queryCount.Where("created_at <= ?", params.Filter.ToDate)
	}

	if params.Filter.ActionBy != nil && *params.Filter.ActionBy != "" {
		query = query.Where("LOWER(action_by) LIKE LOWER(?)", "%"+*params.Filter.ActionBy+"%")
		queryCount = queryCount.Where("LOWER(action_by) LIKE LOWER(?)", "%"+*params.Filter.ActionBy+"%")
	}

	if params.Filter.Status != nil && *params.Filter.Status != "" {
		query = query.Where("status = ?", params.Filter.Status)
		queryCount = queryCount.Where("status = ?", params.Filter.Status)
	}

	if err := query.Order("created_at DESC").Find(&activityLogs).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := queryCount.Count(&total).Error; err != nil {
		return nil, err
	}

	return &ActivityLogResponse{
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(params.PageSize))),
		},
		Data: activityLogs,
	}, nil
}

//encore:api private method=POST path=/api/activity-log/create
func (s *Service) CreateActivityLog(ctx context.Context, logActivity *models.ActivityLog) error {
	// Truncate errorMessage to a maximum length of 255 characters
	truncatedErrorMessage := logActivity.ErrorMessage
	if len(logActivity.ErrorMessage) > 255 {
		truncatedErrorMessage = logActivity.ErrorMessage[:254]
	}

	logActivity.ErrorMessage = truncatedErrorMessage
	if err := s.db.Create(logActivity).Error; err != nil {
		return err
	}
	return nil
}
