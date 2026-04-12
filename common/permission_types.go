package common

import (
	"encore.app/common/constants"
	"encore.dev/types/uuid"
)

type PermissionPreset struct {
	ID          uuid.UUID
	Name        constants.ActionPermission
	Module      constants.ModulePermission
	SubModule   constants.SubModulePermission
	RoleID      uuid.UUID
	GroupRoleID uuid.UUID
	Description string
	Enabled     bool
}
