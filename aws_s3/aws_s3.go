// Service AWS S3 for file upload
package aws_s3

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"encore.app/common"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

//encore:service
type Service struct {
	client *s3.Client
}

var S3Client *s3.Client

var awsConfig struct {
	AWSS3BucketName    string `json:"aws_s3_bucket_name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

// initService initializes the site service.
// It is automatically called by Encore on service startup.
func initService() (*Service, error) {
	common.LoadEnv()

	awsConfig.AWSS3BucketName = os.Getenv("AWS_S3_BUCKET_NAME")
	awsConfig.Region = os.Getenv("AWS_REGION")
	awsConfig.AwsAccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	awsConfig.AwsSecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsConfig.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsConfig.AwsAccessKeyId, awsConfig.AwsSecretAccessKey, "")))
	// config.WithSharedConfigFiles([]string{awsConfig.ConfigPath}),
	// config.WithSharedCredentialsFiles([]string{awsConfig.CredentialsPath}))
	if err != nil {
		log.Println("AWS S3 configuration error! Please check your configuration.")
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	client := s3.NewFromConfig(cfg)

	// Get the first page of results for ListObjectsV2 for a bucket
	_, err = client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(awsConfig.AWSS3BucketName),
	})

	if err != nil {
		log.Println("AWS S3 error!")
		//log.Fatal(err)
	}

	log.Println("AWS S3 successfully connected!")

	S3Client = client

	return &Service{client: client}, nil
}

type Document struct {
	DocPath     string         `json:"document_path" valid:"required~DocPath is required"` // s3 path
	File        multipart.File `json:"file" valid:"required~File is required"`
	Url         string         `json:"url"`
	ContentType string         `json:"content_type,omitempty"`
}

func UploadDocument(document Document) (*Document, error) {
	fmt.Println("aws s3 upload document function")
	fmt.Println(document.File)
	if S3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	file_data, err := ConvertFileToByte(document.File)
	if err != nil {
		return nil, err
	}

	// Reset the file pointer before calling to reuse the multipart.File again
	if err := ResetFilePointer(document.File); err != nil {
		return nil, err
	}

	content_type, err := GetFileContentType(document.File)
	if err != nil {
		return nil, err
	}

	// Upload the file to S3
	upload_input := &s3.PutObjectInput{
		Bucket:      aws.String(awsConfig.AWSS3BucketName), // Your S3 bucket name
		Key:         aws.String(document.DocPath),          // The file name you want to save on S3
		Body:        bytes.NewReader(file_data),            // File content
		ContentType: aws.String(content_type),
	}

	_, err = S3Client.PutObject(context.TODO(), upload_input)
	if err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		return nil, err
	}

	document.Url, err = GenerateFileURL(document.DocPath)
	document.ContentType = content_type
	if err != nil {
		return nil, err
	}
	return &document, nil
}

func DeleteDocument(document_url string) error {
	if S3Client == nil {
		return fmt.Errorf("S3 client not initialized")
	}

	var main_s3_url, err = GenerateFileURL("dummy")
	if err != nil {
		return err
	}

	main_s3_url = strings.Replace(main_s3_url, "dummy", "", 1)
	doc_path := strings.Replace(document_url, main_s3_url, "", 1)
	log.Println(doc_path)

	// Prepare the DeleteObject input
	deleteInput := &s3.DeleteObjectInput{
		Bucket: aws.String(awsConfig.AWSS3BucketName), // Your S3 bucket name
		Key:    aws.String(doc_path),                  // The file path in S3
	}

	// Call the DeleteObject method to delete the file
	_, err = S3Client.DeleteObject(context.TODO(), deleteInput)
	if err != nil {
		log.Printf("Failed to delete file from S3: %v", err)
		return err
	}

	log.Printf("Successfully deleted file %s from S3", document_url)
	return nil
}

func ConvertFileToByte(file multipart.File) ([]byte, error) {
	// Read the file into memory
	var buf bytes.Buffer
	_, err := io.Copy(&buf, file)
	if err != nil {
		return nil, fmt.Errorf("file might be corrupted")
	}
	fileData := buf.Bytes()
	return fileData, nil
}

func GetFileContentType(file multipart.File) (string, error) {
	// Read the first 512 bytes
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		log.Println("Failed to read file: " + err.Error())
		return "", err
	}

	// Detect the content type
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

func ResetFilePointer(file io.ReadSeeker) error {
	// Reset the file pointer before calling to reuse the multipart.File again
	_, err := file.Seek(0, io.SeekStart) // Reset to the start of the file
	if err != nil {
		return fmt.Errorf("failed to reset file pointer: %v", err)
	}
	return nil
}

// Generate a pre-signed URL
func GenerateFileURL(docPath string) (string, error) {

	// Generate a pre-signed URL
	presignClient := s3.NewPresignClient(S3Client)
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(awsConfig.AWSS3BucketName),
		Key:    aws.String(docPath),
	})
	if err != nil {
		return "", err
	}

	// Find the index of '?'
	index := strings.Index(req.URL, "?")
	if index != -1 {
		// Trim everything after '?'
		return req.URL[:index], nil
	}
	// Return the original URL if no query parameters exist
	return req.URL, nil

}

func GenerateCSV(data [][]string) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	// Write CSV data
	for _, row := range data {
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write row to CSV: %v", err)
		}
	}

	// Flush and close writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to finalize CSV: %v", err)
	}

	return buffer.Bytes(), nil
}

func UploadCSV(docPath string, data [][]string) (*string, error) {
	if S3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	// Generate CSV
	csvBytes, err := GenerateCSV(data)
	if err != nil {
		return nil, err
	}

	// Upload the file to S3
	upload_input := &s3.PutObjectInput{
		Bucket: aws.String(awsConfig.AWSS3BucketName), // Your S3 bucket name
		Key:    aws.String(docPath),                   // The file name you want to save on S3 (format ex1: "file_example.ext" ,format ex2: "path_1/file_example.ext")
		Body:   bytes.NewReader(csvBytes),             // File content
		// ContentType: aws.String(content_type),
	}

	_, err = S3Client.PutObject(context.TODO(), upload_input)
	if err != nil {
		log.Printf("Failed to upload file to S3: %v", err)
		return nil, err
	}

	url, err := GenerateFileURL(docPath)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// readCSVFromS3 downloads and parses a CSV file from S3
func ReadCSVFromS3(key string) ([][]string, error) {
	if S3Client == nil {
		return nil, fmt.Errorf("S3 client not initialized")
	}

	// Get the object
	resp, err := S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(awsConfig.AWSS3BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}
	defer resp.Body.Close()

	// Read CSV content
	var records [][]string
	reader := csv.NewReader(resp.Body)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV: %v", err)
		}
		records = append(records, record)
	}

	return records, nil
}
