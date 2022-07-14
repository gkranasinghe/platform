package summary_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/summary"

	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	dataTypesBloodGlucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	"github.com/tidepool-org/platform/pointer"
	userTest "github.com/tidepool-org/platform/user/test"
)

const (
	veryLowBloodGlucose  = 3.0
	lowBloodGlucose      = 3.9
	highBloodGlucose     = 10.0
	veryHighBloodGlucose = 13.9
	units                = "mmol/L"
	requestedAvgGlucose  = 7.0
)

func NewContinuous(units *string, datumTime *time.Time, deviceID *string) *continuous.Continuous {
	datum := continuous.New()
	datum.Glucose = *dataTypesBloodGlucoseTest.NewGlucose(units)
	datum.Type = "cbg"

	datum.Active = true
	datum.ArchivedDataSetID = nil
	datum.ArchivedTime = nil
	datum.CreatedTime = nil
	datum.CreatedUserID = nil
	datum.DeletedTime = nil
	datum.DeletedUserID = nil
	datum.DeviceID = deviceID
	datum.ModifiedTime = nil
	datum.ModifiedUserID = nil
	datum.Time = pointer.FromString(datumTime.Format(time.RFC3339Nano))

	return datum
}

func NewDataSetCGMDataAvg(deviceID string, startTime time.Time, hours float64, reqAvg float64) []*continuous.Continuous {
	requiredRecords := int(hours * 12)

	var dataSetData = make([]*continuous.Continuous, requiredRecords)

	// generate X days of data
	for count := 0; count < requiredRecords; count += 2 {
		randValue := 1 + rand.Float64()*(reqAvg-1)
		glucoseValues := [2]float64{reqAvg + randValue, reqAvg - randValue}

		// this adds 2 entries, one for each side of the average so that the calculated average is the requested value
		for i, glucoseValue := range glucoseValues {
			datumTime := startTime.Add(time.Duration(-(count + i + 1)) * time.Minute * 5)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(glucoseValue)

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

type DataRanges struct {
	Min      float64
	Max      float64
	Padding  float64
	VeryLow  float64
	Low      float64
	High     float64
	VeryHigh float64
}

func NewDataRanges() DataRanges {
	return DataRanges{
		Min:      1,
		Max:      20,
		Padding:  0.01,
		VeryLow:  veryLowBloodGlucose,
		Low:      lowBloodGlucose,
		High:     highBloodGlucose,
		VeryHigh: veryHighBloodGlucose,
	}
}

func NewDataRangesSingle(value float64) DataRanges {
	return DataRanges{
		Min:      value,
		Max:      value,
		Padding:  0,
		VeryLow:  value,
		Low:      value,
		High:     value,
		VeryHigh: value,
	}
}

// creates a dataset with random values evenly divided between ranges
// NOTE: only generates 98.9% CGMUse, due to needing to be divisible by 5
func NewDataSetCGMDataRanges(deviceID string, startTime time.Time, hours float64, ranges DataRanges) []*continuous.Continuous {
	requiredRecords := int(hours * 10)
	var gapCompensation time.Duration

	var dataSetData = make([]*continuous.Continuous, requiredRecords)

	glucoseBrackets := [5][2]float64{
		{ranges.Min, ranges.VeryLow - ranges.Padding},
		{ranges.VeryLow, ranges.Low - ranges.Padding},
		{ranges.Low, ranges.High - ranges.Padding},
		{ranges.High, ranges.VeryHigh - ranges.Padding},
		{ranges.VeryHigh, ranges.Max},
	}

	// generate requiredRecords of data
	for count := 0; count < requiredRecords; count += 5 {
		gapCompensation = 10 * time.Minute * time.Duration(int(float64(count+1)/10))
		for i, bracket := range glucoseBrackets {
			datumTime := startTime.Add(time.Duration(-(count+i+1))*time.Minute*5 - gapCompensation)

			datum := NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			datum.Glucose.Value = pointer.FromFloat64(bracket[0] + (bracket[1]-bracket[0])*rand.Float64())

			dataSetData[requiredRecords-count-i-1] = datum
		}
	}

	return dataSetData
}

var _ = Describe("Summary", func() {
	var ctx context.Context
	var logger *logTest.Logger
	var userID string
	var datumTime time.Time
	var deviceID string
	var err error
	var dataSetCGMData []*continuous.Continuous

	BeforeEach(func() {
		logger = logTest.NewLogger()
		ctx = log.NewContextWithLogger(context.Background(), logger)
		userID = userTest.RandomID()
		deviceID = "SummaryTestDevice"
		datumTime = time.Date(2016, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	})

	Context("GetDuration", func() {
		var libreDatum *continuous.Continuous
		var otherDatum *continuous.Continuous

		It("Returns correct 15 minute duration for AbbottFreeStyleLibre", func() {
			libreDatum = NewContinuous(pointer.FromString(units), &datumTime, &deviceID)
			libreDatum.DeviceID = pointer.FromString("a-AbbottFreeStyleLibre-a")

			duration := summary.GetDuration(libreDatum)
			Expect(duration).To(Equal(15))
		})

		It("Returns correct duration for other devices", func() {
			otherDatum = NewContinuous(pointer.FromString(units), &datumTime, &deviceID)

			duration := summary.GetDuration(otherDatum)
			Expect(duration).To(Equal(5))
		})
	})

	Context("CalculateGMI", func() {
		// input and output examples sourced from https://diabetesjournals.org/care/article/41/11/2275/36593/
		It("Returns correct GMI for medical example 1", func() {
			gmi := summary.CalculateGMI(5.55)
			Expect(gmi).To(Equal(5.7))
		})

		It("Returns correct GMI for medical example 2", func() {
			gmi := summary.CalculateGMI(6.9375)
			Expect(gmi).To(Equal(6.3))
		})

		It("Returns correct GMI for medical example 3", func() {
			gmi := summary.CalculateGMI(8.325)
			Expect(gmi).To(Equal(6.9))
		})

		It("Returns correct GMI for medical example 4", func() {
			gmi := summary.CalculateGMI(9.722)
			Expect(gmi).To(Equal(7.5))
		})

		It("Returns correct GMI for medical example 5", func() {
			gmi := summary.CalculateGMI(11.11)
			Expect(gmi).To(Equal(8.1))
		})

		It("Returns correct GMI for medical example 6", func() {
			gmi := summary.CalculateGMI(12.4875)
			Expect(gmi).To(Equal(8.7))
		})

		It("Returns correct GMI for medical example 7", func() {
			gmi := summary.CalculateGMI(13.875)
			Expect(gmi).To(Equal(9.3))
		})

		It("Returns correct GMI for medical example 8", func() {
			gmi := summary.CalculateGMI(15.2625)
			Expect(gmi).To(Equal(9.9))
		})

		It("Returns correct GMI for medical example 9", func() {
			gmi := summary.CalculateGMI(16.65)
			Expect(gmi).To(Equal(10.5))
		})

		It("Returns correct GMI for medical example 10", func() {
			gmi := summary.CalculateGMI(19.425)
			Expect(gmi).To(Equal(11.7))
		})
	})

	Context("Summary calculations requiring datasets", func() {
		var userSummary *summary.Summary
		Context("CalculateStats", func() {
			It("Returns correct day count when given 2 weeks", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(336))
			})

			It("Returns correct day count when given 1 week", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(168))
			})

			It("Returns correct day count when given 3 weeks", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 504, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(504))
			})

			It("Returns correct record count when given overlapping records", func() {
				var doubledCGMData = make([]*continuous.Continuous, 288*2)

				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)

				// interlace the lists
				for i := 0; i < len(dataSetCGMData); i += 1 {
					doubledCGMData[i*2] = dataSetCGMData[i]
					doubledCGMData[i*2+1] = dataSetCGMDataTwo[i]
				}
				err = userSummary.CalculateStats(doubledCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(24))
				Expect(userSummary.HourlyStats[0].TotalCGMRecords).To(Equal(12))
			})

			It("Returns correct record count when given overlapping records across multiple calculations", func() {
				userSummary = summary.New(userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.Add(15*time.Second), 24, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userSummary.HourlyStats)).To(Equal(24))
				Expect(userSummary.HourlyStats[0].TotalCGMRecords).To(Equal(12))
			})

			It("Returns correct stats when given 1 week, then 1 week more than 2 weeks ahead", func() {
				var lastRecordTime time.Time
				var hourlyStatsLen int
				var newHourlyStatsLen int
				secondDatumTime := datumTime.AddDate(0, 0, 15)
				secondRequestedAvgGlucose := requestedAvgGlucose - 4
				userSummary = summary.New(userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(168))

				By("check total glucose and dates for first batch")
				hourlyStatsLen = len(userSummary.HourlyStats)
				for i := hourlyStatsLen - 1; i >= 0; i-- {
					Expect(userSummary.HourlyStats[i].TotalGlucose).To(Equal(requestedAvgGlucose * 12))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(hourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userSummary.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, secondDatumTime, 168, secondRequestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(528)) // 22 days

				By("check total glucose and dates for second batch")
				newHourlyStatsLen = len(userSummary.HourlyStats)
				expectedNewHourlyStatsLenStart := newHourlyStatsLen - len(dataSetCGMData)/12 // 12 per day, need length without the gap
				for i := newHourlyStatsLen - 1; i >= expectedNewHourlyStatsLenStart; i-- {
					Expect(userSummary.HourlyStats[i].TotalGlucose).To(Equal(secondRequestedAvgGlucose * 12))

					lastRecordTime = secondDatumTime.Add(-time.Hour*time.Duration(newHourlyStatsLen-i-1) - 5*time.Minute)
					Expect(userSummary.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("check total glucose and dates for gap")
				expectedGapEnd := newHourlyStatsLen - expectedNewHourlyStatsLenStart
				for i := hourlyStatsLen; i <= expectedGapEnd; i++ {
					Expect(userSummary.HourlyStats[i].TotalGlucose).To(Equal(float64(0)))
				}
			})

			It("Returns correct stats when given multiple batches in a day", func() {
				var incrementalDatumTime time.Time
				var lastRecordTime time.Time
				userSummary = summary.New(userID)

				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 144, requestedAvgGlucose)
				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(144))

				// TODO move to 0.5 hour to test more cases
				for i := 1; i <= 24; i++ {
					incrementalDatumTime = datumTime.Add(time.Duration(i) * time.Hour)
					dataSetCGMData = NewDataSetCGMDataAvg(deviceID, incrementalDatumTime, 1, float64(i))

					err = userSummary.CalculateStats(dataSetCGMData)

					Expect(err).ToNot(HaveOccurred())
					Expect(len(userSummary.HourlyStats)).To(Equal(144 + i))
					Expect(userSummary.HourlyStats[i].TotalCGMRecords).To(Equal(12))
				}

				for i := 144; i < len(userSummary.HourlyStats); i++ {
					f := fmt.Sprintf("hour %d", i)
					By(f)
					Expect(userSummary.HourlyStats[i].TotalCGMRecords).To(Equal(12))
					Expect(userSummary.HourlyStats[i].TotalCGMMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(time.Hour*time.Duration(i-143) - time.Minute*5)
					Expect(userSummary.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
					Expect(userSummary.HourlyStats[i].TotalGlucose).To(Equal(float64((i - 143) * 12)))

					averageGlucose := userSummary.HourlyStats[i].TotalGlucose / float64(userSummary.HourlyStats[i].TotalCGMRecords)
					Expect(averageGlucose).To(Equal(float64(i - 143)))
				}
			})

			It("Returns correct daily stats for days with different averages", func() {
				var expectedTotalGlucose float64
				var lastRecordTime time.Time
				userSummary = summary.New(userID)
				dataSetCGMDataOne := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -2), 24, requestedAvgGlucose)
				dataSetCGMDataTwo := NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, -1), 24, requestedAvgGlucose+1)
				dataSetCGMDataThree := NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose+2)
				dataSetCGMData = append(dataSetCGMDataOne, dataSetCGMDataTwo...)
				dataSetCGMData = append(dataSetCGMData, dataSetCGMDataThree...)

				err = userSummary.CalculateStats(dataSetCGMData)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(72))

				for i := len(userSummary.HourlyStats) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userSummary.HourlyStats[i].TotalCGMRecords).To(Equal(12))
					Expect(userSummary.HourlyStats[i].TotalCGMMinutes).To(Equal(60))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.HourlyStats)-i-1) - 5*time.Minute)
					Expect(userSummary.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))

					expectedTotalGlucose = (requestedAvgGlucose + float64(i/24)) * 12
					Expect(userSummary.HourlyStats[i].TotalGlucose).To(Equal(expectedTotalGlucose))
				}
			})

			It("Returns correct hourly stats for hours with different Time in Range", func() {
				var lastRecordTime time.Time
				userSummary = summary.New(userID)
				veryLowRange := NewDataRangesSingle(veryLowBloodGlucose - 0.5)
				lowRange := NewDataRangesSingle(lowBloodGlucose - 0.5)
				inRange := NewDataRangesSingle((highBloodGlucose + lowBloodGlucose) / 2)
				highRange := NewDataRangesSingle(highBloodGlucose + 0.5)
				veryHighRange := NewDataRangesSingle(veryHighBloodGlucose + 0.5)

				dataSetCGMDataOne := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-4*time.Hour), 1, veryLowRange)
				dataSetCGMDataTwo := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-3*time.Hour), 1, lowRange)
				dataSetCGMDataThree := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-2*time.Hour), 1, inRange)
				dataSetCGMDataFour := NewDataSetCGMDataRanges(deviceID, datumTime.Add(-1*time.Hour), 1, highRange)
				dataSetCGMDataFive := NewDataSetCGMDataRanges(deviceID, datumTime, 1, veryHighRange)

				// we do this a different way (multiple calls) than the last unit test for extra pattern coverage
				err = userSummary.CalculateStats(dataSetCGMDataOne)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataTwo)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataThree)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataFour)
				Expect(err).ToNot(HaveOccurred())
				err = userSummary.CalculateStats(dataSetCGMDataFive)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(userSummary.HourlyStats)).To(Equal(5))

				By("check record counters for insurance")
				for i := len(userSummary.HourlyStats) - 1; i >= 0; i-- {
					f := fmt.Sprintf("hour %d", i+1)
					By(f)
					Expect(userSummary.HourlyStats[i].TotalCGMRecords).To(Equal(10))
					Expect(userSummary.HourlyStats[i].TotalCGMMinutes).To(Equal(50))

					lastRecordTime = datumTime.Add(-time.Hour*time.Duration(len(userSummary.HourlyStats)-i-1) - time.Minute*5)
					Expect(userSummary.HourlyStats[i].LastRecordTime).To(Equal(lastRecordTime))
				}

				By("very low minutes")
				Expect(userSummary.HourlyStats[0].VeryLowMinutes).To(Equal(50))
				Expect(userSummary.HourlyStats[0].LowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[0].TargetMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[0].HighMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[0].VeryHighMinutes).To(Equal(0))

				By("very low records")
				Expect(userSummary.HourlyStats[0].VeryLowRecords).To(Equal(10))
				Expect(userSummary.HourlyStats[0].LowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[0].TargetRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[0].HighRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[0].VeryHighRecords).To(Equal(0))

				By("low minutes")
				Expect(userSummary.HourlyStats[1].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[1].LowMinutes).To(Equal(50))
				Expect(userSummary.HourlyStats[1].TargetMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[1].HighMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[1].VeryHighMinutes).To(Equal(0))

				By("low records")
				Expect(userSummary.HourlyStats[1].VeryLowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[1].LowRecords).To(Equal(10))
				Expect(userSummary.HourlyStats[1].TargetRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[1].HighRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[1].VeryHighRecords).To(Equal(0))

				By("in-range minutes")
				Expect(userSummary.HourlyStats[2].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[2].LowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[2].TargetMinutes).To(Equal(50))
				Expect(userSummary.HourlyStats[2].HighMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[2].VeryHighMinutes).To(Equal(0))

				By("in-range records")
				Expect(userSummary.HourlyStats[2].VeryLowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[2].LowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[2].TargetRecords).To(Equal(10))
				Expect(userSummary.HourlyStats[2].HighRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[2].VeryHighRecords).To(Equal(0))

				By("high minutes")
				Expect(userSummary.HourlyStats[3].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[3].LowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[3].TargetMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[3].HighMinutes).To(Equal(50))
				Expect(userSummary.HourlyStats[3].VeryHighMinutes).To(Equal(0))

				By("high records")
				Expect(userSummary.HourlyStats[3].VeryLowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[3].LowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[3].TargetRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[3].HighRecords).To(Equal(10))
				Expect(userSummary.HourlyStats[3].VeryHighRecords).To(Equal(0))

				By("very high minutes")
				Expect(userSummary.HourlyStats[4].VeryLowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[4].LowMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[4].TargetMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[4].HighMinutes).To(Equal(0))
				Expect(userSummary.HourlyStats[4].VeryHighMinutes).To(Equal(50))

				By("very high records")
				Expect(userSummary.HourlyStats[4].VeryLowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[4].LowRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[4].TargetRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[4].HighRecords).To(Equal(0))
				Expect(userSummary.HourlyStats[4].VeryHighRecords).To(Equal(10))
			})
		})

		Context("CalculateSummary", func() {
			It("Returns correct time in range for stats", func() {
				var expectedCGMUse float64
				userSummary = summary.New(userID)
				ranges := NewDataRanges()
				dataSetCGMData = NewDataSetCGMDataRanges(deviceID, datumTime, 720, ranges)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(720))

				userSummary.CalculateSummary()
				Expect(userSummary.TotalHours).To(Equal(720))

				stopPoints := []int{1, 7, 14, 30}
				for _, v := range stopPoints {
					periodKey := strconv.Itoa(v) + "d"

					f := fmt.Sprintf("period %s", periodKey)
					By(f)

					Expect(userSummary.Periods[periodKey].TimeInTargetMinutes).To(Equal(240 * v))
					Expect(userSummary.Periods[periodKey].TimeInTargetRecords).To(Equal(48 * v))
					Expect(userSummary.Periods[periodKey].TimeInTargetPercent).To(Equal(0.200))

					Expect(userSummary.Periods[periodKey].TimeInVeryLowMinutes).To(Equal(240 * v))
					Expect(userSummary.Periods[periodKey].TimeInVeryLowRecords).To(Equal(48 * v))
					Expect(userSummary.Periods[periodKey].TimeInVeryLowPercent).To(Equal(0.200))

					Expect(userSummary.Periods[periodKey].TimeInLowMinutes).To(Equal(240 * v))
					Expect(userSummary.Periods[periodKey].TimeInLowRecords).To(Equal(48 * v))
					Expect(userSummary.Periods[periodKey].TimeInLowPercent).To(Equal(0.200))

					Expect(userSummary.Periods[periodKey].TimeInHighMinutes).To(Equal(240 * v))
					Expect(userSummary.Periods[periodKey].TimeInHighRecords).To(Equal(48 * v))
					Expect(userSummary.Periods[periodKey].TimeInHighPercent).To(Equal(0.200))

					Expect(userSummary.Periods[periodKey].TimeInVeryHighMinutes).To(Equal(240 * v))
					Expect(userSummary.Periods[periodKey].TimeInVeryHighRecords).To(Equal(48 * v))
					Expect(userSummary.Periods[periodKey].TimeInVeryHighPercent).To(Equal(0.200))

					// ranges calc only generates 83.3% of an hour, each hour needs to be divisible by 5
					Expect(userSummary.Periods[periodKey].TimeCGMUseMinutes).To(Equal(1200 * v))
					Expect(userSummary.Periods[periodKey].TimeCGMUseRecords).To(Equal(240 * v))

					// this value is a bit funny, its 83.3%, but the missing end of the final day gets compensated off
					// resulting in 83.6% only on the first day
					if v == 1 {
						expectedCGMUse = 0.836
					} else {
						expectedCGMUse = 0.833
					}

					Expect(userSummary.Periods[periodKey].TimeCGMUsePercent).To(BeNumerically("~", expectedCGMUse, 0.001))
				}
			})

			It("Returns correct average glucose for stats", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(336))

				userSummary.CalculateSummary()

				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(Equal(expectedGMI))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(20160))
				Expect(userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(4032))
			})

			It("Correctly removes GMI when CGM use drop below 0.7", func() {
				userSummary = summary.New(userID)
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(336))

				userSummary.CalculateSummary()

				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(Equal(expectedGMI))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(20160))
				Expect(userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(4032))

				// start the real test
				dataSetCGMData = NewDataSetCGMDataAvg(deviceID, datumTime.AddDate(0, 0, 20), 24, requestedAvgGlucose)

				err = userSummary.CalculateStats(dataSetCGMData)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(userSummary.HourlyStats)).To(Equal(720)) // hits 4 days over 30 day cap

				userSummary.CalculateSummary()

				Expect(userSummary.TotalHours).To(Equal(30 * 24)) // 30 days currently capped
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(Equal(requestedAvgGlucose))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.0714, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUseMinutes).To(Equal(1440))
				Expect(userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(288))
			})

		})

		Context("Update", func() {
			var userData []*continuous.Continuous
			var status *summary.UserLastUpdated
			var newDatumTime time.Time

			It("Returns correctly calculated summary with no rolling", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				userSummary = summary.New(userID)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling <100% cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 168, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 7)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(168))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.5, 0.001))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling 100% cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 7)
				userSummary.OutdatedSince = &datumTime
				expectedGMIFirst := summary.CalculateGMI(requestedAvgGlucose - 4)
				expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMIFirst, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 168, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(504))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly non-rolling summary with two 2 week sets", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 24, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 26)
				userSummary.OutdatedSince = &datumTime
				expectedGMISecond := summary.CalculateGMI(requestedAvgGlucose + 4)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(24))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.07142, 0.001))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 252, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())

				// TODO check all other periods
				Expect(userSummary.TotalHours).To(Equal(648)) // 27 days
				Expect(userSummary.Periods["14d"].TimeCGMUseRecords).To(Equal(3024))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.75, 0.001))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose+4, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMISecond, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with rolling dropping cgm use", func() {
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose-4)
				userSummary = summary.New(userID)
				newDatumTime = datumTime.AddDate(0, 0, 7)
				userSummary.OutdatedSince = &datumTime
				expectedGMI := summary.CalculateGMI(requestedAvgGlucose - 4)

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(336))
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 1.0, 0.001))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", requestedAvgGlucose-4, 0.001))
				Expect(*userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNumerically("~", expectedGMI, 0.001))
				Expect(userSummary.OutdatedSince).To(BeNil())

				// start the actual test
				userData = NewDataSetCGMDataAvg(deviceID, newDatumTime, 1, requestedAvgGlucose+4)
				userSummary.OutdatedSince = &datumTime

				status = &summary.UserLastUpdated{
					LastData:   newDatumTime,
					LastUpload: newDatumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())

				expectedNewAvg := (1*(requestedAvgGlucose+4) + 336*(requestedAvgGlucose-4)) / 337

				Expect(userSummary.TotalHours).To(Equal(504)) // 21 days
				Expect(userSummary.Periods["14d"].TimeCGMUsePercent).To(BeNumerically("~", 0.5, 0.005))
				Expect(userSummary.Periods["14d"].AverageGlucose.Value).To(BeNumerically("~", expectedNewAvg, 0.05))
				Expect(userSummary.Periods["14d"].GlucoseManagementIndicator).To(BeNil())
				Expect(userSummary.OutdatedSince).To(BeNil())
			})

			It("Returns correctly calculated summary with userData records before summary LastData", func() {
				summaryLastData := datumTime.AddDate(0, 0, -7)
				userData = NewDataSetCGMDataAvg(deviceID, datumTime, 336, requestedAvgGlucose)
				userSummary = summary.New(userID)
				userSummary.OutdatedSince = &datumTime
				userSummary.LastData = &summaryLastData

				status = &summary.UserLastUpdated{
					LastData:   datumTime,
					LastUpload: datumTime,
				}

				err = userSummary.Update(ctx, status, userData)
				Expect(err).ToNot(HaveOccurred())
				Expect(userSummary.TotalHours).To(Equal(168))
				Expect(userSummary.OutdatedSince).To(BeNil())
			})
		})
	})
})
