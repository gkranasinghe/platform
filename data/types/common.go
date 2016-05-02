package types

import (
	"fmt"
	"reflect"
	"time"

	validator "gopkg.in/bluesuncorp/validator.v8"

	"github.com/tidepool-org/platform/validate"
)

func init() {
	GetPlatformValidator().RegisterValidation(MmolUnitsField.Tag, MmolUnitsValidator)
	GetPlatformValidator().RegisterValidation(MmolOrMgUnitsField.Tag, MmolOrMgUnitsValidator)
	GetPlatformValidator().RegisterValidation(BolusSubTypeField.Tag, BolusSubTypeValidator)
	GetPlatformValidator().RegisterValidation(TimeStringField.Tag, PastTimeStringValidator)
	GetPlatformValidator().RegisterValidation("timeobject", PastTimeObjectValidator)
}

var (
	mmol = "mmol/L"
	mg   = "mg/dL"

	MmolOrMgUnitsField = DatumFieldInformation{
		DatumField: &DatumField{Name: "units"},
		Tag:        "mmolmgunits",
		Message:    fmt.Sprintf("Must be one of %s, %s", mmol, mg),
		Allowed: Allowed{
			mmol:     true,
			"mmol/l": true,
			mg:       true,
			"mg/dl":  true,
		},
	}

	MmolUnitsField = DatumFieldInformation{
		DatumField: &DatumField{Name: "units"},
		Tag:        "mmolunits",
		Message:    fmt.Sprintf("Must be %s", mmol),
		Allowed: Allowed{
			mmol:     true,
			"mmol/l": true,
		},
	}

	BloodGlucoseValueField = FloatDatumField{
		DatumField: &DatumField{Name: "value"},
		Tag:        "bloodglucosevalue",
	}

	mmolBloodGlucoseValueField = FloatDatumField{
		DatumField:        &DatumField{Name: "value"},
		Tag:               "bloodglucosevalue",
		Message:           "Must be between 0.0 and 55.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 55.0},
	}

	mgdlBloodGlucoseValueField = FloatDatumField{
		DatumField:        &DatumField{Name: "value"},
		Tag:               "bloodglucosevalue",
		Message:           "Must be between 0.0 and 1000.0",
		AllowedFloatRange: &AllowedFloatRange{LowerLimit: 0.0, UpperLimit: 1000.0},
	}

	BolusSubTypeField = DatumFieldInformation{
		DatumField: &DatumField{Name: "subType"},
		Tag:        "bolussubtype",
		Message:    "Must be one of normal, square, dual/square",
		Allowed:    Allowed{"normal": true, "square": true, "dual/square": true},
	}

	TimeStringField = DatumFieldInformation{
		DatumField: &DatumField{Name: "time"},
		Tag:        "timestr",
		Message:    "Times need to be ISO 8601 format and not in the future",
	}
)

type BloodGlucoseValidation struct {
	continueValidation    bool
	valueAllowedToBeEmpty bool
	valueErrorPath        string
	Value                 *float64
	returnValue           *float64
	Units                 *string
	returnUnits           *string
}

func NewBloodGlucoseValidation(val *float64, units *string) *BloodGlucoseValidation {
	return &BloodGlucoseValidation{Value: val, Units: units, valueErrorPath: "value", continueValidation: true}
}

func (b *BloodGlucoseValidation) SetValueAllowedToBeEmpty(valueAllowedToBeEmpty bool) *BloodGlucoseValidation {
	b.valueAllowedToBeEmpty = valueAllowedToBeEmpty
	return b
}

func (b *BloodGlucoseValidation) addError(msg, path string, errs validate.ErrorProcessing) {
	//TODO: this needs to be handled with map
	for i := range errs.Errors {
		if errs.Errors[i].Source["pointer"] == errs.BasePath+"/"+path {
			return
		}
	}
	errs.AppendPointerError(
		path,
		"Validation Error",
		msg,
	)
}

func (b *BloodGlucoseValidation) SetValueErrorPath(valueErrorPath string) *BloodGlucoseValidation {
	b.valueErrorPath = valueErrorPath
	return b
}

func (b *BloodGlucoseValidation) normalizeBloodGlucoseUnits(errs validate.ErrorProcessing) {
	if b.Units == nil {

		b.returnUnits = b.Units
		b.continueValidation = false

		b.addError(
			fmt.Sprintf("%s given '%v'", MmolOrMgUnitsField.Message, b.returnUnits),
			MmolOrMgUnitsField.Name,
			errs,
		)
		return
	}

	switch *b.Units {
	case mmol, "mmol/l":
		b.returnUnits = &mmol
	case mg, "mg/dl":
		b.returnUnits = &mg
	default:
		b.returnUnits = b.Units
		b.continueValidation = false
		b.addError(
			fmt.Sprintf("%s given '%v'", MmolOrMgUnitsField.Message, *b.returnUnits),
			MmolOrMgUnitsField.Name,
			errs,
		)
	}
}

func (b *BloodGlucoseValidation) convertBloodGlucoseValueMgToMmol() {
	if !b.valueAllowedToBeEmpty {
		if *b.returnUnits == mg {
			converted := *b.Value / 18.01559
			b.returnValue = &converted
			b.returnUnits = &mmol
		}
	}
}

func (b *BloodGlucoseValidation) validateBloodGlucoseValue(errs validate.ErrorProcessing) {

	if !b.valueAllowedToBeEmpty {
		switch *b.returnUnits {
		case mmol:
			if b.Value == nil {

				b.continueValidation = false
				b.addError(
					fmt.Sprintf("%s given '%v'", mmolBloodGlucoseValueField.Message, b.Value),
					b.valueErrorPath,
					errs,
				)

			} else if *b.Value < mmolBloodGlucoseValueField.LowerLimit || *b.Value > mmolBloodGlucoseValueField.UpperLimit {
				b.continueValidation = false
				b.addError(
					fmt.Sprintf("%s given '%v'", mmolBloodGlucoseValueField.Message, *b.Value),
					b.valueErrorPath,
					errs,
				)
			}

		default:
			if b.Value == nil {
				b.continueValidation = false
				b.addError(
					fmt.Sprintf("%s given '%v'", mgdlBloodGlucoseValueField.Message, b.Value),
					b.valueErrorPath,
					errs,
				)

			} else if *b.Value < mgdlBloodGlucoseValueField.LowerLimit || *b.Value > mgdlBloodGlucoseValueField.UpperLimit {
				b.continueValidation = false
				b.addError(
					fmt.Sprintf("%s given '%v'", mgdlBloodGlucoseValueField.Message, *b.Value),
					b.valueErrorPath,
					errs,
				)
			}
		}
	}

}

func (b *BloodGlucoseValidation) debug() {

	units := "units:nil"
	value := "value:nil"
	if b.returnUnits != nil {
		units = *b.returnUnits
	}
	if b.returnValue != nil {
		value = fmt.Sprintf("%.0f", *b.returnValue)
	}

	origUnits := "units:nil"
	origValue := "value:nil"
	if b.Units != nil {
		origUnits = *b.Units
	}
	if b.Value != nil {
		origValue = fmt.Sprintf("%.0f", *b.Value)
	}

	fmt.Println("# Normalize RETURN:", units, value, "GIVEN: ", origUnits, origValue)
}

func (b *BloodGlucoseValidation) ValidateAndConvertBloodGlucoseValue(errs validate.ErrorProcessing) (*float64, *string) {

	b.returnUnits = b.Units
	b.returnValue = b.Value

	b.normalizeBloodGlucoseUnits(errs)

	if !b.continueValidation {
		//b.debug()
		return b.returnValue, b.returnUnits
	}

	b.validateBloodGlucoseValue(errs)

	if !b.continueValidation {
		//b.debug()
		return b.returnValue, b.returnUnits
	}

	b.convertBloodGlucoseValueMgToMmol()
	//b.debug()
	return b.returnValue, b.returnUnits

}

func BolusSubTypeValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	subType, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = BolusSubTypeField.Allowed[subType]
	return ok
}

func MmolUnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = MmolUnitsField.Allowed[units]
	return ok
}

func MmolOrMgUnitsValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	units, ok := field.Interface().(string)
	if !ok {
		return false
	}
	_, ok = MmolOrMgUnitsField.Allowed[units]
	return ok
}

func PastTimeObjectValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {
	timeObject, ok := field.Interface().(time.Time)
	if !ok {
		return false
	}
	return isTimeObjectValid(timeObject)
}

func isTimeObjectValid(timeObject time.Time) bool {
	return !timeObject.IsZero() && timeObject.Before(time.Now())
}

func PastTimeStringValidator(v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value, field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string) bool {

	timeString, ok := field.Interface().(string)
	if !ok {
		return false
	}
	valid := isTimeStringValid(timeString)
	return valid
}

func isTimeStringValid(timeString string) bool {
	var timeObject time.Time
	timeObject, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		timeObject, err = time.Parse("2006-01-02T15:04:05", timeString)
		if err != nil {
			return false
		}
	}

	return isTimeObjectValid(timeObject)
}