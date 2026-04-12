package middleware

import (
	"encore.app/auth_service"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

func CheckPermission(action_permission constants.ActionPermission, businessID *uuid.UUID, outletID *uuid.UUID) error {
	d := auth.Data()
	user := d.(*models.User)

	if user.GroupRole == nil {
		return &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "User has no role",
		}
	}

	// Check if user has the required permission
	is_permission_granted := false
	for _, permission := range user.GroupRole.Permissions {
		enabled := permission.Enabled
		if permission.Name == action_permission && enabled {
			is_permission_granted = true
			break
		}
	}

	if !is_permission_granted {
		return &errs.Error{
			Code:    errs.PermissionDenied,
			Message: "User does not have required permission",
		}
	}

	if auth_service.IsSanteAdmin() {
		return nil
	}

	if businessID != nil && *businessID != uuid.Nil {
		if err := auth_service.IsAuthWithinBusiness(*businessID); err != nil {
			return err
		}
	}

	if outletID != nil && *outletID != uuid.Nil {
		if err := auth_service.IsAuthWithinOutlet(*outletID); err != nil {
			return err
		}
	}

	return nil
}

func CheckSanteAdmin() error {
	if auth_service.IsSanteAdmin() {
		return nil
	}
	return &errs.Error{
		Code:    errs.PermissionDenied,
		Message: "User is not a SANTE Admin",
	}
}
