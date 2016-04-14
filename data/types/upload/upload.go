package upload

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(deviceTagsField.Tag, DeviceTagsValidator)
	types.GetPlatformValidator().RegisterValidation(timeProcessingField.Tag, TimeProcessingValidator)
	types.GetPlatformValidator().RegisterValidation(deviceManufacturersField.Tag, DeviceManufacturersValidator)
}

type Record struct {
	UploadID            *string   `json:"uploadId" bson:"uploadId" valid:"gt=10"`
	UploadUserID        *string   `json:"byUser" bson:"byUser" valid:"gt=10"`
	Version             *string   `json:"version" bson:"version" valid:"gt=10"`
	ComputerTime        *string   `json:"computerTime" bson:"computerTime" valid:"timestr"`
	DeviceTags          *[]string `json:"deviceTags" bson:"deviceTags" valid:"uploaddevicetags"`
	DeviceManufacturers *[]string `json:"deviceManufacturers" bson:"deviceManufacturers" valid:"uploaddevicemanufacturers"`
	DeviceModel         *string   `json:"deviceModel" bson:"deviceModel" valid:"gt=10"`
	DeviceSerialNumber  *string   `json:"deviceSerialNumber" bson:"deviceSerialNumber" valid:"gt=10"`
	TimeProcessing      *string   `json:"timeProcessing" bson:"timeProcessing" valid:"uploadtimeprocessing"`
	types.Base          `bson:",inline"`
}

const Name = "upload"

var (
	deviceTagsField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceTags"},
		Tag:        "uploaddevicetags",
		Message:    "Must be one of insulin-pump, cgm, bgm",
		Allowed: types.Allowed{
			"insulin-pump": true,
			"cgm":          true,
			"bgm":          true,
		},
	}

	deviceManufacturersField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceManufacturers"},
		Tag:        "uploaddevicemanufacturers",
		Message:    "Must contain at least one manufacturer name",
	}

	timeProcessingField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "timeProcessing"},
		Tag:        "uploadtimeprocessing",
		Message:    "Must be one of across-the-board-timezone, utc-bootstrapping, none",
		Allowed: types.Allowed{
			"across-the-board-timezone": true,
			"utc-bootstrapping":         true,
			"none":                      true,
		},
	}

	computerTimeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "computerTime"},
		Tag:        "timestr",
		Message:    types.TimeStringField.Message,
	}

	uploadIDField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "uploadId"},
		Tag:        "gt",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	uploadUserIDField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "byUser"},
		Tag:        "gt",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	deviceModelField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceModel"},
		Tag:        "gt",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	deviceSerialNumberField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "deviceSerialNumber"},
		Tag:        "gt",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	versionField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "version"},
		Tag:        "gt",
		Message:    "This is a required field need needs to be 10+ characters in length",
	}

	failureReasons = validate.FailureReasons{
		"DeviceTags":          validate.VaidationInfo{FieldName: deviceTagsField.Name, Message: deviceTagsField.Message},
		"TimeProcessing":      validate.VaidationInfo{FieldName: timeProcessingField.Name, Message: timeProcessingField.Message},
		"DeviceManufacturers": validate.VaidationInfo{FieldName: deviceManufacturersField.Name, Message: deviceManufacturersField.Message},
		"ComputerTime":        validate.VaidationInfo{FieldName: computerTimeField.Name, Message: computerTimeField.Message},
		"UploadID":            validate.VaidationInfo{FieldName: uploadIDField.Name, Message: uploadIDField.Message},
		"UploadUserID":        validate.VaidationInfo{FieldName: uploadUserIDField.Name, Message: uploadUserIDField.Message},
		"DeviceModel":         validate.VaidationInfo{FieldName: deviceModelField.Name, Message: deviceModelField.Message},
		"DeviceSerialNumber":  validate.VaidationInfo{FieldName: deviceSerialNumberField.Name, Message: deviceSerialNumberField.Message},
		"Version":             validate.VaidationInfo{FieldName: versionField.Name, Message: versionField.Message},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) *Record {

	record := &Record{
		UploadID:            datum.ToString(uploadIDField.Name, errs),
		ComputerTime:        datum.ToString(computerTimeField.Name, errs),
		UploadUserID:        datum.ToString(uploadUserIDField.Name, errs),
		Version:             datum.ToString(versionField.Name, errs),
		TimeProcessing:      datum.ToString(timeProcessingField.Name, errs),
		DeviceModel:         datum.ToString(deviceModelField.Name, errs),
		DeviceManufacturers: datum.ToStringArray(deviceManufacturersField.Name, errs),
		DeviceTags:          datum.ToStringArray(deviceTagsField.Name, errs),
		DeviceSerialNumber:  datum.ToString(deviceSerialNumberField.Name, errs),
		Base:                types.BuildBase(datum, errs),
	}

	types.GetPlatformValidator().SetFailureReasons(failureReasons).Struct(record, errs)

	return record
}

func TimeProcessingValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	procesingType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = timeProcessingField.Allowed[procesingType]
	return ok
}

func DeviceTagsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	tags, ok := field.Interface().([]string)

	if !ok {
		return false
	}
	if len(tags) == 0 {
		return false
	}
	for i := range tags {
		_, ok = deviceTagsField.Allowed[tags[i]]
		if ok == false {
			break
		}
	}
	return ok
}

func DeviceManufacturersValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	deviceManufacturersField, ok := field.Interface().([]string)
	if !ok {
		return false
	}
	return len(deviceManufacturersField) > 0
}