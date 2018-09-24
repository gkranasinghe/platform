package fetch

import (
	"time"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/activity/physical"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/data/types/device/calibration"
	"github.com/tidepool-org/platform/data/types/food"
	"github.com/tidepool-org/platform/data/types/insulin"
	"github.com/tidepool-org/platform/data/types/state/reported"
	"github.com/tidepool-org/platform/dexcom"
	"github.com/tidepool-org/platform/pointer"
)

// TODO: For now this assumes that the systemTime is close to true UTC time (+/- some small drift).
// However, it is possible for this to NOT be true if the device receives a hard reset.
// Unfortunately, the only way to detect that MIGHT be to look between multiple events.
// If there is a large gap between systemTimes, and a much larger or smaller gap between displayTimes,
// then it MIGHT indicate a hard reset. (It may also simply represent and period of time where the
// device was not in use and displayTime immediately prior to or immediately after period not it use
// were grossly in error.)

const OffsetDuration = 30 * time.Minute // Duration between time zone offsets we scan for

const MaximumOffsets = (14 * time.Hour) / OffsetDuration  // Maximum time zone offset is +14:00
const MinimumOffsets = (-12 * time.Hour) / OffsetDuration // Minimum time zone offset is -12:00

const DailyDuration = 24 * time.Hour
const DailyOffsets = DailyDuration / OffsetDuration

func translateTime(systemTime *time.Time, displayTime *time.Time, datum *types.Base) {
	var clockDriftOffsetDuration time.Duration
	var conversionOffsetDuration time.Duration
	var timeZoneOffsetDuration time.Duration

	delta := displayTime.Sub(*systemTime)
	if delta > 0 {
		offsetCount := time.Duration((float64(delta) + float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount > MaximumOffsets {
			conversionOffsetDuration += DailyDuration
			offsetCount -= DailyOffsets
		}
		timeZoneOffsetDuration = offsetCount * OffsetDuration
	} else if delta < 0 {
		offsetCount := time.Duration((float64(delta) - float64(OffsetDuration)/2) / float64(OffsetDuration))
		clockDriftOffsetDuration = delta - offsetCount*OffsetDuration
		for offsetCount < MinimumOffsets {
			conversionOffsetDuration -= DailyDuration
			offsetCount += DailyOffsets
		}
		timeZoneOffsetDuration = offsetCount * OffsetDuration
	}

	datum.Time = pointer.FromString(systemTime.Format(types.TimeFormat))
	datum.DeviceTime = pointer.FromString(displayTime.UTC().Format(types.DeviceTimeFormat))
	datum.TimeZoneOffset = pointer.FromInt(int(timeZoneOffsetDuration / time.Minute))
	if clockDriftOffsetDuration != 0 {
		datum.ClockDriftOffset = pointer.FromInt(int(clockDriftOffsetDuration / time.Millisecond))
	}
	if conversionOffsetDuration != 0 {
		datum.ConversionOffset = pointer.FromInt(int(conversionOffsetDuration / time.Millisecond))
	}

	if datum.Payload == nil {
		datum.Payload = data.NewBlob()
	}
	(*datum.Payload)["systemTime"] = *systemTime
}

func translateCalibrationToDatum(c *dexcom.Calibration) data.Datum {
	datum := calibration.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	datum.Value = pointer.CloneFloat64(c.Value)
	datum.Units = pointer.CloneString(c.Unit)
	datum.Payload = data.NewBlob()
	if c.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *c.TransmitterID
	}

	translateTime(c.SystemTime, c.DisplayTime, &datum.Base)
	return datum
}

func translateEGVToDatum(e *dexcom.EGV, unit *string, rateUnit *string) data.Datum {
	datum := continuous.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	datum.Value = pointer.CloneFloat64(e.Value)
	datum.Units = pointer.CloneString(unit)
	datum.Payload = data.NewBlob()
	if e.Status != nil {
		(*datum.Payload)["status"] = *e.Status
	}
	if e.Trend != nil {
		(*datum.Payload)["trend"] = *e.Trend
	}
	if e.TrendRate != nil {
		(*datum.Payload)["trendRate"] = *e.TrendRate
		(*datum.Payload)["trendRateUnits"] = *rateUnit
	}
	if e.TransmitterID != nil {
		(*datum.Payload)["transmitterId"] = *e.TransmitterID
	}
	if e.TransmitterTicks != nil {
		(*datum.Payload)["transmitterTicks"] = *e.TransmitterTicks
	}

	switch *unit {
	case dexcom.UnitMgdL:
		if *e.Value < dexcom.EGVValueMinMgdL {
			datum.Annotations = &data.BlobArray{{
				"code":      "bg/out-of-range",
				"value":     "low",
				"threshold": dexcom.EGVValueMinMgdL,
			}}
		} else if *e.Value > dexcom.EGVValueMaxMgdL {
			datum.Annotations = &data.BlobArray{{
				"code":      "bg/out-of-range",
				"value":     "high",
				"threshold": dexcom.EGVValueMaxMgdL,
			}}
		}
	case dexcom.UnitMmolL:
		// TODO: Add annotations
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventCarbsToDatum(e *dexcom.Event) data.Datum {
	datum := food.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if e.Value != nil && e.Unit != nil {
		datum.Nutrition = &food.Nutrition{
			Carbohydrate: &food.Carbohydrate{
				Net:   pointer.CloneFloat64(e.Value),
				Units: pointer.CloneString(e.Unit),
			},
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventExerciseToDatum(e *dexcom.Event) data.Datum {
	datum := physical.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if e.EventSubType != nil {
		switch *e.EventSubType {
		case dexcom.ExerciseLight:
			datum.ReportedIntensity = pointer.FromString(physical.ReportedIntensityLow)
		case dexcom.ExerciseMedium:
			datum.ReportedIntensity = pointer.FromString(physical.ReportedIntensityMedium)
		case dexcom.ExerciseHeavy:
			datum.ReportedIntensity = pointer.FromString(physical.ReportedIntensityHigh)
		}
	}
	if e.Value != nil && e.Unit != nil {
		datum.Duration = &physical.Duration{
			Units: pointer.CloneString(e.Unit),
			Value: pointer.CloneFloat64(e.Value),
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventHealthToDatum(e *dexcom.Event) data.Datum {
	datum := reported.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if e.EventSubType != nil {
		switch *e.EventSubType {
		case dexcom.HealthIllness:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateIllness)}}
		case dexcom.HealthStress:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateStress)}}
		case dexcom.HealthHighSymptoms:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateHyperglycemiaSymptoms)}}
		case dexcom.HealthLowSymptoms:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateHypoglycemiaSymptoms)}}
		case dexcom.HealthCycle:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateCycle)}}
		case dexcom.HealthAlcohol:
			datum.States = &reported.StateArray{{State: pointer.FromString(reported.StateStateAlcohol)}}
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}

func translateEventInsulinToDatum(e *dexcom.Event) data.Datum {
	datum := insulin.New()

	// TODO: Refactor so we don't have to clear these here
	datum.ID = nil
	datum.GUID = nil

	if e.Value != nil && e.Unit != nil {
		datum.Dose = &insulin.Dose{
			Total: pointer.CloneFloat64(e.Value),
			Units: pointer.FromString(insulin.DoseUnitsUnits),
		}
	}

	translateTime(e.SystemTime, e.DisplayTime, &datum.Base)
	return datum
}
