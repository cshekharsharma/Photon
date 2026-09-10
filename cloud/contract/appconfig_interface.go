package contract

import (
	"context"

	"github.com/cshekharsharma/photon/cloud/entity/appconfig"
)

type AppConfigInterface interface {
	FetchConfig(ctx context.Context, input *appconfig.FetchConfigInput) (*appconfig.FetchConfigResult, error)
	WatchConfig(ctx context.Context, input *appconfig.WatchConfigInput, callback func(result *appconfig.FetchConfigResult)) error
	GetCurrentConfiguration(uniqueConfigId string) (*appconfig.FetchConfigResult, error)
	StopWatching(id string) error
	StopAllWatching()
}
