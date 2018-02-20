package prime

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/device"
	"github.com/tidepool-org/platform/structure"
)

const (
	TargetCannula              = "cannula"
	TargetTubing               = "tubing"
	VolumeTargetCannulaMaximum = 10.0
	VolumeTargetCannulaMinimum = 0.0
	VolumeTargetTubingMaximum  = 100.0
	VolumeTargetTubingMinimum  = 0.0
)

func Targets() []string {
	return []string{
		TargetCannula,
		TargetTubing,
	}
}

type Prime struct {
	device.Device `bson:",inline"`

	Target *string  `json:"primeTarget,omitempty" bson:"primeTarget,omitempty"`
	Volume *float64 `json:"volume,omitempty" bson:"volume,omitempty"`
}

func SubType() string {
	return "prime" // TODO: Rename Type to "device/prime"; remove SubType
}

func NewDatum() data.Datum {
	return New()
}

func New() *Prime {
	return &Prime{}
}

func Init() *Prime {
	prime := New()
	prime.Init()
	return prime
}

func (p *Prime) Init() {
	p.Device.Init()
	p.SubType = SubType()

	p.Target = nil
	p.Volume = nil
}

func (p *Prime) Parse(parser data.ObjectParser) error {
	if err := p.Device.Parse(parser); err != nil {
		return err
	}

	p.Target = parser.ParseString("primeTarget")
	p.Volume = parser.ParseFloat("volume")

	return nil
}

func (p *Prime) Validate(validator structure.Validator) {
	if !validator.HasMeta() {
		validator = validator.WithMeta(p.Meta())
	}

	p.Device.Validate(validator)

	if p.SubType != "" {
		validator.String("subType", &p.SubType).EqualTo(SubType())
	}

	validator.String("primeTarget", p.Target).Exists().OneOf(Targets()...)
	if p.Target != nil {
		volumeValidator := validator.Float64("volume", p.Volume)
		switch *p.Target {
		case TargetCannula:
			volumeValidator.InRange(VolumeTargetCannulaMinimum, VolumeTargetCannulaMaximum)
		case TargetTubing:
			volumeValidator.InRange(VolumeTargetTubingMinimum, VolumeTargetTubingMaximum)
		}
	}
}

func (p *Prime) Normalize(normalizer data.Normalizer) {
	if !normalizer.HasMeta() {
		normalizer = normalizer.WithMeta(p.Meta())
	}

	p.Device.Normalize(normalizer)
}
