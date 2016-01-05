package data

type DeviceEvent struct {
	SubType string `json:"subType" valid:"required"`
	Base
}

func BuildDeviceEvent(obj map[string]interface{}) (*DeviceEvent, *DataError) {

	const (
		sub_type_field = "subType"
	)

	base, errs := BuildBase(obj)
	cast := NewCaster(errs)

	deviceEvent := &DeviceEvent{
		SubType: cast.ToString(sub_type_field, obj[sub_type_field]),
		Base:    base,
	}

	_, err := validator.Validate(deviceEvent)
	errs.AppendError(err)
	if errs.IsEmpty() {
		return deviceEvent, nil
	}
	return deviceEvent, errs
}