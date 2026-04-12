package common_operations

import (
	"errors"
	"time"

	"encore.app/common"
	"encore.app/database/models"
	"encore.dev/types/uuid"
	"gorm.io/gorm"
)

func GetDayOfWeek(date time.Time) models.DayOfWeek {
	switch date.Weekday() {
	case time.Monday:
		return models.DayOfWeekMonday
	case time.Tuesday:
		return models.DayOfWeekTuesday
	case time.Wednesday:
		return models.DayOfWeekWednesday
	case time.Thursday:
		return models.DayOfWeekThursday
	case time.Friday:
		return models.DayOfWeekFriday
	case time.Saturday:
		return models.DayOfWeekSaturday
	case time.Sunday:
		return models.DayOfWeekSunday
	}
	return models.DayOfWeekMonday
}

func GetOutletByID(db *gorm.DB, outlet_id uuid.UUID) (*models.Outlet, error) {
	var outlet models.Outlet
	err := db.Model(&models.Outlet{}).Where("id = ?", outlet_id).First(&outlet).Error
	if err != nil {
		return nil, err
	}
	return &outlet, nil
}

func GetOutletGroup(db *gorm.DB, outlet_group_id uuid.UUID) (*models.OutletGroup, error) {
	var outletGroup models.OutletGroup
	err := db.Model(&models.OutletGroup{}).Where("id = ?", outlet_group_id).Preload("Outlets").First(&outletGroup).Error
	if err != nil {
		return nil, err
	}

	return &outletGroup, nil
}

func GetOutletIDsByGroupID(db *gorm.DB, outlet_group_id uuid.UUID) ([]uuid.UUID, error) {
	var outletIDs []uuid.UUID
	outletGroup, err := GetOutletGroup(db, outlet_group_id)
	if err != nil {
		return nil, err
	}

	for _, outlet := range outletGroup.Outlets {
		outletIDs = append(outletIDs, outlet.ID)
	}

	return outletIDs, nil
}

// function to get outlet operation schedule
func GetOutletOperationSchedule(trx *gorm.DB, outlet_id uuid.UUID, additional_query *gorm.DB) (*[]models.OutletOperationSchedule, error) {
	var outletOperationSchedules []models.OutletOperationSchedule
	query := trx.Model(&models.OutletOperationSchedule{}).
		Where("outlet_id = ?", outlet_id).
		Where("is_active = ?", true)
	if additional_query != nil {
		query = query.Where(additional_query)
	}
	result := query.Find(&outletOperationSchedules)

	length := len(outletOperationSchedules)
	if result.Error != nil {
		return nil, result.Error
	}
	if length == 0 {
		outletOperationSchedules, err := GenerateDefaultOutletOperationSchedule(trx, outlet_id)
		if err != nil {
			return nil, err
		}
		ConvertTimeToMalaysiaTimezoneForOperationSchedule(*outletOperationSchedules)
		return outletOperationSchedules, nil
	}

	ConvertTimeToMalaysiaTimezoneForOperationSchedule(outletOperationSchedules)
	return &outletOperationSchedules, nil
}

func ConvertTimeToMalaysiaTimezoneForOperationSchedule(outletOperationSchedules []models.OutletOperationSchedule) {
	for i := range outletOperationSchedules {
		openTime, err := common.GetTimeInMalaysiaTimezone(outletOperationSchedules[i].OpenTime)
		if err != nil {
			continue
		}
		closeTime, err := common.GetTimeInMalaysiaTimezone(outletOperationSchedules[i].CloseTime)
		if err != nil {
			continue
		}
		outletOperationSchedules[i].OpenTime = openTime
		outletOperationSchedules[i].CloseTime = closeTime
	}
}

// function to get start and end time of the outlet
// if all same then return the first one
// if not all same, get the min time among all start time and max time among all end time
func GetOutletStartAndEndTime(trx *gorm.DB, outlet_id uuid.UUID) (time.Time, time.Time, error) {
	outletOperationSchedules, err := GetOutletOperationSchedule(trx, outlet_id, nil)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("outlet operation schedule not found")
	}

	minStartTime := (*outletOperationSchedules)[0].OpenTime
	maxEndTime := (*outletOperationSchedules)[0].CloseTime
	for _, operationSchedule := range *outletOperationSchedules {
		if operationSchedule.OpenTime.Before(minStartTime) {
			minStartTime = operationSchedule.OpenTime
		}
		if operationSchedule.CloseTime.After(maxEndTime) {
			maxEndTime = operationSchedule.CloseTime
		}
	}
	return minStartTime, maxEndTime, nil
}

// function to get outlet pickup time slot
func GetOutletOperationTimeSlot(trx *gorm.DB, outlet_id uuid.UUID, additional_query *gorm.DB, is_asc *bool) (*[]models.OutletOperationTimeSlot, error) {
	var outletOperationTimeSlots []models.OutletOperationTimeSlot
	query := trx.Model(&models.OutletOperationTimeSlot{}).
		Where("outlet_id = ?", outlet_id).
		Where("is_active = ?", true).
		Where("is_pickup_available = ?", true)
	if additional_query != nil {
		query = query.Where(additional_query)
	}
	if is_asc != nil {
		if *is_asc {
			query = query.Order("start_time ASC")
		} else {
			query = query.Order("start_time DESC")
		}
	}
	result := query.Find(&outletOperationTimeSlots)
	if result.Error != nil {
		return nil, result.Error
	}
	// If no records found, generate default pickup time slots
	if len(outletOperationTimeSlots) == 0 {
		// get start and end time of the outlet
		minStartTime, maxEndTime, err := GetOutletStartAndEndTime(trx, outlet_id)
		if err != nil {
			return nil, err
		}
		// generate default pickup time slot
		defaultOperationTimeSlots, err := GenerateDefaultOperationTimeSlot(trx, outlet_id, minStartTime, maxEndTime)
		if err != nil {
			return nil, err
		}
		ConvertTimeToUTCForOperationTimeSlot(*defaultOperationTimeSlots)
		return defaultOperationTimeSlots, nil
	}
	ConvertTimeToUTCForOperationTimeSlot(outletOperationTimeSlots)
	return &outletOperationTimeSlots, nil
}

// function to convert and ensure time in UTC 0
func ConvertTimeToUTCForOperationTimeSlot(outletOperationTimeSlots []models.OutletOperationTimeSlot) {
	for i := range outletOperationTimeSlots {
		outletOperationTimeSlots[i].StartTime = outletOperationTimeSlots[i].StartTime.UTC()
		outletOperationTimeSlots[i].EndTime = outletOperationTimeSlots[i].EndTime.UTC()
	}
}

// function to get outlet pickup time slot
func GetOutletOperationTimeSlotWithAvailability(trx *gorm.DB, outlet_id uuid.UUID, is_asc *bool) (*[]models.OutletOperationTimeSlot, error) {
	var outletOperationTimeSlots []models.OutletOperationTimeSlot
	query := trx.Model(&models.OutletOperationTimeSlot{}).
		Where("outlet_id = ?", outlet_id).
		Where("is_active = ?", true)
	if is_asc != nil {
		if *is_asc {
			query = query.Order("start_time ASC")
		} else {
			query = query.Order("start_time DESC")
		}
	}
	result := query.Find(&outletOperationTimeSlots)
	if result.Error != nil {
		return nil, result.Error
	}
	// If no records found, generate default pickup time slots
	if len(outletOperationTimeSlots) == 0 {
		// get start and end time of the outlet
		minStartTime, maxEndTime, err := GetOutletStartAndEndTime(trx, outlet_id)
		if err != nil {
			return nil, err
		}
		// generate default pickup time slot
		defaultOperationTimeSlots, err := GenerateDefaultOperationTimeSlot(trx, outlet_id, minStartTime, maxEndTime)
		if err != nil {
			return nil, err
		}
		ConvertTimeToMalaysiaTimezoneForOperationTimeSlot(*defaultOperationTimeSlots)
		return defaultOperationTimeSlots, nil
	}
	ConvertTimeToMalaysiaTimezoneForOperationTimeSlot(outletOperationTimeSlots)
	return &outletOperationTimeSlots, nil
}

func ConvertTimeToMalaysiaTimezoneForOperationTimeSlot(outletOperationTimeSlots []models.OutletOperationTimeSlot) {
	for i := range outletOperationTimeSlots {
		startTime, err := common.GetTimeInMalaysiaTimezone(outletOperationTimeSlots[i].StartTime)
		if err != nil {
			continue
		}
		endTime, err := common.GetTimeInMalaysiaTimezone(outletOperationTimeSlots[i].EndTime)
		if err != nil {
			continue
		}
		outletOperationTimeSlots[i].StartTime = startTime
		outletOperationTimeSlots[i].EndTime = endTime
	}
}

// function to generate default outlet operation schedule
func GenerateDefaultOutletOperationSchedule(trx *gorm.DB, outlet_id uuid.UUID) (*[]models.OutletOperationSchedule, error) {
	// if no record found, create a list of 7 days operation schedule
	outletOperationSchedules := []models.OutletOperationSchedule{}
	daysOfWeek := []models.DayOfWeek{models.DayOfWeekMonday, models.DayOfWeekTuesday, models.DayOfWeekWednesday, models.DayOfWeekThursday, models.DayOfWeekFriday, models.DayOfWeekSaturday, models.DayOfWeekSunday}
	for _, dow := range daysOfWeek {
		loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
		if err != nil {
			return nil, err
		}
		openTime := time.Date(2025, 1, 1, 10, 0, 0, 0, loc)
		closeTime := time.Date(2025, 1, 1, 22, 0, 0, 0, loc)

		outletOperationSchedules = append(outletOperationSchedules, models.OutletOperationSchedule{
			OutletID:       outlet_id,
			DayOfWeek:      dow,
			IsActive:       true,
			IsClosed:       false,
			OpenTime:       openTime,
			CloseTime:      closeTime,
			BreakStartTime: nil,
			BreakEndTime:   nil,
			IsBreakActive:  false,
			CreatedAt:      time.Now(),
		})
	}
	result := trx.Create(&outletOperationSchedules)
	if result.Error != nil {
		return nil, result.Error
	}

	return &outletOperationSchedules, nil
}

// function to generate default pickup time slot
func GenerateDefaultOperationTimeSlot(trx *gorm.DB, outlet_id uuid.UUID, start_time time.Time, end_time time.Time) (*[]models.OutletOperationTimeSlot, error) {
	// generate default pickup time slot
	interval := 30 * time.Minute
	var outletOperationTimeSlots []models.OutletOperationTimeSlot
	for start_time.Before(end_time) {
		outletOperationTimeSlots = append(outletOperationTimeSlots, models.OutletOperationTimeSlot{
			OutletID:          outlet_id,
			StartTime:         start_time,
			EndTime:           start_time.Add(interval),
			IsPickupAvailable: true,
			IsActive:          true,
			CreatedAt:         time.Now(),
		})
		start_time = start_time.Add(interval)
	}
	result := trx.Create(&outletOperationTimeSlots)
	if result.Error != nil {
		return nil, result.Error
	}
	return &outletOperationTimeSlots, nil
}

// function to get product that is inactive in business
func GetProductInactiveBusiness(trx *gorm.DB, business_id uuid.UUID) ([]uuid.UUID, error) {
	var productInactiveBusiness []uuid.UUID
	result := trx.Model(&models.Product{}).
		Where("business_id = ?", business_id).
		Where("is_active = ?", false).
		Select("id").
		Find(&productInactiveBusiness)
	if result.Error != nil {
		return nil, result.Error
	}
	return productInactiveBusiness, nil
}

// func to close online order
func ToggleOnlineOrder(trx *gorm.DB, outlet_id uuid.UUID, status models.OutletStatus) error {
	// get all orders with status "new"
	var outlet models.Outlet
	result := trx.Model(&models.Outlet{}).
		Where("id = ?", outlet_id).
		First(&outlet)
	if result.Error != nil {
		return result.Error
	}

	switch status {
	case models.OutletStatusClosed:
		outlet.OnlineOrderEnabled = false
	case models.OutletStatusOpen:
		outlet.OnlineOrderEnabled = true
	default:
		return errors.New("invalid status")
	}

	return trx.Save(&outlet).Error
}
