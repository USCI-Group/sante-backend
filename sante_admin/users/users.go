package users

import (
	"context"
	"errors"
	"math"
	"time"

	"encore.app/auth_service"
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
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//encore:service
type Service struct {
	db *gorm.DB
}

type CreateUserParams struct {
	FirstName        string         `json:"first_name" valid:"required~FirstName is required"`
	Surname          string         `json:"surname" valid:"required~Surname is required"`
	IdentificationNo string         `json:"identification_no" gorm:"type:varchar(255)"`
	Email            string         `json:"email" valid:"required~Email is required,email~Invalid email format"`
	Password         string         `json:"password" valid:"required~Password is required"`
	Phone            string         `json:"phone" valid:"required~Phone is required"`
	EmployeeNo       string         `json:"employee_no"`
	Address          common.Address `json:"address"`
	RoleID           *uuid.UUID     `json:"role_id"`
	BusinessID       *uuid.UUID     `json:"business_id"`
	OutletID         *uuid.UUID     `json:"outlet_id"`
}

type UpdateUserParams struct {
	ID         uuid.UUID      `json:"id"`
	Email      string         `json:"email" valid:"required~Email is required,email~Invalid email format"`
	FirstName  string         `json:"first_name"`
	Password   *string        `json:"password"`
	Surname    string         `json:"surname"`
	Phone      string         `json:"phone"`
	Address    common.Address `json:"address" gorm:"embedded"`
	EmployeeNo string         `json:"employee_no"`
	BusinessID *uuid.UUID     `json:"business_id"`
	OutletID   *uuid.UUID     `json:"outlet_id"`
	RoleID     *uuid.UUID     `json:"role_id"`
}

type GetAllUsersParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Filter   struct {
		BusinessID *uuid.UUID `json:"business_id"`
		OutletID   *uuid.UUID `json:"outlet_id"`
		Name       string     `json:"name"`
	} `json:"filter"`
}

type UsersResponse struct {
	Meta common.Pagination `json:"meta"`
	Data []models.User     `json:"data"`
}

type CreateRoleParams struct {
	Name                    string                             `json:"name" valid:"required~Name is required"`
	Description             string                             `json:"description"`
	BusinessID              *uuid.UUID                         `json:"business_id"`
	RoleType                constants.RoleType                 `json:"role_type" valid:"required~Role type is required"`
	PermissionPresetFormats []constants.PermissionPresetFormat `json:"permission_presets"`
	OutletID                *uuid.UUID                         `json:"outlet_id"`
}

type GetPermissionPresetResponse struct {
	PermissionPresets []constants.PermissionPresetFormat `json:"permission_presets"`
}

type UpdateRoleParams struct {
	ID                      uuid.UUID                          `json:"id"`
	Name                    string                             `json:"name"`
	Description             string                             `json:"description"`
	BusinessID              *uuid.UUID                         `json:"business_id"`
	PermissionPresetFormats []constants.PermissionPresetFormat `json:"permission_presets"`
	RoleType                constants.RoleType                 `json:"role_type"`
}

type GetRolesParams struct {
	BusinessID *uuid.UUID `json:"business_id"`
}

type GetRolesResponse struct {
	Data []models.Role `json:"data"`
}

type RoleOption struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type RoleOptionsResponse struct {
	Data []RoleOption `json:"data"`
}

type GroupRoleOptionParams struct {
	BusinessID *uuid.UUID `json:"business_id"`
}

type DeleteRoleParams struct {
	ID           uuid.UUID `json:"id"`
	Confirmation string    `json:"confirmation" valid:"required~Confirmation is required"`
}

type GetGroupRolesParams struct {
	BusinessID *uuid.UUID `json:"business_id"`
}

type GetGroupRolesResponse struct {
	Data []models.GroupRole `json:"data"`
}

type GetGroupRoleResponse struct {
	ID                      uuid.UUID                          `json:"id"`
	RoleID                  uuid.UUID                          `json:"role_id"`
	BusinessID              *uuid.UUID                         `json:"business_id"`
	PermissionPresetFormats []constants.PermissionPresetFormat `json:"permission_preset_formats"`
	Role                    models.Role                        `json:"role"`
}

type CreateGroupRoleParams struct {
	BusinessID *uuid.UUID `json:"business_id"`
	RoleID     uuid.UUID  `json:"role_id"`
}

type OptionsResponse struct {
	Data []common.Option `json:"data"`
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

	return &Service{db: db}, nil
}

/*

 /$$   /$$
| $$  | $$
| $$  | $$  /$$$$$$$  /$$$$$$   /$$$$$$   /$$$$$$$
| $$  | $$ /$$_____/ /$$__  $$ /$$__  $$ /$$_____/
| $$  | $$|  $$$$$$ | $$$$$$$$| $$  \__/|  $$$$$$
| $$  | $$ \____  $$| $$_____/| $$       \____  $$
|  $$$$$$/ /$$$$$$$/|  $$$$$$$| $$       /$$$$$$$/
 \______/ |_______/  \_______/|__/      |_______/

*/

// CreateUser creates a new user
//
//encore:api auth method=POST path=/api/admin/user/create
func (s *Service) CreateUser(ctx context.Context, params *CreateUserParams) (*models.User, error) {
	if err := middleware.CheckPermission(constants.CreateUserAction, params.BusinessID, params.OutletID); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	// Check if identification number already exists
	// var existingUserWithID models.User
	// if err := s.db.Where("identification_no = ?", params.IdentificationNo).First(&existingUserWithID).Error; err == nil {
	// 	return nil, &errs.Error{
	// 		Code:    errs.AlreadyExists,
	// 		Message: "This identification number is already registered",
	// 	}
	// } else if !errors.Is(err, gorm.ErrRecordNotFound) {
	// 	return nil, err
	// }

	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", params.Email).First(&existingUser).Error; err == nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "This email already registered",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := common.ValidatePassword(params.Password); err != nil {
		return nil, err
	}

	encodedPassword, err := common.EncodePassword(params.Password)
	if err != nil {
		return nil, err
	}

	var groupRoleID *uuid.UUID
	if params.RoleID != nil {
		_, err := s.GetRole(ctx, *params.RoleID)
		if err != nil {
			return nil, err
		}
		groupRoleID = s.GetGroupRoleIdByRole(ctx, *params.RoleID, params.BusinessID)
	}

	user := &models.User{
		FirstName:        params.FirstName,
		Surname:          params.Surname,
		IdentificationNo: params.IdentificationNo,
		Email:            params.Email,
		Phone:            params.Phone,
		Address:          params.Address,
		BusinessID:       params.BusinessID,
		OutletID:         params.OutletID,
		Pwd:              encodedPassword,
		GroupRoleID:      groupRoleID,
		EmployeeNo:       params.EmployeeNo,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

type UpdateUserPasswordParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password" valid:"required~Password is required"`
}

// encore:api auth method=POST path=/api/admin/user/reset-password
func (s *Service) UpdateUserPassword(ctx context.Context, params *UpdateUserPasswordParams) error {
	if err := middleware.CheckPermission(constants.ResetPasswordAction, nil, nil); err != nil {
		return err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return err
	}

	if err := common.ValidatePassword(params.Password); err != nil {
		return err
	}

	encodedPassword, err := common.EncodePassword(params.Password)
	if err != nil {
		return err
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", params.UserID).Error; err != nil {
		return err
	}

	user.Pwd = encodedPassword

	if err := s.db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}

// GetUser retrieves a user by ID
//
//encore:api auth method=GET path=/api/admin/user/get/:id
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.Preload("GroupRole").First(&user, "id = ?", id).Error; err != nil {
		// if errors.Is(err, gorm.ErrRecordNotFound) {
		// 	return nil, nil
		// }
		return nil, err
	}

	if err := middleware.CheckPermission(constants.ReadUserAction, user.BusinessID, user.OutletID); err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
//
//encore:api auth method=POST path=/api/user/update
func (s *Service) UpdateUser(ctx context.Context, params *UpdateUserParams) (*models.User, error) {
	if err := middleware.CheckPermission(constants.UpdateUserAction, params.BusinessID, params.OutletID); err != nil {
		return nil, err
	}

	var groupRoleID *uuid.UUID
	if params.RoleID != nil {
		_, err := s.GetRole(ctx, *params.RoleID)
		if err != nil {
			return nil, err
		}
		groupRoleID = s.GetGroupRoleIdByRole(ctx, *params.RoleID, params.BusinessID)
	}

	userParams := &models.User{
		ID:          params.ID,
		Email:       params.Email,
		FirstName:   params.FirstName,
		Surname:     params.Surname,
		Phone:       params.Phone,
		Address:     params.Address,
		BusinessID:  params.BusinessID,
		OutletID:    params.OutletID,
		GroupRoleID: groupRoleID,
		EmployeeNo:  params.EmployeeNo,
	}

	if params.Password != nil && *params.Password != "" {
		if err := middleware.CheckPermission(constants.ResetPasswordAction, params.BusinessID, params.OutletID); err != nil {
			return nil, err
		}

		encodedPassword, err := common.EncodePassword(*params.Password)
		if err != nil {
			return nil, err
		}
		userParams.Pwd = encodedPassword
	}

	var user models.User
	if err := s.db.First(&user, "id = ?", params.ID).Error; err != nil {
		return nil, err
	}
	// Remove roles from params to prevent direct update
	result := s.db.Model(&user).Updates(userParams)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// DeleteUser soft deletes a user
//
//encore:api auth method=DELETE path=/api/admin/user/delete/:id
func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", id).Error; err != nil {
		return err
	}

	if err := middleware.CheckPermission(constants.DeleteUserAction, user.BusinessID, user.OutletID); err != nil {
		return err
	}

	return s.db.Delete(&models.User{}, "id = ?", id).Error
}

// GetAllUsers retrieves all users for a business
//
//encore:api auth method=POST path=/api/admin/user/get-all
func (s *Service) GetUsers(ctx context.Context, params *GetAllUsersParams) (*UsersResponse, error) {
	if err := middleware.CheckPermission(constants.ReadUserAction, nil, nil); err != nil {
		return nil, err
	}

	d := auth.Data()
	user := d.(*models.User)

	if params.Page == 0 {
		params.Page = 1
	}

	if params.PageSize == 0 {
		params.PageSize = 10
	}

	var users []models.User

	query := s.db.Offset((params.Page - 1) * params.PageSize).Limit(params.PageSize)
	queryCount := s.db.Model(&models.User{})

	query = query.Preload("GroupRole.Role")

	if params.Filter.BusinessID != nil {
		query = query.Where("business_id = ?", params.Filter.BusinessID)
		queryCount = queryCount.Where("business_id = ?", params.Filter.BusinessID)
	}
	if params.Filter.OutletID != nil {
		query = query.Where("outlet_id = ?", params.Filter.OutletID)
		queryCount = queryCount.Where("outlet_id = ?", params.Filter.OutletID)
	}

	if params.Filter.Name != "" {
		query = query.Where("LOWER(first_name) LIKE LOWER(?) OR LOWER(surname) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(employee_no) LIKE LOWER(?) OR LOWER(identification_no) LIKE LOWER(?)", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%")
		queryCount = queryCount.Where("LOWER(first_name) LIKE LOWER(?) OR LOWER(surname) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?) OR LOWER(employee_no) LIKE LOWER(?) OR LOWER(identification_no) LIKE LOWER(?)", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%", "%"+params.Filter.Name+"%")
	}

	if auth_service.IsSanteAdmin() {
		// Sante admin can access all users
	} else if user.OutletID == nil {
		query = query.Where("business_id = ?", user.BusinessID)
		queryCount = queryCount.Where("business_id = ?", user.BusinessID)
	} else {
		query = query.Where("outlet_id = ?", user.OutletID)
		queryCount = queryCount.Where("outlet_id = ?", user.OutletID)
	}

	if err := query.
		Find(&users).Error; err != nil {
		return nil, err
	}

	var total int64
	if err := queryCount.Count(&total).Error; err != nil {
		return nil, err
	}

	return &UsersResponse{
		Meta: common.Pagination{
			Page:       params.Page,
			PageSize:   params.PageSize,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(params.PageSize))),
		},
		Data: users,
	}, nil
}

func GetUserInAuth(ctx context.Context) (*models.User, error) {
	authData := auth.Data()
	userData, ok := authData.(*models.User)
	if !ok {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid auth data type",
		}
	}

	return userData, nil
}

/*

 /$$$$$$$            /$$
| $$__  $$          | $$
| $$  \ $$  /$$$$$$ | $$  /$$$$$$   /$$$$$$$
| $$$$$$$/ /$$__  $$| $$ /$$__  $$ /$$_____/
| $$__  $$| $$  \ $$| $$| $$$$$$$$|  $$$$$$
| $$  \ $$| $$  | $$| $$| $$_____/ \____  $$
| $$  | $$|  $$$$$$/| $$|  $$$$$$$ /$$$$$$$/
|__/  |__/ \______/ |__/ \_______/|_______/

*/

func convertPermissionPresetsToPermissionPresetsFormat(presets []models.PermissionPreset) []constants.PermissionPresetFormat {
	var permissionPresets []constants.PermissionPresetFormat
	for _, preset := range constants.PermissionPresetsArray {
		tempPreset := preset
		tempPreset.Type = "permission_preset"
		for _, p := range presets {
			if p.Name == preset.Name {
				tempPreset.ID = p.ID
				tempPreset.Enabled = p.Enabled
				break
			}
		}
		permissionPresets = append(permissionPresets, tempPreset)
	}
	return permissionPresets
}

func convertPermissionsToPermissionPresetsFormat(presets []models.Permission) []constants.PermissionPresetFormat {
	var permissionPresets []constants.PermissionPresetFormat
	for _, preset := range constants.PermissionPresetsArray {
		tempPreset := preset
		tempPreset.Type = "permission_preset"
		for _, p := range presets {
			if p.Name == preset.Name {
				tempPreset.ID = p.ID
				tempPreset.GroupRoleID = p.GroupRoleID
				tempPreset.Enabled = p.Enabled
				break
			}
		}
		permissionPresets = append(permissionPresets, tempPreset)
	}
	return permissionPresets
}

// CreateRole creates a new role
//
//encore:api auth method=POST path=/api/admin/role/create
func (s *Service) CreateRole(ctx context.Context, params *CreateRoleParams) (*models.Role, error) {
	if err := middleware.CheckPermission(constants.CreateRoleAction, nil, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	// Check if role already exists
	var existingRole models.Role
	if err := s.db.Where("name = ?", params.Name).First(&existingRole).Error; err == nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "A role with this name already exists",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := &models.Role{
		Name:        params.Name,
		Description: params.Description,
		BusinessID:  params.BusinessID,
		RoleType:    params.RoleType,
	}

	if err := s.db.Create(role).Error; err != nil {
		return nil, err
	}

	for _, preset := range params.PermissionPresetFormats {
		newPreset := models.PermissionPreset{
			Name:      preset.Name,
			Enabled:   preset.Enabled,
			Module:    preset.Module,
			SubModule: preset.SubModule,
			RoleID:    role.ID,
		}
		if err := s.db.Create(&newPreset).Error; err != nil {
			return nil, err
		}

	}

	// Create group role with permissions from presets for any sante admin role
	if params.RoleType == constants.RoleTypeAdmin {
		s.CreateGroupRole(ctx, &CreateGroupRoleParams{
			RoleID: role.ID,
		})
	}

	return role, nil
}

// GetPermissionPreset gets permission presets for a role
//
//encore:api auth method=GET path=/api/admin/role/get-permission-preset
func (s *Service) GetPermissionPreset(ctx context.Context) (GetPermissionPresetResponse, error) {
	permissionPresets := constants.PermissionPresetsArray
	return GetPermissionPresetResponse{
		PermissionPresets: permissionPresets,
	}, nil
}

// UpdateRole updates an existing role
//
//encore:api auth method=POST path=/api/admin/role/update
func (s *Service) UpdateRole(ctx context.Context, params *UpdateRoleParams) (*UpdateRoleParams, error) {
	if err := middleware.CheckPermission(constants.UpdateRoleAction, nil, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	// Check if role exists
	var role models.Role
	if err := s.db.First(&role, "id = ?", params.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Role not found",
			}
		}
		return nil, err
	}

	// Update role
	role.Name = params.Name
	role.Description = params.Description
	role.RoleType = params.RoleType
	if err := s.db.Save(&role).Error; err != nil {
		return nil, err
	}

	is_update_permission_preset := false
	is_update_permission := false
	is_update_permissions_by_all_business := false

	if auth_service.IsSanteAdmin() {
		is_update_permission_preset = true
		if params.RoleType == constants.RoleTypeAdmin {
			is_update_permission = true
		} else if params.RoleType == constants.RoleTypeBusiness {
			is_update_permission = true
		} else {
			is_update_permissions_by_all_business = true
		}
	} else {
		if params.RoleType == constants.RoleTypeBusiness {
			is_update_permission_preset = true
			is_update_permission = true
		} else if params.RoleType == constants.RoleTypeGeneral {
			is_update_permission = true
		}
	}

	if params.RoleType == constants.RoleTypeBusiness {
		is_within_business_err := auth_service.IsAuthWithinBusiness(*params.BusinessID)
		if is_within_business_err == nil {
			is_update_permission_preset = true
			is_update_permission = true
		}
	} else {
		if auth_service.IsSanteAdmin() {
			is_update_permission_preset = true
		}
		if params.RoleType == constants.RoleTypeAdmin {
			is_update_permission = true
		}
	}

	if is_update_permission_preset {
		for _, preset := range params.PermissionPresetFormats {
			var existingPreset models.PermissionPreset
			err := s.db.First(&existingPreset, "id = ?", preset.ID).Error

			// If the preset does not exist, create it
			if err != nil {
				existingPreset.Name = preset.Name
				existingPreset.Enabled = preset.Enabled
				existingPreset.Module = preset.Module
				existingPreset.SubModule = preset.SubModule
				existingPreset.RoleID = role.ID

				if err := s.db.Create(&existingPreset).Error; err != nil {
					return nil, err
				}
			} else {
				// If the preset exists, update it
				existingPreset.Enabled = preset.Enabled
				if err := s.db.Save(&existingPreset).Error; err != nil {
					return nil, err
				}
			}
		}
	}

	if is_update_permission {
		groupRoleID := s.GetGroupRoleIdByRole(ctx, role.ID, params.BusinessID)

		s.DeletePermissions(ctx, *groupRoleID)

		s.UpdatePermissions(ctx, GetGroupRoleResponse{
			ID:                      *groupRoleID,
			PermissionPresetFormats: params.PermissionPresetFormats,
		})
	}

	if is_update_permissions_by_all_business {
		var groupRoleIDs []uuid.UUID
		if err := s.db.Model(&models.GroupRole{}).Where("role_id = ?", role.ID).Pluck("id", &groupRoleIDs).Error; err != nil {
			return nil, err
		}
		for _, groupRoleID := range groupRoleIDs {
			s.DeletePermissions(ctx, groupRoleID)
			s.UpdatePermissions(ctx, GetGroupRoleResponse{
				ID:                      groupRoleID,
				PermissionPresetFormats: params.PermissionPresetFormats,
			})
		}

	}

	updatedRole, err := s.GetRole(ctx, role.ID)
	if err != nil {
		return nil, err
	}

	return updatedRole, nil
}

// GetRole gets a role by ID
//
//encore:api auth method=GET path=/api/admin/role/get/:id
func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (*UpdateRoleParams, error) {
	if err := middleware.CheckPermission(constants.ReadRoleAction, nil, nil); err != nil {
		return nil, err
	}

	var role models.Role
	if err := s.db.Preload("PermissionPresets").First(&role, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Role not found",
			}
		}
		return nil, err
	}

	permissionPresetFormats := convertPermissionPresetsToPermissionPresetsFormat(role.PermissionPresets)

	response := &UpdateRoleParams{
		ID:                      role.ID,
		Name:                    role.Name,
		Description:             role.Description,
		BusinessID:              role.BusinessID,
		PermissionPresetFormats: permissionPresetFormats,
		RoleType:                role.RoleType,
	}

	return response, nil
}

// GetRoles gets all roles
//
//encore:api auth method=POST path=/api/admin/role/get-all
func (s *Service) GetRoles(ctx context.Context, params *GetRolesParams) (*GetRolesResponse, error) {
	if err := middleware.CheckPermission(constants.ReadRoleAction, nil, nil); err != nil {
		return nil, err
	}
	if !auth_service.IsSanteAdmin() {
		params.BusinessID = auth_service.GetUserBusinessID()
	}

	var roles []models.Role
	query := s.db
	if params.BusinessID != nil {
		query = query.Where("(business_id = ? OR business_id IS NULL) AND role_type != ?", params.BusinessID, constants.RoleTypeAdmin)
	} else {
		// query = query.Where("business_id IS NULL")
	}

	if err := query.Order("created_at asc").Find(&roles).Error; err != nil {
		return nil, err
	}

	return &GetRolesResponse{
		Data: roles,
	}, nil
}

// GetRoleOptions retrieves a list of roles with minimal info for dropdown options
//
//encore:api auth method=POST path=/api/admin/role/options
func (s *Service) GetRoleOptions(ctx context.Context, params *GroupRoleOptionParams) (*RoleOptionsResponse, error) {
	var roles []RoleOption
	query := s.db.Model(&models.Role{}).Select("id", "name")

	if params.BusinessID != nil {
		query = query.Where("(role_type NOT IN (?, ?) OR business_id = ? OR role_type IS NULL)", constants.RoleTypeAdmin, constants.RoleTypeBusiness, params.BusinessID)
	} else {
		query = query.Where("role_type = ?", constants.RoleTypeAdmin)
	}

	if err := query.Order("created_at asc").Find(&roles).Error; err != nil {
		return nil, err
	}

	return &RoleOptionsResponse{
		Data: roles,
	}, nil
}

// DeleteRole deletes a role
//
//encore:api auth method=POST path=/api/admin/role/delete
func (s *Service) DeleteRole(ctx context.Context, params *DeleteRoleParams) error {
	if err := middleware.CheckPermission(constants.DeleteRoleAction, nil, nil); err != nil {
		return err
	}

	var role models.Role
	if err := s.db.First(&role, "id = ?", params.ID).Error; err != nil {
		return err
	}

	if params.Confirmation != "delete "+role.Name {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Confirmation is incorrect, please enter 'delete " + role.Name + "'",
		}
	}

	// Delete associated permissions first
	if err := s.db.Where("role_id = ?", params.ID).Delete(&models.PermissionPreset{}).Error; err != nil {
		return err
	}

	// Delete the role
	return s.db.Delete(&role).Error
}

/*
  /$$$$$$                                                /$$$$$$$            /$$
 /$$__  $$                                              | $$__  $$          | $$
| $$  \__/  /$$$$$$    /$$$$$$  /$$   /$$  /$$$$$$       | $$  \ $$  /$$$$$$ | $$  /$$$$$$   /$$$$$$$
| $$ /$$$$ /$$__  $$ /$$__  $$| $$  | $$ /$$__  $$      | $$$$$$$/ /$$__  $$| $$ /$$__  $$ /$$_____/
| $$|_  $$| $$  \__/| $$  \ $$| $$  | $$| $$  \ $$      | $$__  $$| $$  \ $$| $$| $$$$$$$$|  $$$$$$
| $$  \ $$| $$      | $$  | $$| $$  | $$| $$  | $$      | $$  \ $$| $$  | $$| $$| $$_____/ \____  $$
|  $$$$$$/| $$      |  $$$$$$/|  $$$$$$/| $$$$$$$/      | $$  | $$|  $$$$$$/| $$|  $$$$$$$ /$$$$$$$/
|__/  |__/ \______/ |__/ \_______/|__/      |__/       |__/  |__/ \______/ |__/ \_______/|_______/

*/

// GetGroupRoles gets all group roles for a business
//
//encore:api auth method=POST path=/api/admin/group-role/get-all
func (s *Service) GetGroupRoles(ctx context.Context, params *GetGroupRolesParams) (*GetGroupRolesResponse, error) {
	if err := middleware.CheckPermission(constants.ReadRoleAction, params.BusinessID, nil); err != nil {
		return nil, err
	}

	var groupRoles []models.GroupRole
	query := s.db
	if params.BusinessID != nil {
		query = query.Where("business_id = ?", params.BusinessID)
	} else {
		query = query.Where("business_id IS NULL")
	}

	if err := query.Preload("Role").Find(&groupRoles).Error; err != nil {
		return nil, err
	}

	return &GetGroupRolesResponse{
		Data: groupRoles,
	}, nil
}

// GetGroupRole gets a group role by ID
//
//encore:api auth method=GET path=/api/admin/group-role/get/:id
func (s *Service) GetGroupRole(ctx context.Context, id uuid.UUID) (*GetGroupRoleResponse, error) {
	var groupRole models.GroupRole
	if err := s.db.Preload("Role").Preload("Permissions").First(&groupRole, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Group role not found",
			}
		}
		return nil, err
	}

	if err := middleware.CheckPermission(constants.ReadRoleAction, groupRole.BusinessID, nil); err != nil {
		return nil, err
	}

	permissionsPresets := convertPermissionsToPermissionPresetsFormat(groupRole.Permissions)

	response := &GetGroupRoleResponse{
		ID:                      groupRole.ID,
		RoleID:                  groupRole.RoleID,
		BusinessID:              groupRole.BusinessID,
		PermissionPresetFormats: permissionsPresets,
		Role:                    *groupRole.Role,
	}

	return response, nil
}

// CreateGroupRole creates a new group role for a business
//
//encore:api auth method=POST path=/api/admin/group-role/create
func (s *Service) CreateGroupRole(ctx context.Context, params *CreateGroupRoleParams) (*models.GroupRole, error) {
	// Check if role exists
	var role models.Role
	if err := s.db.First(&role, "id = ?", params.RoleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "Role not found",
			}
		}
		return nil, err
	}

	// Check if group role already exists
	var existingGroupRole models.GroupRole
	query := s.db.Where("role_id = ?", params.RoleID)
	if params.BusinessID != nil {
		// Check if business exists
		var business models.Business
		if err := s.db.First(&business, "id = ?", params.BusinessID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &errs.Error{
					Code:    errs.NotFound,
					Message: "Business not found",
				}
			}
			return nil, err
		}

		query = query.Where("business_id = ?", params.BusinessID)
	}

	if err := query.First(&existingGroupRole).Error; err == nil {
		return nil, &errs.Error{
			Code:    errs.AlreadyExists,
			Message: "Group role already exists",
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	groupRole := &models.GroupRole{
		BusinessID: params.BusinessID,
		RoleID:     params.RoleID,
	}

	if err := s.db.Create(groupRole).Error; err != nil {
		return nil, err
	}

	// Get permission presets for this role
	var permissionPresets []models.PermissionPreset
	s.db.Where("role_id = ?", params.RoleID).Find(&permissionPresets)

	// Create permissions from presets
	for _, preset := range permissionPresets {
		permission := &models.Permission{
			Name:        preset.Name,
			Module:      preset.Module,
			SubModule:   preset.SubModule,
			GroupRoleID: groupRole.ID,
			Enabled:     preset.Enabled,
		}
		s.db.Create(permission)
	}

	return groupRole, nil
}

// SetPermissions sets permissions for a group role
//
//encore:api auth method=POST path=/api/admin/group-role/update-permissions
func (s *Service) UpdatePermissions(ctx context.Context, params GetGroupRoleResponse) (*GetGroupRoleResponse, error) {
	if err := middleware.CheckPermission(constants.UpdateRoleAction, params.BusinessID, nil); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(params); err != nil {
		return nil, err
	}

	// Get existing permissions for this group role
	var existingPermissions []models.Permission
	if err := s.db.Where("group_role_id = ?", params.ID).Find(&existingPermissions).Error; err != nil {
		return nil, err
	}

	// Update or create permissions
	for _, permission := range params.PermissionPresetFormats {
		var existingPermission models.Permission
		err := s.db.First(&existingPermission, "id = ?", permission.ID).Error

		// If the preset does not exist, create it
		if err != nil {
			existingPermission.GroupRoleID = params.ID
			existingPermission.Name = permission.Name
			existingPermission.Module = permission.Module
			existingPermission.SubModule = permission.SubModule
			existingPermission.Enabled = permission.Enabled

			if err := s.db.Create(&existingPermission).Error; err != nil {
				return nil, err
			}
		} else {
			// If the preset exists, update it
			existingPermission.Enabled = permission.Enabled
			if err := s.db.Save(&existingPermission).Error; err != nil {
				return nil, err
			}
		}
	}

	return s.GetGroupRole(ctx, params.ID)
}

// GetGroupRoleByRole gets a group role by role ID and business ID
func (s *Service) GetGroupRoleIdByRole(ctx context.Context, roleID uuid.UUID, businessID *uuid.UUID) *uuid.UUID {
	var groupRole models.GroupRole

	_, err := s.GetRole(ctx, roleID)
	if err != nil {
		return nil
	}

	if businessID != nil {
		err = s.db.Where("role_id = ? AND business_id = ?", roleID, businessID).
			First(&groupRole).Error
	} else {
		err = s.db.Where("role_id = ?", roleID).
			First(&groupRole).Error
	}

	if err != nil {
		groupRole, err := s.CreateGroupRole(ctx, &CreateGroupRoleParams{
			RoleID:     roleID,
			BusinessID: businessID,
		})
		if err != nil {
			return nil
		}

		return &groupRole.ID
	}

	return &groupRole.ID
}

// DeletePermissions deletes permissions for a group role
func (s *Service) DeletePermissions(ctx context.Context, groupRoleID uuid.UUID) error {
	if err := s.db.Unscoped().Where("group_role_id = ?", groupRoleID).Delete(&models.Permission{}).Error; err != nil {
		return err
	}

	return nil
}

//encore:api auth method=GET path=/api/admin/user/get-with-outlet-group-role/:businessID
func (s *Service) GetUserWithOutletGroupRole(ctx context.Context, businessID uuid.UUID) (*OptionsResponse, error) {
	var users []models.User
	if err := s.db.Model(&models.User{}).
		Preload("GroupRole.Role").
		Joins("LEFT JOIN group_roles ON users.group_role_id = group_roles.id").
		Joins("LEFT JOIN roles ON group_roles.role_id = roles.id").
		Where("users.business_id = ?", businessID).
		Where("roles.has_outlet_group = ?", true).
		Select("users.id, users.first_name, users.surname, roles.name, users.group_role_id").
		Find(&users).Error; err != nil {
		return nil, err
	}

	var options []common.Option
	for _, user := range users {
		options = append(options, common.Option{
			ID:   user.ID,
			Name: user.FirstName + " " + user.Surname + " (" + user.GroupRole.Role.Name + ")",
		})
	}

	return &OptionsResponse{
		Data: options,
	}, nil
}

/*
    $$$$$$\    $$\                $$$$$$\   $$$$$$\
   $$  __$$\   $$ |              $$  __$$\ $$  __$$\
   $$ /  \__|$$$$$$\    $$$$$$\  $$ /  \__|$$ /  \__|
   \$$$$$$\  \_$$  _|   \____$$\ $$$$\     $$$$\
    \____$$\   $$ |     $$$$$$$ |$$  _|    $$  _|
   $$\   $$ |  $$ |$$\ $$  __$$ |$$ |      $$ |
   \$$$$$$  |  \$$$$  |\$$$$$$$ |$$ |      $$ |
    \______/    \____/  \_______|\__|      \__|

*/

type GetAllUsersInOutletRequest struct {
	OutletID uuid.UUID `json:"outlet_id"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type GetAllUsersInOutletResponse struct {
	Meta common.Pagination `json:"meta"`
	//Data []models.User     `json:"data"`
	Data []CustomUsersInOutlet `json:"data"`
}

type CustomUsersInOutlet struct {
	ID             uuid.UUID             `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID     *uuid.UUID            `json:"business_id" gorm:"type:uuid"`
	OutletID       *uuid.UUID            `json:"outlet_id" gorm:"type:uuid"`
	FirstName      string                `json:"first_name" gorm:"type:varchar(255)" valid:"required~FirstName is required"`
	Surname        string                `json:"surname" gorm:"type:varchar(255)" valid:"required~Surname is required"`
	GroupRoleID    *uuid.UUID            `json:"group_role_id" gorm:"type:uuid"`
	EmployeeNo     string                `json:"employee_no" gorm:"type:varchar(255)" valid:"required~Employee No is required"`
	Email          string                `json:"email" gorm:"type:varchar(255)" valid:"required~Email is required"`
	Phone          string                `json:"phone" gorm:"type:varchar(255)"`
	Address        common.Address        `json:"address" gorm:"embedded"`
	Status         constants.UserStatus  `json:"status" gorm:"type:varchar(255);default:ACTIVE"`
	Business       *models.Business      `gorm:"foreignKey:BusinessID"`
	Outlet         *models.Outlet        `gorm:"foreignKey:OutletID"`
	GroupRole      *models.GroupRole     `gorm:"foreignKey:GroupRoleID"`
	FCMDeviceToken *string               `json:"fcm_device_token" gorm:"type:varchar(255)"`
	OutletGroups   []*models.OutletGroup `gorm:"many2many:outlet_groups_users;"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      *time.Time            `json:"updated_at"`
	LastLoginAt    *time.Time            `json:"last_login_at" gorm:"-"`
	LastLogoutAt   *time.Time            `json:"last_logout_at" gorm:"-"`
}

// API to get all users in a outlet in staff management
//
//encore:api auth method=POST path=/api/admin/staff/get-all-users-in-outlet
func (s *Service) GetAllUsersInOutlet(ctx context.Context, req *GetAllUsersInOutletRequest) (*GetAllUsersInOutletResponse, error) {
	var users []CustomUsersInOutlet
	query := s.db.Model(&models.User{}).
		Select("id", "business_id", "outlet_id", "first_name", "surname", "group_role_id",
			"employee_no", "email", "phone", "street_line1", "street_line2", "street_line3",
			"city", "state", "postal_code", "country", "status", "fcm_device_token",
			"created_at", "updated_at").
		Where("outlet_id = ?", req.OutletID).Order("created_at ASC")

	var total int64
	var totalPages int32
	query.Count(&total)
	totalPages = int32(math.Ceil(float64(total) / float64(req.PageSize)))

	query = query.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize)

	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}

	// Populate LastLoginAt and LastLogoutAt from activity logs
	for i := range users {
		userID := users[i].ID

		// get group role
		groupRoleID := *users[i].GroupRoleID
		groupRole, err := common_operations.GetGroupRoleByID(s.db, groupRoleID, true)
		if err != nil {
			return nil, err
		}
		users[i].GroupRole = groupRole

		// Get latest Staff Login
		var latestLogin models.ActivityLog
		loginErr := s.db.Where("action_by_user_id = ? AND activity = ?", userID, constants.LOG_ACTION_STAFF_LOGIN).
			Order("created_at DESC").
			First(&latestLogin).Error

		// Get latest Staff Logout
		var latestLogout models.ActivityLog
		logoutErr := s.db.Where("action_by_user_id = ? AND activity = ?", userID, constants.LOG_ACTION_STAFF_LOGOUT).
			Order("created_at DESC").
			First(&latestLogout).Error

		// Set LastLoginAt if login exists
		if loginErr == nil {
			users[i].LastLoginAt = &latestLogin.CreatedAt
		}

		// Set LastLogoutAt if logout exists
		if logoutErr == nil {
			users[i].LastLogoutAt = &latestLogout.CreatedAt
		}

		// If both exist, prioritize the more recent one
		if loginErr == nil && logoutErr == nil {
			if latestLogin.CreatedAt.After(latestLogout.CreatedAt) {
				// Login is more recent, keep LastLoginAt set
				users[i].LastLogoutAt = nil // or keep it, depending on requirement
			} else {
				// Logout is more recent, keep LastLogoutAt set
				users[i].LastLoginAt = nil // or keep it, depending on requirement
			}
		}
	}

	return &GetAllUsersInOutletResponse{
		Meta: common.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: int(totalPages),
		},
		Data: users,
	}, nil
}
