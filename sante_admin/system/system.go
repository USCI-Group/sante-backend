package system

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/common_operations"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/firebase"
	"encore.app/middleware"
	"encore.dev/beta/errs"
	"github.com/asaskevich/govalidator"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

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

type SystemData struct {
	InfoType    constants.SystemDataInfoType `json:"info_type" gorm:"type:varchar(255)" valid:"required~Info type is required"`
	InfoValue   string                       `json:"info_value" gorm:"type:varchar(500)" valid:"required~Info value is required"`
	Expiry      *time.Time                   `json:"expiry" gorm:"type:timestamp with time zone"`
	IsEncrypted bool                         `json:"is_encrypted" gorm:"type:boolean;default:false"`
}

// Define a struct to match the JSON input
type AppVersionRequest struct {
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
	Environment        string    `json:"environment"`
	IsActive           bool      `json:"is_active"`
	AppName            string    `json:"app_name"`           // Firebase App Name
	AppID              string    `json:"app_id"`             // Firebase App ID
	UploadToFirebase   bool      `json:"upload_to_firebase"` // Upload to Firebase
}

// encore:api auth method=POST path=/api/system/store-system-data
func (s *Service) storeSystemData(ctx context.Context, request *SystemData) error {
	if err := middleware.CheckSanteAdmin(); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(request); err != nil {
		return err
	}

	if request.IsEncrypted {
		encryptedInfoValue, err := common.EncryptText(request.InfoValue)
		if err != nil {
			return err
		}
		request.InfoValue = encryptedInfoValue
	}

	systemData := &models.SystemData{
		InfoType:    request.InfoType,
		InfoValue:   request.InfoValue,
		Expiry:      request.Expiry,
		IsEncrypted: request.IsEncrypted,
	}

	err := s.db.Create(systemData).Error
	if err != nil {
		return err
	}

	return nil
}

// API to create app version
//
//encore:api auth raw method=POST path=/api/system/app/create-app-version
func (s *Service) CreateAppVersion(w http.ResponseWriter, r *http.Request) {

	err := middleware.CheckPermission(constants.CreateAppVersionAction, nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Set request body size limit to handle large APK files (up to 100MB)
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB limit for request body

	// Set a larger max memory for multipart parsing
	// This allocates memory for the entire multipart form, including files
	// Setting it to 50MB should be sufficient for 32MB APK + metadata
	err = r.ParseMultipartForm(50 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form (file may be too large): "+err.Error(), http.StatusBadRequest)
		return
	}

	jsonData := r.FormValue("json_data")
	if jsonData == "" {
		http.Error(w, "JSON data is required", http.StatusBadRequest)
		return
	}

	var req AppVersionRequest
	err = json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		http.Error(w, "Invalid JSON data: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AppPackageName == "" || req.Platform == "" || req.VersionName == "" ||
		req.VersionCode == "" || req.ReleaseNote == "" || req.DownloadURL == "" ||
		req.MinimumVersionName == "" || req.MinimumVersionCode == "" ||
		(req.UploadToFirebase && (req.AppName == "" || req.AppID == "")) {
		http.Error(w, "All required fields must be provided, and if upload to firebase is true, app name, app id and project number are required", http.StatusBadRequest)
		return
	}

	// get uploaded file
	file, header, err := r.FormFile("app_file")
	if err != nil {
		fmt.Println("error getting file", err)
	}

	app_version := models.AppVersion{
		AppPackageName:     req.AppPackageName,
		Platform:           models.AppPlatform(req.Platform),
		VersionName:        req.VersionName,
		VersionCode:        req.VersionCode,
		MinimumVersionName: req.MinimumVersionName,
		MinimumVersionCode: req.MinimumVersionCode,
		ReleaseNote:        req.ReleaseNote,
		DownloadURL:        req.DownloadURL,
		MandatoryUpdate:    req.MandatoryUpdate,
		ReleaseDate:        req.ReleaseDate,
		Environment:        models.Environment(req.Environment),
		IsActive:           req.IsActive,
	}

	// Process file upload if file exists
	if file != nil && header != nil && !req.UploadToFirebase {
		defer file.Close()

		var document aws_s3.Document
		document.File = file
		file_extension := filepath.Ext(header.Filename)
		document.DocPath = fmt.Sprintf("business/%s/app/%d%s",
			req.AppPackageName,
			time.Now().UnixMilli(),
			file_extension)
		document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

		// Upload the file to S3
		document_res, err := aws_s3.UploadDocument(document)
		if err != nil {
			errs.HTTPError(w, err)
			return
		}

		// Override download url with s3 url and update app_version so it gets saved to DB
		req.DownloadURL = document_res.Url
		app_version.DownloadURL = document_res.Url
	}

	// process upload to firebase (binary file as POST body per Firebase API)
	if req.UploadToFirebase {
		if file == nil || header == nil {
			http.Error(w, "App file is required for Firebase upload", http.StatusBadRequest)
			return
		}
		defer file.Close()
		uploadResponse, err := firebase.UploadAppRelease(file, header.Filename, req.AppID)
		if err != nil {
			errs.HTTPError(w, err)
			return
		}
		uploadResponseData, err := firebase.ExtractInformationFromUploadResponse(uploadResponse)
		if err != nil {
			errs.HTTPError(w, err)
			return
		}

		app_version.FirebaseProjectID = &uploadResponseData.ProjectID
		app_version.FirebaseAppID = &uploadResponseData.AppID
	}

	// Save to database
	if err := s.db.Create(&app_version).Error; err != nil {
		http.Error(w, "Failed to create app version: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("App version created successfully"))
}

// encore:api auth method=GET path=/api/system/system-data/:info_type
func (s *Service) GetSystemDataValue(ctx context.Context, info_type string) (*models.SystemData, error) {
	if err := middleware.CheckSanteAdmin(); err != nil {
		return nil, err
	}

	systemDataValue := common_operations.GetSystemDataValue(s.db, constants.SystemDataInfoType(info_type))
	if systemDataValue == "" {
		return nil, &errs.Error{
			Code:    errs.NotFound,
			Message: "System data not found",
		}
	}

	systemData := &models.SystemData{
		InfoType:  constants.SystemDataInfoType(info_type),
		InfoValue: systemDataValue,
	}

	return systemData, nil
}

// encore:api private method=GET path=/api/private/system/system-data/:infoType
func (s *Service) GetSystemDataValuePrivate(ctx context.Context, infoType string) (*models.SystemData, error) {
	systemData := &models.SystemData{}
	err := s.db.Where("info_type = ?", infoType).First(systemData).Error
	if err != nil {
		return nil, err
	}

	if systemData.IsEncrypted {
		decryptedInfoValue, err := common.DecryptText(systemData.InfoValue)
		if err != nil {
			return nil, err
		}
		systemData.InfoValue = decryptedInfoValue
	}
	return systemData, nil
}

// AppReleasesResponse wraps raw JSON from Firebase App Distribution API.
type FirebaseAppDistributionRawResponse struct {
	Data json.RawMessage `json:"data"`
}

// API to get app releases from firebase
//
//encore:api auth method=GET path=/api/system/firebase/app/releases/:app_id
func (s *Service) GetAppReleases(ctx context.Context, app_id string) (*FirebaseAppDistributionRawResponse, error) {
	body, err := firebase.GetAppReleasesInformation(app_id)
	if err != nil {
		return nil, err
	}

	return &FirebaseAppDistributionRawResponse{
		Data: body,
	}, nil
}

type ListProjectAppsRequest struct {
	// project identifier can be project number or project name
	// example: 460043168578 or sante-pos
	ProjectIdentifier string  `json:"project_identifier"`
	Platform          *string `json:"platform"`
}

// API to list app ids from firebase
// POST with JSON body. Send at least {} - request body is required.
// Example: {"project_identifier": "460043168578" or "sante-pos", "platform": "android" or "ios" or "web"}
//
//encore:api auth method=POST path=/api/system/firebase/project/apps/list
func (s *Service) ListProjectApps(ctx context.Context, req *ListProjectAppsRequest) (*FirebaseAppDistributionRawResponse, error) {
	body, err := firebase.ListProjectApps(&req.ProjectIdentifier, req.Platform)
	if err != nil {
		return nil, err
	}
	return &FirebaseAppDistributionRawResponse{
		Data: body,
	}, nil
}

// API to get all projects from firebase
// EXAMPLE: /api/system/firebase/project/all/10
//
//encore:api auth method=GET path=/api/system/firebase/project/all/:page_size
func (s *Service) GetAllProjects(ctx context.Context, page_size int) (*FirebaseAppDistributionRawResponse, error) {
	body, err := firebase.GetAllProjects(page_size)
	if err != nil {
		return nil, err
	}
	return &FirebaseAppDistributionRawResponse{
		Data: body,
	}, nil
}
