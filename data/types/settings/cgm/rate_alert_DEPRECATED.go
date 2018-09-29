package cgm

import (
	"github.com/tidepool-org/platform/data"
	dataBloodGlucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/structure"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

const (
	RateDEPRECATEDMgdLThree  = 3.0
	RateDEPRECATEDMgdLTwo    = 2.0
	RateDEPRECATEDMmolLThree = 0.16652243973136602
	RateDEPRECATEDMmolLTwo   = 0.11101495982091067
)

type RateAlertDEPRECATED struct {
	Enabled *bool    `json:"enabled,omitempty" bson:"enabled,omitempty"`
	Rate    *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
}

func (r *RateAlertDEPRECATED) Parse(parser data.ObjectParser) {
	r.Enabled = parser.ParseBoolean("enabled")
	r.Rate = parser.ParseFloat("rate")
}

func (r *RateAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	validator.Bool("enabled", r.Enabled).Exists()
	validator.Float64("rate", r.Rate).Exists()
}

func (r *RateAlertDEPRECATED) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		r.Rate = dataBloodGlucose.NormalizeValueForUnits(r.Rate, units)
	}
}

type FallRateAlertDEPRECATED struct {
	RateAlertDEPRECATED `bson:",inline"`
}

func ParseFallRateAlertDEPRECATED(parser data.ObjectParser) *FallRateAlertDEPRECATED {
	if parser.Object() == nil {
		return nil
	}
	datum := NewFallRateAlertDEPRECATED()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewFallRateAlertDEPRECATED() *FallRateAlertDEPRECATED {
	return &FallRateAlertDEPRECATED{}
}

func (f *FallRateAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	f.RateAlertDEPRECATED.Validate(validator, units)

	if rates := f.RatesForUnits(units); len(rates) > 0 {
		validator.Float64("rate", f.Rate).OneOf(rates...)
	}
}

func (f *FallRateAlertDEPRECATED) RatesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return []float64{-RateDEPRECATEDMgdLThree, -RateDEPRECATEDMgdLTwo}
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return []float64{-RateDEPRECATEDMmolLThree, -RateDEPRECATEDMmolLTwo}
		}
	}
	return nil
}

type RiseRateAlertDEPRECATED struct {
	RateAlertDEPRECATED `bson:",inline"`
}

func ParseRiseRateAlertDEPRECATED(parser data.ObjectParser) *RiseRateAlertDEPRECATED {
	if parser.Object() == nil {
		return nil
	}
	datum := NewRiseRateAlertDEPRECATED()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewRiseRateAlertDEPRECATED() *RiseRateAlertDEPRECATED {
	return &RiseRateAlertDEPRECATED{}
}

func (r *RiseRateAlertDEPRECATED) Validate(validator structure.Validator, units *string) {
	r.RateAlertDEPRECATED.Validate(validator, units)

	if rates := r.RatesForUnits(units); len(rates) > 0 {
		validator.Float64("rate", r.Rate).OneOf(rates...)
	}
}

func (r *RiseRateAlertDEPRECATED) RatesForUnits(units *string) []float64 {
	if units != nil {
		switch *units {
		case dataBloodGlucose.MgdL, dataBloodGlucose.Mgdl:
			return []float64{RateDEPRECATEDMgdLTwo, RateDEPRECATEDMgdLThree}
		case dataBloodGlucose.MmolL, dataBloodGlucose.Mmoll:
			return []float64{RateDEPRECATEDMmolLTwo, RateDEPRECATEDMmolLThree}
		}
	}
	return nil
}

type RateAlertsDEPRECATED struct {
	FallRateAlert *FallRateAlertDEPRECATED `json:"fallRate,omitempty" bson:"fallRate,omitempty"`
	RiseRateAlert *RiseRateAlertDEPRECATED `json:"riseRate,omitempty" bson:"riseRate,omitempty"`
}

func ParseRateAlertsDEPRECATED(parser data.ObjectParser) *RateAlertsDEPRECATED {
	if parser.Object() == nil {
		return nil
	}
	datum := NewRateAlertsDEPRECATED()
	datum.Parse(parser)
	parser.ProcessNotParsed()
	return datum
}

func NewRateAlertsDEPRECATED() *RateAlertsDEPRECATED {
	return &RateAlertsDEPRECATED{}
}

func (r *RateAlertsDEPRECATED) Parse(parser data.ObjectParser) {
	r.FallRateAlert = ParseFallRateAlertDEPRECATED(parser.NewChildObjectParser("fallRate"))
	r.RiseRateAlert = ParseRiseRateAlertDEPRECATED(parser.NewChildObjectParser("riseRate"))
}

func (r *RateAlertsDEPRECATED) Validate(validator structure.Validator, units *string) {
	if r.FallRateAlert != nil {
		r.FallRateAlert.Validate(validator.WithReference("fallRate"), units)
	} else {
		validator.WithReference("fallRate").ReportError(structureValidator.ErrorValueNotExists())
	}
	if r.RiseRateAlert != nil {
		r.RiseRateAlert.Validate(validator.WithReference("riseRate"), units)
	} else {
		validator.WithReference("riseRate").ReportError(structureValidator.ErrorValueNotExists())
	}
}

func (r *RateAlertsDEPRECATED) Normalize(normalizer data.Normalizer, units *string) {
	if normalizer.Origin() == structure.OriginExternal {
		if r.FallRateAlert != nil {
			r.FallRateAlert.Normalize(normalizer, units)
		}
		if r.RiseRateAlert != nil {
			r.RiseRateAlert.Normalize(normalizer, units)
		}
	}
}
