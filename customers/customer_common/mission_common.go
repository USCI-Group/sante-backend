package customer_common

import (
	"time"

	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

// ========================================
// MISSION
// ========================================

// create a valid output for mission status
func ProperNameForMissionStatus(status models.CustomerMissionStatus) string {
	switch status {
	case models.CustomerMissionStatusInProgress:
		return "In Progress"
	case models.CustomerMissionStatusCompleted:
		return "Completed"
	case models.CustomerMissionStatusExpired:
		return "Expired"
	case models.CustomerMissionStatusCancelled:
		return "Cancelled"
	}
	return "Unknown"
}

// get missions by business id
func GetMissionsByBusinessID(trx *gorm.DB, business_id uuid.UUID, offset int, page_size int, isPreload bool) ([]models.Mission, error) {
	var missions []models.Mission
	query := trx.Model(&models.Mission{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Order("created_at DESC").
		Offset(offset).
		Limit(page_size)
	if isPreload {
		query = query.Preload("MissionCriteria").
			Preload("MissionRewards")
	}
	result := query.Find(&missions)
	if result.Error != nil {
		return nil, result.Error
	}
	return missions, nil
}

// get missions count by business id
func GetMissionsCountByBusinessID(trx *gorm.DB, business_id uuid.UUID) (int64, error) {
	var count int64
	result := trx.Model(&models.Mission{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", true).
		Count(&count)
	return count, result.Error
}

// get mission by id
func GetMissionByID(
	trx *gorm.DB,
	mission_id uuid.UUID,
	preloadProducts bool,
	preloadOutlets bool,
	preloadMemberships bool,
) (*models.Mission, error) {
	var mission models.Mission
	query := trx.Model(&models.Mission{}).
		Where("id = ?", mission_id).
		Preload("MissionCriteria").
		Preload("MissionCriteria.Membership").
		Preload("MissionRewards").
		Preload("MissionRewards.PointRule").
		Preload("MissionRewards.Voucher")
	if preloadProducts {
		query = query.Preload("MissionCriteria.Product")
	}
	if preloadOutlets {
		query = query.Preload("MissionCriteria.Outlet")
	}
	if preloadMemberships {
		query = query.Preload("MissionCriteria.Membership")
	}
	result := query.First(&mission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &mission, nil
}

func GetDistinctCustomerMissionsByCustomerID(trx *gorm.DB, customer_id uuid.UUID) ([]models.CustomerMission, error) {
	var customerMissions []models.CustomerMission
	// Use a subquery to get the latest started_at for each mission_id
	result := trx.Model(&models.CustomerMission{}).
		Where("customer_id = ?", customer_id).
		Where("(mission_id, started_at) IN (SELECT mission_id, MAX(started_at) FROM customer_missions WHERE customer_id = ? GROUP BY mission_id)", customer_id).
		Preload("Mission").
		Preload("Mission.MissionCriteria").
		Preload("Mission.MissionRewards").
		Order("started_at DESC").
		Find(&customerMissions)

	if result.Error != nil {
		return []models.CustomerMission{}, nil
	}
	return customerMissions, nil
}

// get customer mission by customer id (list)
func GetCustomerMissionsByCustomerID(trx *gorm.DB, customer_id uuid.UUID) ([]models.CustomerMission, error) {
	var customerMissions []models.CustomerMission
	result := trx.Model(&models.CustomerMission{}).
		Where("customer_id = ?", customer_id).
		Preload("Mission").
		Preload("Mission.MissionCriteria").
		Preload("Mission.MissionRewards").
		Order("started_at DESC").
		Find(&customerMissions)
	if result.Error != nil {
		return []models.CustomerMission{}, nil
	}
	return customerMissions, nil
}

// get customer mission by customer id and mission id (single)
func GetCustomerMissionByCustomerIDAndMissionID(trx *gorm.DB, customer_id uuid.UUID, mission_id uuid.UUID) (*models.CustomerMission, error) {
	var customerMission models.CustomerMission
	result := trx.Model(&models.CustomerMission{}).
		Where("customer_id = ? AND mission_id = ?", customer_id, mission_id).
		First(&customerMission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerMission, nil
}

// get customer mission by customer id and mission id (single)
func GetCustomerMissionByID(trx *gorm.DB, customer_mission_id uuid.UUID) (*models.CustomerMission, error) {
	var customerMission models.CustomerMission
	result := trx.Model(&models.CustomerMission{}).
		Where("id = ?", customer_mission_id).
		First(&customerMission)
	if result.Error != nil {
		return nil, result.Error
	}
	return &customerMission, nil
}

// get customer mission criteria progress by customer mission id (list)
func GetCustomerMissionCriteriaProgressByCustomerMissionID(trx *gorm.DB, customer_mission_id uuid.UUID) ([]models.CustomerMissionCriteriaProgress, error) {
	var customerMissionCriteriaProgress []models.CustomerMissionCriteriaProgress
	result := trx.Model(&models.CustomerMissionCriteriaProgress{}).
		Where("customer_mission_id = ?", customer_mission_id).
		Preload("MissionCriteria").
		Find(&customerMissionCriteriaProgress)
	if result.Error != nil {
		return nil, result.Error
	}
	return customerMissionCriteriaProgress, nil
}

// calculate customer mission progress
func CalculateCustomerMissionProgress(trx *gorm.DB, customerMissionID uuid.UUID) (float64, error) {
	customerMissionCriteriaProgress, err := GetCustomerMissionCriteriaProgressByCustomerMissionID(trx, customerMissionID)
	if err != nil {
		return 0, err
	}

	var missionCompleted float64 = 0
	var missionTotal float64 = 0
	for _, cmcp := range customerMissionCriteriaProgress {
		mc := cmcp.MissionCriteria
		// only location based criteria is not counted for mission progress
		// this criteria is not independent criteria, it is dependent on the other criteria to be satisfied
		// It like add-on criteria on top of the other criteria to be satisfied
		if mc.CriteriaType == models.MissionCriteriaTypeLocationBased {
			continue
		}
		// add all mission criteria count as total
		missionTotal++
		if cmcp.IsSatisfied && mc.IsActive {
			missionCompleted++
		}
	}

	if missionTotal == 0 {
		return 0, nil
	}

	// calculate mission progress
	missionProgress := missionCompleted / missionTotal
	return missionProgress, nil
}

// get mission rewards by mission id (list)
func GetMissionRewardsByMissionIDs(trx *gorm.DB, mission_ids []uuid.UUID, preloadPointRule bool, preloadVoucher bool) ([]models.MissionReward, error) {
	var missionRewards []models.MissionReward
	query := trx.Model(&models.MissionReward{}).
		Where("mission_id IN ?", mission_ids).
		Where("is_active = ?", true).
		Order("created_at DESC")
	if preloadPointRule {
		query = query.Preload("PointRule")
	}
	if preloadVoucher {
		query = query.Preload("Voucher")
	}
	result := query.Find(&missionRewards)
	if result.Error != nil {
		return nil, result.Error
	}
	return missionRewards, nil
}

// update mission reward grant
func CreateMissionRewardGrant(trx *gorm.DB, mission_reward_id uuid.UUID, customer_mission_id uuid.UUID) error {
	missionRewardGrant := models.MissionRewardGrant{
		CustomerMissionID: customer_mission_id,
		MissionRewardID:   mission_reward_id,
		GrantedAt:         time.Now(),
		CreatedAt:         time.Now(),
	}
	result := trx.Create(&missionRewardGrant)
	if result.Error != nil {
		return result.Error
	}
	if result.Error != nil {
		return result.Error
	}
	return nil
}
