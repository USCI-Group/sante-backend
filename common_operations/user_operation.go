package common_operations

import (
	"fmt"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// func to get customer data from auth data
func GetUserDataFromAuthData(authData func() any) (*models.User, error) {
	userData, ok := authData().(*models.User)
	if !ok {
		return nil, fmt.Errorf("invalid auth data type")
	}
	return userData, nil
}

func GetUserByID(db *gorm.DB, userID uuid.UUID, preloadGroupRole bool, preloadGroupRolePermissions bool, preloadGroupRoleRole bool) (*models.User, error) {
	var user models.User
	query := db.Model(&models.User{}).Where("id = ?", userID)
	if preloadGroupRole {
		query = query.Preload("GroupRole")
	}
	if preloadGroupRolePermissions {
		query = query.Preload("GroupRole.Permissions")
	}
	if preloadGroupRoleRole {
		query = query.Preload("GroupRole.Role")
	}
	result := query.First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
