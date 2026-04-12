package constants

type UnitMeasurement string

const (
	UnitMeasurementKilogram   UnitMeasurement = "kg"
	UnitMeasurementGram       UnitMeasurement = "g"
	UnitMeasurementLiter      UnitMeasurement = "L"
	UnitMeasurementMilliliter UnitMeasurement = "ml"
	UnitMeasurementCentimeter UnitMeasurement = "cm"
	UnitMeasurementMeter      UnitMeasurement = "m"
	UnitMeasurementInch       UnitMeasurement = "inch"
	UnitMeasurementPiece      UnitMeasurement = "pcs"
)

// IsLargeUnitMeasurement checks if the given unit measurement is considered a large unit
func IsLargeUnitMeasurement(unit UnitMeasurement) bool {
	switch unit {
	case UnitMeasurementKilogram, UnitMeasurementLiter, UnitMeasurementMeter:
		return true
	default:
		return false
	}
}

// GetSmallUnitFromLarge returns the corresponding small unit for a large unit measurement
func GetSmallUnitFromLarge(unit UnitMeasurement) UnitMeasurement {
	switch unit {
	case UnitMeasurementKilogram:
		return UnitMeasurementGram
	case UnitMeasurementLiter:
		return UnitMeasurementMilliliter
	case UnitMeasurementMeter:
		return UnitMeasurementCentimeter
	default:
		return unit
	}
}

// GetLargeUnitFromSmall returns the corresponding large unit for a small unit measurement
func GetLargeUnitFromSmall(unit UnitMeasurement) UnitMeasurement {
	switch unit {
	case UnitMeasurementGram:
		return UnitMeasurementKilogram
	case UnitMeasurementMilliliter:
		return UnitMeasurementLiter
	case UnitMeasurementCentimeter:
		return UnitMeasurementMeter
	default:
		return unit
	}
}

func ConvertToLargeUnit(originalUnit UnitMeasurement, quantity float32) float32 {
	if IsLargeUnitMeasurement(originalUnit) {
		return quantity
	}

	smallUnit := GetSmallUnitFromLarge(originalUnit)
	smallUnitQuantity := GetSmallUnitQuantity(smallUnit)

	return quantity / smallUnitQuantity
}

func ConvertToSmallUnit(originalUnit UnitMeasurement, quantity float32) float32 {
	if !IsLargeUnitMeasurement(originalUnit) {
		return quantity
	}

	smallUnit := GetSmallUnitFromLarge(originalUnit)
	smallUnitQuantity := GetSmallUnitQuantity(smallUnit)

	return quantity * smallUnitQuantity
}

func ConvertToTargetUnit(originalUnit UnitMeasurement, quantity float32, targetUnit UnitMeasurement) float32 {
	if IsLargeUnitMeasurement(targetUnit) {
		return ConvertToLargeUnit(originalUnit, quantity)
	}

	return ConvertToSmallUnit(originalUnit, quantity)
}

func GetSmallUnitQuantity(unit UnitMeasurement) float32 {
	switch unit {
	case UnitMeasurementGram:
		return 1000
	case UnitMeasurementMilliliter:
		return 1000
	case UnitMeasurementCentimeter:
		return 100
	case UnitMeasurementMeter:
		return 100
	case UnitMeasurementInch:
		return 1
	case UnitMeasurementPiece:
		return 1
	default:
		return 1
	}
}
