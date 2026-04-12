package common

import (
	"mime/multipart"
	"time"

	"encore.dev/types/uuid"
)

type Address struct {
	StreetLine1 string `json:"street_line1" gorm:"type:varchar(150)"`
	StreetLine2 string `json:"street_line2,omitempty" gorm:"type:varchar(150)"`
	StreetLine3 string `json:"street_line3,omitempty" gorm:"type:varchar(150)"`
	City        string `json:"city" gorm:"type:varchar(50)"`
	State       string `json:"state" gorm:"type:varchar(50)"`
	PostalCode  string `json:"postal_code" gorm:"type:varchar(20)"`
	Country     string `json:"country" gorm:"type:varchar(50)"`
}

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	Total      int64 `json:"total"`
}

type TokenInfo struct {
	UserID      uuid.UUID `json:"user_id"`
	TokenExpiry time.Time `json:"token_expiry"`
	LoginType   string    `json:"login_type"`
}

// basic response for all api
type BasicResponse struct {
	Message string `json:"message"`
}

type OutletModifierOption struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Price    float32   `json:"price"`
	IsActive bool      `json:"is_active"`
}

type Option struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type OptionWithValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type FileUploadRequest struct {
	ModelID uuid.UUID             `json:"model_id"`
	File    multipart.File        `json:"file"`
	Header  *multipart.FileHeader `json:"header"`
}
