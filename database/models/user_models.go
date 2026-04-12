package models

import (
	"time"

	"encore.app/common"
	"encore.app/common/constants"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID            `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	BusinessID       *uuid.UUID           `json:"business_id" gorm:"type:uuid"`
	OutletID         *uuid.UUID           `json:"outlet_id" gorm:"type:uuid"`
	FirstName        string               `json:"first_name" gorm:"type:varchar(255)" valid:"required~FirstName is required"`
	Surname          string               `json:"surname" gorm:"type:varchar(255)" valid:"required~Surname is required"`
	GroupRoleID      *uuid.UUID           `json:"group_role_id" gorm:"type:uuid"`
	IdentificationNo string               `json:"identification_no" gorm:"type:varchar(255)" valid:"required~Identification No is required"`
	EmployeeNo       string               `json:"employee_no" gorm:"type:varchar(255)" valid:"required~Employee No is required"`
	Email            string               `json:"email" gorm:"type:varchar(255)" valid:"required~Email is required"`
	Pwd              string               `json:"-" gorm:"type:varchar(255)"`
	Phone            string               `json:"phone" gorm:"type:varchar(255)"`
	Address          common.Address       `json:"address" gorm:"embedded"`
	Status           constants.UserStatus `json:"status" gorm:"type:varchar(255);default:ACTIVE"`
	Business         *Business            `gorm:"foreignKey:BusinessID"`
	Outlet           *Outlet              `gorm:"foreignKey:OutletID"`
	GroupRole        *GroupRole           `gorm:"foreignKey:GroupRoleID"`
	FCMDeviceToken   *string              `json:"fcm_device_token" gorm:"type:varchar(255)"`
	OutletGroups     []*OutletGroup       `gorm:"many2many:outlet_groups_users;"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        *time.Time           `json:"updated_at"`
	DeletedAt        gorm.DeletedAt       `json:"deleted_at,omitempty"`
}

type Role struct {
	ID                uuid.UUID          `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name              string             `json:"name" gorm:"type:varchar(255)"`
	Description       string             `json:"description" gorm:"type:varchar(255)"`
	RoleType          constants.RoleType `json:"role_type" gorm:"type:varchar(255)"`
	PermissionPresets []PermissionPreset `json:"permission_presets"`
	BusinessID        *uuid.UUID         `json:"business_id" gorm:"type:uuid;index;nullable"`
	HasOutletGroup    bool               `json:"has_outlet_group" gorm:"type:boolean;default:false"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         *time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt     `json:"deleted_at,omitempty"`
}

// GroupRole is a role that is assigned to users within a business(or group that could be the sante Admin)
// If the businessID is null, then the group role is a global role(SANTE Admin)
type GroupRole struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	RoleID      uuid.UUID      `json:"role_id" gorm:"type:uuid;index" valid:"required~Role is required"`
	BusinessID  *uuid.UUID     `json:"business_id" gorm:"type:uuid;index;nullable"`
	Permissions []Permission   `json:"permissions" gorm:"foreignKey:GroupRoleID"`
	Role        *Role          `gorm:"foreignKey:RoleID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
}

type Permission struct {
	ID          uuid.UUID                     `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name        constants.ActionPermission    `json:"name" gorm:"type:varchar(255);index"`
	Module      constants.ModulePermission    `json:"module" gorm:"type:varchar(255)"`
	SubModule   constants.SubModulePermission `json:"sub_module" gorm:"type:varchar(255)"`
	GroupRoleID uuid.UUID                     `json:"group_role_id" gorm:"type:uuid;index" valid:"required~GroupRoleID is required"`
	Enabled     bool                          `json:"enabled" gorm:"type:boolean;default:false"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   *time.Time                    `json:"updated_at"`
	DeletedAt   gorm.DeletedAt                `json:"deleted_at,omitempty"`
}

type PermissionPreset struct {
	ID        uuid.UUID                     `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	Name      constants.ActionPermission    `json:"name" gorm:"type:varchar(255)"`
	Module    constants.ModulePermission    `json:"module" gorm:"type:varchar(255)"`
	SubModule constants.SubModulePermission `json:"sub_module" gorm:"type:varchar(255)"`
	RoleID    uuid.UUID                     `json:"role_id" gorm:"type:uuid;index"`
	Enabled   bool                          `json:"enabled" gorm:"type:boolean;default:false"`
	CreatedAt time.Time                     `json:"created_at"`
	UpdatedAt *time.Time                    `json:"updated_at"`
	DeletedAt gorm.DeletedAt                `json:"deleted_at,omitempty"`
}

type UserToken struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid; default:gen_random_uuid()"`
	UserID       uuid.UUID      `json:"user_id" gorm:"type:uuid;index"`
	RefreshToken string         `json:"refresh_token" gorm:"type:varchar(255)"`
	DeviceID     *string        `json:"device_id" gorm:"type:varchar(255)"`
	ExpiredAt    time.Time      `json:"expired_at" gorm:"type:timestamp with time zone"`
	IsStaffLogin bool           `json:"is_staff_login" gorm:"type:boolean;default:false"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty"`
}
