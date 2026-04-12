package models

import (
	"time"

	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type AppPlatform string

const (
	PlatformAndroid AppPlatform = "android"
	PlatformIOS     AppPlatform = "ios"
)

type Environment string

const (
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentDevelopment Environment = "development"
)

type AppVersion struct {
	ID                 uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	AppPackageName     string         `json:"app_package_name" gorm:"type:varchar(255)"`
	Platform           AppPlatform    `json:"platform" gorm:"type:varchar(100)"`
	VersionName        string         `json:"version_name" gorm:"type:varchar(255)"`
	VersionCode        string         `json:"version_code" gorm:"type:varchar(255)"`
	MinimumVersionName string         `json:"minimum_version_name" gorm:"type:varchar(255)"`
	MinimumVersionCode string         `json:"minimum_version_code" gorm:"type:varchar(255)"`
	ReleaseNote        string         `json:"release_note" gorm:"type:text"`
	DownloadURL        string         `json:"download_url" gorm:"type:varchar(255)"`
	MandatoryUpdate    bool           `json:"mandatory_update" gorm:"type:boolean"`
	ReleaseDate        time.Time      `json:"release_date" gorm:"type:timestamp with time zone"`
	Environment        Environment    `json:"environment" gorm:"type:varchar(100)"`
	IsActive           bool           `json:"is_active" gorm:"type:boolean;default:true"`
	FirebaseProjectID  *string        `json:"firebase_project_id" gorm:"type:varchar(255)"`
	FirebaseAppID      *string        `json:"firebase_app_id" gorm:"type:varchar(255)"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          *time.Time     `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at,omitempty"`
}
