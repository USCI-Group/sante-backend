package products

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"encore.app/aws_s3"
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/common_operations"
	"encore.app/database"
	"encore.app/database/models"
	"encore.app/middleware"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
	"github.com/asaskevich/govalidator"
	googleUUID "github.com/google/uuid" // For generating new UUIDs
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

// request for add category from business owner / admin
type AddProductCategoryRequest struct {
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Description string    `json:"description" valid:"required~Description is required"`
}

// response for get all product category
type GetAllProductCategoryResponse struct {
	Message    string                   `json:"message"`
	BusinessID uuid.UUID                `json:"business_id"`
	Meta       common.Pagination        `json:"meta"`
	Data       []models.ProductCategory `json:"data"`
}

// request for update product category
type UpdateProductCategoryRequest struct {
	BusinessID  uuid.UUID `json:"business_id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Description string    `json:"description" valid:"required~Description is required"`
}

// request for delete product category
type DeleteProductCategoryRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	CategoryID uuid.UUID `json:"category_id"`
}

// request for add product sub category
type AddProductSubCategoryRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	//ProductCategoryID uuid.UUID `json:"product_category_id"`
	Name        string `json:"name" valid:"required~Name is required"`
	Description string `json:"description" valid:"required~Description is required"`
}

// response for get all product sub category
type ProductSubCategory struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	BusinessID  uuid.UUID `json:"business_id"`
	//ProductCategoryID uuid.UUID `json:"product_category_id"`
}

// response for get all product sub category
type GetAllProductSubCategoryResponse struct {
	Message string               `json:"message"`
	Data    []ProductSubCategory `json:"data"`
	Meta    common.Pagination    `json:"meta"`
}

// request for update product sub category data
type UpdateProductSubCategoryRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	//ProductCategoryID    uuid.UUID `json:"product_category_id"`
	ProductSubCategoryID uuid.UUID `json:"product_sub_category_id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
}

type ProductsCategory struct {
	ID          *uuid.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

type ProductsSubCategory struct {
	ID          *uuid.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
}

type ProductCustomizationGroups struct {
	ID          *uuid.UUID                    `json:"id"`
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	InputType   string                        `json:"input_type"`
	IsRequired  bool                          `json:"is_required"`
	Options     []ProductCustomizationOptions `json:"options"`
}

type ProductCustomizationOptions struct {
	ID              *uuid.UUID `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	PriceAdjustment float32    `json:"price_adjustment"`
	IsDefault       bool       `json:"is_default"`
}

// Add Product Request
type AddProductRequest struct {
	// Product Table
	BusinessID  uuid.UUID `json:"business_id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Description string    `json:"description" valid:"required~Description is required"`
	Price       float32   `json:"price"`

	//tags
	Categories    []ProductsCategory    `json:"categories"`
	SubCategories []ProductsSubCategory `json:"sub_categories"`

	IsGrabFood     bool                  `json:"is_grab_food"`
	IsStoreOutlet  bool                  `json:"is_store_outlet"`
	IsShopeeFood   bool                  `json:"is_shopee_food"`
	GrabFoodInfo   models.GrabFoodInfo   `json:"grab_food_info"`
	ShopeeFoodInfo models.ShopeeFoodInfo `json:"shopee_food_info"`
	SortOrder      int                   `json:"sort_order"`

	// Product Customization Groups
	CustomizationGroups []ProductCustomizationGroups `json:"customization_groups"`

	// Modifier Groups
	ModifierGroups []ModifierGroup `json:"modifier_groups"`

	// Product Ingredients
	Ingredients []ProductIngredientMapping `json:"ingredients"`
}

type ProductIngredient struct {
	// ingredient id
	ID uuid.UUID `json:"id"`
	//Name     string    `json:"name"`
	// unit here is use for product ingredient mapping
	Unit string `json:"unit"`
	// quantity here is use for product ingredient mapping
	Quantity float32 `json:"quantity"`
}

type GetAllProductsResponse struct {
	Message string               `json:"message"`
	Data    []ProductInfoMapping `json:"data"`
	Meta    common.Pagination    `json:"meta"`
}

type GetProductResponse struct {
	Message string             `json:"message"`
	Data    ProductInfoMapping `json:"data"`
}

type ProductInfoMapping struct {
	ID                  uuid.UUID                    `json:"id"`
	BusinessID          uuid.UUID                    `json:"business_id"`
	Name                string                       `json:"name"`
	Description         string                       `json:"description"`
	Cost                float32                      `json:"cost"`
	BasePrice           float32                      `json:"base_price"`
	Price               float32                      `json:"price"`
	ImageURL            string                       `json:"image_url"`
	IsActive            bool                         `json:"is_active"`
	IsStoreOutlet       bool                         `json:"is_store_outlet"`
	IsGrabFood          bool                         `json:"is_grab_food"`
	IsShopeeFood        bool                         `json:"is_shopee_food"`
	GrabFoodInfo        models.GrabFoodInfo          `json:"grab_food_info"`
	ShopeeFoodInfo      models.ShopeeFoodInfo        `json:"shopee_food_info"`
	SortOrder           int                          `json:"sort_order"`
	ModifierOptionsID   *uuid.UUID                   `json:"modifier_options_id"`
	CreatedAt           time.Time                    `json:"created_at"`
	ProductCategory     []ProductsCategory           `json:"product_category"`
	ProductSubCategory  []ProductsSubCategory        `json:"product_sub_category"`
	CustomizationGroups []ProductCustomizationGroups `json:"customization_groups"`
	ModifierGroups      []ModifierGroup              `json:"modifier_groups"`
	Ingredients         []ProductIngredientMapping   `json:"ingredients"`
}

type ProductIngredientMapping struct {
	ID             uuid.UUID `json:"id"`
	ProductID      uuid.UUID `json:"product_id"`
	IngredientID   uuid.UUID `json:"ingredient_id"`
	IngredientName string    `json:"name"`
	Unit           string    `json:"unit"`
	Quantity       float32   `json:"quantity"`
}

type GetAllProductsByCategoriesResponse struct {
	Message string           `json:"message"`
	Data    []CustomProducts `json:"data"`
	//Data    []models.Product `json:"data"`
}

type CustomProducts struct {
	ID                  uuid.UUID                    `json:"id"`
	BusinessID          uuid.UUID                    `json:"business_id"`
	Name                string                       `json:"name"`
	Description         string                       `json:"description"`
	Cost                float32                      `json:"cost"`
	BasePrice           float32                      `json:"base_price"`
	Price               float32                      `json:"price"`
	IsGrabFood          bool                         `json:"is_grab_food"`
	IsShopeeFood        bool                         `json:"is_shopee_food"`
	GrabFoodInfo        models.GrabFoodInfo          `json:"grab_food_info"`
	ShopeeFoodInfo      models.ShopeeFoodInfo        `json:"shopee_food_info"`
	ProductCategory     []ProductsCategory           `json:"product_category"`
	ProductSubCategory  []ProductsSubCategory        `json:"product_sub_category"`
	CustomizationGroups []ProductCustomizationGroups `json:"customization_groups"`
}

type UpdateProductRequest struct {
	BusinessID     uuid.UUID             `json:"business_id"`
	SortOrder      int                   `json:"sort_order"`
	Name           string                `json:"name" valid:"required~Name is required"`
	Description    string                `json:"description"`
	Cost           float32               `json:"cost"`
	BasePrice      float32               `json:"base_price"`
	Price          float32               `json:"price"`
	IsActive       bool                  `json:"is_active"`
	IsStoreOutlet  bool                  `json:"is_store_outlet"`
	IsGrabFood     bool                  `json:"is_grab_food"`
	IsShopeeFood   bool                  `json:"is_shopee_food"`
	GrabFoodInfo   models.GrabFoodInfo   `json:"grab_food_info"`
	ShopeeFoodInfo models.ShopeeFoodInfo `json:"shopee_food_info"`

	//tags
	Categories    []ProductsCategory    `json:"categories"`
	SubCategories []ProductsSubCategory `json:"sub_categories"`

	// Product Customization Groups
	CustomizationGroups []ProductCustomizationGroups `json:"customization_groups"`

	// Modifier Groups
	ModifierGroups []ModifierGroup `json:"modifier_groups"`

	// Product Ingredients
	Ingredients []ProductIngredientMapping `json:"ingredients"`

	ModifierOptionsID *uuid.UUID `json:"modifier_options_id"`
}

type CreateIngredientRequest struct {
	BusinessID   uuid.UUID                 `json:"business_id"`
	Name         string                    `json:"name" valid:"required~Name is required"`
	Description  string                    `json:"description"`
	Unit         constants.UnitMeasurement `json:"unit" valid:"required~Unit is required"`
	Quantity     float32                   `json:"quantity" valid:"required~Quantity is required"`
	PricePerUnit float32                   `json:"price_per_unit" valid:"required~Price Per Unit is required"`
	SortOrder    int                       `json:"sort_order"`
}

type UpdateIngredientRequest struct {
	BusinessID   uuid.UUID                 `json:"business_id"`
	Name         string                    `json:"name"`
	Description  string                    `json:"description"`
	Unit         constants.UnitMeasurement `json:"unit"`
	Quantity     float32                   `json:"quantity"`
	PricePerUnit float32                   `json:"price_per_unit"`
	SortOrder    int                       `json:"sort_order"`
}

type GetAllIngredientsResponse struct {
	Message string              `json:"message"`
	Data    []models.Ingredient `json:"data"`
}

type RecipeStep struct {
	Name        string `json:"name" valid:"required~Name is required"`
	Instruction string `json:"instruction" valid:"required~Instruction is required"`
	Precedence  int    `json:"precedence" valid:"required~Precedence is required"`
}

type CreateRecipeRequest struct {
	BusinessID  uuid.UUID    `json:"business_id"`
	ProductID   uuid.UUID    `json:"product_id"`
	Name        string       `json:"name" valid:"required~Name is required"`
	Description string       `json:"description" valid:"required~Description is required"`
	Steps       []RecipeStep `json:"steps" valid:"required~Steps is required"`
}

type CustomRecipeStep struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name" valid:"required~Name is required"`
	Instruction string    `json:"instruction" valid:"required~Instruction is required"`
	Precedence  int       `json:"precedence" valid:"required~Precedence is required"`
}

type GetRecipeResponse struct {
	Message     string             `json:"message"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Steps       []CustomRecipeStep `json:"steps"`
}

type ProductWithRecipe struct {
	ID                uuid.UUID          `json:"id"`
	ProductName       string             `json:"product_name"`
	RecipeName        string             `json:"recipe_name"`
	RecipeDescription string             `json:"recipe_description"`
	Steps             []CustomRecipeStep `json:"steps"`
}

type GetAllRecipesResponse struct {
	Message string              `json:"message"`
	Data    []ProductWithRecipe `json:"data"`
	Meta    common.Pagination   `json:"meta"`
}

type CreateModifierGroupRequest struct {
	BusinessID      uuid.UUID        `json:"business_id"`
	Name            string           `json:"name" valid:"required~Name is required"`
	InputType       models.InputType `json:"input_type" valid:"required~Input Type is required"`
	ModifierOptions []ModifierOption `json:"modifier_options"`
}

type ModifierOption struct {
	ID                 *uuid.UUID                         `json:"id"`
	Name               string                             `json:"name" valid:"required~Name is required"`
	PriceAdjustment    float32                            `json:"price_adjustment"`
	SortOrder          int                                `json:"sort_order"`
	IsActive           bool                               `json:"is_active"`
	IngredientMappings []models.ModifierIngredientMapping `json:"ingredient_mappings"`
}

type ModifierIngredientMapping struct {
	ID                uuid.UUID                 `json:"id"`
	ModifierOptionsID uuid.UUID                 `json:"modifier_options_id"`
	IngredientID      uuid.UUID                 `json:"ingredient_id"`
	Unit              constants.UnitMeasurement `json:"unit"`
	Quantity          float32                   `json:"quantity"`
}

type GetAllModifierGroupsResponse struct {
	Message string          `json:"message"`
	Data    []ModifierGroup `json:"data"`
}

type ModifierGroup struct {
	ID           uuid.UUID        `json:"id"`
	Name         string           `json:"name"`
	InputType    models.InputType `json:"input_type"`
	BusinessID   uuid.UUID        `json:"business_id"`
	MaxSelection int              `json:"max_selection"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    *time.Time       `json:"updated_at"`
	//ModifierOptions []ModifierOption `json:"modifier_options"`
	ModifierOptions []models.ModifierOptions `json:"modifier_options"`
}

type SyncProductToOutletRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	OutletID   uuid.UUID `json:"outlet_id"`
	ProductID  uuid.UUID `json:"product_id"`
	IsAdd      bool      `json:"is_add"`
}

// testing
type GetAllProductsFromOutletResponse struct {
	Message string                         `json:"message"`
	Data    []ProductInfoMappingFromOutlet `json:"data"`
	Meta    common.Pagination              `json:"meta"`
}

type ProductInfoMappingFromOutlet struct {
	ID                 uuid.UUID                  `json:"id"`
	BusinessID         uuid.UUID                  `json:"business_id"`
	Name               string                     `json:"name"`
	Description        string                     `json:"description"`
	Cost               float32                    `json:"cost"`
	BasePrice          float32                    `json:"base_price"`
	Price              float32                    `json:"price"`
	ImageURL           string                     `json:"image_url"`
	IsActiveInBusiness bool                       `json:"is_active_in_business"`
	IsActiveInOutlet   bool                       `json:"is_active_in_outlet"`
	IsGrabFood         bool                       `json:"is_grab_food"`
	IsShopeeFood       bool                       `json:"is_shopee_food"`
	CreatedAt          time.Time                  `json:"created_at"`
	ProductCategory    []ProductsCategory         `json:"product_category"`
	ProductSubCategory []ProductsSubCategory      `json:"product_sub_category"`
	ModifierGroups     []ModifierGroup            `json:"modifier_groups"`
	Ingredients        []ProductIngredientMapping `json:"ingredients"`
}

type SyncModifierOptionToOutletRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	OutletID   uuid.UUID `json:"outlet_id"`
}

type GetOutletModifierOptionsResponse struct {
	Data []common.OutletModifierOption `json:"data"`
}

type GetModifierOptionsResponse struct {
	Data []models.ModifierOptions `json:"data"`
}

// API to add category
//
//encore:api auth method=POST path=/api/products/add-category
func (s *Service) AddProductCategory(ctx context.Context, req *AddProductCategoryRequest) (*models.ProductCategory, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.CreateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// add product category
	productCategory := models.ProductCategory{
		Name:        req.Name,
		Description: req.Description,
		BusinessID:  &req.BusinessID,
		CreatedAt:   time.Now(),
		UpdatedAt:   nil,
		DeletedAt:   gorm.DeletedAt{},
	}

	//create product category
	result := s.db.Create(&productCategory)
	if result.Error != nil {
		return nil, result.Error
	}

	return &productCategory, nil
}

// API to get product category
//
//encore:api auth method=GET path=/api/business/products/get-product-category/:id/:page/:page_size
func (s *Service) GetAllProductCategory(ctx context.Context, id uuid.UUID, page int, page_size int) (*GetAllProductCategoryResponse, error) {
	// default page
	if page <= 0 {
		page = 1
	}

	// default page size
	if page_size <= 0 {
		page_size = 10
	}

	err := middleware.CheckPermission(constants.ReadProductAction, &id, nil)
	if err != nil {
		return nil, err
	}

	// get all product category
	var totalProductCategory int64
	err = s.db.Model(&models.ProductCategory{}).Where("business_id = ?", id).Count(&totalProductCategory).Error
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * page_size
	totalPages := (totalProductCategory / int64(page_size)) + 1

	var productCategoryList []models.ProductCategory
	err = s.db.Model(&models.ProductCategory{}).Where("business_id = ?", id).Offset(offset).Limit(page_size).Find(&productCategoryList).Error
	if err != nil {
		return nil, err
	}

	sort.Slice(productCategoryList, func(i, j int) bool {
		return productCategoryList[i].Name < productCategoryList[j].Name
	})

	return &GetAllProductCategoryResponse{
		Message:    "Product category fetched successfully",
		BusinessID: id,
		Meta: common.Pagination{
			Page:       page,
			PageSize:   page_size,
			TotalPages: int(totalPages),
			Total:      int64(totalProductCategory),
		},
		Data: productCategoryList,
	}, nil
}

// API to update product category
//
//encore:api auth method=PUT path=/api/products/update-product-category
func (s *Service) UpdateProductCategory(ctx context.Context, req *UpdateProductCategoryRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.UpdateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check product category existence
	var productCategory models.ProductCategory
	err = s.db.Model(&models.ProductCategory{}).Where("id = ? AND business_id = ? ", req.CategoryID, req.BusinessID).First(&productCategory).Error
	if err != nil {
		return nil, err
	}

	// update product category
	updateData := make(map[string]interface{})
	common.HandleFieldUpdate(&productCategory.Name, &req.Name, "name", updateData, false)
	common.HandleFieldUpdate(&productCategory.Description, &req.Description, "description", updateData, false)

	result := s.db.Model(&models.ProductCategory{}).Where("id = ? AND business_id = ?", req.CategoryID, req.BusinessID).Updates(updateData)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &common.BasicResponse{
		Message: "Product category updated successfully",
	}, nil
}

//encore:api auth raw method=POST path=/api/products/category/upload-image
func (s *Service) UploadProductCategoryImage(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		http.Error(w, "Product Category ID is required", http.StatusBadRequest)
		return
	}

	temp_uuid, err := googleUUID.Parse(id)
	if err != nil {
		http.Error(w, "Invalid Product Category ID format", http.StatusBadRequest)
		return
	}

	productCategory := models.ProductCategory{}
	err = s.db.Model(&models.ProductCategory{}).Where("id = ?", temp_uuid).First(&productCategory).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	err = s.db.Model(&models.Business{}).Where("id = ?", productCategory.BusinessID).First(&business).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(1 << 20); err != nil { // Limit to 1 MB
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
	file_name := productCategory.Name + "_" + time.Now().Format("20060102150405") + file_extension
	document.DocPath = "business/" + business.RegistrationNumber + "/images/product-categories/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// remove old image from s3
	if productCategory.ImageURL != "" {
		aws_s3.DeleteDocument(productCategory.ImageURL)
	}

	productCategory.ImageURL = document_res.Url

	err = s.db.Save(&productCategory).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}
	fmt.Println("productCategory.ImageURL -------------> ", productCategory.ImageURL)

	w.WriteHeader(http.StatusOK)
}

//encore:api auth raw method=POST path=/api/products/category/upload-banner
func (s *Service) UploadProductCategoryBanner(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		http.Error(w, "Product Category ID is required", http.StatusBadRequest)
		return
	}

	productCategory := models.ProductCategory{}
	if err := s.db.Model(&models.ProductCategory{}).Where("id = ?", id).First(&productCategory).Error; err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	if err := s.db.Model(&models.Business{}).Where("id = ?", productCategory.BusinessID).First(&business).Error; err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(1 << 20); err != nil { // Limit to 1 MB
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
	file_name := productCategory.Name + "_" + time.Now().Format("20060102150405") + file_extension
	document.DocPath = "business/" + business.RegistrationNumber + "/images/product-categories/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// remove old image from s3
	if productCategory.BannerURL != "" {
		aws_s3.DeleteDocument(productCategory.BannerURL)
	}

	productCategory.BannerURL = document_res.Url

	if err = s.db.Save(&productCategory).Error; err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// API to delete product category
//
//encore:api auth method=DELETE path=/api/products/delete-product-category/:business_id/:product_id
func (s *Service) DeleteProductCategory(ctx context.Context, business_id uuid.UUID, product_id uuid.UUID) (*common.BasicResponse, error) {
	fmt.Println("business_id ------>", business_id)
	fmt.Println("product_id ------>", product_id)

	err := middleware.CheckPermission(constants.DeleteProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	// delete product category
	result := s.db.Model(&models.ProductCategory{}).Where("id= ? AND business_id = ?", product_id, business_id).Delete(&models.ProductCategory{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &common.BasicResponse{
		Message: "Product category deleted successfully",
	}, nil
}

// API to add sub category
//
//encore:api auth method=POST path=/api/products/add-sub-category
func (s *Service) AddProductSubCategory(ctx context.Context, req *AddProductSubCategoryRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.CreateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business).Error
	if err != nil {
		return nil, err
	}

	//add product sub category
	productSubCategory := models.ProductSubCategory{
		Name:        req.Name,
		Description: req.Description,
		//ProductCategoryID: req.ProductCategoryID,
		BusinessID: &req.BusinessID,
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
		DeletedAt:  gorm.DeletedAt{},
	}

	//create product sub category
	result := s.db.Create(&productSubCategory)
	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Product sub category added successfully",
	}, nil
}

// API to get all product sub category
//
//encore:api auth method=GET path=/api/products/get-all-product-sub-category/:id/:page/:page_size
func (s *Service) GetAllProductSubCategory(ctx context.Context, id uuid.UUID, page int, page_size int) (*GetAllProductSubCategoryResponse, error) {

	if page <= 0 {
		page = 1
	}
	if page_size <= 0 {
		page_size = 10
	}

	err := middleware.CheckPermission(constants.ReadProductAction, &id, nil)
	if err != nil {
		return nil, err
	}

	var totalProductSubCategory int64
	err = s.db.Model(&models.ProductSubCategory{}).Where("business_id = ?", id).Count(&totalProductSubCategory).Error
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * page_size
	totalPages := (totalProductSubCategory / int64(page_size)) + 1

	fmt.Println("totalPages ------>", totalPages)
	fmt.Println("totalProductSubCategory ------>", totalProductSubCategory)
	fmt.Println("offset ------>", offset)
	fmt.Println("page_size ------>", page_size)

	// get all product sub category
	var productSubCategoryList []ProductSubCategory
	err = s.db.Model(&models.ProductSubCategory{}).Where("business_id = ?", id).Limit(page_size).Offset(offset).Find(&productSubCategoryList).Error
	if err != nil {
		return nil, err
	}

	// sort responseData by name
	sort.Slice(productSubCategoryList, func(i, j int) bool {
		return productSubCategoryList[i].Name < productSubCategoryList[j].Name
	})

	fmt.Println("productSubCategoryList ------>", productSubCategoryList)

	return &GetAllProductSubCategoryResponse{
		Message: "Product sub category fetched successfully",
		Data:    productSubCategoryList,
		Meta: common.Pagination{
			Page:       page,
			PageSize:   page_size,
			TotalPages: int(totalPages),
			Total:      int64(totalProductSubCategory),
		},
	}, nil
}

// API to get all product sub category
//
//encore:api auth method=GET path=/api/products/get-all-product-sub-category/:id
func (s *Service) GetAllProductSubCategoryWithoutPagination(ctx context.Context, id uuid.UUID) (*GetAllProductSubCategoryResponse, error) {

	err := middleware.CheckPermission(constants.ReadProductAction, &id, nil)
	if err != nil {
		return nil, err
	}

	var totalProductSubCategory int64
	err = s.db.Model(&models.ProductSubCategory{}).Where("business_id = ?", id).Count(&totalProductSubCategory).Error
	if err != nil {
		return nil, err
	}

	// get all product sub category
	var productSubCategoryList []ProductSubCategory
	err = s.db.Model(&models.ProductSubCategory{}).Where("business_id = ?", id).Find(&productSubCategoryList).Error
	if err != nil {
		return nil, err
	}
	// sort responseData by name
	sort.Slice(productSubCategoryList, func(i, j int) bool {
		return productSubCategoryList[i].Name < productSubCategoryList[j].Name
	})
	fmt.Println("productSubCategoryList ------>", productSubCategoryList)

	return &GetAllProductSubCategoryResponse{
		Message: "Product sub category fetched successfully",
		Data:    productSubCategoryList,
		Meta: common.Pagination{
			Total: int64(totalProductSubCategory),
		},
	}, nil
}

// API to update product sub category
//
//encore:api auth method=PUT path=/api/products/update-product-sub-category
func (s *Service) UpdateProductSubCategory(ctx context.Context, req *UpdateProductSubCategoryRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.UpdateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check product sub category existence
	var productSubCategory models.ProductSubCategory
	err = s.db.Model(&models.ProductSubCategory{}).Where("id = ? AND business_id = ?", req.ProductSubCategoryID, req.BusinessID).First(&productSubCategory).Error
	if err != nil {
		return nil, err
	}

	updateData := make(map[string]interface{})
	common.HandleFieldUpdate(&productSubCategory.Name, &req.Name, "name", updateData, false)
	common.HandleFieldUpdate(&productSubCategory.Description, &req.Description, "description", updateData, false)

	result := s.db.Model(&models.ProductSubCategory{}).Where("id = ? AND business_id = ?", req.ProductSubCategoryID, req.BusinessID).Updates(updateData)
	if result.Error != nil {
		return nil, result.Error
	}

	fmt.Println("result.RowsAffected -----------> ", result.RowsAffected)

	return &common.BasicResponse{
		Message: "Product sub category updated successfully",
	}, nil

}

// API to delete product sub category
//
//encore:api auth method=DELETE path=/api/products/delete-product-sub-category/:business_id/:product_sub_category_id
func (s *Service) DeleteProductSubCategory(ctx context.Context, business_id uuid.UUID, product_sub_category_id uuid.UUID) (*common.BasicResponse, error) {

	err := middleware.CheckPermission(constants.DeleteProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	// delete product category
	result := s.db.Model(&models.ProductSubCategory{}).Where("id= ? AND business_id = ?", product_sub_category_id, business_id).Delete(&models.ProductSubCategory{})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return &common.BasicResponse{
		Message: "Product sub category deleted successfully",
	}, nil
}

// API to add product
// If the id of category or sub category are not present in the database, then just remain empty.
// It will automatically add the new tags based on provided name and description in the database.
//
//encore:api auth method=POST path=/api/products/add-product
func (s *Service) AddProduct(ctx context.Context, req *AddProductRequest) (*models.Product, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	if err := validatePlatformInfo(req); err != nil {
		return nil, err
	}

	err := middleware.CheckPermission(constants.CreateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id= ?", req.BusinessID).First(&business).Error
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// check product category tags need to create or not
	var productCategoryList []models.ProductCategory
	for _, category := range req.Categories {
		// if category id is nil, then create new product category in database and add to productCategoryList
		if category.ID == nil {
			newCat := models.ProductCategory{
				Name:        category.Name,
				Description: category.Description,
				BusinessID:  &req.BusinessID,
			}
			result := tx.Create(&newCat)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			productCategoryList = append(productCategoryList, newCat)
		} else {
			// get product category without add to database
			var productCategory models.ProductCategory
			result := tx.Model(&models.ProductCategory{}).Where("id = ?", category.ID).First(&productCategory)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			productCategoryList = append(productCategoryList, productCategory)
		}

	}

	// check product sub category tags need to create or not
	var productSubCategoryList []models.ProductSubCategory
	for _, subCategory := range req.SubCategories {
		if subCategory.ID == nil {
			newSubCat := models.ProductSubCategory{
				Name:        subCategory.Name,
				Description: subCategory.Description,
				//ProductCategoryID: productCategoryList[0].ID,
				BusinessID: &req.BusinessID,
			}
			result := tx.Create(&newSubCat)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			productSubCategoryList = append(productSubCategoryList, newSubCat)
		} else {
			var productSubCategory models.ProductSubCategory
			result := tx.Model(&models.ProductSubCategory{}).Where("id = ?", subCategory.ID).First(&productSubCategory)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
			productSubCategoryList = append(productSubCategoryList, productSubCategory)
		}
	}

	if req.Price < 0 {
		tx.Rollback()
		return nil, errors.New("price cannot be negative")
	}
	// add product
	product := models.Product{
		Name:           req.Name,
		Description:    req.Description,
		Price:          req.Price,
		BasePrice:      req.Price,
		BusinessID:     &req.BusinessID,
		IsGrabFood:     req.IsGrabFood,
		IsStoreOutlet:  req.IsStoreOutlet,
		IsShopeeFood:   req.IsShopeeFood,
		GrabFoodInfo:   req.GrabFoodInfo,
		ShopeeFoodInfo: req.ShopeeFoodInfo,
		SortOrder:      req.SortOrder,
	}
	result := tx.Create(&product)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// add product category mapping
	var productCatMappingList []models.ProductCategoryMapping
	for _, category := range productCategoryList {

		for _, subCategory := range productSubCategoryList {
			productCatMapping := models.ProductCategoryMapping{
				ProductID:            product.ID,
				ProductCategoryID:    category.ID,
				ProductSubCategoryID: &subCategory.ID,
			}
			productCatMappingList = append(productCatMappingList, productCatMapping)
		}
	}

	// add product category mapping to database
	batchSize := 100
	result = tx.CreateInBatches(productCatMappingList, batchSize)
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// Add verification step
	var count int64
	err = tx.Model(&models.ProductCategoryMapping{}).Where("product_id = ?", product.ID).Count(&count).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// add product modifier groups
	if len(req.ModifierGroups) > 0 {
		for _, modifierGroup := range req.ModifierGroups {
			newModifierGroup := models.ProductModifierMapping{
				ProductID:       product.ID,
				ModifierGroupID: modifierGroup.ID,
				MaxSelection:    modifierGroup.MaxSelection,
				CreatedAt:       time.Now(),
				UpdatedAt:       nil,
			}
			result = tx.Create(&newModifierGroup)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
		}
	}

	// add product ingredients
	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			newIngredientMapping := models.ProductIngredientMapping{
				ProductID: product.ID,
				// noted that .ID is due to from fontend, the ID === IngredientID
				IngredientID: ingredient.ID,
				Unit:         constants.UnitMeasurement(ingredient.Unit),
				Quantity:     ingredient.Quantity,
				CreatedAt:    time.Now(),
				UpdatedAt:    nil,
				DeletedAt:    gorm.DeletedAt{},
			}
			result = tx.Create(&newIngredientMapping)
			if result.Error != nil {
				tx.Rollback()
				return nil, result.Error
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &product, nil
}

func validatePlatformInfo(req *AddProductRequest) error {
	if req.IsGrabFood {
		if req.GrabFoodInfo.GrabFoodPrice == nil || *req.GrabFoodInfo.GrabFoodPrice <= 0 {
			return errors.New("grabfood info is required if grabfood is enabled")
		}
	}
	if req.IsShopeeFood {
		if req.ShopeeFoodInfo.ShopeeFoodPrice == nil || *req.ShopeeFoodInfo.ShopeeFoodPrice <= 0 {
			return errors.New("shopeefood info is required if shopeefood is enabled")
		}
	}
	return nil
}

//encore:api auth raw method=POST path=/api/products/upload/image
func (s *Service) UploadProductImage(w http.ResponseWriter, req *http.Request) {
	product_id := req.FormValue("id")
	if product_id == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	temp_uuid, err := googleUUID.Parse(product_id)
	if err != nil {
		http.Error(w, "Invalid Product ID format", http.StatusBadRequest)
		return
	}

	product := models.Product{}
	err = s.db.Model(&models.Product{}).Where("id = ?", temp_uuid).First(&product).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	err = s.db.Model(&models.Business{}).Where("id = ?", product.BusinessID).First(&business).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(2 << 20); err != nil { // Limit to 2 MB
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
	file_name := product.Name + "_" + time.Now().Format("20060102150405") + file_extension
	document.DocPath = "business/" + business.RegistrationNumber + "/images/products/" + file_name
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// remove old image from s3
	if product.ImageURL != "" {
		aws_s3.DeleteDocument(product.ImageURL)
	}

	product.ImageURL = document_res.Url

	err = s.db.Save(&product).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}
	/* err = s.db.Model(&models.Product{}).Where("id = ?", product.ID).Update("image_url", product.ImageURL).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	} */
	fmt.Println("product.ImageURL -------------> ", product.ImageURL)

	w.WriteHeader(http.StatusOK)
}

// API to get products based on categories
//
//encore:api auth method=GET path=/api/business/products/get-products-by-categories/:id/:product_category_id
func (s *Service) GetAllProductsByCategories(ctx context.Context, id uuid.UUID, product_category_id uuid.UUID) (*GetAllProductsByCategoriesResponse, error) {
	err := middleware.CheckPermission(constants.ReadProductAction, &id, nil)
	if err != nil {
		return nil, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id=?", id).First(&business).Error
	if err != nil {
		return nil, err
	}

	var productCategoryMappingList []models.ProductCategoryMapping
	err = s.db.Model(&models.ProductCategoryMapping{}).Where("product_category_id = ?", product_category_id).Find(&productCategoryMappingList).Error
	if err != nil {
		return nil, err
	}

	// get all product ids
	var productIDs []uuid.UUID
	for _, mapping := range productCategoryMappingList {
		productIDs = append(productIDs, mapping.ProductID)
	}

	// get all products based on product ids
	var products []CustomProducts
	err = s.db.Model(&models.Product{}).Where("id IN (?)", productIDs).Find(&products).Error
	if err != nil {
		return nil, err
	}

	fmt.Println("---------------start-----------------")
	for i, product := range products {
		var productSubCategories []models.ProductCategoryMapping
		err = s.db.Model(&models.ProductCategoryMapping{}).Where("product_id = ?", product.ID).Find(&productSubCategories).Error
		if err != nil {
			return nil, err
		}

		// Create a map to track unique subcategories
		uniqueSubCategories := make(map[uuid.UUID]bool)

		for _, productSubCategory := range productSubCategories {
			var productSubCategoryDetail models.ProductSubCategory
			err = s.db.Model(&models.ProductSubCategory{}).Where("id = ?", productSubCategory.ProductSubCategoryID).First(&productSubCategoryDetail).Error
			if err != nil {
				return nil, err
			}

			if !uniqueSubCategories[productSubCategoryDetail.ID] {
				uniqueSubCategories[productSubCategoryDetail.ID] = true
				products[i].ProductSubCategory = append(products[i].ProductSubCategory, ProductsSubCategory{
					ID:          &productSubCategoryDetail.ID,
					Name:        productSubCategoryDetail.Name,
					Description: productSubCategoryDetail.Description,
				})
			}
		}

	}

	return &GetAllProductsByCategoriesResponse{
		Message: "Products fetched successfully",
		Data:    products,
	}, nil

}

// API to get all products based on business id with pagination
//
//encore:api auth method=GET path=/api/business/products/get-all-products/:business_id/:page/:page_size
func (s *Service) GetAllProducts(ctx context.Context, business_id uuid.UUID, page int, page_size int) (*GetAllProductsResponse, error) {
	return s.getAllProducts(business_id, page, page_size, nil)
}

type GetAllProductsWithFiltersRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	Search     *string   `json:"search"`
}

// API to get all products based on business id with pagination
//
//encore:api auth method=POST path=/api/business/products/with-filters
func (s *Service) GetAllProductsWithFilters(ctx context.Context, req *GetAllProductsWithFiltersRequest) (*GetAllProductsResponse, error) {
	return s.getAllProducts(req.BusinessID, req.Page, req.PageSize, req.Search)
}

func (s *Service) getAllProducts(business_id uuid.UUID, page int, page_size int, search *string) (*GetAllProductsResponse, error) {
	// default page
	if page <= 0 {
		page = 1
	}

	// default page size
	if page_size <= 0 {
		page_size = 10
	}

	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id=?", business_id).First(&business).Error
	if err != nil {
		return nil, err
	}

	productQuery := s.db.Model(&models.Product{}).Where("business_id = ?", business_id)
	if search != nil {
		productQuery = productQuery.Where("LOWER(name) LIKE LOWER(?)", "%"+*search+"%")
	}

	var totalProducts int64
	err = productQuery.Count(&totalProducts).Error
	if err != nil {
		return nil, err
	}
	offset := (page - 1) * page_size
	//totalPages := int((totalProducts + int64(page_size) - 1) / int64(page_size))
	totalPages := int(math.Ceil(float64(totalProducts) / float64(page_size)))
	// get all products
	var products []models.Product
	err = productQuery.
		Order("created_at ASC").
		Offset(offset).
		Limit(page_size).Find(&products).Error
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return &GetAllProductsResponse{
			Message: "No products found",
			Data:    nil,
			Meta: common.Pagination{
				Page:     page,
				PageSize: page_size,
				Total:    int64(totalPages),
			},
		}, nil
	}

	// get all product ids
	var productIDs []uuid.UUID
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	// get all product category mapping
	var mappings []models.ProductCategoryMapping
	err = s.db.Preload("ProductCategory").Preload("ProductSubCategory").Preload("Product").Where("product_id IN (?)", productIDs).Order("created_at ASC").Find(&mappings).Error
	if err != nil {
		return nil, err
	}

	productMap := make(map[uuid.UUID]*ProductInfoMapping)
	for _, m := range mappings {
		//fmt.Println("\nm.ProductID -------------> ", m.ProductID)
		productID := m.ProductID
		if _, exists := productMap[productID]; !exists {
			productMap[productID] = &ProductInfoMapping{
				ID:                 productID,
				Name:               m.Product.Name,
				Description:        m.Product.Description,
				Cost:               m.Product.Cost,
				BasePrice:          m.Product.BasePrice,
				Price:              m.Product.Price,
				ImageURL:           m.Product.ImageURL,
				IsActive:           m.Product.IsActive,
				IsStoreOutlet:      m.Product.IsStoreOutlet,
				IsGrabFood:         m.Product.IsGrabFood,
				IsShopeeFood:       m.Product.IsShopeeFood,
				GrabFoodInfo:       m.Product.GrabFoodInfo,
				ShopeeFoodInfo:     m.Product.ShopeeFoodInfo,
				SortOrder:          m.Product.SortOrder,
				ModifierOptionsID:  m.Product.ModifierOptionsID,
				ProductCategory:    []ProductsCategory{},
				ProductSubCategory: []ProductsSubCategory{},
				ModifierGroups:     []ModifierGroup{},
				CreatedAt:          m.Product.CreatedAt,
				BusinessID:         *m.Product.BusinessID,
			}
		}

		// Add product category to product mapping
		if m.ProductCategory != (models.ProductCategory{}) {
			cat := ProductsCategory{
				ID:          &m.ProductCategory.ID,
				Name:        m.ProductCategory.Name,
				Description: m.ProductCategory.Description,
			}
			//check duplicate category
			found := false
			for _, existingCat := range productMap[productID].ProductCategory {
				fmt.Println("\nexistingCat.ID -------------> ", existingCat.ID)
				fmt.Println("cat.ID -------------> ", cat.ID)
				if *existingCat.ID == *cat.ID {
					fmt.Println("existingCat.ID == cat.ID")
					found = true
					break
				}
			}
			if !found {
				productMap[productID].ProductCategory = append(productMap[productID].ProductCategory, cat)
			}
		}

		// Add product sub category to product mapping
		// (Note: m.ProductSubCategory.ID will be zero if no sub-category exists.)
		if m.ProductSubCategory.ID != (uuid.UUID{}) {
			sub := ProductsSubCategory{
				ID:          &m.ProductSubCategory.ID,
				Name:        m.ProductSubCategory.Name,
				Description: m.ProductSubCategory.Description,
			}
			// Check for duplicates.
			found := false
			for _, existingSub := range productMap[productID].ProductSubCategory {
				if *existingSub.ID != uuid.Nil && *existingSub.ID == *sub.ID {
					found = true
					break
				}
			}
			if !found {
				productMap[productID].ProductSubCategory = append(productMap[productID].ProductSubCategory, sub)
			}
		}

	}

	// add modifier groups to productMap
	for _, productID := range productIDs {
		var modifierMapping []models.ProductModifierMapping
		err = s.db.Where("product_id = ?", productID).Find(&modifierMapping).Error
		if err != nil {
			return nil, err

		}

		// add modifier groups to products
		for _, mapping := range modifierMapping {
			var modifierGroup models.ModifierGroups
			err = s.db.Where("id = ?", mapping.ModifierGroupID).First(&modifierGroup).Error
			if err != nil {
				return nil, err
			}

			var modifierOptions []models.ModifierOptions
			err = s.db.Where("modifier_group_id = ?", mapping.ModifierGroupID).Find(&modifierOptions).Error
			if err != nil {
				return nil, err
			}
			productMap[productID].ModifierGroups = append(productMap[productID].ModifierGroups, ModifierGroup{
				ID:              mapping.ModifierGroupID,
				Name:            modifierGroup.Name,
				InputType:       modifierGroup.InputType,
				BusinessID:      modifierGroup.BusinessID,
				MaxSelection:    mapping.MaxSelection,
				CreatedAt:       modifierGroup.CreatedAt,
				UpdatedAt:       modifierGroup.UpdatedAt,
				ModifierOptions: modifierOptions,
			})
		}

	}

	// get product ingredients
	for _, product := range productIDs {
		var productIngredientMapping []models.ProductIngredientMapping
		err = s.db.Preload("Ingredient").Where("product_id = ?", product).Find(&productIngredientMapping).Error
		if err != nil {
			return nil, err
		}

		for _, ingredient := range productIngredientMapping {
			productMap[product].Ingredients = append(productMap[product].Ingredients, ProductIngredientMapping{
				ID:             ingredient.ID,
				ProductID:      ingredient.ProductID,
				IngredientID:   ingredient.IngredientID,
				IngredientName: ingredient.Ingredient.Name,
				Unit:           string(ingredient.Unit),
				Quantity:       ingredient.Quantity,
			})
		}
	}
	// convert productMap to slice for response
	responseData := make([]ProductInfoMapping, 0, len(productMap))
	for _, mapping := range productMap {
		responseData = append(responseData, *mapping)
	}

	// Sort by CreatedAt (oldest to newest), then by ID for tie-breaking
	sort.Slice(responseData, func(i, j int) bool {
		if responseData[i].CreatedAt.Equal(responseData[j].CreatedAt) {
			// If creation times are equal, sort by ID to ensure consistent ordering
			return responseData[i].ID.String() < responseData[j].ID.String()
		}
		return responseData[i].CreatedAt.Before(responseData[j].CreatedAt)
	})

	return &GetAllProductsResponse{
		Message: "Products fetched successfully",
		Data:    responseData,
		Meta: common.Pagination{
			Page:       page,
			PageSize:   page_size,
			TotalPages: totalPages,
			Total:      int64(totalProducts),
		},
	}, nil
}

// API to get all products based on business id
//
//encore:api auth method=GET path=/api/business/products/get-all-products/:business_id
func (s *Service) GetAllProductsWithoutPagination(ctx context.Context, business_id uuid.UUID) (*GetAllProductsResponse, error) {
	isPermitted1 := true
	isPermitted2 := true

	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		isPermitted1 = false
	}
	err = middleware.CheckPermission(constants.ReadFinanceTransactionAction, &business_id, nil)
	if err != nil {
		isPermitted2 = false
	}
	if !isPermitted1 && !isPermitted2 {
		return nil, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id=?", business_id).First(&business).Error
	if err != nil {
		return nil, err
	}

	var totalProducts int64
	err = s.db.Model(&models.Product{}).Where("business_id = ?", business_id).Count(&totalProducts).Error
	if err != nil {
		return nil, err
	}

	// get all products
	var products []models.Product
	err = s.db.Model(&models.Product{}).Where("business_id = ?", business_id).Find(&products).Error
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return &GetAllProductsResponse{
			Message: "No products found",
			Data:    nil,
		}, nil
	}

	// get all product ids
	var productIDs []uuid.UUID
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}

	// get all product category mapping
	var mappings []models.ProductCategoryMapping
	err = s.db.Preload("ProductCategory").Preload("ProductSubCategory").Preload("Product").Where("product_id IN (?)", productIDs).Find(&mappings).Error
	if err != nil {
		return nil, err
	}

	productMap := make(map[uuid.UUID]*ProductInfoMapping)
	for _, m := range mappings {
		//fmt.Println("\nm.ProductID -------------> ", m.ProductID)
		productID := m.ProductID
		if _, exists := productMap[productID]; !exists {
			productMap[productID] = &ProductInfoMapping{
				ID:                  productID,
				BusinessID:          *m.Product.BusinessID,
				Name:                m.Product.Name,
				Description:         m.Product.Description,
				Cost:                m.Product.Cost,
				BasePrice:           m.Product.BasePrice,
				Price:               m.Product.Price,
				ImageURL:            m.Product.ImageURL,
				IsActive:            m.Product.IsActive,
				CreatedAt:           m.Product.CreatedAt,
				IsGrabFood:          m.Product.IsGrabFood,
				IsShopeeFood:        m.Product.IsShopeeFood,
				GrabFoodInfo:        m.Product.GrabFoodInfo,
				ShopeeFoodInfo:      m.Product.ShopeeFoodInfo,
				ProductCategory:     []ProductsCategory{},
				ProductSubCategory:  []ProductsSubCategory{},
				CustomizationGroups: []ProductCustomizationGroups{},
				ModifierGroups:      []ModifierGroup{},
			}
		}

		// Add product category to product mapping
		if m.ProductCategory != (models.ProductCategory{}) {
			cat := ProductsCategory{
				ID:          &m.ProductCategory.ID,
				Name:        m.ProductCategory.Name,
				Description: m.ProductCategory.Description,
			}
			//check duplicate category
			found := false
			for _, existingCat := range productMap[productID].ProductCategory {
				fmt.Println("\nexistingCat.ID -------------> ", existingCat.ID)
				fmt.Println("cat.ID -------------> ", cat.ID)
				if *existingCat.ID == *cat.ID {
					fmt.Println("existingCat.ID == cat.ID")
					found = true
					break
				}
			}
			fmt.Println("finish for llop")
			if !found {
				productMap[productID].ProductCategory = append(productMap[productID].ProductCategory, cat)
			}
		}

		// Add product sub category to product mapping
		// (Note: m.ProductSubCategory.ID will be zero if no sub-category exists.)
		if m.ProductSubCategory.ID != (uuid.UUID{}) {
			sub := ProductsSubCategory{
				ID:          &m.ProductSubCategory.ID,
				Name:        m.ProductSubCategory.Name,
				Description: m.ProductSubCategory.Description,
			}
			// Check for duplicates.
			found := false
			for _, existingSub := range productMap[productID].ProductSubCategory {
				if *existingSub.ID != uuid.Nil && *existingSub.ID == *sub.ID {
					found = true
					break
				}
			}
			if !found {
				productMap[productID].ProductSubCategory = append(productMap[productID].ProductSubCategory, sub)
			}
		}

	}

	// add modifier groups to productMap
	for _, productID := range productIDs {
		var modifierMapping []models.ProductModifierMapping
		err = s.db.Where("product_id = ?", productID).Find(&modifierMapping).Error
		if err != nil {
			return nil, err

		}

		// add modifier groups to products
		for _, mapping := range modifierMapping {
			var modifierGroup models.ModifierGroups
			err = s.db.Where("id = ?", mapping.ModifierGroupID).First(&modifierGroup).Error
			if err != nil {
				return nil, err
			}

			var modifierOptions []models.ModifierOptions
			err = s.db.Where("modifier_group_id = ?", mapping.ModifierGroupID).Find(&modifierOptions).Error
			if err != nil {
				return nil, err
			}
			productMap[productID].ModifierGroups = append(productMap[productID].ModifierGroups, ModifierGroup{
				ID:              mapping.ModifierGroupID,
				Name:            modifierGroup.Name,
				InputType:       modifierGroup.InputType,
				BusinessID:      modifierGroup.BusinessID,
				CreatedAt:       modifierGroup.CreatedAt,
				UpdatedAt:       modifierGroup.UpdatedAt,
				ModifierOptions: modifierOptions,
			})
		}

	}

	// convert productMap to slice for response
	responseData := make([]ProductInfoMapping, 0, len(productMap))
	for _, mapping := range productMap {
		responseData = append(responseData, *mapping)
	}

	// sort responseData by name
	sort.Slice(responseData, func(i, j int) bool {
		return responseData[i].Name < responseData[j].Name
	})

	return &GetAllProductsResponse{
		Message: "Products fetched successfully",
		Data:    responseData,
	}, nil
}

// API to get product by id
//
//encore:api auth method=GET path=/api/products/get-product/:business_id/:product_id
func (s *Service) GetProduct(ctx context.Context, business_id uuid.UUID, product_id uuid.UUID) (*GetProductResponse, error) {

	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	// check product id existence
	var product models.Product
	err = s.db.Model(&models.Product{}).Where("id= ? AND business_id = ?", product_id, business_id).First(&product).Error
	if err != nil {
		return nil, err
	}

	var productCategoryMapping []models.ProductCategoryMapping
	err = s.db.Preload("ProductCategory").Preload("ProductSubCategory").Preload("Product").Where("product_id = ?", product_id).Find(&productCategoryMapping).Error
	if err != nil {
		return nil, err
	}

	var productCategoryList []ProductsCategory
	var productSubCategoryList []ProductsSubCategory
	for _, mapping := range productCategoryMapping {
		// add product category to list
		if mapping.ProductCategory != (models.ProductCategory{}) {
			// check if product category is already in the list
			found := false
			for _, existingCategory := range productCategoryList {
				if *existingCategory.ID == mapping.ProductCategory.ID {
					found = true
					break
				}
			}
			if !found {
				productCategoryList = append(productCategoryList, ProductsCategory{
					ID:          &mapping.ProductCategory.ID,
					Name:        mapping.ProductCategory.Name,
					Description: mapping.ProductCategory.Description,
				})
			}
		}
		// add product sub category to list
		if mapping.ProductSubCategory != (models.ProductSubCategory{}) {
			// check if product sub category is already in the list
			found := false
			for _, existingSubCategory := range productSubCategoryList {
				if *existingSubCategory.ID == mapping.ProductSubCategory.ID {
					found = true
					break
				}
			}
			if !found {
				productSubCategoryList = append(productSubCategoryList, ProductsSubCategory{
					ID:          &mapping.ProductSubCategory.ID,
					Name:        mapping.ProductSubCategory.Name,
					Description: mapping.ProductSubCategory.Description,
				})
			}
		}
	}
	return &GetProductResponse{
		Message: "Product fetched successfully",
		Data: ProductInfoMapping{
			ID:                 product.ID,
			Name:               product.Name,
			Description:        product.Description,
			Cost:               product.Cost,
			BasePrice:          product.BasePrice,
			Price:              product.Price,
			IsGrabFood:         product.IsGrabFood,
			IsShopeeFood:       product.IsShopeeFood,
			GrabFoodInfo:       product.GrabFoodInfo,
			ShopeeFoodInfo:     product.ShopeeFoodInfo,
			ProductCategory:    productCategoryList,
			ProductSubCategory: productSubCategoryList,
		},
	}, nil
}

// API to update product
// For this API, accuracy of the product category id and product sub category id are crucial.
// For this API, accuracy of the product category name and product sub category name are crucial.
// Inaccurate name and id may lead to wrong product category mapping and duplication of category and sub category.
// Flow for the API:
// 1. Check if the product category id and product sub category id are exist in the database.
// 2. If category/sub category id is nil, create new product category and product sub category.
// 3. store the category/sub category to a list
// 4. Update the product category mapping and store in database
// FUTURE ENHANCEMENT: User can add new category and sub category while updating the product
//
//encore:api auth method=PUT path=/api/products/update-product/:product_id
func (s *Service) UpdateProduct(ctx context.Context, product_id uuid.UUID, req *UpdateProductRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	if err := validatePlatformInfo(&AddProductRequest{
		IsGrabFood:     req.IsGrabFood,
		IsShopeeFood:   req.IsShopeeFood,
		GrabFoodInfo:   req.GrabFoodInfo,
		ShopeeFoodInfo: req.ShopeeFoodInfo,
	}); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	err := middleware.CheckPermission(constants.UpdateProductAction, &req.BusinessID, nil)
	if err != nil {
		return &common.BasicResponse{
			Message: "You are not allowed to update product",
		}, err
	}

	// check business id existence
	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ? ", req.BusinessID).First(&business).Error
	if err != nil {
		return &common.BasicResponse{
			Message: "Business not found",
		}, err
	}

	// check product id existence
	var product models.Product
	err = s.db.Model(&models.Product{}).Where("id = ? AND business_id = ?", product_id, req.BusinessID).First(&product).Error
	if err != nil {
		return &common.BasicResponse{
			Message: "Product not found",
		}, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	//Update product name, description, cost, base price, price
	updateProductData := make(map[string]interface{})
	common.HandleFieldUpdate(&product.Name, &req.Name, "name", updateProductData, true)
	common.HandleFieldUpdate(&product.Description, &req.Description, "description", updateProductData, true)
	common.HandleFieldUpdate(&product.Cost, &req.Cost, "cost", updateProductData, true)
	common.HandleFieldUpdate(&product.BasePrice, &req.BasePrice, "base_price", updateProductData, true)
	common.HandleFieldUpdate(&product.Price, &req.Price, "price", updateProductData, true)
	common.HandleFieldUpdate(&product.IsActive, &req.IsActive, "is_active", updateProductData, true)
	common.HandleFieldUpdate(&product.IsGrabFood, &req.IsGrabFood, "is_grab_food", updateProductData, true)
	common.HandleFieldUpdate(&product.IsStoreOutlet, &req.IsStoreOutlet, "is_store_outlet", updateProductData, true)
	common.HandleFieldUpdate(&product.IsShopeeFood, &req.IsShopeeFood, "is_shopee_food", updateProductData, true)
	common.HandleFieldUpdate(&product.SortOrder, &req.SortOrder, "sort_order", updateProductData, true)

	// query all modifier groups based on product id in productmodifier_mapping table
	var productModifierMappings []models.ProductModifierMapping
	err = tx.Model(&models.ProductModifierMapping{}).Where("product_id = ?", product_id).Find(&productModifierMappings).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product modifier mapping",
		}, err
	}

	// Initialize map with all existing mappings marked as invalid (false)
	isValidModifierMapping := make(map[uuid.UUID]bool)
	for _, mapping := range productModifierMappings {
		isValidModifierMapping[mapping.ModifierGroupID] = false
	}

	// Mark mappings as valid (true) if they exist in request
	for _, modifierGroup := range req.ModifierGroups {
		if _, exists := isValidModifierMapping[modifierGroup.ID]; exists {
			isValidModifierMapping[modifierGroup.ID] = true
		}
	}

	// Delete invalid mappings (those not in request)
	for modifierGroupID, isValid := range isValidModifierMapping {
		if !isValid {
			err = tx.Where("product_id = ? AND modifier_group_id = ?",
				product_id, modifierGroupID).
				Unscoped().Delete(&models.ProductModifierMapping{}).Error
			if err != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to delete product modifier mapping",
				}, err
			}
		}
	}

	// Create new mappings for groups in request that weren't previously mapped
	existingGroupIDs := make(map[uuid.UUID]bool)
	for _, mapping := range productModifierMappings {
		existingGroupIDs[mapping.ModifierGroupID] = true
	}

	for _, modifierGroup := range req.ModifierGroups {
		if !existingGroupIDs[modifierGroup.ID] {
			err = tx.Create(&models.ProductModifierMapping{
				ProductID:       product_id,
				ModifierGroupID: modifierGroup.ID,
				MaxSelection:    modifierGroup.MaxSelection,
				CreatedAt:       time.Now(),
				UpdatedAt:       nil,
				DeletedAt:       gorm.DeletedAt{},
			}).Error
			if err != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to create product modifier mapping",
				}, err
			}
		} else {
			err = tx.Model(&models.ProductModifierMapping{}).Where("product_id = ? AND modifier_group_id = ?", product_id, modifierGroup.ID).Update("max_selection", modifierGroup.MaxSelection).Error
			if err != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to update product modifier mapping",
				}, err
			}
		}
	}

	// query all ingredient based on product id in productingredient_mapping table
	var productIngredientMappings []models.ProductIngredientMapping
	err = tx.Model(&models.ProductIngredientMapping{}).Where("product_id = ?", product_id).Find(&productIngredientMappings).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product ingredient mapping",
		}, err
	}

	// Create a map of all requested ingredient IDs
	requestedIngredientIDs := make(map[uuid.UUID]bool)
	for _, ingredient := range req.Ingredients {
		requestedIngredientIDs[ingredient.IngredientID] = true
	}

	// Delete mappings that exist in DB but aren't in request
	if len(req.Ingredients) > 0 {
		// Get all current mapping IDs to delete
		var mappingsToDelete []uuid.UUID
		for _, mapping := range productIngredientMappings {
			if !requestedIngredientIDs[mapping.IngredientID] {
				mappingsToDelete = append(mappingsToDelete, mapping.ID)
			}
		}

		// Delete in bulk if there are mappings to delete
		if len(mappingsToDelete) > 0 {
			err = tx.Where("id IN (?)", mappingsToDelete).Unscoped().Delete(&models.ProductIngredientMapping{}).Error
			if err != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to delete product ingredient mapping",
				}, err
			}
		}
	} else {
		// If no ingredients in request, delete all mappings
		err = tx.Where("product_id = ?", product_id).Unscoped().Delete(&models.ProductIngredientMapping{}).Error
		if err != nil {
			tx.Rollback()
			return &common.BasicResponse{
				Message: "Failed to delete product ingredient mapping",
			}, err
		}
	}

	// if there is no ingredient in the request, delete all existing ingredient mapping
	if len(req.Ingredients) == 0 {
		err = tx.Where("product_id = ?", product_id).Unscoped().Delete(&models.ProductIngredientMapping{}).Error
		if err != nil {
			tx.Rollback()
		}
	}

	// if there is ingredient in the request, update the ingredient
	if len(req.Ingredients) > 0 {
		for _, ingredient := range req.Ingredients {
			// this faster the process if product ingredients id is included in the request
			var productIngredientMapping models.ProductIngredientMapping
			err = tx.Where("id = ?", ingredient.ID).First(&productIngredientMapping).Error
			if err == nil {
				// mean ingredient is already mapped to the product
				fmt.Println("ingredient is already mapped to the product")
				// update the ingredient
				updateData := make(map[string]interface{})
				common.HandleFieldUpdate(&ingredient.Unit, &ingredient.Unit, "unit", updateData, false)
				common.HandleFieldUpdate(&ingredient.Quantity, &ingredient.Quantity, "quantity", updateData, false)
				if len(updateData) > 0 {
					result := tx.Model(&models.ProductIngredientMapping{}).Where("id = ? AND product_id = ? AND ingredient_id = ?", ingredient.ID, product_id, ingredient.IngredientID).Save(&updateData)
					if result.Error != nil {
						tx.Rollback()
						return &common.BasicResponse{
							Message: "Failed to update product ingredient mapping",
						}, result.Error
					}
				}
				// it will not enter the following code and continue to the next iteration
				continue
			}

			// Check if this ingredient is already mapped to the product
			var existingMapping models.ProductIngredientMapping
			err = tx.Where("product_id = ? AND ingredient_id = ?", product_id, ingredient.ID).First(&existingMapping).Error

			if err != nil {
				// Create new mapping if it doesn't exist
				fmt.Println("enter create new ingredient mapping")
				productIngredientMapping := models.ProductIngredientMapping{
					ProductID:    product_id,
					IngredientID: ingredient.ID,
					Unit:         constants.UnitMeasurement(ingredient.Unit),
					Quantity:     ingredient.Quantity,
					CreatedAt:    time.Now(),
					UpdatedAt:    nil,
					DeletedAt:    gorm.DeletedAt{},
				}
				err = tx.Create(&productIngredientMapping).Error
				if err != nil {
					tx.Rollback()
					return &common.BasicResponse{
						Message: "Failed to create product ingredient mapping",
					}, err
				}
			} else {
				// Update existing mapping
				fmt.Println("enter update ingredient")
				updateData := make(map[string]interface{})
				common.HandleFieldUpdate(&ingredient.Unit, &ingredient.Unit, "unit", updateData, false)
				common.HandleFieldUpdate(&ingredient.Quantity, &ingredient.Quantity, "quantity", updateData, false)

				if len(updateData) > 0 {
					result := tx.Model(&models.ProductIngredientMapping{}).
						Where("product_id = ? AND ingredient_id = ?", product_id, ingredient.ID).
						Updates(updateData)
					if result.Error != nil {
						tx.Rollback()
						return &common.BasicResponse{
							Message: "Failed to update product ingredient mapping",
						}, result.Error
					}
				}
			}
		}
	}

	if len(updateProductData) > 0 {
		result := tx.Model(&product).Updates(updateProductData)
		if result.Error != nil {
			tx.Rollback()
			return &common.BasicResponse{
				Message: "Failed to update product",
			}, result.Error
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return &common.BasicResponse{
				Message: "Product not found",
			}, result.Error
		}
	}

	product.ModifierOptionsID = req.ModifierOptionsID
	product.GrabFoodInfo = req.GrabFoodInfo
	product.ShopeeFoodInfo = req.ShopeeFoodInfo
	tx.Save(&product)

	// Update product category mapping
	// get existing product category mapping based on product id
	var productCategoryMapping []models.ProductCategoryMapping
	err = tx.Preload("ProductCategory").Preload("ProductSubCategory").Preload("Product").Where("product_id = ?", product_id).Find(&productCategoryMapping).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product category mapping",
		}, err
	}

	// get all product category based on business id
	var productCategoryList []models.ProductCategory
	err = tx.Where("business_id = ?", req.BusinessID).Order("created_at ASC").Find(&productCategoryList).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product category",
		}, err
	}
	productCategoryListOldSize := len(productCategoryList)

	// add product category to database if the id in req is null
	fmt.Println("Start check if product category is exist in database")
	for _, category := range req.Categories {
		// nil means product category is not exist in database
		if category.ID == nil {
			// add product category
			productCategory := models.ProductCategory{
				Name:        category.Name,
				Description: category.Description,
				BusinessID:  &req.BusinessID,
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   gorm.DeletedAt{},
			}
			result := tx.Create(&productCategory)
			if result.Error != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to add product category",
				}, result.Error
			}
			// add new created product category to product category list
			productCategoryList = append(productCategoryList, productCategory)
		}
	}

	// Update product sub category mapping
	var productSubCategoryList []models.ProductSubCategory
	err = tx.Where("business_id = ?", req.BusinessID).Order("created_at ASC").Find(&productSubCategoryList).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product sub category",
		}, err
	}
	productSubCategoryListOldSize := len(productSubCategoryList)
	// check if product sub category is exist in database
	for _, subCategory := range req.SubCategories {
		if subCategory.ID == nil {
			// add product sub category
			productSubCategory := models.ProductSubCategory{
				Name:        subCategory.Name,
				Description: subCategory.Description,
				BusinessID:  &req.BusinessID,
				CreatedAt:   time.Now(),
				UpdatedAt:   nil,
				DeletedAt:   gorm.DeletedAt{},
			}
			result := tx.Create(&productSubCategory)
			if result.Error != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to add product sub category",
				}, result.Error
			}
			// add new created product sub category to product sub category list
			productSubCategoryList = append(productSubCategoryList, productSubCategory)
		}
	}

	// get product sub category number
	var productSubCategoryNumber int64
	err = tx.Model(&models.ProductCategoryMapping{}).Distinct("product_sub_category_id").Where("product_id = ?", product_id).Count(&productSubCategoryNumber).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product sub category number",
		}, err
	}

	// get product category number
	var productCategoryNumber int64
	err = tx.Model(&models.ProductCategoryMapping{}).Distinct("product_category_id").Where("product_id = ?", product_id).Count(&productCategoryNumber).Error
	if err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to get product category number",
		}, err
	}

	// use in categories == nil no need refresh the index
	var curNewCategoryIndex int = productCategoryListOldSize
	for _, category := range req.Categories {

		// if is newly created category
		if category.ID == nil {
			// add existing and newly created sub category to the newly created category
			newCategoryID := productCategoryList[curNewCategoryIndex].ID
			// increase the index of the newly created category so that it can be used to get the next newly created category
			curNewCategoryIndex++

			for _, subCategory := range req.SubCategories {

				// sub category is not nil means sub category is exist in database
				if subCategory.ID != nil {
					// add new sub category to the newly created category

					/* fmt.Println("product_id -------------> ", product_id)
					fmt.Println("categoryID -------------> ", newCategoryID)
					fmt.Println("subCategory.ID -------------> ", subCategory.ID) */
					productCategoryMapping := models.ProductCategoryMapping{
						ProductID:            product_id,
						ProductCategoryID:    newCategoryID,
						ProductSubCategoryID: subCategory.ID,
					}
					tx.Create(&productCategoryMapping)
				} else {
					// sub category is nil means sub category is newly created
					for _, existingSubCategory := range productSubCategoryList {
						if existingSubCategory.Name == subCategory.Name {
							productCategoryMapping := models.ProductCategoryMapping{
								ProductID:            product_id,
								ProductCategoryID:    newCategoryID,
								ProductSubCategoryID: &existingSubCategory.ID,
							}
							tx.Create(&productCategoryMapping)
						}
					}

				}

				// sub category is nil means sub category is newly created

			}
		} else {
			// not newly created category
			// use in categories != nil
			// put here due to every category has all the same sub category, therefore starting index of new sub category is tally with the req.SubCategories
			var curNewSubCategoryIndex int = productSubCategoryListOldSize
			for _, subCategory := range req.SubCategories {
				// sub category is not nil means sub category is exist in database
				if subCategory.ID != nil {
					// check if the category exist in the mapping (if not exist, link the category and sub category together)
					var productCategoryMapping models.ProductCategoryMapping
					err = tx.Where("product_id = ? AND product_category_id = ? AND product_sub_category_id = ?", product_id, category.ID, subCategory.ID).First(&productCategoryMapping).Error
					if err == nil {
					} else {
						// category not exist in the mapping
						productCategoryMapping := models.ProductCategoryMapping{
							ProductID:            product_id,
							ProductCategoryID:    *category.ID,
							ProductSubCategoryID: subCategory.ID,
						}
						tx.Create(&productCategoryMapping)
					}

				} else {
					// sub category is nil means sub category is newly created
					productCategoryMapping := models.ProductCategoryMapping{
						ProductID:            product_id,
						ProductCategoryID:    *category.ID,
						ProductSubCategoryID: &productSubCategoryList[curNewSubCategoryIndex].ID,
					}
					tx.Create(&productCategoryMapping)
					curNewSubCategoryIndex++
				}
			}
		}

	}

	// check if the product category mapping tallyness is correct
	if len(req.Categories) <= int(productCategoryNumber) {
		// delete the additional category
		requestedCategoryIDs := make(map[uuid.UUID]bool)
		// map the requested category id to the map
		for _, reqCategory := range req.Categories {
			if reqCategory.ID != nil {
				requestedCategoryIDs[*reqCategory.ID] = true
			}
		}
		var currentCategoryMapping []models.ProductCategoryMapping
		err = tx.Where("product_id = ?", product_id).Find(&currentCategoryMapping).Error
		if err != nil {
			tx.Rollback()
			return &common.BasicResponse{
				Message: "Failed to get product category mapping",
			}, err
		}

		var mappingsToDelete []uuid.UUID
		for _, mapping := range currentCategoryMapping {
			exists := requestedCategoryIDs[mapping.ProductCategoryID]
			if !exists {
				mappingsToDelete = append(mappingsToDelete, mapping.ID)
			}
		}
		if len(mappingsToDelete) > 0 {
			result := tx.Where("product_id = ? AND id IN (?)", product_id, mappingsToDelete).Unscoped().Delete(&models.ProductCategoryMapping{})
			if result.Error != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to delete product category mapping",
				}, result.Error
			}
		}

	}

	// check if the product sub category mapping tallyness is correct
	if len(req.SubCategories) <= int(productSubCategoryNumber) {
		// delete the additional sub category
		//( sub category is unqiue in the sense of in the mapping, the sub category will belongs to category. )
		// Delete the additional sub category will remove the combination of category and sub category.
		requestedSubCategoryIDs := make(map[uuid.UUID]bool)
		for _, reqSubCategory := range req.SubCategories {
			if reqSubCategory.ID != nil {
				requestedSubCategoryIDs[*reqSubCategory.ID] = true
			}
		}
		var currentSubCategoryMapping []models.ProductCategoryMapping
		err = tx.Where("product_id = ?", product_id).Find(&currentSubCategoryMapping).Error
		if err != nil {
			tx.Rollback()
			return &common.BasicResponse{
				Message: "Failed to get product sub category mapping",
			}, err
		}

		var mappingsToDelete []uuid.UUID
		for _, mapping := range currentSubCategoryMapping {
			exists := requestedSubCategoryIDs[*mapping.ProductSubCategoryID]
			if !exists {
				mappingsToDelete = append(mappingsToDelete, mapping.ID)
			}
		}
		if len(mappingsToDelete) > 0 {
			result := tx.Where("product_id = ? AND id IN (?)", product_id, mappingsToDelete).Unscoped().Delete(&models.ProductCategoryMapping{})
			if result.Error != nil {
				tx.Rollback()
				return &common.BasicResponse{
					Message: "Failed to delete product sub category mapping",
				}, result.Error
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return &common.BasicResponse{
			Message: "Failed to commit transaction",
		}, err
	}

	return &common.BasicResponse{
		Message: "Product updated successfully",
	}, nil
}

// Need change the permision action in future
// API to delete product
//
//encore:api auth method=DELETE path=/api/products/delete-product/:business_id/:product_id
func (s *Service) DeleteProduct(ctx context.Context, business_id uuid.UUID, product_id uuid.UUID) (*common.BasicResponse, error) {
	err := middleware.CheckPermission(constants.DeleteProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	var business models.Business
	err = s.db.Model(&models.Business{}).Where("id = ?", business_id).First(&business).Error
	if err != nil {
		return nil, err
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// delete product category mapping
	result := tx.Model(&models.ProductCategoryMapping{}).Where("product_id = ?", product_id).Unscoped().Delete(&models.ProductCategoryMapping{})
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, result.Error
	}

	// delete product modifier mapping
	result = tx.Model(&models.ProductModifierMapping{}).Where("product_id = ?", product_id).Unscoped().Delete(&models.ProductModifierMapping{})
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// delete product ingredient mapping
	result = tx.Model(&models.ProductIngredientMapping{}).Where("product_id = ?", product_id).Unscoped().Delete(&models.ProductIngredientMapping{})
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// delete product
	result = tx.Model(&models.Product{}).Where("id= ? AND business_id = ?", product_id, business_id).Unscoped().Delete(&models.Product{})
	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, result.Error
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Product deleted successfully",
	}, nil
}

// upload image api
//
//encore:api public raw method=POST path=/api/products/upload-image
func UploadImage(w http.ResponseWriter, r *http.Request) {
	// parse the multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// get the file from the form
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()

	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		http.Error(w, "Failed to create uploads directory", http.StatusInternalServerError)
		return
	}

	dstPath := fmt.Sprintf("%s/%s", uploadsDir, handler.Filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer dst.Close()

	// copy the uploaded file to the new file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	//_, err = aws_s3.UploadImageToS3Beta(dstPath)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	w.WriteHeader(http.StatusInternalServerError)
	//	w.Write([]byte(err.Error()))
	//	return
	//}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("File uploaded successfully"))
}

// API to create ingredient
//
//encore:api auth method=POST path=/api/ingredients/create
func (s *Service) CreateIngredient(ctx context.Context, req *CreateIngredientRequest) (*models.Ingredient, error) {
	// validate request struct
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.CreateIngredientAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check if business exists
	var business models.Business
	result := s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business)
	if result.Error != nil {
		return nil, result.Error
	}

	// create ingredient
	ingredient := models.Ingredient{
		BusinessID:   req.BusinessID,
		Name:         req.Name,
		Description:  req.Description,
		Unit:         req.Unit,
		Quantity:     req.Quantity,
		PricePerUnit: req.PricePerUnit,
		SortOrder:    req.SortOrder,
	}

	// create ingredient
	result = s.db.Create(&ingredient)
	if result.Error != nil {
		return nil, result.Error
	}

	// success response
	return &ingredient, nil
}

// API to update ingredient
//
//encore:api auth method=PUT path=/api/ingredients/update/:ingredient_id
func (s *Service) UpdateIngredient(ctx context.Context, ingredient_id uuid.UUID, req *UpdateIngredientRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return nil, err
	}
	// check permission from middleware
	err := middleware.CheckPermission(constants.UpdateIngredientAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	// check if business exists
	var business models.Business
	result := s.db.Model(&models.Business{}).Where("id = ?", req.BusinessID).First(&business)
	if result.Error != nil {
		return nil, result.Error
	}

	//query old data from db
	var oldIngredient models.Ingredient
	result = s.db.Model(&models.Ingredient{}).Where("id=? AND business_id=?", ingredient_id, req.BusinessID).First(&oldIngredient)
	if result.Error != nil {
		return nil, result.Error
	}

	// ensure there is data to update
	result = s.db.Model(&models.Ingredient{}).Where("id=? AND business_id=?", ingredient_id, req.BusinessID).Updates(req)
	if result.Error != nil {
		return nil, result.Error
	}

	// Update all stock where ingredient_id matches the updated ingredient
	s.db.Model(&models.Stock{}).
		Where("ingredient_id = ?", ingredient_id).
		Updates(map[string]interface{}{
			"small_scale_unit": constants.GetSmallUnitFromLarge(req.Unit),
			"large_scale_unit": constants.GetLargeUnitFromSmall(req.Unit),
		})

	// success response
	return &common.BasicResponse{
		Message: "Ingredient updated successfully",
	}, nil
}

// API to delete ingredient
//
//encore:api auth method=DELETE path=/api/ingredients/delete/:ingredient_id
func (s *Service) DeleteIngredient(ctx context.Context, ingredient_id uuid.UUID) (*common.BasicResponse, error) {

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if trx.Error != nil {
		return nil, trx.Error
	}

	var ingredient models.Ingredient
	result := trx.Model(&models.Ingredient{}).Where("id=?", ingredient_id).First(&ingredient)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.DeleteIngredientAction, &ingredient.BusinessID, nil)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// check if ingredient is used in any product
	var productIngredientMapping []models.ProductIngredientMapping
	result = trx.Model(&models.ProductIngredientMapping{}).Where("ingredient_id=?", ingredient_id).Preload("Product").Find(&productIngredientMapping)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	if len(productIngredientMapping) > 0 {
		// ingredient is in use, cannot delete
		trx.Rollback()
		productNames := make([]string, len(productIngredientMapping))
		for i, mapping := range productIngredientMapping {
			productNames[i] = mapping.Product.Name
		}
		productNamesString := strings.Join(productNames, ", ")
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: fmt.Sprintf("Ingredient is used in the following products: %s.\nPlease unlink it from these products before deleting", productNamesString),
		}
	}

	// check if ingredient is used in any modifier
	var modifierIngredientMapping []models.ModifierIngredientMapping
	result = trx.Model(&models.ModifierIngredientMapping{}).
		Where("ingredient_id=?", ingredient_id).
		Find(&modifierIngredientMapping)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}
	if len(modifierIngredientMapping) > 0 {
		// Fetch modifier option names while trx is still valid (before Rollback)
		optionIDs := make([]uuid.UUID, 0, len(modifierIngredientMapping))
		seenIDs := make(map[uuid.UUID]struct{})
		for _, mapping := range modifierIngredientMapping {
			if _, ok := seenIDs[mapping.ModifierOptionsID]; !ok {
				seenIDs[mapping.ModifierOptionsID] = struct{}{}
				optionIDs = append(optionIDs, mapping.ModifierOptionsID)
			}
		}
		var options []models.ModifierOptions
		result = trx.Model(&models.ModifierOptions{}).Where("id IN ?", optionIDs).Find(&options)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
		optionNameByID := make(map[uuid.UUID]string)
		for _, opt := range options {
			optionNameByID[opt.ID] = opt.Name
		}
		modifierOptionNames := make([]string, 0, len(optionIDs))
		for _, id := range optionIDs {
			if name := optionNameByID[id]; name != "" {
				modifierOptionNames = append(modifierOptionNames, name)
			}
		}
		modifierOptionNamesString := strings.Join(modifierOptionNames, ", ")
		trx.Rollback()
		return nil, &errs.Error{
			Code:    errs.FailedPrecondition,
			Message: fmt.Sprintf("Ingredient is used in the following modifier options: %s.\nPlease unlink it from these modifier options before deleting", modifierOptionNamesString),
		}
	}

	result = trx.Model(&models.Ingredient{}).Where("id=?", ingredient_id).Delete(&models.Ingredient{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// delete all stock where ingredient_id matches the deleted ingredient
	result = trx.Model(&models.Stock{}).Where("ingredient_id=?", ingredient_id).Delete(&models.Stock{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	result = trx.Commit()
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// check if business exists
	return &common.BasicResponse{
		Message: "Ingredient deleted successfully",
	}, nil
}

// API to get all ingredients
//
//encore:api auth method=GET path=/api/ingredients/get-all/:business_id
func (s *Service) GetAllIngredients(ctx context.Context, business_id uuid.UUID) (*GetAllIngredientsResponse, error) {
	// check permission from middleware
	err := middleware.CheckPermission(constants.ReadIngredientAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	if business_id == uuid.Nil {
		return nil, nil
	}

	var ingredients []models.Ingredient
	result := s.db.Model(&models.Ingredient{}).Where("business_id=?", business_id).Order("sort_order ASC").Find(&ingredients)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetAllIngredientsResponse{
		Message: "Ingredients fetched successfully",
		Data:    ingredients,
	}, nil
}

//encore:api auth raw method=POST path=/api/ingredients/upload/image
func (s *Service) UploadIngredientImage(w http.ResponseWriter, req *http.Request) {
	ingredient_id := req.FormValue("id")
	if ingredient_id == "" {
		http.Error(w, "Ingredient ID is required", http.StatusBadRequest)
		return
	}

	temp_uuid, err := googleUUID.Parse(ingredient_id)
	if err != nil {
		http.Error(w, "Invalid Ingredient ID format", http.StatusBadRequest)
		return
	}

	ingredient := models.Ingredient{}
	err = s.db.Model(&models.Ingredient{}).Where("id = ?", temp_uuid).First(&ingredient).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	business := models.Business{}
	err = s.db.Model(&models.Business{}).Where("id = ?", ingredient.BusinessID).First(&business).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	// Validate file size
	if err := req.ParseMultipartForm(2 << 20); err != nil { // Limit to 2 MB
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
	filename := ingredient.Name + "_" + ingredient.ID.String()[:6] + file_extension
	filename = strings.ReplaceAll(filename, " ", "_")
	document.DocPath = "business/" + business.RegistrationNumber + "/images/ingredients/" + filename
	document.DocPath = strings.ReplaceAll(document.DocPath, " ", "_")

	// Upload the file to S3
	document_res, err := aws_s3.UploadDocument(document)
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	ingredient.ImageURL = document_res.Url

	err = s.db.Save(&ingredient).Error
	if err != nil {
		errs.HTTPError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// API to create recipe
//
//encore:api auth method=POST path=/api/recipes/create-recipe
func (s *Service) CreateRecipe(ctx context.Context, req *CreateRecipeRequest) (*common.BasicResponse, error) {

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.CreateIngredientAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	if trx.Error != nil {
		return nil, trx.Error
	}

	//create recipe
	recipe := models.Recipe{
		ProductID:   req.ProductID,
		Name:        req.Name,
		Description: req.Description,
	}
	result := trx.Create(&recipe)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, step := range req.Steps {
		recipeStep := models.RecipeStep{
			RecipeID:    recipe.ID,
			Name:        step.Name,
			Instruction: step.Instruction,
			Precedence:  step.Precedence,
		}
		result := trx.Create(&recipeStep)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	result = trx.Commit()
	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Recipe created successfully",
	}, nil
}

// API to get a recipe by id
//
//encore:api auth method=GET path=/api/recipes/get-recipe/:business_id/:product_id/:recipe_id
func (s *Service) GetRecipe(ctx context.Context, business_id uuid.UUID, product_id uuid.UUID, recipe_id uuid.UUID) (*GetRecipeResponse, error) {

	// check permission from middleware
	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	var recipe models.Recipe
	result := s.db.Model(&models.Recipe{}).Where("id=? AND product_id=?", recipe_id, product_id).First(&recipe)
	if result.Error != nil {
		return nil, result.Error
	}

	var steps []CustomRecipeStep
	result = s.db.Model(&models.RecipeStep{}).Where("recipe_id = ?", recipe_id).Order("precedence ASC").Find(&steps)
	if result.Error != nil {
		return nil, result.Error
	}

	return &GetRecipeResponse{
		Message:     "Recipe fetched successfully",
		Name:        recipe.Name,
		Description: recipe.Description,
		Steps:       steps,
	}, nil

	// get recipe
}

// API to get all recipes by business id
//
//encore:api auth method=GET path=/api/recipes/get-all-recipes/:business_id/:page/:page_size
func (s *Service) GetAllRecipe(ctx context.Context, business_id uuid.UUID, page int, page_size int) (*GetAllRecipesResponse, error) {

	if page <= 1 {
		page = 1
	}
	if page_size <= 1 {
		page_size = 10
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.ReadIngredientAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	// get total product

	var totalProduct int64
	count_result := s.db.Model(&models.Product{}).Where("business_id = ?", business_id).Count(&totalProduct).Error
	if count_result != nil {
		return nil, count_result
	}
	offset := (page - 1) * page_size
	totalPages := int64(totalProduct/int64(page_size)) + 1

	var products []ProductWithRecipe
	result := s.db.Model(&models.Product{}).
		Select("*, name as product_name").
		Where("business_id = ?", business_id).
		Limit(page_size).
		Offset(offset).
		Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}

	for i := range products { // Use index to modify the original slice
		var recipe models.Recipe
		result = s.db.Model(&models.Recipe{}).Where("product_id = ?", products[i].ID).First(&recipe)

		if result.Error == nil {
			var steps []models.RecipeStep
			result = s.db.Model(&models.RecipeStep{}).Where("recipe_id = ?", recipe.ID).Order("precedence ASC").Find(&steps)
			if result.Error == nil {
				products[i].RecipeName = recipe.Name
				products[i].RecipeDescription = recipe.Description
				for _, step := range steps {
					products[i].Steps = append(products[i].Steps, CustomRecipeStep{
						ID:          step.ID,
						Name:        step.Name,
						Instruction: step.Instruction,
						Precedence:  step.Precedence,
					})
				}
			}
		}
	}

	return &GetAllRecipesResponse{
		Message: "Recipes fetched successfully",
		Data:    products,
		Meta: common.Pagination{
			Page:       page,
			PageSize:   page_size,
			TotalPages: int(totalPages),
			Total:      int64(totalProduct),
		},
	}, nil

}

// API to create modifier group
//
//encore:api auth method=POST path=/api/modifiers/create-modifier-group
func (s *Service) CreateModifierGroup(ctx context.Context, req *CreateModifierGroupRequest) (*common.BasicResponse, error) {

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.CreateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if trx.Error != nil {
		return nil, trx.Error
	}

	modifierGroup := models.ModifierGroups{
		BusinessID: req.BusinessID,
		Name:       req.Name,
		InputType:  req.InputType,
	}

	result := trx.Create(&modifierGroup)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, option := range req.ModifierOptions {
		modifierOption := models.ModifierOptions{
			ModifierGroupID: modifierGroup.ID,
			Name:            option.Name,
			PriceAdjustment: option.PriceAdjustment,
			SortOrder:       option.SortOrder,
		}
		result := trx.Create(&modifierOption)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	result = trx.Commit()
	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Modifier group created successfully",
	}, nil
}

// API to get all modifier groups
//
//encore:api auth method=GET path=/api/modifiers/get-all-modifier-groups/:business_id
func (s *Service) GetAllModifierGroups(ctx context.Context, business_id uuid.UUID) (*GetAllModifierGroupsResponse, error) {

	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	var modifierGroups []models.ModifierGroups
	result := s.db.Model(&models.ModifierGroups{}).Where("business_id = ?", business_id).Find(&modifierGroups)
	if result.Error != nil {
		return nil, result.Error
	}

	// pre-allocate response slice
	modifierGroupsResponse := make([]ModifierGroup, 0, len(modifierGroups))

	groupIDs := make([]uuid.UUID, len(modifierGroups))
	for i, group := range modifierGroups {
		groupIDs[i] = group.ID
	}

	var allOptions []models.ModifierOptions
	if len(groupIDs) > 0 {
		result := s.db.Model(&models.ModifierOptions{}).Where("modifier_group_id IN (?)", groupIDs).Preload("IngredientMappings.Ingredient").Find(&allOptions)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	//optionsByGroup := make(map[uuid.UUID][]ModifierOption)
	optionsByGroup := make(map[uuid.UUID][]models.ModifierOptions)
	for _, option := range allOptions {
		optionsByGroup[option.ModifierGroupID] = append(optionsByGroup[option.ModifierGroupID], models.ModifierOptions{
			ID:                 option.ID,
			Name:               option.Name,
			PriceAdjustment:    option.PriceAdjustment,
			IsActive:           option.IsActive,
			IngredientMappings: option.IngredientMappings,
		})
	}

	for _, group := range modifierGroups {
		modifierGroupsResponse = append(modifierGroupsResponse, ModifierGroup{
			ID:              group.ID,
			Name:            group.Name,
			InputType:       group.InputType,
			BusinessID:      group.BusinessID,
			CreatedAt:       group.CreatedAt,
			UpdatedAt:       group.UpdatedAt,
			ModifierOptions: optionsByGroup[group.ID],
		})
	}

	return &GetAllModifierGroupsResponse{
		Message: "Modifier groups fetched successfully",
		Data:    modifierGroupsResponse,
	}, nil

}

// API to update modifier group
//
//encore:api auth method=PUT path=/api/modifiers/update-modifier-group/:modifier_group_id
func (s *Service) UpdateModifierGroup(ctx context.Context, modifier_group_id uuid.UUID, req *CreateModifierGroupRequest) (*common.BasicResponse, error) {

	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	// check permission from middleware
	err := middleware.CheckPermission(constants.UpdateProductAction, &req.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	if trx.Error != nil {
		return nil, trx.Error
	}

	var modifierGroup models.ModifierGroups
	result := trx.Model(&models.ModifierGroups{}).Where("id = ? AND business_id = ?", modifier_group_id, req.BusinessID).First(&modifierGroup)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	modifierGroup.Name = req.Name
	if req.InputType == models.InputTypeRadio {
		modifierGroup.InputType = models.InputTypeRadio
	} else if req.InputType == models.InputTypeCheckbox {
		modifierGroup.InputType = models.InputTypeCheckbox
	} else {
		trx.Rollback()
		return nil, fmt.Errorf("invalid input type")
	}

	result = trx.Save(&modifierGroup)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, option := range req.ModifierOptions {
		if option.ID == nil || option.ID.String() == "" {
			modifierOption := models.ModifierOptions{
				ModifierGroupID: modifier_group_id,
				Name:            option.Name,
				PriceAdjustment: option.PriceAdjustment,
				SortOrder:       option.SortOrder,
			}
			result = trx.Create(&modifierOption)
			if result.Error != nil {
				trx.Rollback()
				return nil, result.Error
			}
			continue
		}
		var modifierOption models.ModifierOptions
		result = trx.Model(models.ModifierOptions{}).Where("id=? AND modifier_group_id = ?", option.ID, modifier_group_id).First(&modifierOption)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}

		modifierOption.Name = option.Name
		modifierOption.PriceAdjustment = option.PriceAdjustment
		modifierOption.SortOrder = option.SortOrder
		modifierOption.IsActive = option.IsActive

		result = trx.Save(&modifierOption)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	result = trx.Commit()
	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Modifier group updated successfully",
	}, nil

}

// API to delete modifier group
//
//encore:api auth method=DELETE path=/api/modifiers/delete-modifier-group/:modifier_group_id
func (s *Service) DeleteModifierGroup(ctx context.Context, modifier_group_id uuid.UUID) (*common.BasicResponse, error) {
	authData := auth.Data()
	userData, ok := authData.(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}
	err := middleware.CheckPermission(constants.DeleteProductAction, userData.BusinessID, nil)
	if err != nil {
		return nil, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	// WARNING: Uncomment this to allow deleting all data related to modifier (modifiers in past orders will also be deleted)

	// // Delete all SelectedModifierGroup records for this modifier group
	// result := trx.Where("modifier_group_id = ?", modifier_group_id).Unscoped().Delete(&models.SelectedModifierGroup{})
	// if result.Error != nil {
	// 	trx.Rollback()
	// 	return nil, result.Error
	// }

	// // First, get all modifier options for this group
	// var modifierOptions []models.ModifierOptions
	// result = trx.Where("modifier_group_id = ?", modifier_group_id).Find(&modifierOptions)
	// if result.Error != nil {
	// 	trx.Rollback()
	// 	return nil, result.Error
	// }

	// // Delete all ModifierIngredientMapping records for these modifier options
	// for _, option := range modifierOptions {
	// 	result = trx.Where("modifier_options_id = ?", option.ID).Unscoped().Delete(&models.ModifierIngredientMapping{})
	// 	if result.Error != nil {
	// 		trx.Rollback()
	// 		return nil, result.Error
	// 	}
	// }

	// // Now delete the modifier options
	result := trx.Model(&models.ModifierOptions{}).Where("modifier_group_id = ?", modifier_group_id).Unscoped().Delete(&models.ModifierOptions{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	result = trx.Model(&models.ModifierGroups{}).Where("id = ? AND business_id = ?", modifier_group_id, userData.BusinessID).Unscoped().Delete(&models.ModifierGroups{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	// Delete all ProductModifierMapping records for this modifier group
	result = trx.Model(&models.ProductModifierMapping{}).Where("modifier_group_id = ?", modifier_group_id).Unscoped().Delete(&models.ProductModifierMapping{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	result = trx.Commit()

	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Modifier group deleted successfully",
	}, nil

}

type AssignIngredientsToModifierRequest struct {
	ModifierIngredientMappings []models.ModifierIngredientMapping `json:"modifier_ingredient_mappings" validate:"required"`
}

// api to save modifier ingredients
//
//encore:api auth method=POST path=/api/modifiers/assign-ingredients
func (s *Service) AssignIngredientsToModifier(ctx context.Context, req *AssignIngredientsToModifierRequest) (*common.BasicResponse, error) {
	if _, err := govalidator.ValidateStruct(req); err != nil {
		return &common.BasicResponse{
			Message: err.Error(),
		}, err
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	// Get all modifier_options_ids from the request
	modifierOptionsID := req.ModifierIngredientMappings[0].ModifierOptionsID

	// Delete all existing ModifierIngredientMapping records for these modifier options
	result := trx.Where("modifier_options_id = ?", modifierOptionsID).Unscoped().Delete(&models.ModifierIngredientMapping{})
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, mapping := range req.ModifierIngredientMappings {
		modifierIngredientMapping := models.ModifierIngredientMapping{
			ModifierOptionsID: mapping.ModifierOptionsID,
			IngredientID:      mapping.IngredientID,
			Quantity:          mapping.Quantity,
			Unit:              mapping.Unit,
		}
		result := trx.Create(&modifierIngredientMapping)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	result = trx.Commit()
	if result.Error != nil {
		return nil, result.Error
	}

	return &common.BasicResponse{
		Message: "Ingredients assigned to modifier successfully",
	}, nil
}

// api/function to sync product to outlet
//
//encore:api auth method=POST path=/api/products/sync-product-to-outlet
func (s *Service) SyncProductToOutlet(ctx context.Context, req *SyncProductToOutletRequest) (*common.BasicResponse, error) {

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()

	if trx.Error != nil {
		return nil, trx.Error
	}

	var products []models.Product
	result := trx.Model(&models.Product{}).Where("business_id = ?", req.BusinessID).Find(&products)
	if result.Error != nil {
		trx.Rollback()
		return nil, result.Error
	}

	for _, product := range products {
		// check if product is already synced to outlet
		var outletProduct models.OutletProduct
		result = trx.Model(&models.OutletProduct{}).Where("outlet_id = ? AND product_id = ?", req.OutletID, product.ID).First(&outletProduct)
		if result.Error == nil {
			if !product.IsActive {
				// delete outlet product
				trx.Unscoped().Where("outlet_id = ? AND product_id = ?", req.OutletID, product.ID).Delete(&models.OutletProduct{})
			}
			continue
		}

		outletProduct = models.OutletProduct{
			OutletID:  req.OutletID,
			ProductID: product.ID,
			IsActive:  product.IsActive,
		}

		result = trx.Create(&outletProduct)
		if result.Error != nil {
			trx.Rollback()
			return nil, result.Error
		}
	}

	err := trx.Commit().Error
	if err != nil {
		return nil, err
	}

	return &common.BasicResponse{
		Message: "Product synced to outlet successfully",
	}, nil

}

// Internal version that accepts a transaction
func SyncProductToOutletWithTx(ctx context.Context, req *SyncProductToOutletRequest, trx *gorm.DB) (*common.BasicResponse, error) {

	// If a specific product ID is provided, sync only that product
	if req.ProductID != uuid.Nil {
		var product models.Product
		result := trx.Model(&models.Product{}).Where("business_id = ? AND id = ?", req.BusinessID, req.ProductID).First(&product)
		if result.Error != nil {
			return nil, result.Error
		}

		// Check if product is already synced to outlet
		var outletProduct models.OutletProduct
		result = trx.Model(&models.OutletProduct{}).Where("outlet_id = ? AND product_id = ?", req.OutletID, product.ID).First(&outletProduct)

		// Add or remove product based on IsAdd flag
		if req.IsAdd {
			// If product is not yet synced and we want to add it
			if result.Error != nil {
				outletProduct = models.OutletProduct{
					OutletID:  req.OutletID,
					ProductID: product.ID,
					IsActive:  product.IsActive,
				}
				result = trx.Create(&outletProduct)
				if result.Error != nil {
					return nil, result.Error
				}
			}
		} else {
			// If product exists and we want to remove it
			if result.Error == nil {
				result = trx.Where("outlet_id = ? AND product_id = ?", req.OutletID, product.ID).Delete(&models.OutletProduct{})
				if result.Error != nil {
					return nil, result.Error
				}
			}
		}

		var message string
		if req.IsAdd {
			message = "Product added successfully to outlet"
		} else {
			message = "Product removed successfully from outlet"
		}
		return &common.BasicResponse{
			Message: message,
		}, nil
	}

	// Otherwise, sync all products from the business to the outlet
	var products []models.Product
	result := trx.Model(&models.Product{}).
		Where("business_id = ?", req.BusinessID).
		Where("is_active = ?", true).
		Find(&products)
	if result.Error != nil {
		return nil, result.Error
	}

	for _, product := range products {
		// Check if product is already synced to outlet
		var outletProduct models.OutletProduct
		result = trx.Model(&models.OutletProduct{}).Where("outlet_id = ? AND product_id = ? ", req.OutletID, product.ID).First(&outletProduct)
		if result.Error == nil {
			// Product already exists in outlet
			if !product.IsActive {
				// delete outlet product
				trx.Unscoped().Where("outlet_id = ? AND product_id = ?", req.OutletID, product.ID).Delete(&models.OutletProduct{})
			}
			continue
		}

		// Add the product to outlet
		outletProduct = models.OutletProduct{
			OutletID:  req.OutletID,
			ProductID: product.ID,
			IsActive:  true,
		}

		result = trx.Create(&outletProduct)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	return &common.BasicResponse{
		Message: "All products synced to outlet successfully",
	}, nil
}

// api to get outlet product which use in pos due to need to show all products in pos with the active status of product in a outlet
//
//encore:api auth method=GET path=/api/products/get-outlet-product/:outlet_id
func (s *Service) GetOutletAllProductsWithoutPagination(ctx context.Context, outlet_id uuid.UUID) (*GetAllProductsFromOutletResponse, error) {

	err := middleware.CheckPermission(constants.ReadProductAction, nil, &outlet_id)
	if err != nil {
		return nil, err
	}

	userData, ok := auth.Data().(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}

	trx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			trx.Rollback()
		}
	}()
	// check business id existence
	var business models.Business
	err = trx.Model(&models.Business{}).Where("id=?", userData.BusinessID).First(&business).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// sync product to outlet if not exist
	// sync business products to outlet
	_, err = SyncProductToOutletWithTx(ctx, &SyncProductToOutletRequest{
		BusinessID: *userData.BusinessID,
		OutletID:   outlet_id,
		ProductID:  uuid.Nil,
		IsAdd:      false,
	}, trx)

	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// get all outlet products
	var outletProducts []models.OutletProduct
	err = trx.Preload("Product").Where("outlet_id = ?", outlet_id).Find(&outletProducts).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	if len(outletProducts) == 0 {
		return &GetAllProductsFromOutletResponse{
			Message: "No products found",
			Data:    nil,
		}, nil
	}

	var outletProductIDs []uuid.UUID
	for _, outletProduct := range outletProducts {
		outletProductIDs = append(outletProductIDs, outletProduct.ProductID)
	}

	// get all product category mapping
	var mappings []models.ProductCategoryMapping
	err = trx.Preload("ProductCategory").Preload("ProductSubCategory").Preload("Product").Where("product_id IN (?)", outletProductIDs).Find(&mappings).Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	productMap := make(map[uuid.UUID]*ProductInfoMappingFromOutlet)
	for _, m := range mappings {
		productID := m.ProductID
		if _, exists := productMap[productID]; !exists {

			// get product active status of a outlet
			var outletProduct models.OutletProduct
			err = trx.Select("is_active").Where("outlet_id = ? AND product_id = ?", outlet_id, productID).First(&outletProduct).Error
			if err != nil {
				trx.Rollback()
				return nil, err
			}

			productMap[productID] = &ProductInfoMappingFromOutlet{
				ID:                 productID,
				BusinessID:         *m.Product.BusinessID,
				Name:               m.Product.Name,
				Description:        m.Product.Description,
				Cost:               m.Product.Cost,
				BasePrice:          m.Product.BasePrice,
				Price:              m.Product.Price,
				ImageURL:           m.Product.ImageURL,
				IsActiveInBusiness: m.Product.IsActive,
				IsActiveInOutlet:   outletProduct.IsActive,
				IsGrabFood:         m.Product.IsGrabFood,
				IsShopeeFood:       m.Product.IsShopeeFood,
				CreatedAt:          m.Product.CreatedAt,
				ProductCategory:    []ProductsCategory{},
				ProductSubCategory: []ProductsSubCategory{},
				ModifierGroups:     []ModifierGroup{},
			}
		}

		// Add product category to product mapping
		if m.ProductCategory != (models.ProductCategory{}) {
			cat := ProductsCategory{
				ID:          &m.ProductCategory.ID,
				Name:        m.ProductCategory.Name,
				Description: m.ProductCategory.Description,
			}
			//check duplicate category
			found := false
			for _, existingCat := range productMap[productID].ProductCategory {
				if *existingCat.ID == *cat.ID {
					found = true
					break
				}
			}
			if !found {
				productMap[productID].ProductCategory = append(productMap[productID].ProductCategory, cat)
			}
		}

		// Add product sub category to product mapping
		// (Note: m.ProductSubCategory.ID will be zero if no sub-category exists.)
		if m.ProductSubCategory.ID != (uuid.UUID{}) {
			sub := ProductsSubCategory{
				ID:          &m.ProductSubCategory.ID,
				Name:        m.ProductSubCategory.Name,
				Description: m.ProductSubCategory.Description,
			}
			// Check for duplicates.
			found := false
			for _, existingSub := range productMap[productID].ProductSubCategory {
				if *existingSub.ID != uuid.Nil && *existingSub.ID == *sub.ID {
					found = true
					break
				}
			}
			if !found {
				productMap[productID].ProductSubCategory = append(productMap[productID].ProductSubCategory, sub)
			}
		}

	}
	// add modifier groups to productMap
	for _, productID := range outletProductIDs {
		var modifierMapping []models.ProductModifierMapping
		err = trx.Where("product_id = ?", productID).Find(&modifierMapping).Error
		if err != nil {
			trx.Rollback()
			return nil, err

		}

		// add modifier groups to products
		for _, mapping := range modifierMapping {
			var modifierGroup models.ModifierGroups
			err = trx.Where("id = ?", mapping.ModifierGroupID).First(&modifierGroup).Error
			if err != nil {
				trx.Rollback()
				return nil, err
			}

			var modifierOptions []models.ModifierOptions
			err = trx.Where("modifier_group_id = ?", mapping.ModifierGroupID).
				Where("is_active = ?", true).
				Find(&modifierOptions).Error
			if err != nil {
				trx.Rollback()
				return nil, err
			}
			productMap[productID].ModifierGroups = append(productMap[productID].ModifierGroups, ModifierGroup{
				ID:         mapping.ModifierGroupID,
				Name:       modifierGroup.Name,
				InputType:  modifierGroup.InputType,
				BusinessID: modifierGroup.BusinessID,
				// max selection is from product modifier mapping
				MaxSelection:    mapping.MaxSelection,
				CreatedAt:       modifierGroup.CreatedAt,
				UpdatedAt:       modifierGroup.UpdatedAt,
				ModifierOptions: modifierOptions,
			})
		}

	}

	// get all product that is inactive in business
	productsInactiveBusiness, err := common_operations.GetProductInactiveBusiness(trx, business.ID)
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	// delete product that is inactive in business from productMap
	for _, productID := range productsInactiveBusiness {
		delete(productMap, productID)
	}

	// convert productMap to slice for response (MOVED AFTER DELETION)
	responseData := make([]ProductInfoMappingFromOutlet, 0, len(productMap))
	for _, mapping := range productMap {
		responseData = append(responseData, *mapping)
	}

	// sort responseData by created at
	sort.Slice(responseData, func(i, j int) bool {
		return responseData[i].CreatedAt.Before(responseData[j].CreatedAt)
	})

	err = trx.Commit().Error
	if err != nil {
		trx.Rollback()
		return nil, err
	}

	return &GetAllProductsFromOutletResponse{
		Message: "Products fetched successfully",
		Data:    responseData,
	}, nil

}

func SyncModifierOptionToOutlet(ctx context.Context, req *SyncModifierOptionToOutletRequest, trx *gorm.DB) error {

	var modifierOptions []models.ModifierOptions
	err := trx.Model(&models.ModifierOptions{}).Joins("JOIN modifier_groups ON modifier_options.modifier_group_id = modifier_groups.id").
		Where("modifier_groups.business_id = ?", req.BusinessID).
		Where("modifier_options.is_active = ?", true).
		Find(&modifierOptions).Error
	if err != nil {
		return err
	}

	for _, modifierOption := range modifierOptions {
		var outletModifierOption models.OutletModifierOption
		err = trx.Model(&models.OutletModifierOption{}).Where("outlet_id = ? AND modifier_options_id = ?", req.OutletID, modifierOption.ID).First(&outletModifierOption).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			// Product doesn't exist in outlet, create it
			outletModifierOption = models.OutletModifierOption{
				OutletID:          req.OutletID,
				ModifierOptionsID: modifierOption.ID,
				IsActive:          true,
			}
			err = trx.Create(&outletModifierOption).Error
			if err != nil {
				return err
			}
		}

	}

	return nil
}

//encore:api auth method=GET path=/api/products/get-outlet-modifier-options/:outlet_id
func (s *Service) GetOutletModifierOptions(ctx context.Context, outlet_id uuid.UUID) (*GetOutletModifierOptionsResponse, error) {

	outlet := models.Outlet{}
	err := s.db.Model(&models.Outlet{}).Where("id = ?", outlet_id).First(&outlet).Error
	if err != nil {
		return nil, err
	}

	err = SyncModifierOptionToOutlet(ctx, &SyncModifierOptionToOutletRequest{
		BusinessID: outlet.BusinessID,
		OutletID:   outlet_id,
	}, s.db)

	if err != nil {
		return nil, err
	}

	var modifierOptions []models.OutletModifierOption
	err = s.db.Model(&models.OutletModifierOption{}).Where("outlet_id = ?", outlet_id).Preload("ModifierOptions").Find(&modifierOptions).Error
	if err != nil {
		return nil, err
	}

	modifierOptionsResponse := make([]common.OutletModifierOption, 0, len(modifierOptions))
	for _, modifierOption := range modifierOptions {
		modifierOptionsResponse = append(modifierOptionsResponse, common.OutletModifierOption{
			ID:       modifierOption.ID,
			Name:     modifierOption.ModifierOptions.Name,
			Price:    modifierOption.ModifierOptions.PriceAdjustment,
			IsActive: modifierOption.IsActive,
		})
	}

	return &GetOutletModifierOptionsResponse{
		Data: modifierOptionsResponse,
	}, nil
}

//encore:api auth method=GET path=/api/products/get-modifier-options/:business_id
func (s *Service) GetModifierOptions(ctx context.Context, business_id uuid.UUID) (*GetModifierOptionsResponse, error) {
	err := middleware.CheckPermission(constants.ReadProductAction, &business_id, nil)
	if err != nil {
		return nil, err
	}

	var modifierOptions []models.ModifierOptions
	err = s.db.Model(&models.ModifierOptions{}).
		Joins("JOIN modifier_groups ON modifier_options.modifier_group_id = modifier_groups.id").
		Where("modifier_groups.business_id = ?", business_id).
		Where("is_active = ?", true).
		Order("sort_order").
		Find(&modifierOptions).Error
	if err != nil {
		return nil, err
	}

	return &GetModifierOptionsResponse{
		Data: modifierOptions,
	}, nil
}
