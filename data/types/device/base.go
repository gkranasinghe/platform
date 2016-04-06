package device

import (
	"reflect"

	"github.com/tidepool-org/platform/Godeps/_workspace/src/gopkg.in/bluesuncorp/validator.v8"
	"github.com/tidepool-org/platform/data/types"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	types.GetPlatformValidator().RegisterValidation(subTypeField.Tag, SubTypeValidator)
	types.GetPlatformValidator().RegisterValidation(statusField.Tag, StatusValidator)
}

type Base struct {
	SubType    *string `json:"subType" bson:"subType" valid:"devicesubtype"`
	types.Base `bson:",inline"`
}

const Name = "deviceEvent"

var (
	subTypeField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "subType"},
		Tag:        "devicesubtype",
		Message:    "Must be one of alarm, calibration, status, prime, timeChange, reservoirChange",
		Allowed: types.Allowed{
			"alarm":           true,
			"calibration":     true,
			"status":          true,
			"prime":           true,
			"timeChange":      true,
			"reservoirChange": true,
		},
	}

	statusField = types.DatumFieldInformation{
		DatumField: &types.DatumField{Name: "status"},
		Tag:        "devicestatus",
		Message:    "Must be one of suspended, resumed",
		Allowed: types.Allowed{
			"suspended": true,
			"resumed":   true,
		},
	}
)

func Build(datum types.Datum, errs validate.ErrorProcessing) interface{} {

	base := Base{
		SubType: datum.ToString(subTypeField.Name, errs),
		Base:    types.BuildBase(datum, errs),
	}

	switch *base.SubType {
	case "alarm":
		return base.makeAlarm(datum, errs)
	case "calibration":
		return base.makeCalibration(datum, errs)
	case "status":
		return base.makeStatus(datum, errs)
	case "prime":
		return base.makePrime(datum, errs)
	case "timeChange":
		return base.makeTimeChange(datum, errs)
	case "reservoirChange":
		return base.makeReservoirChange(datum, errs)
	}
	return nil
}

func StatusValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	status, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = statusField.Allowed[status]
	return ok
}

func SubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = subTypeField.Allowed[subType]
	return ok
}