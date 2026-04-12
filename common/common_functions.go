package common

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"encore.app/common/constants"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Check if list contains role
func IsRoleInList(roles []constants.UserRole, role constants.UserRole) bool {
	for _, each_role := range roles {
		if role == each_role {
			return true
		}
	}
	return false
}

// Check if roles doesn't contain any admin roles
func ContainsGeneralStaffRole(roles []constants.UserRole) bool {
	for _, role := range roles {
		if !IsRoleInList(constants.SanteAdminRoles, role) {
			return true
		}
	}
	return false
}

func LoadEnv() {
	environment := os.Getenv("ENV")
	log.Println("ENV:", environment)

	if environment == "local" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func IsProduction() bool {
	return os.Getenv("ENV") == "production"
}

func IsStaging() bool {
	return os.Getenv("ENV") == "staging"
}

func IsLocal() bool {
	return os.Getenv("ENV") == "local"
}

func GetOrigin() string {
	LoadEnv()
	return os.Getenv("ORIGIN")
}

func GetSanteGoogleAuthClientID() string {
	LoadEnv()
	return os.Getenv("SANTE_GOOGLE_AUTH_CLIENT_ID")
}

type TokenOptions struct {
	UserID              uuid.UUID
	LoginType           *string
	ExpiryTimeInMinutes time.Duration
}

func GetJWTSecretKey() string {
	LoadEnv()
	return os.Getenv("JWT_SECRET_KEY")
}

func GenerateJWTToken(options TokenOptions) (*string, error) {
	jwtSecretKey := GetJWTSecretKey()

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        options.UserID.String(),
		"exp":        time.Now().Add(options.ExpiryTimeInMinutes * time.Minute).Unix(),
		"login_type": options.LoginType,
	})

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return nil, err
	}
	return &signedToken, nil
}

func GetRedisAddress() string {
	address := "localhost:6379"
	if !IsLocal() {
		log.Println("Not local")
		address = "redis:6379"
	}
	return address
}

func GetBundleID() string {
	bundleID := "com.afed.pretzley.stg"
	if IsProduction() {
		bundleID = "com.afed.pretzley"
	}
	return bundleID
}

// DecodeToken decodes a JWT token and returns the claims
func DecodeToken(tokenString string) (*TokenInfo, error) {
	jwtSecretKey := GetJWTSecretKey()

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "unexpected signing method",
			}
		}
		return []byte(jwtSecretKey), nil
	})

	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid token",
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid token claims",
		}
	}

	// Extract claims into TokenInfo struct
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid user_id claim",
		}
	}

	expiry, ok := claims["exp"].(float64)
	if !ok {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid token_expiry claim",
		}
	}

	loginType, ok := claims["login_type"].(string)
	if !ok {
		loginType = ""
	}

	parsedUUID, err := googleUuid.Parse(userID)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "invalid user_id format",
		}
	}
	return &TokenInfo{
		UserID:      uuid.UUID(parsedUUID),
		TokenExpiry: time.Unix(int64(expiry), 0),
		LoginType:   loginType,
	}, nil

}

// EncodePassword hashes a password using bcrypt
func EncodePassword(password string) (string, error) {
	if password == "" {
		return "", &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "password cannot be empty",
		}
	}

	// Generate hash from password with default cost
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", &errs.Error{
			Code:    errs.Internal,
			Message: "failed to hash password",
		}
	}

	return string(hashedPassword), nil
}

// validatePassword checks if the password meets minimum security requirements:
// - At least 8 characters long
// - Contains at least one uppercase letter
// - Contains at least one lowercase letter
// - Contains at least one number
// - Contains at least one special character
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Password must be at least 8 characters long",
		}
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasNumber = true
		case char == '!' || char == '@' || char == '#' || char == '$' || char == '%' || char == '^' || char == '&' || char == '*':
			hasSpecial = true
		}
	}

	if !hasUpper {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Password must contain at least one uppercase letter",
		}
	}
	if !hasLower {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Password must contain at least one lowercase letter",
		}
	}
	if !hasNumber {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Password must contain at least one number",
		}
	}
	if !hasSpecial {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Password must contain at least one special character (!@#$%^&*)",
		}
	}

	return nil
}

// HandleFieldUpdate updates a field in the updateData map based on the newVal and nullable flag
func HandleFieldUpdate[T any](oldVal *T, newVal *T, fieldName string, updateData map[string]interface{}, nullable bool) {
	if newVal == nil {
		fmt.Printf("%s is nil\n", fieldName)
		if nullable {
			updateData[fieldName] = gorm.Expr("NULL")
		} else {
			updateData[fieldName] = oldVal
		}
	} else {
		fmt.Printf("%s is not nil\n", fieldName)
		switch v := any(*newVal).(type) {
		case int:
			if v >= 0 {
				updateData[fieldName] = v
			} else {
				updateData[fieldName] = oldVal
			}
		case float32:
			if v >= 0 {
				updateData[fieldName] = v
			} else {
				updateData[fieldName] = oldVal
			}
		case string:
			if v != "" {
				updateData[fieldName] = v
			} else {
				updateData[fieldName] = oldVal
			}
		case bool:
			updateData[fieldName] = v
		default:
			updateData[fieldName] = oldVal
		}
	}
}

// GenerateVoucherCode generates a random voucher code if custom code is not provided
func GenerateVoucherCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = byte(rand.Intn(26) + 65)
	}

	return string(code)
}

// GenerateOrderNumber based on outlet daily order number count (reset daily)
func GenerateOrderNumber(currentCount int) string {

	return fmt.Sprintf("%d", currentCount+1)
}

// Encrypt text using AES
func EncryptText(text string) (string, error) {
	jwtSecretKey := GetJWTSecretKey()
	key, err := hex.DecodeString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))

	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(cryptoRand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt text using AES
func DecryptText(encryptedText string) (string, error) {
	jwtSecretKey := GetJWTSecretKey()
	key, err := hex.DecodeString(jwtSecretKey)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func SetEnvByEnvKey(key string, envValue string) {
	os.Setenv(key, envValue)
}
func GetEnvByEnvKey(key string) string {
	return os.Getenv(key)
}

// convert float32 to decimal.Decimal
func ConvertFloat32ToDecimal(value float32) decimal.Decimal {
	return decimal.NewFromFloat(float64(value))
}

// convert float64 to decimal.Decimal
func ConvertFloat64ToDecimal(value float64) decimal.Decimal {
	return decimal.NewFromFloat(value)
}

// convert int to decimal.Decimal
func ConvertIntToDecimal(value int) decimal.Decimal {
	return decimal.NewFromInt(int64(value))
}

// convert decimal.Decimal to float32
func ConvertDecimalToFloat32(value decimal.Decimal) float32 {
	convertedValue, _ := value.Float64()
	return float32(convertedValue)
}

// convert decimal.Decimal to float64
func ConvertDecimalToFloat64(value decimal.Decimal) float64 {
	convertedValue, _ := value.Float64()
	return convertedValue
}

// convert decimal.Decimal to int
func ConvertDecimalToInt(value decimal.Decimal) int {
	return int(value.IntPart())
}

// get date in format YYYY-MM-DD
func GetDateInFormatYYYYMMDD(dateString string) (time.Time, error) {
	return time.Parse("2006-01-02", dateString)
}

// get date in format YYYY-MM-DD HH:MM:SS
func GetDateInFormatYYYYMMDDHHMMSS(dateString string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", dateString)
}

// get date in format YYYY-MM-DD HH:MM:SS and start time of the day
func GetDateInFormatYYYYMMDDHHMMSSStartOfDay(dateString string) (time.Time, error) {
	date, err := GetDateInFormatYYYYMMDD(dateString)
	if err != nil {
		return time.Time{}, err
	}
	return date.AddDate(0, 0, 0).Add(time.Second), nil
}

// get date in format YYYY-MM-DD HH:MM:SS and end time of the day
func GetDateInFormatYYYYMMDDHHMMSSEndOfDay(dateString string) (time.Time, error) {
	date, err := GetDateInFormatYYYYMMDD(dateString)
	if err != nil {
		return time.Time{}, err
	}
	return date.AddDate(0, 0, 1).Add(-time.Second), nil
}

func GetStartDateOfMonth(date time.Time) (time.Time, error) {
	date = date.In(time.FixedZone("Malaysia Time", 8*60*60))
	date = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	date = date.UTC()
	return date, nil
}

// get start of day in Malaysia timezone
// example: startOfDay 2026-01-08 00:00:00 +0800 +08
func GetStartOfDay(date time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		return time.Time{}, err
	}
	dateInMY := date.In(loc)
	startOfDay := time.Date(dateInMY.Year(), dateInMY.Month(), dateInMY.Day(), 0, 0, 0, 0, loc)
	return startOfDay, nil
}

// get end of day in Malaysia timezone
// example: endOfDay 2026-01-09 23:59:59 +0800 +08
func GetEndOfDay(date time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		return time.Time{}, err
	}
	dateInMY := date.In(loc)
	endOfDay := time.Date(dateInMY.Year(), dateInMY.Month(), dateInMY.Day(), 23, 59, 59, 0, loc)
	return endOfDay, nil
}

func GetTimeInMalaysiaTimezone(date time.Time) (time.Time, error) {
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		return time.Time{}, err
	}
	return date.In(loc), nil
}

// calculate max page number
func CalculateMaxPageNumber(totalCount int64, pageSize int) int {
	maxPage := int(math.Ceil(float64(totalCount) / float64(pageSize)))
	return maxPage
}

// GeneralSearch is a function that returns a function that can be used to search for a string in a table
// tableName is the name of the table
// tableAttribute is the attribute of the table
// searchKey is the key to search for
// if empty string is provided, it will return the original db
func GeneralSearch(searchKey string, tableName string, tableAttribute string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if searchKey == "" {
			return db
		}
		return db.Where("LOWER("+tableName+"."+tableAttribute+") LIKE ?", "%"+strings.ToLower(searchKey)+"%")
	}
}

func IsTokenExpired(expiryTime *time.Time) bool {
	if expiryTime == nil {
		return true
	}
	// Check if the token is 5 minutes to expire
	return time.Now().Add(time.Minute * 5).After(*expiryTime)
}

func StringToUUID(uuidString string) (uuid.UUID, error) {
	temp_uuid, err := googleUuid.Parse(uuidString)
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.UUID(temp_uuid), nil
}

func Float32Ptr(f float32) *float32 {
	return &f
}

func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ArrayContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GenerateRandom6DigitCode() (string, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := fmt.Sprintf("%06d", r.Intn(1000000))
	return code, nil
}

func GenerateUniqueVoucherCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	now := time.Now()
	year := fmt.Sprintf("%02d", now.Year()%100)
	day := fmt.Sprintf("%03d", now.YearDay())

	// Generate 7 random alphanumeric characters
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 7)
	for i := range code {
		code[i] = chars[r.Intn(len(chars))]
	}

	//last character
	checksum := byte((r.Intn(26) + 65))

	return string(code) + year + day + string(checksum)
}

func GetStartOfDayByTruncate(date time.Time) time.Time {
	// Extract date components in the original timezone
	year, month, day := date.Date()
	// Create new time at 00:00:00 in the same timezone
	startDate := time.Date(year, month, day, 0, 0, 0, 0, date.Location())
	return startDate
}

func GetEndOfDayByTruncate(date time.Time) time.Time {
	// Extract date components in the original timezone
	year, month, day := date.Date()
	// Create new time at 23:59:59.999999999 in the same timezone
	endDate := time.Date(year, month, day, 23, 59, 59, 999999999, date.Location())
	return endDate
}

func FlattenObject(obj map[string]interface{}) []interface{} {
	result := []interface{}{}

	for _, v := range obj {
		if nested, ok := v.(map[string]interface{}); ok {
			result = append(result, FlattenObject(nested)...)
		} else {
			result = append(result, v)
		}
	}

	return result
}

func LevenshteinDistance(a, b string) int {
	lenA := len(a)
	lenB := len(b)

	dp := make([][]int, lenA+1)
	for i := range dp {
		dp[i] = make([]int, lenB+1)
	}

	// Base cases
	for i := 0; i <= lenA; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		dp[0][j] = j
	}

	// Fill DP table
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + MinThree(
					dp[i-1][j],   // delete
					dp[i][j-1],   // insert
					dp[i-1][j-1], // replace
				)
			}
		}
	}

	return dp[lenA][lenB]
}

func MinThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Geocode address to coordinates using OpenStreetMap Nominatim
func GeocodeAddress(address string) (float64, float64, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"

	// Build query parameters
	params := url.Values{}
	params.Add("q", address)
	params.Add("format", "json")
	params.Add("limit", "1")

	// Make request
	resp, err := http.Get(baseURL + "?" + params.Encode())
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
		case http.StatusForbidden:
			return 0, 0, &errs.Error{
				Code:    errs.PermissionDenied,
				Message: "Permission denied to geocode address",
			}
		default:
			return 0, 0, &errs.Error{
				Code:    errs.Internal,
				Message: "Failed to geocode address: " + resp.Status,
			}
		}
	}

	// Parse response
	var results []struct {
		Lat string `json:"lat"`
		Lon string `json:"lon"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return 0, 0, err
	}

	if len(results) == 0 {
		return 0, 0, fmt.Errorf("no coordinates found for address: %s", address)
	}

	lat, err := strconv.ParseFloat(results[0].Lat, 64)
	if err != nil {
		return 0, 0, err
	}

	lon, err := strconv.ParseFloat(results[0].Lon, 64)
	if err != nil {
		return 0, 0, err
	}

	return lat, lon, nil
}

func GetPossiblePhoneNumbers(phoneNumber string) []string {
	phoneNumbers := []string{}

	hasPlus6 := strings.HasPrefix(phoneNumber, "+6")
	has6 := strings.HasPrefix(phoneNumber, "6")

	if hasPlus6 {
		phoneNumbers = append(phoneNumbers, strings.TrimPrefix(phoneNumber, "+"))
		phoneNumbers = append(phoneNumbers, strings.TrimPrefix(phoneNumber, "+6"))
	} else if has6 {
		phoneNumbers = append(phoneNumbers, "+"+phoneNumber)
		phoneNumbers = append(phoneNumbers, strings.TrimPrefix(phoneNumber, "6"))
	} else {
		phoneNumbers = append(phoneNumbers, "+6"+phoneNumber)
		phoneNumbers = append(phoneNumbers, "6"+phoneNumber)
	}
	phoneNumbers = append(phoneNumbers, phoneNumber)

	return phoneNumbers
}

func SanitizePhoneNumber(phoneNumber string) string {
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ".", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "(", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, ")", "")

	hasPlus6 := strings.HasPrefix(phoneNumber, "+6")
	has6 := strings.HasPrefix(phoneNumber, "6")

	if hasPlus6 {
		return phoneNumber
	} else if has6 {
		return "+" + phoneNumber
	} else {
		return "+6" + phoneNumber
	}
}

func ClearCountryCodeFromPhoneNumber(phoneNumber string) string {
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	return phoneNumber
}
