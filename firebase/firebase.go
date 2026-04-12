package firebase

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"encore.app/common"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

var firebaseApp *firebase.App                  // this is for pos app
var messagingClient *messaging.Client          // this is for pos app

// type of firebase app used
type FirebaseAppType string

const (
	FirebaseAppTypePOS FirebaseAppType = "pos"
)

//encore:service
type Service struct {
	// Add your dependencies here
}

type FirebaseTestRequest struct {
	TargetToken string             `json:"target_token"`
	Title       string             `json:"title"`
	Body        string             `json:"body"`
	OrderID     *uuid.UUID         `json:"order_id"`
	ActionURL   *string            `json:"action_url"`
	CustomData  *map[string]string `json:"custom_data"`
}

//go:embed sante-pos-firebase-adminsdk-fbsvc-37ed27e7ef.json
var firebaseCredentialsJSON []byte

//go:embed sante-pos-prod-firebase-adminsdk-fbsvc-239d0a6066.json
var firebaseCredentialsJSONProd []byte

// initService initializes the site service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	common.LoadEnv()

	var posCredentials []byte
	if common.IsProduction() {
		updatedFirebaseCredentialsJSON := replacePlaceholders(string(firebaseCredentialsJSONProd), FirebaseAppTypePOS)
		posCredentials = []byte(updatedFirebaseCredentialsJSON)
	} else {
		updatedFirebaseCredentialsJSON := replacePlaceholders(string(firebaseCredentialsJSON), FirebaseAppTypePOS)
		posCredentials = []byte(updatedFirebaseCredentialsJSON)
	}
	opt := option.WithCredentialsJSON(posCredentials)

	posApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	firebaseApp = posApp

	posClient, err := posApp.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting Messaging client: %v\n", err)
	}
	messagingClient = posClient
	fmt.Println("Firebase Messaging client initialized successfully!")

	return &Service{}, nil
}

func replacePlaceholders(jsonString string, appType FirebaseAppType) string {
	//log.Printf("FIREBASE_PRIVATE_KEY: %v\n", os.Getenv("FIREBASE_PRIVATE_KEY"))
	//log.Printf("FIREBASE_PRIVATE_KEY_ID: %v\n", os.Getenv("FIREBASE_PRIVATE_KEY_ID"))
	//jsonString = strings.Replace(jsonString, "{{.FIREBASE_PRIVATE_KEY_ID}}", os.Getenv("FIREBASE_PRIVATE_KEY_ID"), 1)
	// Get the private key and properly escape it for JSON
	switch appType {
	case FirebaseAppTypePOS:
		firebasePrivateKeyID := os.Getenv("FIREBASE_PRIVATE_KEY_ID")
		jsonString = strings.Replace(jsonString, "{{.FIREBASE_PRIVATE_KEY_ID}}", firebasePrivateKeyID, 1)
	}
	return jsonString
}

// StringPtr returns a pointer to the given string value.
// This is a helper function for passing string literals as *string arguments.
func StringPtr(s string) *string {
	return &s
}

// send notification to a single device
func SendNotification(
	ctx context.Context,
	targetToken string,
	title string,
	body string,
	notificationIDs []uuid.UUID,
) error {
	if messagingClient == nil {
		return fmt.Errorf("messaging client not initialized")
	}
	data := map[string]string{
		"notification_id": "91683fd9-2528-4960-969b-a34515f4f989",
	}
	if len(notificationIDs) > 0 {
		notificationID := notificationIDs[0]
		data["notification_id"] = notificationID.String()
		log.Printf("notification_id: %v\n", notificationID.String())
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: targetToken, // Send to a specific device token
		// Alternatively, use Topic or Condition
		// Topic: "news",
		// Condition: "'stock_up' in topics && ('industry' in topics || 'finance' in topics')",
	}

	response, err := messagingClient.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}

	fmt.Printf("Successfully sent message: %v\n", response)
	return nil
}

// send notification to multiple devices
func SendNotificationToMultipleDevices(
	ctx context.Context,
	deviceTokens []string,
	title string,
	body string,
	notificationIDs []uuid.UUID, // pass in first index if need one only
	orderID *uuid.UUID,
	actionURL *string,
	notificationType models.NotificationType,
	firebaseAppType FirebaseAppType,
	customData *map[string]string, // custom data (optional and it will override the default data)
) error {
	var client *messaging.Client
	//var app *firebase.App
	switch firebaseAppType {
	case FirebaseAppTypePOS:
		if messagingClient == nil {
			return fmt.Errorf("POS messaging client not initialized")
		}
		if firebaseApp == nil {
			return fmt.Errorf("POS firebase app not initialized")
		}
		//app = firebaseApp
		client = messagingClient
		SendNotificationPerDevicePerNotification(
			ctx,
			deviceTokens,
			title,
			body,
			notificationIDs,
			orderID,
			actionURL,
			client,
			notificationType,
		)
	default:
		return nil
	}
	return nil
}

// this for now use for POS only
func SendNotificationPerDevicePerNotification(
	ctx context.Context,
	deviceTokens []string,
	title string,
	body string,
	notificationIDs []uuid.UUID,
	orderID *uuid.UUID,
	actionURL *string,
	client *messaging.Client,
	notificationType models.NotificationType,
) error {
	// Send one message per device per notification
	for i, notificationID := range notificationIDs {
		data := map[string]string{
			"notification_id": notificationID.String(),
		}
		if orderID != nil {
			data["order_id"] = orderID.String()
		}
		if actionURL != nil {
			data["action_url"] = *actionURL
		}
		data["notification_type"] = string(notificationType)
		message := &messaging.Message{
			Token: deviceTokens[i],
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}
		// Send individual message
		response, err := client.Send(ctx, message)
		if err != nil {
			fmt.Printf("error sending notification %d to device %s: %v", i+1, deviceTokens[i], err)
			continue
		}
		fmt.Printf("Successfully sent message %d to device %s: %v\n", i+1, deviceTokens[i], response)
	}
	return nil
}

// this for now use for Membership App only
func SendSingleNotificationToMultipleDevices(
	ctx context.Context,
	deviceTokens []string,
	title string,
	body string,
	notificationID uuid.UUID,
	orderID *uuid.UUID,
	actionURL *string,
	client *messaging.Client,
	notificationType models.NotificationType,
	customData *map[string]string, // custom data (optional and it will override the default data)
) error {
	// Send one message per device and max one notification per device
	for i, deviceToken := range deviceTokens {
		dataToSend := map[string]string{}
		if customData != nil {
			dataToSend = *customData
		} else {
			dataToSend = map[string]string{
				"notification_id": notificationID.String(),
			}
			if orderID != nil {
				dataToSend["order_id"] = orderID.String()
			}
			if actionURL != nil {
				dataToSend["action_url"] = *actionURL
			}
			dataToSend["notification_type"] = string(notificationType)
		}
		badge := 1
		message := &messaging.Message{
			Token: deviceTokens[i],
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			FCMOptions: &messaging.FCMOptions{
				AnalyticsLabel: "notification_label",
			},
			APNS: &messaging.APNSConfig{
				Headers: map[string]string{
					"apns-push-type": "alert",
					"apns-priority":  "10",
				},
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: title,
							Body:  body,
						},
						Badge:          &badge,
						Sound:          "pretzley_notification.caf",
						MutableContent: true,
					},
				},
			},
			Data: dataToSend,
		}
		// Send individual message
		response, err := client.Send(ctx, message)
		if err != nil {
			fmt.Printf("error sending notification to device %s: %v", deviceToken, err)
			continue
		}
		fmt.Printf("Successfully sent message to device %s: %v\n", deviceToken, response)
	}
	return nil
}

// notification api
//
//encore:api private method=POST path=/api/firebase/notification
func (s *Service) FirebaseNotificationAPI(ctx context.Context, req FirebaseTestRequest) error {
	err := SendNotification(ctx, req.TargetToken, req.Title, req.Body, []uuid.UUID{})
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}
	return nil
}

// test notification api
//
//encore:api auth method=POST path=/api/firebase/test-notification
func (s *Service) TestNotificationAPI(ctx context.Context, req FirebaseTestRequest) error {
	err := SendNotification(ctx, req.TargetToken, req.Title, req.Body, []uuid.UUID{})
	if err != nil {
		return fmt.Errorf("error sending notification: %v", err)
	}
	return nil
}


// TestNotificationAPIMembershipApp removed as it was specific to the membership app

// VerifyFirebaseToken removed as it was specific to the membership app

func getFirebaseCredentials() []byte {
	if common.IsProduction() {
		return firebaseCredentialsJSONProd
	} else {
		return firebaseCredentialsJSON
	}
}

func getFirebaseCredentialsFromJSON(json []byte) (*google.Credentials, error) {
	creds, err := google.CredentialsFromJSON(context.Background(), json, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	return creds, nil
}

// Get app release info
func GetAppReleasesInformation(
	appID string, // firebase app id example: 1:460043168578:android:2d90ee756cfe9d14bc76db
) ([]byte, error) {
	credentials := getFirebaseCredentials()
	creds, err := getFirebaseCredentialsFromJSON(credentials)
	if err != nil {
		return nil, err
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	// App Distribution API requires projectNumber (numeric), not project_id (string).
	// Format: projects/{projectNumber}/apps/{appId} - appId format: 1:projectNumber:platform:hash
	parts := strings.Split(appID, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid app ID format")
	}
	projectNumber := parts[1]
	url := fmt.Sprintf("https://firebaseappdistribution.googleapis.com/v1/projects/%s/apps/%s/releases", projectNumber, appID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	//fmt.Println(string(body))
	return body, nil
}

// UploadAppRelease uploads a binary (APK/AAB) to Firebase App Distribution.
// file: the binary file content (io.Reader)
// filename: original filename for X-Goog-Upload-File-Name header (e.g. "release.apk")
// appID: Firebase app ID format 1:projectNumber:platform:hash
func UploadAppRelease(file io.Reader, filename string, appID string) ([]byte, error) {
	credentials := getFirebaseCredentials()
	creds, err := getFirebaseCredentialsFromJSON(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	parts := strings.Split(appID, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid app ID format")
	}
	projectNum := parts[1]

	url := fmt.Sprintf("https://firebaseappdistribution.googleapis.com/upload/v1/projects/%s/apps/%s/releases:upload", projectNum, appID)
	log.Println("url", url)
	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, file)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token.AccessToken)
	httpReq.Header.Set("X-Goog-Upload-Protocol", "raw")
	httpReq.Header.Set("X-Goog-Upload-File-Name", filename)
	response, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil

}

type UploadAppReleaseResponse struct {
	ProjectID string `json:"project_id"`
	AppID     string `json:"app_id"`
}

// example response:
//
//	{
//		"name": "projects/460043168578/apps/1:460043168578:android:2d90ee756cfe9d14bc76db/releases/-/operations/f08ba92d8d17f0569dc8b121de225524c46b8ea7a0dcbcd1cc40b4e61a495916"
//	}
func ExtractInformationFromUploadResponse(uploadResponse []byte) (*UploadAppReleaseResponse, error) {
	var uploadResponseMap map[string]string
	err := json.Unmarshal(uploadResponse, &uploadResponseMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload response: %w", err)
	}
	name := uploadResponseMap["name"]
	parts := strings.Split(name, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid name format")
	}
	return &UploadAppReleaseResponse{
		ProjectID: parts[1],
		AppID:     parts[3],
	}, nil
}

// get app id from credentials
// platform: androidApps / iosApps / webApps / macOsApps / tvOsApps
func ListProjectApps(projectIdentifier *string, platform *string) ([]byte, error) {
	credentials := getFirebaseCredentials()
	creds, err := getFirebaseCredentialsFromJSON(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	var url string
	// if projectNumber is empty, it will list all apps in the project
	// if projectNumber is not empty, it will list all apps in the project with the given project number
	if projectIdentifier == nil || *projectIdentifier == "" {
		url = "https://firebase.googleapis.com/v1beta1/projects:searchApps"
	} else {
		url = fmt.Sprintf("https://firebase.googleapis.com/v1beta1/projects/%s:searchApps", *projectIdentifier)
	}

	if (platform != nil && *platform != "") && (projectIdentifier != nil && *projectIdentifier != "") {
		platformToLower := strings.ToLower(*platform)
		url = fmt.Sprintf("%s?filter=platform=%s", url, platformToLower)
	}
	log.Println("url", url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	// 3. Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// 4. Parse the response
	// Response will contain a list of 'AppMetadata' objects
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// get all projects from firebase
func GetAllProjects(pageSize int) ([]byte, error) {
	credentials := getFirebaseCredentials()
	creds, err := getFirebaseCredentialsFromJSON(credentials)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}
	url := "https://firebase.googleapis.com/v1beta1/projects"
	if pageSize > 0 {
		url = fmt.Sprintf("%s?pageSize=%d", url, pageSize)
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return body, nil
}

// appDistributionListReleasesResponse is the Firebase App Distribution API list releases response.
type appDistributionListReleasesResponse struct {
	Releases      []AppDistributionRelease `json:"releases"`
	NextPageToken string                   `json:"nextPageToken"`
}

// AppDistributionRelease represents a single release from Firebase App Distribution.
type AppDistributionRelease struct {
	Name               string `json:"name"`
	DisplayVersion     string `json:"displayVersion"`
	BuildVersion       string `json:"buildVersion"`
	CreateTime         string `json:"createTime"`
	BinaryDownloadURI  string `json:"binaryDownloadUri"`
	FirebaseConsoleURI string `json:"firebaseConsoleUri"`
	TestingURI         string `json:"testingUri"`
	ExpireTime         string `json:"expireTime"`
}

// GenerateDownloadURLFromFirebase returns the binary download URL of the latest release for the given app.
// Latest is determined by createTime (most recent first).
func GenerateDownloadURLFromFirebase(projectID string, appID string) (*AppDistributionRelease, error) {
	releasesResp, err := GetAppReleasesInformation(appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app releases information: %w", err)
	}
	var listResp appDistributionListReleasesResponse
	if err := json.Unmarshal(releasesResp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal releases: %w", err)
	}
	if len(listResp.Releases) == 0 {
		return nil, fmt.Errorf("no releases found for app %s", appID)
	}
	latest := findLatestRelease(listResp.Releases)
	if latest == nil {
		return nil, fmt.Errorf("could not determine latest release")
	}
	if latest.BinaryDownloadURI == "" {
		return nil, fmt.Errorf("latest release has no binary download URI")
	}
	return latest, nil
}

// findLatestRelease returns the release with the most recent createTime.
func findLatestRelease(releases []AppDistributionRelease) *AppDistributionRelease {
	if len(releases) == 0 {
		return nil
	}
	var latest *AppDistributionRelease
	var latestTime time.Time
	for i := range releases {
		t, err := time.Parse(time.RFC3339, releases[i].CreateTime)
		if err != nil {
			continue
		}
		if latest == nil || t.After(latestTime) {
			latestTime = t
			latest = &releases[i]
		}
	}
	return latest
}

// GetLatestReleaseInfo returns the latest release info (displayVersion, buildVersion, download URL, etc.) for the given app.
// Latest is determined by createTime (most recent first).
func GetLatestReleaseInfo(appID string) (*AppDistributionRelease, error) {
	releasesResp, err := GetAppReleasesInformation(appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app releases information: %w", err)
	}
	var listResp appDistributionListReleasesResponse
	if err := json.Unmarshal(releasesResp, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal releases: %w", err)
	}
	if len(listResp.Releases) == 0 {
		return nil, fmt.Errorf("no releases found for app %s", appID)
	}
	latest := findLatestRelease(listResp.Releases)
	if latest == nil {
		return nil, fmt.Errorf("could not determine latest release")
	}
	return latest, nil
}
