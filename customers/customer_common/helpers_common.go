package customer_common

import (
	"encoding/json"
	"errors"
	"reflect"

	"encore.app/database/models"
	"encore.dev/beta/errs"
	"encore.dev/types/uuid"
)

// ========================================
// Helper functions
// ========================================

// convert any to map[uuid.UUID]bool
func ConvertAnyToMapUUIDBool(any any) (map[uuid.UUID]bool, error) {
	mapUUIDBool := make(map[uuid.UUID]bool)
	json.Unmarshal([]byte(any.(string)), &mapUUIDBool)
	return mapUUIDBool, nil
}

// to determine the prefix of the key in redis
func GetKeyByAction(customer_id string, action models.EmailActionType) (string, error) {
	var key string
	if action == models.EmailActionTypeForgetPassword {
		key = "forget_password_code_" + customer_id
	} else if action == models.EmailActionTypeVerifyEmail {
		key = "verification_code_" + customer_id
	} else {
		return "", &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid action",
		}
	}
	return key, nil
}

// converter from src to dest with the same name
func AutoMapStructFields(src any, dest interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dest)

	if dstVal.Kind() != reflect.Ptr {
		return errors.New("destination must be a pointer")
	}

	dstVal = dstVal.Elem()
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	dstType := dstVal.Type()

	for i := range dstType.NumField() {
		dstField := dstType.Field(i)
		dstFieldVal := dstVal.Field(i)

		if !dstFieldVal.CanSet() {
			continue
		}

		srcFieldValue := srcVal.FieldByName(dstField.Name)
		if !srcFieldValue.IsValid() {
			continue
		}

		// Handle slice type conversion
		if srcFieldValue.Type().Kind() == reflect.Slice && dstFieldVal.Type().Kind() == reflect.Slice {
			if err := mapSliceField(srcFieldValue, dstFieldVal); err == nil {
				continue // Successfully mapped
			}
		}

		// Original assignable check
		if srcFieldValue.Type().AssignableTo(dstFieldVal.Type()) {
			dstFieldVal.Set(srcFieldValue)
		}
	}

	return nil
}

func mapSliceField(srcSlice, dstSlice reflect.Value) error {
	srcType := srcSlice.Type().Elem()
	dstType := dstSlice.Type().Elem()

	// Check if elements are structs with compatible fields
	if srcType.Kind() == reflect.Struct && dstType.Kind() == reflect.Struct {
		result := reflect.MakeSlice(dstSlice.Type(), srcSlice.Len(), srcSlice.Cap())

		for i := range srcSlice.Len() {
			srcElem := srcSlice.Index(i)
			dstElem := result.Index(i)

			// Recursively map struct fields
			if err := AutoMapStructFields(srcElem.Interface(), dstElem.Addr().Interface()); err != nil {
				return err
			}
		}

		dstSlice.Set(result)
		return nil
	}

	return errors.New("cannot map slice types")
}
