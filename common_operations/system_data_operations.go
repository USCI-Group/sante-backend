package common_operations

import (
	"encore.app/common"
	"encore.app/common/constants"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

func GetSystemDataValue(db *gorm.DB, infoType constants.SystemDataInfoType) string {
	systemData := &models.SystemData{}
	err := db.Where("info_type = ?", infoType).First(systemData).Error
	if err != nil {
		return ""
	}

	if systemData.IsEncrypted {
		decryptedInfoValue, err := common.DecryptText(systemData.InfoValue)
		if err != nil {
			return ""
		}
		return decryptedInfoValue
	}
	return systemData.InfoValue
}

// Write log to activity log table
func WriteLog(
	trx *gorm.DB,
	activity constants.LogAction,
	status constants.ActivityLogStatus,
	details string, // json string of the details
	actionByUserID *uuid.UUID, // may be user id or system id
	actionBy string, // may be user name or system name
	errorMessage string, // may be error message or system message
) error {
	logActivity := &models.ActivityLog{
		Activity:       activity,
		Status:         status,
		Details:        details,
		ActionByUserID: *actionByUserID,
		ActionBy:       actionBy,
		ErrorMessage:   errorMessage,
	}
	err := trx.Create(logActivity).Error
	if err != nil {
		return err
	}
	return nil
}
