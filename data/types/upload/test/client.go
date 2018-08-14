package test

import (
	dataTest "github.com/tidepool-org/platform/data/test"
	dataTypesUpload "github.com/tidepool-org/platform/data/types/upload"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/test"
	testInternet "github.com/tidepool-org/platform/test/internet"
)

func NewClient() *dataTypesUpload.Client {
	datum := dataTypesUpload.NewClient()
	datum.Name = pointer.FromString(testInternet.NewReverseDomain())
	datum.Version = pointer.FromString(testInternet.NewSemanticVersion())
	datum.Private = dataTest.NewBlob()
	return datum
}

func CloneClient(datum *dataTypesUpload.Client) *dataTypesUpload.Client {
	if datum == nil {
		return nil
	}
	clone := dataTypesUpload.NewClient()
	clone.Name = test.CloneString(datum.Name)
	clone.Version = test.CloneString(datum.Version)
	clone.Private = dataTest.CloneBlob(datum.Private)
	return clone
}
