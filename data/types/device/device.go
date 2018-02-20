package device

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
)

type Device struct {
	types.Base `bson:",inline"`

	SubType string `json:"subType,omitempty" bson:"subType,omitempty"`
}

type Meta struct {
	Type    string `json:"type,omitempty"`
	SubType string `json:"subType,omitempty"`
}

func Type() string {
	return "deviceEvent"
}

func (d *Device) Init() {
	d.Base.Init()
	d.Type = Type()

	d.SubType = ""
}

func (d *Device) Meta() interface{} {
	return &Meta{
		Type:    d.Type,
		SubType: d.SubType,
	}
}

func (d *Device) Parse(parser data.ObjectParser) error {
	parser.SetMeta(d.Meta())

	return d.Base.Parse(parser)
}

func (d *Device) Validate(validator structure.Validator) {
	d.Base.Validate(validator)

	if d.Type != "" {
		validator.String("type", &d.Type).EqualTo(Type())
	}

	validator.String("subType", &d.SubType).Exists().NotEmpty()
}

func (d *Device) IdentityFields() ([]string, error) {
	identityFields, err := d.Base.IdentityFields()
	if err != nil {
		return nil, err
	}

	if d.SubType == "" {
		return nil, errors.New("sub type is empty")
	}

	return append(identityFields, d.SubType), nil
}
