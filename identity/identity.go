package identity

import (
	"context"
	"net/http"
	"os"
	"time"

	"encore.dev/beta/errs"
)

// StudentResponse defines the format returned by UCSI's Custom API
type StudentResponse struct {
	Status    string `json:"status"`
	Name      string `json:"name"`
	Grade     string `json:"grade"`
	StudentID string `json:"student_id"`
}

type VerifyRequest struct {
	StudentID   string `json:"student_id"`
	ParentPhone string `json:"parent_phone"`
}

type VerifyResponse struct {
	IsValid bool             `json:"is_valid"`
	Student *StudentResponse `json:"student,omitempty"`
}

// VerifyStudent verifies a student via the UCSI Custom API
//
//encore:api public method=POST path=/api/identity/verify
func VerifyStudent(ctx context.Context, req *VerifyRequest) (*VerifyResponse, error) {
	apiURL := os.Getenv("UCSI_DBS_API_URL")
	apiKey := os.Getenv("UCSI_DBS_API_KEY")

	if apiURL == "" {
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "UCSI DBS API URL not configured",
		}
	}

	// In a real implementation, we would make a secure HTTP call here.
	// For now, we'll implement a mock that correlates to the "Test Student"
	// until Nathan provides the actual endpoint.
	
	client := &http.Client{Timeout: 5 * time.Second}
	_ = client
	_ = apiKey

	// MOCK LOGIC for development/UAT
	if req.StudentID == "1002XXX" || req.StudentID == "TEST123" {
		return &VerifyResponse{
			IsValid: true,
			Student: &StudentResponse{
				Status:    "active",
				Name:      "Alvin Tan (Mock)",
				Grade:     "Primary 4",
				StudentID: req.StudentID,
			},
		}, nil
	}

	return &VerifyResponse{IsValid: false}, nil
}

// GetStudentDetails proxies the call to IT's custom API to fetch student info
func GetStudentDetails(ctx context.Context, studentID string) (*StudentResponse, error) {
	// Implementation for background lookup if needed by the Wallet system
	return &StudentResponse{Name: "Mock Student", StudentID: studentID}, nil
}
