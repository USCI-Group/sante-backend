package common_operations

import (
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

func GetAllPermissionsOfUser(db *gorm.DB, userID uuid.UUID, preloadGroupRole bool, preloadGroupRolePermissions bool, preloadGroupRoleRole bool) (*models.GroupRole, error) {
	user, err := GetUserByID(db, userID, preloadGroupRole, preloadGroupRolePermissions, preloadGroupRoleRole)
	if err != nil {
		return nil, err
	}
	return user.GroupRole, nil
}
