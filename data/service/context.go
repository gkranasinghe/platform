package service

import (
	"github.com/tidepool-org/platform/auth"
	dataClient "github.com/tidepool-org/platform/data/client"
	"github.com/tidepool-org/platform/data/deduplicator"
	dataSource "github.com/tidepool-org/platform/data/source"
	dataStoreDEPRECATED "github.com/tidepool-org/platform/data/storeDEPRECATED"
	"github.com/tidepool-org/platform/metric"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/service"
	syncTaskStore "github.com/tidepool-org/platform/synctask/store"
)

type Context interface {
	service.Context

	AuthClient() auth.Client
	MetricClient() metric.Client
	PermissionClient() permission.Client

	DataDeduplicatorFactory() deduplicator.Factory

	DataRepository() dataStoreDEPRECATED.DataRepository
	SyncTaskRepository() syncTaskStore.SyncTaskRepository

	DataClient() dataClient.Client
	DataSourceClient() dataSource.Client
}

type HandlerFunc func(context Context)
