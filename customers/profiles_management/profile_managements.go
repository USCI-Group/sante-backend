package profiles_management

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encore.app/auth_service"
	"encore.app/aws_s3"
	"encore.app/customers/customer_common"
	"encore.app/database"
	"encore.app/database/models"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

// request for read profile all data
type GetProfileRequest struct {
	Token string `json:"token"`
}

// response for read profile all data
type GetProfileResponse struct {
	Response int    `json:"response"`
	Message  string `json:"message"`
	Customer SafeCustomer
}

type RankInfo struct {
	RankName          string             `json:"rank_name"`
	RankImageURL      string             `json:"rank_image_url"`
	OverallGoal       int                `json:"overall_goal"`
	OverallProgress   int                `json:"overall_progress"`
	DetailToNextRanks []DetailToNextRank `json:"detail_to_next_ranks"`
}

type DetailToNextRank struct {
	ProductID   *uuid.UUID `json:"product_id"`
	Description string     `json:"description"`
	Progress    float32    `json:"progress"`
	Goal        float32    `json:"goal"`
}

// request for update profile data
type UpdateProfileRequest struct {
	FirstName              string `json:"first_name"`
	LastName               string `json:"last_name"`
	Email                  string `json:"email"`
	Phone                  string `json:"phone"`
	DOB                    string `json:"dob"`
	IsNewsletterSubscribed bool   `json:"is_newsletter_subscribed"`
}

// response for update profile data
type UpdateProfileResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// customer data safe
type SafeCustomer struct {
	FirstName              string    `json:"first_name"`
	LastName               string    `json:"last_name"`
	Email                  string    `json:"email"`
	Phone                  string    `json:"phone"`
	Address                string    `json:"address"`
	DOB                    string    `json:"dob"`
	ProfilePicture         string    `json:"profile_picture"`
	IsNewsletterSubscribed bool      `json:"is_newsletter_subscribed"`
	CreatedAt              time.Time `json:"created_at"`
}

type CheckLatestVersionRequest struct {
	Platform           string `json:"platform"`
	AppPackageName     string `json:"app_package_name"`
	CurrentVersionName string `json:"current_version_name"`
	CurrentVersionCode string `json:"current_version_code"`
}

type CheckLatestVersionResponse struct {
	Message   string   `json:"message"`
	HasUpdate bool     `json:"has_update"`
	AppInfo   *AppInfo `json:"app_info"`
}

type AppInfo struct {
	AppPackageName     string    `json:"app_package_name"`
	Platform           string    `json:"platform"`
	VersionName        string    `json:"version_name"`
	VersionCode        string    `json:"version_code"`
	MinimumVersionName string    `json:"minimum_version_name"`
	MinimumVersionCode string    `json:"minimum_version_code"`
	ReleaseNote        string    `json:"release_note"`
	DownloadURL        string    `json:"download_url"`
	MandatoryUpdate    bool      `json:"mandatory_update"`
	ReleaseDate        time.Time `json:"release_date"`
	IsActive           bool      `json:"is_active"`
}

var secretsKeys struct {
	jwtSecretKey string
}

// initService initializes the user service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: database.GetSanteDB().Stdlib(),
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	secretsKeys.jwtSecretKey = os.Getenv("JWT_SECRET_KEY")
	return &Service{db: db}, nil
}

// need change to auth in future
// get profile data of user
//
//encore:api auth method=GET path=/api/customers/profiles/get-profile
func (s *Service) GetProfile(ctx context.Context) (*GetProfileResponse, error) {

	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return nil, err
	}

	var customer models.Customer
	if err := s.db.Where("id = ?", user.ID).First(&customer).Error; err != nil {
		return nil, err
	}

	return &GetProfileResponse{
		Response: http.StatusOK,
		Message:  "Profile data fetched successfully",
		Customer: SafeCustomer{
			FirstName:              customer.FirstName,
			LastName:               customer.LastName,
			Email:                  customer.Email,
			Phone:                  customer.PhoneNumber,
			Address:                customer.Email,
			DOB:                    customer.DateOfBirth.Format("2006-01-02"),
			ProfilePicture:         customer.ProfilePicture,
			IsNewsletterSubscribed: customer.IsNewsletterSubscribed,
			CreatedAt:              customer.CreatedAt,
		},
	}, nil
}

// need to change to auth middleware
// update profile data
//
//encore:api auth method=POST path=/api/customers/profiles/update-profile
func (s *Service) UpdateProfile(ctx context.Context, req *UpdateProfileRequest) (*UpdateProfileResponse, error) {
	user, err := auth_service.GetMe(ctx)
	if err != nil {
		return nil, err
	}

	var customer models.Customer
	err = s.db.Model(&customer).Where("id = ? ", user.ID).Updates(map[string]interface{}{
		"first_name":               req.FirstName,
		"last_name":                req.LastName,
		"email":                    req.Email,
		"phone_number":             req.Phone,
		"date_of_birth":            req.DOB,
		"is_newsletter_subscribed": req.IsNewsletterSubscribed,
	}).Error
	if err != nil {
		return nil, err
	}

	return &UpdateProfileResponse{
		Status:  http.StatusOK,
		Message: "Profile data updated successfully",
	}, nil

}

// update profile picture
//
//encore:api auth raw method=POST path=/api/customers/profiles/update-profile-picture
func (s *Service) UpdateProfilePicture(w http.ResponseWriter, req *http.Request) {
	customer_id := req.FormValue("customer_id")
	business_id := req.FormValue("business_id")

	if customer_id == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}
	if business_id == "" {
		http.Error(w, "Business ID is required", http.StatusBadRequest)
		return
	}

	customer := models.Customer{}
	err := s.db.Where("id = ?", customer_id).First(&customer).Error
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	business := models.Business{}
	err = s.db.Where("id = ?", business_id).First(&business).Error
	if err != nil {
		http.Error(w, "Business not found", http.StatusNotFound)
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
	file_name := fmt.Sprintf("%s_%s_%d%s", customer.FirstName, customer.ID.String(), currentTime, file_extension)
	document.DocPath = "business/" + business.RegistrationNumber + "/customer/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// remove old image from s3
	if customer.ProfilePicture != "" {
		aws_s3.DeleteDocument(customer.ProfilePicture)
	}

	customer.ProfilePicture = document_res.Url

	err = s.db.Save(&customer).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Check latest version of app
//
//encore:api auth method=POST path=/api/customers/mobile/membership/app/version/check
func (s *Service) CheckLatestVersion(ctx context.Context, req *CheckLatestVersionRequest) (*CheckLatestVersionResponse, error) {
	appVersion, err := customer_common.GetLastestAppVersionByPlatformAndAppPackageName(
		s.db,
		req.Platform,
		req.AppPackageName,
	)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "App version not found",
		}
	}
	platform := string(appVersion.Platform)
	hasUpdate := false
	if req.CurrentVersionName < appVersion.VersionName || req.CurrentVersionCode < appVersion.VersionCode {
		hasUpdate = true
	}
	return &CheckLatestVersionResponse{
		Message:   "Latest app version checked successfully",
		HasUpdate: hasUpdate,
		AppInfo: &AppInfo{
			AppPackageName:     appVersion.AppPackageName,
			Platform:           platform,
			VersionName:        appVersion.VersionName,
			VersionCode:        appVersion.VersionCode,
			MinimumVersionName: appVersion.MinimumVersionName,
			MinimumVersionCode: appVersion.MinimumVersionCode,
			ReleaseNote:        appVersion.ReleaseNote,
			DownloadURL:        appVersion.DownloadURL,
			MandatoryUpdate:    appVersion.MandatoryUpdate,
			ReleaseDate:        appVersion.ReleaseDate,
			IsActive:           appVersion.IsActive,
		},
	}, nil
}
