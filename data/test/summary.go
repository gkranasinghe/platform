package test

import (
	"github.com/tidepool-org/platform/data/summary"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
)

func RandomSummary() *summary.Summary {
	var datum = summary.Summary{
		LastUpdatedDate:          test.RandomTime(),
		HasLastUploadDate:        test.RandomBool(),
		FirstData:                test.RandomTime(),
		LastData:                 pointer.FromTime(test.RandomTime()),
		LastUploadDate:           test.RandomTime(),
		OutdatedSince:            nil,
		TotalHours:               test.RandomIntFromRange(0, 2160),
		HighGlucoseThreshold:     test.RandomFloat64FromRange(5, 10),
		VeryHighGlucoseThreshold: test.RandomFloat64FromRange(10, 20),
		LowGlucoseThreshold:      test.RandomFloat64FromRange(3, 5),
		VeryLowGlucoseThreshold:  test.RandomFloat64FromRange(0, 3),
	}

	// we only make 2, as its lighter and 2 vs 14 vs 90 isn't very different here.
	datum.HourlyStats = make([]*summary.Stats, 2)
	for i := 0; i < 2; i++ {
		datum.HourlyStats[i] = &summary.Stats{
			Date:            test.RandomTime(),
			TargetMinutes:   test.RandomIntFromRange(0, 1440),
			TargetRecords:   test.RandomIntFromRange(0, 288),
			LowMinutes:      test.RandomIntFromRange(0, 1440),
			LowRecords:      test.RandomIntFromRange(0, 288),
			VeryLowMinutes:  test.RandomIntFromRange(0, 1440),
			VeryLowRecords:  test.RandomIntFromRange(0, 288),
			HighMinutes:     test.RandomIntFromRange(0, 1440),
			HighRecords:     test.RandomIntFromRange(0, 288),
			VeryHighMinutes: test.RandomIntFromRange(0, 1440),
			VeryHighRecords: test.RandomIntFromRange(0, 288),
			TotalGlucose:    test.RandomFloat64FromRange(0, 10000),
			TotalCGMMinutes: test.RandomIntFromRange(0, 1440),
			TotalCGMRecords: test.RandomIntFromRange(0, 288),
			LastRecordTime:  test.RandomTime(),
		}
	}

	datum.Periods = make(map[string]*summary.Period)
	for _, period := range []string{"1d", "7d", "14d", "30d"} {
		datum.Periods[period] = &summary.Period{
			GlucoseManagementIndicator:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 20)),
			HasGlucoseManagementIndicator: test.RandomBool(),

			AverageGlucose: &summary.Glucose{
				Value: test.RandomFloat64FromRange(1, 30),
				Units: "mmol/L",
			},
			HasAverageGlucose: test.RandomBool(),

			TimeCGMUsePercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeCGMUsePercent: test.RandomBool(),
			TimeCGMUseMinutes:    test.RandomIntFromRange(0, 129600),
			TimeCGMUseRecords:    test.RandomIntFromRange(0, 25920),

			TimeInTargetPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInTargetPercent: test.RandomBool(),
			TimeInTargetMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInTargetRecords:    test.RandomIntFromRange(0, 25920),

			TimeInLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInLowPercent: test.RandomBool(),
			TimeInLowMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryLowPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryLowPercent: test.RandomBool(),
			TimeInVeryLowMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInVeryLowRecords:    test.RandomIntFromRange(0, 25920),

			TimeInHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInHighPercent: test.RandomBool(),
			TimeInHighMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInHighRecords:    test.RandomIntFromRange(0, 25920),

			TimeInVeryHighPercent:    pointer.FromFloat64(test.RandomFloat64FromRange(0, 1)),
			HasTimeInVeryHighPercent: test.RandomBool(),
			TimeInVeryHighMinutes:    test.RandomIntFromRange(0, 129600),
			TimeInVeryHighRecords:    test.RandomIntFromRange(0, 25920),
		}
	}

	return &datum
}
